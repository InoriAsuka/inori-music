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
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/queue_drawer.dart';
import 'package:inori_music/src/shared/desktop_integration.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/system_titlebar_provider.dart';
import 'package:inori_music/src/shared/theme/artwork_overlay_skin.dart'
    show CoverPaletteAccent;
import 'package:inori_music/src/shared/theme/skin_provider.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';
import 'package:inori_music/src/shared/widgets/shell_chrome.dart';
import 'package:inori_music/src/shared/widgets/sidebar_group_collapse_provider.dart';
import 'package:inori_music/src/shared/widgets/sidebar_playlists_section.dart';

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
  /// server's catalog vs. the things this account has accumulated. Feeds
  /// [_navItems], which is what the narrower layouts' flat
  /// NavigationBar/NavigationRail render (they have no section-header
  /// concept) *and* what [_selectedIndex]/[_onItemTapped] dispatch off for
  /// those two layouts.
  ///
  /// v5.33.0 gives the desktop sidebar its own, differently-organised
  /// EchoMusic-style groups (see [_desktopDiscoverItems]/
  /// [_desktopLibraryItems] below) without touching this pair — the phase's
  /// own scope note is explicit that only `_DesktopSidebar` (>=1200dp)
  /// changes, and mobile/tablet keep exactly the six destinations they had
  /// before (a real content freeze, not just "the widget code is
  /// untouched": an earlier pass folded the desktop-only regrouping into
  /// this shared pair, which silently changed the phone bottom bar from six
  /// destinations to seven — caught by
  /// `test/shell_scaffold_nav_test.dart`'s own "mobile bottom bar carries
  /// all six destinations" regression guard, which is exactly what it
  /// exists to catch).
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

  /// The desktop sidebar's own "发现音乐" group — EchoMusic's `Sidebar.vue`
  /// puts 为您推荐/探索发现 here (see for_you_screen.dart/explore_screen.dart),
  /// not the catalog-browsing destinations [_discoverItems] holds for the
  /// narrower layouts. A separate function/list from [_discoverItems]
  /// rather than a shared one specifically so mobile/tablet's own six
  /// destinations (via [_navItems]) can stay completely unmodified — see
  /// this class's own doc comment above for why that separation is load
  /// bearing, not incidental.
  List<_NavItem> _desktopDiscoverItems(AppLocalizations t) => [
    _NavItem(
      label: t.forYou,
      icon: Icons.auto_awesome_outlined,
      route: AppRoutes.forYou,
    ),
    _NavItem(
      label: t.sidebarExplore,
      icon: Icons.explore_outlined,
      route: AppRoutes.explore,
    ),
  ];

  /// The desktop sidebar's own "我的乐库" group. EchoMusic's own 我的乐库
  /// (favourites/personal FM/cloud drive/history) has no catalog-browsing
  /// entries in it at all; 私人FM/我的云盘 have no counterpart in this app's
  /// backend and are deliberately excluded (see requirement.md's v5.33.0
  /// entry), which would leave this group with only favourites/history —
  /// thin enough that artists/albums/search are folded in here too, so the
  /// desktop sidebar's skeleton reads as populated rather than an
  /// afterthought, at the cost of no longer matching EchoMusic's own
  /// category boundary exactly. Catalog playlists are *not* folded in here
  /// even though [_libraryItems] (the mobile/tablet equivalent) still
  /// carries them as a flat item — the desktop sidebar surfaces both
  /// playlist subsystems through its own dedicated tabbed section
  /// (`SidebarPlaylistsSection`) instead, see [_DesktopSidebar].
  List<_NavItem> _desktopLibraryItems(AppLocalizations t) => [
    _NavItem(
      label: t.favorites,
      icon: Icons.favorite_outline,
      route: AppRoutes.favorites,
    ),
    _NavItem(label: t.history, icon: Icons.history, route: AppRoutes.history),
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

  /// The desktop sidebar's own flat item list — [_desktopDiscoverItems] +
  /// [_desktopLibraryItems] concatenated, the same "flatten the groups for
  /// index purposes" shape [_navItems] already established, just over a
  /// different pair of lists. This is what [_DesktopSidebar]'s own
  /// `onItemTapped`/`selectedIndex` are computed against — a completely
  /// separate index domain from [_navItems], not a shared one, which is
  /// exactly what lets the two layouts' destinations differ without either
  /// one's index scheme corrupting the other's.
  List<_NavItem> _desktopNavItems(AppLocalizations t) => [
    ..._desktopDiscoverItems(t),
    ..._desktopLibraryItems(t),
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
      // A completely separate flat item list/index from `items`/
      // `selectedIndex` above (guest mode excepted — see the ternary
      // below) — see [_desktopNavItems]'s own doc comment for why sharing
      // one list between mobile/tablet and the desktop sidebar is exactly
      // the mistake this phase's own field report caught.
      final desktopItems = isGuest ? items : _desktopNavItems(t);
      final desktopSelectedIndex = isGuest
          ? selectedIndex
          : _selectedIndex(context, desktopItems);
      return _DesktopLayout(
        // Two guest destinations under a section header would be more chrome
        // than content, so guest mode gets one unlabelled group.
        navGroups: isGuest
            ? [(header: null, collapseKey: null, items: desktopItems)]
            : [
                (
                  header: t.discover,
                  collapseKey: 'discover',
                  items: _desktopDiscoverItems(t),
                ),
                (
                  header: t.library,
                  collapseKey: 'library',
                  items: _desktopLibraryItems(t),
                ),
              ],
        selectedIndex: desktopSelectedIndex,
        onItemTapped: (i) => _onItemTapped(context, desktopItems, i),
        bottomBar: bottomBar,
        isGuest: isGuest,
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
    required this.isGuest,
  });

  final List<_NavGroup> navGroups;
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;
  final Widget child;
  final Widget bottomBar;

  /// Threaded down to [_DesktopSidebar] so it can keep the guest shell to
  /// its own explicitly simple shape (two nav items + the account card +,
  /// where it's meaningful, the download entry) instead of the full
  /// EchoMusic skeleton (collapsible groups, the playlist tabs section, the
  /// device entry) — v5.33.0's own scope note: a guest account has no
  /// server-side playlists to preview and no output-device capability
  /// question that differs from a signed-in session, so none of that
  /// skeleton earns its keep here.
  final bool isGuest;

  /// Fixed sidebar column width — named so [QueueDrawer]'s own positioning
  /// below can reference the exact same figure the sidebar's `SizedBox`
  /// uses, rather than a second, independently-typed `220` that could drift
  /// out of sync with it.
  static const _sidebarWidth = 220.0;

  @override
  Widget build(BuildContext context) {
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
    // v5.31.0 drops the floating-panel-with-margin shape entirely (see
    // _DesktopSidebar's own doc comment for the EchoMusic reference this now
    // follows) — the sidebar sits flush against the window's left, top, and
    // bottom edges with no Padding wrapper at all, so there is no longer a
    // panel corner for macOS's native traffic lights to straddle and nothing
    // here needs to special-case which corner they land near. That removes
    // the whole v5.30.0-v5.30.7 lineage of gutter-margin bugs by construction
    // rather than by patching the margin math again.
    return Scaffold(
      body: Stack(
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              SizedBox(
                width: _sidebarWidth,
                child: _DesktopSidebar(
                  navGroups: navGroups,
                  selectedIndex: selectedIndex,
                  onItemTapped: onItemTapped,
                  isGuest: isGuest,
                ),
              ),
              Expanded(
                child: ShellChrome(
                  // The sidebar above still spans the window's full height and
                  // therefore still owns the top-left corner macOS's native
                  // traffic lights land in — every DesktopAppBar rendered by a
                  // screen in this content column would otherwise reserve
                  // gutter space for lights that are nowhere near it (this
                  // column starts at x=220, not the window edge).
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
          // EchoMusic's MainLayout.vue paints a `layout-accent-gradient` div
          // that spans *both* columns, positioned above the Row in source
          // order (i.e. above it in paint order too) — its own comment there
          // reads "主题色顶部渐变氛围层（横跨侧栏与内容，盖住中缝避免出现分隔白线）"
          // ("theme-colour top gradient ambience layer, spanning the sidebar
          // and content, covering the seam so no dividing white line
          // appears"). Same idea here: the sidebar's hairline border is a
          // real, deliberate seam (see _DesktopSidebar), but a single flat
          // hairline between two otherwise-identical flat fills can still
          // read as a harder break than intended at the very top of the
          // window, where both columns' own top chrome (the sidebar's drag
          // strip/wordmark, the content column's DesktopAppBar) sit right
          // next to each other with nothing else to soften the join. A
          // low-opacity wash across the top of *both* columns papers over
          // that without touching the hairline itself.
          const Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: _LayoutAccentGradient(),
          ),
          // The mini bar's queue button (see mini_player_bar.dart's
          // _openQueue) toggles queueDrawerOpenProvider rather than
          // navigating away — this is what actually renders that drawer.
          // Scoped to start at _sidebarWidth, not left: 0, so it only ever
          // overlays the content column: the sidebar's own nav stays fully
          // clickable while the drawer is open, matching Spotify/EchoMusic
          // (their queue panels dock over the page, not over their own nav
          // rail). Always mounted (not wrapped in `if (open)`) — see
          // QueueDrawer's own doc comment for why its AnimationController
          // needs to outlive a single open/close cycle.
          const Positioned(
            top: 0,
            left: _sidebarWidth,
            right: 0,
            bottom: 0,
            child: QueueDrawer(),
          ),
        ],
      ),
    );
  }
}

