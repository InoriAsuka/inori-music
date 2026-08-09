// ignore_for_file: unnecessary_non_null_assertion
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, TargetPlatform;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:window_manager/window_manager.dart' show DragToMoveArea;

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/system_titlebar_provider.dart';
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/glass_panel.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';
import 'package:inori_music/src/shared/widgets/shell_chrome.dart';

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

    // One shared instance for every breakpoint (mobile/tablet/desktop alike)
    // since v5.30.6 — MiniPlayerBar decides its own cover-vs-no-cover,
    // shuffle/repeat-flanked-transport-vs-not shape internally from its own
    // measured width now (see that class's doc comment on why the
    // `showNowPlaying` flag this used to need per layout was retired). A
    // brief v5.30.5 detour docked the cover+title block at the sidebar's own
    // foot instead (see the since-deleted SidebarNowPlaying) on the theory
    // that the field report's red-boxed layout meant "move it out of the
    // bar"; the user's actual ask was that the cover stay *with* the
    // transport controls, just not spread across the full window width, so
    // v5.30.6 reverted that and left the four-region layout itself
    // (full-height sidebar, bar scoped to the content column) untouched.
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

/// Whether macOS's native traffic-light buttons need dedicated layout
/// treatment from this shell's own chrome, rather than sitting inside a real
/// OS title bar. Shared by [_DesktopLayout] (which corner the sidebar panel
/// floats flush against) and [_DesktopSidebar] (title row inset, panel
/// corner radius) so the two can never independently drift out of sync —
/// before v5.30.7 folded this into one place, that was exactly how the
/// v5.30.0 gutter bug happened (see [ShellChrome]'s doc comment). The same
/// three-part gate `FullPlayerScreen._needsMacTrafficLightGutter` uses for
/// its own, unrelated call site (that screen has no sidebar ancestor to
/// share this with — see that function's own doc comment).
bool _macTrafficLightGutterNeeded(WidgetRef ref) =>
    DesktopIntegration.isDesktop &&
    defaultTargetPlatform == TargetPlatform.macOS &&
    !ref.watch(systemTitleBarProvider);

