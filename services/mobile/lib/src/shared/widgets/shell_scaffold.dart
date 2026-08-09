// ignore_for_file: unnecessary_non_null_assertion
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/glass_panel.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';

/// Adaptive shell scaffold:
/// - Mobile (<600dp): BottomNavigationBar + MiniPlayerBar
/// - Tablet (600–1199dp): NavigationRail + MiniPlayerBar
/// - Desktop (≥1200dp): Permanent NavigationDrawer + MiniPlayerBar
///
/// Desktop keyboard shortcuts:
/// - Space — toggle play/pause
/// - ← / MediaTrackPrevious — previous
/// - → / MediaTrackNext — next
class ShellScaffold extends ConsumerStatefulWidget {
  const ShellScaffold({super.key, required this.child});

  final Widget child;

  @override
  ConsumerState<ShellScaffold> createState() => _ShellScaffoldState();
}

class _ShellScaffoldState extends ConsumerState<ShellScaffold> {
  /// Navigation split into the two groups EchoMusic uses — browsing the
  /// server's catalog vs. the things this account has accumulated. The
  /// grouping is only rendered by the desktop sidebar; the narrower layouts
  /// flatten it back into a single strip (NavigationBar/NavigationRail have
  /// no section-header concept), which is why [_navItems] concatenates them
  /// and every index-based helper works off that flat list.
  ///
  /// `AppRoutes.playlists` is new here: the route and its screen already
  /// existed but nothing anywhere in the app linked to them, so catalog
  /// playlists were unreachable outside a deep link.
  List<_NavItem> _discoverItems(AppLocalizations t) => [
    _NavItem(
      label: t.artists,
      icon: Icons.people_outline,
      route: AppRoutes.artists,
    ),
    _NavItem(
      label: t.albums,
      icon: Icons.album_outlined,
      route: AppRoutes.albums,
    ),
    _NavItem(label: t.search, icon: Icons.search, route: AppRoutes.search),
  ];

  List<_NavItem> _libraryItems(AppLocalizations t) => [
    _NavItem(
      label: t.favorites,
      icon: Icons.favorite_outline,
      route: AppRoutes.favorites,
    ),
    _NavItem(label: t.history, icon: Icons.history, route: AppRoutes.history),
    _NavItem(
      label: t.playlists,
      icon: Icons.queue_music_outlined,
      route: AppRoutes.playlists,
    ),
  ];

  List<_NavItem> _navItems(AppLocalizations t) => [
    ..._discoverItems(t),
    ..._libraryItems(t),
  ];

  /// Guest mode is a local-files player with no server catalog behind it, so
  /// it gets the same responsive shell with a much shorter destination list
  /// rather than the bare Scaffold it used to fall back to — that left guest
  /// mode with no navigation chrome at all, and made Settings reachable only
  /// from the local library screen's own app bar. Mirrors Spotube, which does
  /// not switch to a separate UI when you skip sign-in either: the local
  /// library is a destination inside the same shell.
  List<_NavItem> _guestItems(AppLocalizations t) => [
    _NavItem(
      label: t.localLibrary,
      icon: Icons.library_music_outlined,
      route: AppRoutes.localLibrary,
    ),
    _NavItem(
      label: t.settings,
      icon: Icons.settings_outlined,
      route: AppRoutes.settings,
    ),
  ];

  late final HardwareKeyboard _keyboard;

  @override
  void initState() {
    super.initState();
    _keyboard = HardwareKeyboard.instance;
    _keyboard.addHandler(_handleKey);
  }

  @override
  void dispose() {
    _keyboard.removeHandler(_handleKey);
    super.dispose();
  }

  bool _handleKey(KeyEvent event) {
    if (event is! KeyDownEvent) return false;
    final notifier = ref.read(playerProvider.notifier);
    switch (event.logicalKey) {
      case LogicalKeyboardKey.space:
        notifier.togglePlayPause();
        return true;
      case LogicalKeyboardKey.arrowLeft:
      case LogicalKeyboardKey.mediaTrackPrevious:
        notifier.previous();
        return true;
      case LogicalKeyboardKey.arrowRight:
      case LogicalKeyboardKey.mediaTrackNext:
        notifier.next();
        return true;
      default:
        return false;
    }
  }