/// Height of [_LayoutAccentGradient]'s wash. A fixed pixel count rather than
/// a fraction of window height: what it needs to visually bridge is the
/// sidebar's own top chrome — drag strip + wordmark row + account card +
/// divider + the first section header — which is fixed-height regardless of
/// how tall the window is, so a taller window does not need a
/// proportionally taller wash to cover the same seam. Measured directly
/// (`shell_scaffold_nav_test.dart`'s drag-strip tests pin the inputs this
/// depends on) rather than estimated: on the worst case, macOS's 48px strip,
/// that stack puts the first nav tile's own centre at y=242 as of v5.33.0's
/// account card redesign (up from y=226 pre-v5.33.0 — the filled card's own
/// margin/padding, see `_accountCard`, adds real height the old bare row
/// didn't have; 24px shorter, y=218, on other desktop platforms —
/// `_DesktopSidebar._dragStripHeight` is the only thing that differs
/// between the two). 250 clears the macOS figure with a small margin, so
/// the wash — deliberately — reaches slightly past the sidebar's own
/// header block into the first nav row rather than stopping exactly at its
/// boundary. That overlap is what a "still clickable through the wash"
/// test needs in order to be testing anything real — a gradient that only
/// ever covered blank space could never prove [IgnorePointer] below is
/// doing its job.
const _layoutAccentGradientHeight = 250.0;