class _DesktopLayout extends ConsumerWidget {
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
  Widget build(BuildContext context, WidgetRef ref) {
    // Four regions laid out side by side rather than nested, matching the
    // v5.30.0 field report's red-boxed reference layout: the sidebar spans
    // the window's *full* height — including the "now playing" tile docked
    // at its own foot, see _DesktopSidebar — and the player bar is scoped to
    // the content column on the right instead of running full-width beneath
    // both. Before v5.30.5 the sidebar and the (then full-width) bar were
    // both nested inside one outer Column — sidebar-height above, bar below
    // — which made the sidebar end wherever the bar's height left off,
    // exactly the "sidebar shorter than the window, bar spanning underneath
    // it too" shape the field report's photo ruled out.
    //
    // macOS's native traffic lights are painted by the OS at a fixed window
    // coordinate, independent of anything Flutter draws — they do not move
    // just because this panel has a margin. Floating the panel 8px in on
    // every side (the v5.30.0-v5.30.6 shape) put the panel's own top-left
    // corner seam directly under the lights, so they sat *outside* the
    // panel, straddling its rounded edge (the v5.30.7 field report's
    // screenshot). The fix is not to add more margin — that just moves the
    // seam, it can never get *under* a fixed-position overlay — it's to
    // remove the margin on the two sides the lights are actually next to, so
    // the panel's own corner is no longer there to straddle: the lights end
    // up sitting inside the panel with the window's own native inset around
    // them, exactly like Apple Music's own sidebar. Right and bottom keep
    // their float — nothing sits at the window's top-right or bottom-left
    // corners that needs the same treatment.
    final flushTopLeft = _macTrafficLightGutterNeeded(ref);
    return Scaffold(
      body: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: flushTopLeft
                ? const EdgeInsets.fromLTRB(0, 0, 8, 8)
                : const EdgeInsets.fromLTRB(8, 8, 8, 8),
            child: SizedBox(
              width: 220,
              child: _DesktopSidebar(
                navGroups: navGroups,
                selectedIndex: selectedIndex,
                onItemTapped: onItemTapped,
              ),
            ),
          ),
          Expanded(
            child: ShellChrome(
              // The sidebar above already reserves room for macOS's native
              // traffic lights at the window's top-left corner (see
              // _DesktopSidebar's _macTitleRowTopInset) — every DesktopAppBar
              // rendered by a screen in this content column would otherwise
              // double-reserve it, which is exactly the coordinate bug the
              // field report traced: the gutter lived on the AppBar, which by
              // then only ever rendered at x >= 236, nowhere near the lights.
              reservesTrafficLightGutter: true,
              child: Column(
                children: [
                  Expanded(child: child),
                  bottomBar,
                ],
              ),
            ),
          ),
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

  /// macOS positions its native traffic-light buttons at a fixed offset from
  /// the window's top-left corner, independent of whatever Flutter renders
  /// underneath (a system constant, not something this app controls): 20pt
  /// in from the left edge, 12pt button diameter, 20pt centre-to-centre
  /// spacing, vertical centre ~20pt down from the window's top edge. That
  /// puts their lowest point at window-coordinate y = 20 + 12/2 = 26.
  ///
  /// Through v5.30.6 this panel floated 8px in from the window edge (see
  /// _DesktopLayout's old unconditional outer Padding), so window-coordinate
  /// y=26 sat only 26 - 8 = 18px into the panel's own local coordinate
  /// space, and this inset was 30 (18 rounded up past the minimum, leaving
  /// roughly a button-diameter of daylight under the cluster). v5.30.7
  /// floats the panel flush against the top-left corner instead (see
  /// _DesktopLayout's doc comment on why — the lights sat *outside* the
  /// panel's rounded corner, not just close to it), which moves the panel's
  /// local origin from window-y=8 to window-y=0. Re-deriving from that same
  /// window-coordinate fact: 26 - 0 = 26px is the new bare minimum, plus the
  /// same ~12px buffer the old value used (30 - 18 = 12) = 38. That also
  /// matches the simpler cross-check of just preserving the wordmark row's
  /// prior *absolute* on-screen position, which was already visually tuned:
  /// old panel margin (8) + old inset (30) = 38 from the true window top.
  /// Both derivations agree, hence 38 rather than a fresh guess.
  static const _macTitleRowTopInset = 38.0;

  /// The pre-v5.30.5 inset, kept everywhere the traffic lights aren't a
  /// concern: non-macOS (the caption-button trio there is ordinary Flutter
  /// content, not a native overlay) or the user opted into the real OS title
  /// bar, where the lights live inside actual native chrome above this panel
  /// rather than over it.
  static const _defaultTitleRowTopInset = 24.0;

  /// macOS's own window corner radius (Big Sur onward settled on roughly
  /// 10-12pt at standard scale — Apple has never published an exact value,
  /// this is the commonly-measured figure other custom-chrome Mac apps use).
  /// Matching it on the panel's flush top-left corner (see
  /// [_macTrafficLightGutterNeeded]) is what makes that corner read as part
  /// of the window's own frame instead of a second, competing curve sitting
  /// right at the edge.
  static const _macWindowCornerRadius = 12.0;

  /// This panel's usual uniform rounding everywhere the flush-corner
  /// treatment above doesn't apply.
  static const _panelCornerRadius = 16.0;

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

    final needsTrafficLightGutter = _macTrafficLightGutterNeeded(ref);

    final titleRow = Padding(
      padding: EdgeInsets.fromLTRB(
        16,
        needsTrafficLightGutter
            ? _macTitleRowTopInset
            : _defaultTitleRowTopInset,
        16,
        12,
      ),
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
    );

    // GlassPanel (Material inside, not a plain coloured Container): ListTile
    // paints its selected tile colour and ink splashes onto the nearest
    // Material ancestor, so a ColoredBox in between swallows both — the same
    // defect the sidebar had in v5.22.0. Zero padding here: the sidebar
    // already manages every inset itself (title row, account block, list
    // rows), and GlassPanel's own default padding would double up with all
    // of them rather than just framing the floating panel's edge.
    return GlassPanel(
      padding: EdgeInsets.zero,
      borderRadius: _panelCornerRadius,
      // Only the flush corner needs to differ — see _DesktopLayout's doc
      // comment on why that corner alone loses its margin, and
      // _macWindowCornerRadius's on why it borrows the window's own radius
      // rather than the panel's usual one. The other three corners are
      // untouched by any of this, so they keep _panelCornerRadius exactly as
      // every other platform/setting combination already does.
      borderRadiusOverride: needsTrafficLightGutter
          ? const BorderRadius.only(
              topLeft: Radius.circular(_macWindowCornerRadius),
              topRight: Radius.circular(_panelCornerRadius),
              bottomLeft: Radius.circular(_panelCornerRadius),
              bottomRight: Radius.circular(_panelCornerRadius),
            )
          : null,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Draggable so the blank space the traffic lights sit over still
          // lets the user move the window, matching Apple Music's own
          // sidebar header — before this fix nothing in the sidebar was a
          // drag region at all, since DragToMoveArea only ever wrapped
          // DesktopAppBar's row in the content column to the right.
          // Mac-only: Windows/Linux keep the plain, non-draggable title row
          // they've always had, since their own drag region and window
          // buttons already live in that content-column bar.
          needsTrafficLightGutter ? DragToMoveArea(child: titleRow) : titleRow,
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

  /// Horizontal gap between the panel's own edge and the tile's selected
  /// pill. Without this the pill (painted across the ListTile's *own*
  /// bounds, which used to run edge-to-edge inside the panel) shared exactly
  /// the same x-coordinates as GlassPanel's hairline border, so a selected
  /// row read as spilling out of the panel rather than floating inside it
  /// (v5.30.7 field report).
  static const _horizontalInset = 8.0;

  /// [ListTileThemeData.contentPadding] (skin_definition.dart) already
  /// spends 16px of horizontal inset on every ListTile in the app. Applying
  /// that on top of [_horizontalInset] would put this tile's icon/label 24px
  /// from the panel edge while _SectionHeader/_AccountBlock's own text sits
  /// at a flat 16px — a column of sidebar content whose left edge zigzags
  /// between two indents. Overriding it down to 16 - _horizontalInset here
  /// keeps the *total* (outer inset + this) at exactly 16px, so the icon
  /// lines up with every header above it; only the pill's own outer edge
  /// (drawn at the ListTile's full bounds, unaffected by contentPadding)
  /// picks up the new breathing room.
  static const _contentPadding = EdgeInsets.symmetric(
    horizontal: 16 - _horizontalInset,
  );

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(
        horizontal: _horizontalInset,
        vertical: 2,
      ),
      child: ListTile(
        dense: true,
        contentPadding: _contentPadding,
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
      ),
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