  int _selectedIndex(BuildContext context, List<_NavItem> items) {
    final location = GoRouterState.of(context).matchedLocation;
    for (var i = 0; i < items.length; i++) {
      if (location.startsWith(items[i].route)) return i;
    }
    return 0;
  }

  void _onItemTapped(BuildContext context, List<_NavItem> items, int index) {
    context.go(items[index].route);
  }

  @override
  Widget build(BuildContext context) {
    // Surfaces PlayerNotifier.playTrack failures here rather than at each of
    // its several call sites (local library, track list tiles, search,
    // queue navigation) — this is the one widget guaranteed to wrap all of
    // them, in both guest and logged-in layouts, with a Scaffold/
    // ScaffoldMessenger already available.
    ref.listen(playerProvider.select((s) => s.playbackError), (previous, next) {
      if (next == null) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(next.message)));
    });

    const bottomBar = MiniPlayerBar();

    // Guest mode drops the whole server-catalog nav (Artists/Albums/Search/
    // Favorites/History) — meaningless without an account — but keeps the
    // shell itself; see [_guestItems].
    final isGuest = ref.watch(authProvider).valueOrNull?.isGuest ?? false;

    final t = AppLocalizations.of(context);
    final items = isGuest ? _guestItems(t) : _navItems(t);
    final width = MediaQuery.sizeOf(context).width;
    final selectedIndex = _selectedIndex(context, items);

    if (width >= 1200) {
      return _DesktopLayout(
        // Two guest destinations under a section header would be more chrome
        // than content, so guest mode gets one unlabelled group.
        navGroups: isGuest
            ? [(header: null, items: items)]
            : [
                (header: t.discover, items: _discoverItems(t)),
                (header: t.library, items: _libraryItems(t)),
              ],
        selectedIndex: selectedIndex,
        onItemTapped: (i) => _onItemTapped(context, items, i),
        bottomBar: bottomBar,
        child: widget.child,
      );
    } else if (width >= 600) {
      return _TabletLayout(
        navItems: items,
        selectedIndex: selectedIndex,
        onItemTapped: (i) => _onItemTapped(context, items, i),
        bottomBar: bottomBar,
        child: widget.child,
      );
    } else {
      return _MobileLayout(
        navItems: items,
        selectedIndex: selectedIndex,
        onItemTapped: (i) => _onItemTapped(context, items, i),
        bottomBar: bottomBar,
        child: widget.child,
      );
    }
  }
}

// ---------------------------------------------------------------------------
// Mobile layout
// ---------------------------------------------------------------------------

class _MobileLayout extends StatelessWidget {
  const _MobileLayout({
    required this.navItems,
    required this.selectedIndex,
    required this.onItemTapped,
    required this.child,
    required this.bottomBar,
  });

  final List<_NavItem> navItems;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;
  final Widget child;
  final Widget bottomBar;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Expanded(child: child),
          bottomBar,
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: selectedIndex,
        onDestinationSelected: onItemTapped,
        // Past four destinations, always-on labels stop fitting at phone
        // widths — Material's own answer for a crowded bar is to label only
        // the selected one. Guest mode has two, so it keeps full labels.
        labelBehavior: navItems.length > 4
            ? NavigationDestinationLabelBehavior.onlyShowSelected
            : NavigationDestinationLabelBehavior.alwaysShow,
        destinations: navItems
            .map(
              (item) => NavigationDestination(
                icon: Icon(item.icon),
                label: item.label,
              ),
            )
            .toList(),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Tablet layout
// ---------------------------------------------------------------------------

class _TabletLayout extends StatelessWidget {
  const _TabletLayout({
    required this.navItems,
    required this.selectedIndex,
    required this.onItemTapped,
    required this.child,
    required this.bottomBar,
  });