/// Max opacity at the gradient's top edge, tuned by the *ambient skin's* own
/// brightness rather than the cover-backdrop luminance
/// `artwork_overlay_skin.dart` branches on — that file's `isLightBackdrop`
/// answers "is the artwork backdrop behind full-bleed player/lyrics content
/// light", which has no counterpart here: this shell's sidebar and content
/// surfaces always render in the ambient skin's own colours (see
/// _DesktopSidebar's doc comment on why this file never wraps itself in an
/// overlay SkinScope), so it is *that* brightness the wash sits on top of,
/// not any particular track's cover. A saturated accent at a given alpha
/// reads as a clean tint over Moonlit Indigo's dark ground but as a visibly
/// dirty smudge over Sakura Dusk's near-white one (more of the wash's own
/// hue survives against a background with little of its own colour to
/// assert against it) — halving the ceiling on a light base keeps the same
/// wash from crossing that line. Exposed as a top-level function (rather
/// than inlined) so a test can assert both branches directly without
/// pumping a widget, the same pattern `full_player_screen.dart`'s
/// `playerArtworkSize`/`playerControlWidth` already use.
double layoutAccentGradientMaxAlpha(Brightness ambientBrightness) =>
    ambientBrightness == Brightness.dark ? 0.24 : 0.12;

/// The colour [_LayoutAccentGradient] washes down from. Cover-driven when a
/// palette is available — [CoverPaletteAccent.accentOverArtwork] is the same
/// swatch-preference order `artworkOverlaySkin` picks for its own on-artwork
/// controls — and the ambient skin's own accent otherwise, so nothing playing
/// (or a palette that hasn't resolved yet) still renders a sensible colour
/// instead of a transparent gap. A free function for the same
/// pump-free-testability reason as [layoutAccentGradientMaxAlpha].
Color layoutAccentGradientColor({
  required CoverPalette? palette,
  required Color fallback,
}) => palette?.accentOverArtwork ?? fallback;

/// Test hook for [_LayoutAccentGradient] — same reasoning as
/// [desktopSidebarKey]: the widget itself is private, so a test needs a
/// public marker to locate, size, and hit-test it.
const layoutAccentGradientKey = Key('layout-accent-gradient');

/// EchoMusic's `MainLayout.vue` paints a `layout-accent-gradient` div above
/// both the sidebar and the content column — see _DesktopLayout's doc
/// comment for the source comment explaining why (masking the seam between
/// two otherwise-flat, adjacent fills). This is that div: a purely decorative
/// top-anchored wash, low-opacity enough to read as ambience rather than a
/// colour block, that never intercepts a pointer event meant for whatever it
/// sits above (the sidebar's title row, account block, and first nav rows all
/// render underneath the top of this band).
///
/// A dedicated [ConsumerWidget] rather than inlined into _DesktopLayout.build
/// so watching the current track's palette only rebuilds this ~200px strip,
/// not the whole sidebar/content/player-bar tree underneath it.
class _LayoutAccentGradient extends ConsumerWidget {
  const _LayoutAccentGradient();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final mediaItem = ref.watch(playerProvider.select((s) => s.mediaItem));
    final albumId = mediaItem?.extras?['albumId'] as String?;
    final localArtUri = mediaItem?.artUri;
    final palette = ref
        .watch(
          coverPaletteProvider((albumId: albumId, localArtUri: localArtUri)),
        )
        .valueOrNull;
    final accent = layoutAccentGradientColor(
      palette: palette,
      fallback: context.skinColors.sakuraPink,
    );
    final maxAlpha = layoutAccentGradientMaxAlpha(Theme.of(context).brightness);