  final List<_NavItem> navItems;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;
  final Widget child;
  final Widget bottomBar;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Expanded(
            child: Row(
              children: [
                NavigationRail(
                  selectedIndex: selectedIndex,
                  onDestinationSelected: onItemTapped,
                  labelType: NavigationRailLabelType.all,
                  // The rail's counterpart to the desktop sidebar's account
                  // block: at this width there's no room for the name, but
                  // Settings still needs a way in (see _AccountBlock).
                  // Suppressed when Settings is already one of the
                  // destinations, which is the case in guest mode.
                  trailing: navItems.any((i) => i.route == AppRoutes.settings)
                      ? null
                      : Expanded(
                          child: Align(
                            alignment: Alignment.bottomCenter,
                            child: Padding(
                              padding: const EdgeInsets.only(bottom: 12),
                              child: IconButton(
                                icon: const Icon(Icons.settings_outlined),
                                tooltip: AppLocalizations.of(context).settings,
                                onPressed: () => context.go(AppRoutes.settings),
                              ),
                            ),
                          ),
                        ),
                  destinations: navItems
                      .map(
                        (item) => NavigationRailDestination(
                          icon: Icon(item.icon),
                          label: Text(item.label),
                        ),
                      )
                      .toList(),
                ),
                const VerticalDivider(thickness: 0.5, width: 0.5),
                Expanded(child: child),
              ],
            ),
          ),
          bottomBar,
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Desktop layout
// ---------------------------------------------------------------------------

class _DesktopLayout extends StatelessWidget {
  const _DesktopLayout({
    required this.navGroups,
    required this.selectedIndex,
    required this.onItemTapped,
    required this.child,
    required this.bottomBar,
  });

  final List<_NavGroup> navGroups;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;
  final Widget child;
  final Widget bottomBar;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Expanded(
            child: Row(
              children: [
                // Floating, not flush: Apple Music's desktop sidebar sits in
                // its own inset rounded panel rather than a Material sheet
                // butted up against a divider — the margin plus GlassPanel's
                // own hairline border does the job the VerticalDivider used
                // to do, so the divider is gone rather than duplicating it.
                Padding(
                  padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
                  child: SizedBox(
                    width: 220,
                    child: _DesktopSidebar(
                      navGroups: navGroups,
                      selectedIndex: selectedIndex,
                      onItemTapped: onItemTapped,
                    ),
                  ),
                ),
                Expanded(child: child),
              ],
            ),
          ),
          bottomBar,
        ],
      ),
    );
  }
}

class _DesktopSidebar extends ConsumerWidget {
  const _DesktopSidebar({
    required this.navGroups,
    required this.selectedIndex,
    required this.onItemTapped,
  });

  final List<_NavGroup> navGroups;

  /// Index into the *flattened* item list, matching the order [navGroups]
  /// are laid out in.
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // Flat index that keeps walking across group boundaries, so the sidebar's
    // notion of "item 4" stays the same as the bottom bar's and the rail's.
    var flatIndex = 0;
    final rows = <Widget>[];
    for (final group in navGroups) {
      final header = group.header;
      if (header != null) rows.add(_SectionHeader(label: header));
      for (final item in group.items) {
        // Snapshot per row: the callback would otherwise close over the loop
        // counter itself and every tile would report the final index.
        final index = flatIndex;
        rows.add(
          _SidebarTile(
            item: item,
            isSelected: index == selectedIndex,
            onTap: () => onItemTapped(index),
          ),
        );
        flatIndex++;
      }
      rows.add(const SizedBox(height: 8));
    }

    // GlassPanel (Material inside, not a plain coloured Container): ListTile
    // paints its selected tile colour and ink splashes onto the nearest
    // Material ancestor, so a ColoredBox in between swallows both — the same
    // defect the sidebar had in v5.22.0. Zero padding here: the sidebar
    // already manages every inset itself (title row, account block, list
    // rows), and GlassPanel's own default padding would double up with all
    // of them rather than just framing the floating panel's edge.
    return GlassPanel(
      padding: EdgeInsets.zero,
      borderRadius: 16,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 24, 16, 12),
            child: Row(
              children: [
                const InoriMark(size: 22),
                const SizedBox(width: 8),
                // Expanded, because the wordmark at titleMedium/w700 already
                // exceeds the 220px sidebar's content width in some text
                // scales/themes — unbounded it overflows the Row.
                Expanded(
                  child: Text(
                    'Inori Music',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: context.skinColors.onSurface,
                      fontWeight: FontWeight.w700,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ),
          const _AccountBlock(),
          const Divider(),
          Expanded(child: ListView(children: rows)),
        ],
      ),
    );
  }
}