    return IgnorePointer(
      key: layoutAccentGradientKey,
      child: SizedBox(
        height: _layoutAccentGradientHeight,
        width: double.infinity,
        child: DecoratedBox(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                accent.withValues(alpha: maxAlpha),
                accent.withValues(alpha: 0),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// Test hook for locating the sidebar's own root surface — the same purpose
/// [MiniPlayerBar.contentKey] serves for that bar. Needed because [_DesktopSidebar]
/// itself is library-private (tests live in a different library and cannot
/// spell the type), and because the sidebar no longer renders a distinctively-
/// typed [GlassPanel] the way it did through v5.30.7.
const desktopSidebarKey = Key('desktop-sidebar');

/// Test hook for the blank drag strip at the sidebar's own top — lets a test
/// measure its height directly instead of reverse-engineering it from the
/// wordmark row's position.
const desktopSidebarDragStripKey = Key('desktop-sidebar-drag-strip');

class _DesktopSidebar extends ConsumerWidget {
  const _DesktopSidebar({
    required this.navGroups,
    required this.selectedIndex,
    required this.onItemTapped,
    required this.isGuest,
  });

  final List<_NavGroup> navGroups;

  /// Index into the *flattened* item list, matching the order [navGroups]
  /// are laid out in.
  final int selectedIndex;
  final ValueChanged<int> onItemTapped;

  /// See [_DesktopLayout.isGuest]'s own doc comment — gates the v5.33.0
  /// EchoMusic-skeleton additions (playlist tabs, download/device entries)
  /// off entirely rather than rendering them empty; the nav groups
  /// themselves and the account card above are unaffected by this flag
  /// (they already had their own guest-appropriate shapes before this
  /// phase).
  final bool isGuest;

  /// EchoMusic's `Sidebar.vue` renders a blank `drag-region` div as the
  /// sidebar's very first child — `isMac ? 'h-12' : 'h-6'` (48px / 24px) —
  /// rather than computing a top inset from the traffic lights' exact
  /// geometry and pushing the wordmark row down by whatever clears them (the
  /// v5.30.0-v5.30.7 approach this replaces; see git history for the
  /// now-deleted `_macTitleRowTopInset`, which had grown a multi-paragraph
  /// derivation of a single pixel constant — a sign the approach itself, not
  /// just the number, was the wrong shape). Matching EchoMusic directly is
  /// both simpler and more robust: macOS's lights (20pt from the left edge,
  /// 12pt diameter, vertical centre ~20pt down the window, so their lowest
  /// point sits at window-coordinate y=26) comfortably fit inside a 48px
  /// band without this file ever needing to know their exact geometry, and
  /// it no longer has to be re-derived every time the sidebar's own margin
  /// changes — which is exactly what forced two rewrites of that constant
  /// between v5.30.0 and v5.30.7.
  static const _macDragStripHeight = 48.0;

  /// Non-macOS desktop platforms get a smaller band purely for EchoMusic
  /// visual parity and a convenient drag handle — nothing sits under it
  /// there that needs dodging (Windows/Linux's own window-caption buttons
  /// live in the content column's [DesktopAppBar], not the sidebar).
  static const _dragStripHeight = 24.0;

  /// The wordmark row's own padding — fixed regardless of platform now that
  /// traffic-light clearance lives in the drag strip above it (see
  /// [_macDragStripHeight]) instead of in this row's own top inset. This is
  /// the same constant v5.30.0 shipped before v5.30.5 first taught this row
  /// to inflate its own top inset for the lights; it never actually needed
  /// to change once something else took over clearing them.
  static const _titleRowPadding = EdgeInsets.fromLTRB(16, 24, 16, 12);

  /// Whether this shell draws its own drag strip at all, instead of relying
  /// on a real OS-drawn title bar to make the window draggable. Mirrors
  /// [DesktopAppBar]'s own `useSystemTitleBar` branch for the same reason:
  /// once the system title bar is on, that real bar already spans the
  /// *entire* window width — sidebar included — and already handles
  /// dragging, so stacking this shell's own blank strip underneath it would
  /// just be dead space with nothing left to do. Desktop-gated for the same
  /// reason [DesktopAppBar] checks `DesktopIntegration.isDesktop` before
  /// ever touching `DragToMoveArea` — that widget assumes a real native
  /// window exists underneath it.
  static bool _usesCustomChrome(WidgetRef ref) =>
      DesktopIntegration.isDesktop && !ref.watch(systemTitleBarProvider);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final collapsedGroups = ref.watch(sidebarGroupCollapseProvider);

    // Flat index that keeps walking across group boundaries, so the sidebar's
    // notion of "item 4" stays the same as the bottom bar's and the rail's.
    // Collapsing a group only hides its *rows* below — flatIndex still
    // advances through every item in it exactly as if it were expanded, so
    // collapse state can never desync this sidebar's index scheme from the
    // one _navItems (and therefore the mobile/tablet layouts, and
    // onItemTapped's own route lookup) use.
    var flatIndex = 0;
    final rows = <Widget>[];
    for (final group in navGroups) {
      final header = group.header;
      final collapseKey = group.collapseKey;
      final collapsed =
          collapseKey != null && collapsedGroups.contains(collapseKey);
      if (header != null) {
        rows.add(
          _SectionHeader(
            label: header,
            collapsed: collapsed,
            onToggle: collapseKey == null
                ? null
                : () => ref
                      .read(sidebarGroupCollapseProvider.notifier)
                      .setCollapsed(collapseKey, !collapsed),
          ),
        );
      }
      for (final item in group.items) {
        // Snapshot per row: the callback would otherwise close over the loop
        // counter itself and every tile would report the final index.
        final index = flatIndex;
        if (!collapsed) {
          rows.add(
            _SidebarTile(
              item: item,
              isSelected: index == selectedIndex,
              onTap: () => onItemTapped(index),
            ),
          );
        }
        flatIndex++;
      }
      rows.add(const SizedBox(height: 8));
    }

    final usesCustomChrome = _usesCustomChrome(ref);
    final isMac = defaultTargetPlatform == TargetPlatform.macOS;

    final titleRow = Padding(
      padding: _titleRowPadding,
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

    // EchoMusic's Sidebar.vue: `h-full flex flex-col bg-bg-sidebar border-r
    // border-[var(--border-subtle)]` — a flush, edge-to-edge column with a
    // single hairline on the side facing the content column; no margin, no
    // rounding, no shadow. Replaces the floating GlassPanel this sidebar used
    // through v5.30.7: a panel with its own margin and rounded corners reads
    // as a second surface competing with the content column instead of one
    // continuous shell (the user's own "割裂感" field report), and floating
    // it needed increasingly special-cased macOS-only corner/margin logic
    // just to keep the native traffic lights from straddling its edge — see
    // the now-deleted `_macWindowCornerRadius`/`GlassPanel.borderRadiusOverride`.
    // Flush removes the seam those tricks kept patching instead of removing.
    //
    // Material(type: transparency), not a bare DecoratedBox: ListTile paints
    // its selected tile colour and ink splashes onto the nearest Material
    // ancestor, so anything opaque in between swallows both silently — the
    // same defect this sidebar had in v5.22.0. GlassPanel used to supply that
    // Material internally; now that it's gone from this call site, this
    // widget has to supply one itself. `type: transparency` paints nothing of
    // its own, so the DecoratedBox below still fully controls the visible
    // fill and hairline.
    return DecoratedBox(
      key: desktopSidebarKey,
      decoration: BoxDecoration(
        color: context.skinColors.surface,
        border: Border(
          right: BorderSide(
            color: context.skinColors.outlineVariant,
            width: 0.8,
          ),
        ),
      ),
      child: Material(
        type: MaterialType.transparency,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (usesCustomChrome)
              DragToMoveArea(
                child: SizedBox(
                  key: desktopSidebarDragStripKey,
                  width: double.infinity,
                  height: isMac ? _macDragStripHeight : _dragStripHeight,
                ),
              ),
            titleRow,
            const _AccountBlock(),
            const Divider(),
            Expanded(
              child: ListView(
                children: [
                  ...rows,
                  // Guest mode keeps the simple two-item shell (see
                  // [isGuest]'s own doc comment) — a guest account has no
                  // server-side playlists to preview at all, so this
                  // section would only ever show its own empty state,
                  // which is not what "范围限制" asked for here.
                  if (!isGuest) ...[
                    const Divider(height: 24),
                    const SidebarPlaylistsSection(),
                  ],
                ],
              ),
            ),
            if (!isGuest) const Divider(height: 1),
            if (!isGuest) const _SidebarBottomEntries(),
          ],
        ),
      ),
    );
  }
}

/// The sidebar's own footer — EchoMusic's `Sidebar.vue` pins 下载/装置 below
/// the scrollable nav content rather than letting them scroll away with it.
/// Guest-gated at the call site above, not here: a guest account has
/// nothing downloaded through this app's own account-gated download flow
/// (`_OfflineLibrarySection` in settings_screen.dart is itself `if
/// (!isGuest)`-gated for the same reason) and the output-device question
/// does not differ for a guest session, so there is nothing this footer
/// would show that isn't already covered by keeping the guest shell simple.
class _SidebarBottomEntries extends ConsumerWidget {
  const _SidebarBottomEntries();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    // Capability-driven, not a hardcoded hide: PlaybackCapabilities.
    // outputDeviceSelection is false under the current just_audio engine
    // (see just_audio_engine.dart's own capabilities getter, which
    // explains exactly why: "just_audio exposes none of the output chain").
    // That is *why* this row does not render today, not a TODO standing in
    // for a real check — once a future engine (media_kit, then the
    // in-house one) reports outputDeviceSelection: true, this appears with
    // no code change here at all. This is playback_engine.dart's own
    // stated rule ("a control the engine cannot honour must not be shown")
    // applied to sidebar chrome instead of a settings screen for the first
    // time.
    final canSelectOutputDevice = ref.watch(
      playbackCapabilitiesProvider.select((c) => c.outputDeviceSelection),
    );

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _SidebarBottomEntry(
            icon: Icons.download_outlined,
            label: t.downloads,
            // Reuses the existing Offline Library section in Settings
            // (download_notifier.dart's downloadProvider is what actually
            // manages downloads) rather than building a second downloads
            // screen — this is a shortcut into where that capability
            // already lives, not a new subsystem.
            onTap: () => context.go(AppRoutes.settings),
          ),
          if (canSelectOutputDevice)
            _SidebarBottomEntry(
              icon: Icons.speaker_outlined,
              label: t.outputDevice,
              // The actual output-device picker UI is out of scope for
              // this phase (it has nothing to drive yet — see the doc
              // comment above); Settings is where that control will live
              // once an engine actually reports this capability, matching
              // the download entry's own destination above.
              onTap: () => context.go(AppRoutes.settings),
            ),
        ],
      ),
    );
  }
}

class _SidebarBottomEntry extends StatelessWidget {
  const _SidebarBottomEntry({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(
        horizontal: _SidebarTile._horizontalInset,
        vertical: 2,
      ),
      child: ListTile(
        dense: true,
        contentPadding: _SidebarTile._contentPadding,
        leading: Icon(
          icon,
          size: 20,
          color: context.skinColors.onSurfaceVariant,
        ),
        title: Text(
          label,
          style: TextStyle(
            fontSize: 13,
            color: context.skinColors.onSurfaceVariant,
          ),
        ),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        onTap: onTap,
      ),
    );
  }
}

/// Shared "filled rounded card" chrome both [_AccountBlock] and
/// [_GuestSignInPrompt] sit inside — EchoMusic's own account block is a
/// distinct surface (a lightly filled, rounded rectangle) rather than a
/// bare row painted directly on the sidebar's own background, which is what
/// this file rendered through v5.32.0 (v5.33.0 field report: "整块应该是一个
/// 带浅填充的圆角卡片，不是裸的一行"). Pulled out once, with its own margin
/// replacing each variant's separate outer `Padding`, so the two variants
/// can't drift apart on inset/radius/fill even though their content
/// differs.
Widget _accountCard({required BuildContext context, required Widget child}) {
  return Container(
    margin: const EdgeInsets.fromLTRB(12, 0, 12, 8),
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
    decoration: BoxDecoration(
      color: context.skinColors.surfaceContainer,
      borderRadius: BorderRadius.circular(12),
    ),
    child: child,
  );
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

    return _accountCard(
      context: context,
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
                fontWeight: FontWeight.w600,
                color: context.skinColors.onSurface,
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
/// it — v5.33.0 changes only this widget's own chrome (the shared
/// [_accountCard] fill, and a second, primary "未登录" line above the
/// existing [AppLocalizations.tapToSignIn] prompt) and leaves that
/// `onTap` untouched. Settings is deliberately not duplicated here (unlike
/// the signed-in branch above): guest mode already carries Settings as its
/// own nav destination (see [_ShellScaffoldState._guestItems]).
class _GuestSignInPrompt extends ConsumerWidget {
  const _GuestSignInPrompt();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    return _accountCard(
      context: context,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(8),
          onTap: () => ref.read(authProvider.notifier).exitGuestMode(),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 2),
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
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        t.notSignedIn,
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: context.skinColors.onSurface,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        t.tapToSignIn,
                        style: TextStyle(
                          fontSize: 11,
                          color: context.skinColors.onSurfaceVariant,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
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

/// A nav group's header row — plain text (as before v5.33.0) when
/// [onToggle] is null (guest mode's single unlabelled group never collapses
/// — see [_ShellScaffoldState.build]'s guest branch, which passes
/// `collapseKey: null`), otherwise a tappable row with a chevron that flips
/// to reflect [collapsed]. The chevron's own rotation is what actually
/// carries the state visually — [InkWell] here needs the ambient
/// `Material(type: transparency)` [_DesktopSidebar] already wraps its whole
/// column in (the same requirement [_SidebarTile]'s own selected-pill ink
/// has), not a Material of its own.
class _SectionHeader extends StatelessWidget {
  const _SectionHeader({
    required this.label,
    this.collapsed = false,
    this.onToggle,
  });
  final String label;
  final bool collapsed;
  final VoidCallback? onToggle;

  @override
  Widget build(BuildContext context) {
    final row = Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
      child: Row(
        children: [
          Expanded(
            child: Text(
              label.toUpperCase(),
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                letterSpacing: 0.8,
                color: context.skinColors.outline,
              ),
            ),
          ),
          if (onToggle != null)
            AnimatedRotation(
              // 0 turns (pointing down) expanded, a quarter-turn
              // counter-clockwise (pointing right) collapsed — the same
              // expanded/collapsed convention a Material ExpansionTile
              // chevron uses, so this reads as "foldable" on sight without
              // needing a second visual cue.
              turns: collapsed ? -0.25 : 0,
              duration: const Duration(milliseconds: 150),
              child: Icon(
                Icons.expand_more,
                size: 16,
                color: context.skinColors.outline,
              ),
            ),
        ],
      ),
    );
    if (onToggle == null) return row;
    return InkWell(onTap: onToggle, child: row);
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

  /// Horizontal gap between the sidebar's own edge and the tile's selected
  /// pill. Without this the pill (painted across the ListTile's *own*
  /// bounds, which used to run edge-to-edge inside the panel) shared exactly
  /// the same x-coordinates as the sidebar's own hairline border (v5.30.7's
  /// GlassPanel border, now v5.31.0's flush DecoratedBox border — the inset
  /// this creates matters the same way regardless of which one is drawing
  /// it), so a selected row read as spilling out of the sidebar rather than
  /// sitting inside it (v5.30.7 field report).
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
///
/// [collapseKey] is a stable, untranslated identifier
/// ([SidebarGroupCollapseNotifier] persists collapsed state keyed by it, so
/// a language switch can't reset — or silently merge — a group's fold
/// state the way keying off the translated [header] text would) — null
/// alongside a null [header] for guest mode's single group, which never
/// collapses.
typedef _NavGroup = ({
  String? header,
  String? collapseKey,
  List<_NavItem> items,
});