/// Signed-in identity plus the app's only entry point to Settings outside
/// guest mode — before this the `/settings` route existed but was reachable
/// only from the guest-mode local library screen, so an account holder had no
/// way in at all. EchoMusic's equivalent block also carries membership
/// level/VIP badges; those have no counterpart in this project's user model,
/// so only the name and the settings affordance are kept.
///
/// Guest mode used to render this the same way with `t.guest` ("Guest") as
/// the "username" — a placeholder dressed up as an identity, with nothing to
/// tap. EchoMusic's own account block treats "not logged in" as its own
/// state instead: the primary line reads "未登录" and a muted second line
/// invites the tap. v5.30.0 follows that: guest mode gets a distinct,
/// tappable login prompt rather than a fake account row.
class _AccountBlock extends ConsumerWidget {
  const _AccountBlock();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auth = ref.watch(authProvider).valueOrNull;
    final isGuest = auth?.isGuest ?? false;

    if (isGuest) return const _GuestSignInPrompt();

    final t = AppLocalizations.of(context);
    final username = auth?.username ?? '';
    final initial = username.isEmpty ? '?' : username[0].toUpperCase();

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 0, 8, 8),
      child: Row(
        children: [
          CircleAvatar(
            radius: 14,
            backgroundColor: context.skinColors.sakuraPinkDark,
            child: Text(
              initial,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: context.skinColors.onSurface,
              ),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              username,
              style: TextStyle(
                fontSize: 13,
                color: context.skinColors.onSurfaceVariant,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          IconButton(
            icon: Icon(
              Icons.settings_outlined,
              size: 18,
              color: context.skinColors.onSurfaceVariant,
            ),
            tooltip: t.settings,
            onPressed: () => context.go(AppRoutes.settings),
          ),
        ],
      ),
    );
  }
}

/// The guest-mode account block: a tappable row inviting sign-in rather than
/// a placeholder name. Tapping it calls the same [AuthNotifier.exitGuestMode]
/// the Settings screen's own "登录" button already calls — the router's
/// `isPastGate` redirect then takes over and lands on `/login`, so this is a
/// second entry point to one action rather than a second implementation of
/// it. Settings is deliberately not duplicated here (unlike the signed-in
/// branch above): guest mode already carries Settings as its own nav
/// destination (see [_ShellScaffoldState._guestItems]).
class _GuestSignInPrompt extends ConsumerWidget {
  const _GuestSignInPrompt();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(12),
          onTap: () => ref.read(authProvider.notifier).exitGuestMode(),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 4),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 14,
                  backgroundColor: context.skinColors.surfaceVariant,
                  child: Icon(
                    Icons.person_outline,
                    size: 16,
                    color: context.skinColors.onSurfaceVariant,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    t.tapToSignIn,
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: context.skinColors.onSurfaceVariant,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Icon(
                  Icons.chevron_right,
                  size: 18,
                  color: context.skinColors.onSurfaceVariant,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({required this.label});
  final String label;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
      child: Text(
        label.toUpperCase(),
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          letterSpacing: 0.8,
          color: context.skinColors.outline,
        ),
      ),
    );
  }
}

class _SidebarTile extends StatelessWidget {
  const _SidebarTile({
    required this.item,
    required this.isSelected,
    required this.onTap,
  });

  final _NavItem item;
  final bool isSelected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      dense: true,
      leading: Icon(
        item.icon,
        color: isSelected
            ? context.skinColors.sakuraPinkLight
            : context.skinColors.onSurfaceVariant,
      ),
      title: Text(
        item.label,
        style: TextStyle(
          color: isSelected
              ? context.skinColors.onSurface
              : context.skinColors.onSurfaceVariant,
          fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
        ),
      ),
      selected: isSelected,
      selectedTileColor: context.skinColors.sakuraPinkDark.withValues(
        alpha: 0.3,
      ),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      onTap: onTap,
    );
  }
}

// ---------------------------------------------------------------------------

class _NavItem {
  const _NavItem({
    required this.label,
    required this.icon,
    required this.route,
  });
  final String label;
  final IconData icon;
  final String route;
}

/// A sidebar section. A null [header] renders the items with no section
/// label — used by guest mode, whose two destinations don't need grouping.
typedef _NavGroup = ({String? header, List<_NavItem> items});
