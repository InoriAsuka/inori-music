// shell_scaffold_nav_test.dart
//
// Covers the v5.22.0 navigation rework: the desktop sidebar's grouped
// sections and account block, and the fact that every destination — including
// the previously entry-less catalog Playlists route — is reachable at each
// breakpoint.
//
// v5.30.5 adds: the four-region desktop layout (full-height sidebar, player
// bar scoped to the content column), the mobile/tablet regression guard for
// the bar keeping its now-playing info, and the macOS traffic-light gutter
// handoff from DesktopAppBar to the sidebar itself.
//
// v5.30.6 reverts one v5.30.5 decision: that phase had docked a
// SidebarNowPlaying cover+title block at the sidebar's own foot and left the
// desktop player bar with mode controls only, on the theory that the
// v5.30.0 field report's red-boxed layout meant "move the cover out of the
// bar". The user's actual ask was that the cover stay *with* the transport
// controls, just not spread across the full window width — so
// SidebarNowPlaying is gone and the desktop bar carries its own cover+title
// again, same as every other breakpoint. The four-region layout itself
// (full-height sidebar, bar scoped to the content column) is unaffected —
// see mini_player_bar_desktop_test.dart for the bar's own wide-shape
// coverage.
//
import 'package:audio_service/audio_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/widgets/glass_panel.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';
import 'package:inori_music/src/shared/widgets/shell_scaffold.dart';

// ---------------------------------------------------------------------------
// Stubs — neither auth nor the player may touch the network/audio stack here.
// ---------------------------------------------------------------------------
class _StubAuthNotifier extends AuthNotifier {
  _StubAuthNotifier(this._state);
  final AuthState _state;

  @override
  Future<AuthState> build() async => _state;
}

class _StubPlayerNotifier extends PlayerNotifier {
  // Optional rather than required: the existing `_StubPlayerNotifier.new`
  // tear-off (a zero-arg `PlayerNotifier Function()`) stays valid against a
  // constructor with only optional parameters, so every pre-v5.30.7 call
  // site here keeps its exact prior behaviour (an empty PlayerState) with no
  // changes of its own.
  _StubPlayerNotifier([pstate.PlayerState? state])
    : _state = state ?? pstate.PlayerState();
  final pstate.PlayerState _state;

  @override
  pstate.PlayerState build() => _state;
}

/// The routes the shell navigates between. Bodies are placeholders — this
/// exercises the shell chrome, not the screens.
///
/// `/player` is registered as a top-level sibling of the `ShellRoute`, not
/// nested inside it — matching production's actual router.dart, where
/// FullPlayerScreen is a full-bleed route with no shell chrome behind it.
/// It needs to exist here at all only because the player bar's own
/// title/artist tap and its queue button both make a real
/// `context.push(AppRoutes.player)` call that needs somewhere to land.
GoRouter _router({String initialLocation = AppRoutes.artists}) => GoRouter(
  initialLocation: initialLocation,
  routes: [
    GoRoute(
      path: AppRoutes.player,
      builder: (_, _) =>
          Scaffold(body: Center(child: Text('body:${AppRoutes.player}'))),
    ),
    ShellRoute(
      builder: (context, state, child) => ShellScaffold(child: child),
      routes: [
        for (final path in [
          AppRoutes.artists,
          AppRoutes.albums,
          AppRoutes.search,
          AppRoutes.favorites,
          AppRoutes.history,
          AppRoutes.playlists,
          AppRoutes.settings,
          AppRoutes.localLibrary,
        ])
          GoRoute(
            path: path,
            builder: (_, _) =>
                Scaffold(body: Center(child: Text('body:$path'))),
          ),
        // v5.30.7 — the mini player bar's title/artist now link here (see
        // MiniPlayerBar's own doc comment). Real router.dart nests these
        // under /albums and /artists as ShellRoute children; registered as
        // siblings here instead purely so this test file's route list stays
        // a flat, scannable array like every other entry above it — go_router
        // resolves `push()` against the whole tree regardless of nesting
        // shape, so this is not a meaningfully different route graph for
        // what these tests actually exercise.
        GoRoute(
          path: AppRoutes.albumDetail,
          builder: (_, state) =>
              Scaffold(body: Text('album:${state.pathParameters['id']}')),
        ),
        GoRoute(
          path: AppRoutes.artistDetail,
          builder: (_, state) =>
              Scaffold(body: Text('artist:${state.pathParameters['id']}')),
        ),
      ],
    ),
  ],
);

const _signedIn = AuthState(
  status: AuthStatus.authenticated,
  username: 'inori',
  userId: 'u-1',
);
const _guest = AuthState(status: AuthStatus.guest);

Widget _buildApp(
  GoRouter router, {
  AuthState auth = _signedIn,
  pstate.PlayerState? playerState,
}) => ProviderScope(
  overrides: [
    authProvider.overrideWith(() => _StubAuthNotifier(auth)),
    playerProvider.overrideWith(() => _StubPlayerNotifier(playerState)),
  ],
  child: MaterialApp.router(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    routerConfig: router,
  ),
);

/// Drives the shell into the layout branch that renders `_DesktopSidebar`
/// (>= 1200 logical px wide).
void _useDesktopWindow(WidgetTester tester) {
  tester.view.devicePixelRatio = 1.0;
  tester.view.physicalSize = const Size(1400, 1000);
  addTearDown(tester.view.reset);
}

void main() {
  testWidgets('desktop sidebar groups destinations under section headers', (
    tester,
  ) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    expect(find.text('DISCOVER'), findsOneWidget);
    expect(find.text('LIBRARY'), findsOneWidget);
    for (final label in [
      'Artists',
      'Albums',
      'Search',
      'Favorites',
      'History',
      'Playlists',
    ]) {
      expect(
        find.text(label),
        findsOneWidget,
        reason: '$label must have exactly one sidebar entry',
      );
    }
  });

  testWidgets('sidebar account block shows the signed-in username', (
    tester,
  ) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    expect(find.text('inori'), findsOneWidget);
    // Avatar initial, derived from the username.
    expect(find.text('I'), findsOneWidget);
  });

  testWidgets('sidebar settings button navigates to /settings', (tester) async {
    _useDesktopWindow(tester);
    final router = _router();
    await tester.pumpWidget(_buildApp(router));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.settings_outlined));
    await tester.pumpAndSettle();

    expect(find.text('body:${AppRoutes.settings}'), findsOneWidget);
  });

  testWidgets('Playlists is reachable from the sidebar', (tester) async {
    // Regression guard: /playlists and PlaylistsScreen both existed before
    // v5.22.0 but nothing in the app linked to them.
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Playlists'));
    await tester.pumpAndSettle();

    expect(find.text('body:${AppRoutes.playlists}'), findsOneWidget);
  });

  testWidgets('selection tracks the active route across group boundaries', (
    tester,
  ) async {
    // The sidebar renders two groups but indexes them as one flat list; an
    // off-by-one there would highlight the wrong row for anything in the
    // second group.
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('History'));
    await tester.pumpAndSettle();

    final selected = tester
        .widgetList<ListTile>(find.byType(ListTile))
        .where((t) => t.selected)
        .toList();
    expect(selected, hasLength(1));
    expect((selected.single.title! as Text).data, 'History');
  });

  testWidgets('mobile bottom bar carries all six destinations', (tester) async {
    tester.view.devicePixelRatio = 1.0;
    tester.view.physicalSize = const Size(420, 900);
    addTearDown(tester.view.reset);

    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    final bar = tester.widget<NavigationBar>(find.byType(NavigationBar));
    expect(bar.destinations, hasLength(6));
    expect(
      bar.labelBehavior,
      NavigationDestinationLabelBehavior.onlyShowSelected,
      reason: 'Six labelled destinations do not fit at phone widths',
    );
  });

  // -------------------------------------------------------------------------
  // v5.23.0 — guest mode gets the same shell, not a bare Scaffold
  // -------------------------------------------------------------------------

  Widget guestApp() =>
      _buildApp(_router(initialLocation: AppRoutes.localLibrary), auth: _guest);

  testWidgets('guest mode gets navigation chrome instead of a bare body', (
    tester,
  ) async {
    // Before v5.23.0 the guest branch returned a Scaffold with no nav at all,
    // leaving Settings reachable only from the local library's own app bar.
    tester.view.devicePixelRatio = 1.0;
    tester.view.physicalSize = const Size(420, 900);
    addTearDown(tester.view.reset);

    await tester.pumpWidget(guestApp());
    await tester.pumpAndSettle();

    final bar = tester.widget<NavigationBar>(find.byType(NavigationBar));
    expect(bar.destinations, hasLength(2));
    expect(
      bar.labelBehavior,
      NavigationDestinationLabelBehavior.alwaysShow,
      reason: 'Two destinations comfortably fit full labels',
    );
    expect(find.text('Local Library'), findsOneWidget);
    expect(find.text('Settings'), findsOneWidget);
    // No server-catalog destinations leak into guest mode.
    expect(find.text('Artists'), findsNothing);
    expect(find.text('Favorites'), findsNothing);
  });

  testWidgets('guest mode can reach Settings from the nav', (tester) async {
    tester.view.devicePixelRatio = 1.0;
    tester.view.physicalSize = const Size(420, 900);
    addTearDown(tester.view.reset);

    await tester.pumpWidget(guestApp());
    await tester.pumpAndSettle();

    await tester.tap(find.text('Settings'));
    await tester.pumpAndSettle();

    expect(find.text('body:${AppRoutes.settings}'), findsOneWidget);
  });

  testWidgets(
    'guest sidebar shows a sign-in prompt instead of a placeholder name, '
    'without a second settings button',
    (tester) async {
      // v5.30.0: the account block used to render `t.guest` ("Guest") as if
      // it were a real username. EchoMusic's own account block treats "not
      // logged in" as a distinct, tappable state instead — this sidebar now
      // does the same.
      _useDesktopWindow(tester);
      await tester.pumpWidget(guestApp());
      await tester.pumpAndSettle();

      expect(find.text('Guest'), findsNothing);
      expect(find.text('Tap to sign in'), findsOneWidget);
      // Settings is already a destination here, so the account block must not
      // duplicate it — the only settings glyph on screen is the nav row's.
      expect(find.byIcon(Icons.settings_outlined), findsOneWidget);
      expect(
        find.descendant(
          of: find.byType(ListTile),
          matching: find.byIcon(Icons.settings_outlined),
        ),
        findsOneWidget,
      );
      // Two destinations don't warrant a section header.
      expect(find.text('DISCOVER'), findsNothing);
      expect(find.text('LIBRARY'), findsNothing);
    },
  );

  testWidgets('the desktop sidebar floats with a margin instead of sitting '
      'flush against the window edge', (tester) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    final topLeft = tester.getTopLeft(find.byType(GlassPanel));
    expect(
      topLeft.dx,
      greaterThan(0),
      reason: 'A flush sidebar would start at x=0',
    );
    expect(
      topLeft.dy,
      greaterThan(0),
      reason: 'A flush sidebar would start at y=0',
    );
  });

  // -------------------------------------------------------------------------
  // v5.30.5 — four-region desktop layout (field-report follow-up: the
  // sidebar and the player bar were fighting over the same strip of window
  // beneath the sidebar's old, shorter-than-full-height shape).
  // -------------------------------------------------------------------------

  testWidgets(
    'the desktop sidebar spans the window\'s full height, not just the row '
    'above the player bar',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final sidebarHeight = tester.getSize(find.byType(GlassPanel)).height;
      // Window height (1000) minus the 8px top/bottom margins _DesktopLayout
      // applies around the sidebar — critically, *not* also minus the player
      // bar's own height, which is what the old nested-Column shape reduced
      // it by.
      expect(
        sidebarHeight,
        closeTo(1000 - 8 - 8, 1),
        reason:
            'A sidebar still sharing height with the player bar below it '
            'would be noticeably shorter than window height minus its own '
            'margins',
      );
    },
  );

  testWidgets('the player bar no longer runs full-width beneath the sidebar', (
    tester,
  ) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    final sidebarRight = tester.getTopRight(find.byType(GlassPanel)).dx;
    final barLeft = tester.getTopLeft(find.byKey(MiniPlayerBar.contentKey)).dx;
    expect(
      barLeft,
      greaterThanOrEqualTo(sidebarRight),
      reason:
          'The bar must be scoped to the content column to the right of '
          'the sidebar, not span underneath it',
    );
  });

  testWidgets(
    'the desktop player bar carries its own now-playing info; the sidebar '
    'foot does not (v5.30.6 revert of the v5.30.5 SidebarNowPlaying detour)',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      expect(
        find.descendant(
          of: find.byKey(MiniPlayerBar.contentKey),
          matching: find.byType(MiniPlayerArtwork),
        ),
        findsOneWidget,
        reason: 'The cover+title block belongs with the transport controls',
      );
      // Nowhere in the sidebar's own floating panel (list, account block,
      // title row) duplicates it.
      expect(
        find.descendant(
          of: find.byType(GlassPanel),
          matching: find.byType(MiniPlayerArtwork),
        ),
        findsNothing,
      );
      // This window is wide enough that the bar's own measured width also
      // crosses its wide-shape threshold, so shuffle/repeat flank the
      // transport group too — see MiniPlayerBar's doc comment on the
      // width-driven switch that replaced the old showNowPlaying flag.
      expect(
        find.descendant(
          of: find.byKey(MiniPlayerBar.contentKey),
          matching: find.byIcon(Icons.shuffle),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: find.byKey(MiniPlayerBar.contentKey),
          matching: find.byIcon(Icons.repeat),
        ),
        findsOneWidget,
      );
    },
  );

  testWidgets('tapping the desktop player bar\'s cover opens the full player', (
    tester,
  ) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    // v5.30.7: the cover carries its own dedicated GestureDetector now
    // (see _nowPlayingInfo in mini_player_bar.dart) rather than the whole
    // content row sharing one outer InkWell — the field report was
    // explicit that only the cover should do this, since the title/artist
    // became links to their own detail pages (see the "does not open the
    // player" case right below) and the old shared InkWell would have
    // fought with that.
    await tester.tap(find.byType(MiniPlayerArtwork));
    await tester.pumpAndSettle();

    expect(find.text('body:${AppRoutes.player}'), findsOneWidget);
  });

  testWidgets(
    'tapping the desktop player bar\'s title/artist text does not open the '
    'full player — only the cover does',
    (tester) async {
      // Regression guard for the v5.30.7 field report ("只通过点击封面展开"):
      // before this phase the entire content row (including the title text)
      // shared one InkWell that opened the player, which would have fought
      // with the same tap turning into an album/artist navigation instead.
      _useDesktopWindow(tester);
      final mediaItem = MediaItem(
        id: 'track-1',
        title: 'Idol',
        artist: 'Yoasobi',
        // Deliberately no albumId/artistId — this proves the *player*
        // navigation stays off regardless of whether the text is itself
        // linkable, rather than conflating the two behaviours.
      );
      await tester.pumpWidget(
        _buildApp(
          _router(),
          playerState: pstate.PlayerState(
            queue: [mediaItem],
            currentIndex: 0,
            mediaItem: mediaItem,
            playbackState: PlaybackState(playing: false),
          ),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Idol'));
      await tester.pumpAndSettle();
      expect(find.text('body:${AppRoutes.player}'), findsNothing);

      await tester.tap(find.text('Yoasobi'));
      await tester.pumpAndSettle();
      expect(find.text('body:${AppRoutes.player}'), findsNothing);
    },
  );

  testWidgets(
    'the desktop player bar\'s title links to the album and the artist '
    'links to the artist, when the current track carries those ids',
    (tester) async {
      _useDesktopWindow(tester);
      final mediaItem = MediaItem(
        id: 'track-1',
        title: 'Idol',
        artist: 'Yoasobi',
        extras: {'albumId': 'album-1', 'artistId': 'artist-1'},
      );
      await tester.pumpWidget(
        _buildApp(
          _router(),
          playerState: pstate.PlayerState(
            queue: [mediaItem],
            currentIndex: 0,
            mediaItem: mediaItem,
            playbackState: PlaybackState(playing: false),
          ),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Idol'));
      await tester.pumpAndSettle();
      expect(find.text('album:album-1'), findsOneWidget);
    },
  );

  testWidgets(
    'a local (guest-mode) track with no ids renders plain, unlinked title '
    'text in the desktop player bar',
    (tester) async {
      _useDesktopWindow(tester);
      final mediaItem = MediaItem(
        id: 'local:track-1',
        title: 'Local Track',
        artist: 'Unknown Artist',
      );
      await tester.pumpWidget(
        _buildApp(
          _router(),
          playerState: pstate.PlayerState(
            queue: [mediaItem],
            currentIndex: 0,
            mediaItem: mediaItem,
            playbackState: PlaybackState(playing: false),
          ),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Local Track'));
      await tester.pumpAndSettle();

      // No crash, and definitely no navigation to a route this track has no
      // id for.
      expect(tester.takeException(), isNull);
      expect(find.text('Local Track'), findsOneWidget);
    },
  );

  testWidgets(
    'the sidebar\'s selected pill sits inset from the panel edge rather '
    'than spanning it edge-to-edge',
    (tester) async {
      // v5.30.7 field report: the selected pill's rounded background used
      // to run flush against GlassPanel's own hairline border, reading as
      // if it were spilling out of the panel.
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final panelLeft = tester.getTopLeft(find.byType(GlassPanel)).dx;
      final panelWidth = tester.getSize(find.byType(GlassPanel)).width;
      final selectedTile = find
          .byWidgetPredicate((w) => w is ListTile && w.selected)
          .first;
      final tileRect = tester.getRect(selectedTile);

      expect(
        tileRect.left,
        greaterThan(panelLeft),
        reason: 'The pill must not start flush at the panel\'s own left edge',
      );
      expect(
        tileRect.right,
        lessThan(panelLeft + panelWidth),
        reason:
            'The pill must not extend all the way to the panel\'s right edge',
      );
    },
  );

  testWidgets(
    'the desktop player bar\'s queue button opens the full player, since no '
    'standalone queue view exists outside it',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.queue_music));
      await tester.pumpAndSettle();

      expect(find.text('body:${AppRoutes.player}'), findsOneWidget);
    },
  );

  testWidgets(
    'mobile\'s bottom bar keeps its now-playing info block (regression '
    'guard)',
    (tester) async {
      tester.view.devicePixelRatio = 1.0;
      tester.view.physicalSize = const Size(420, 900);
      addTearDown(tester.view.reset);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      expect(
        find.descendant(
          of: find.byKey(MiniPlayerBar.contentKey),
          matching: find.byType(MiniPlayerArtwork),
        ),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: find.byKey(MiniPlayerBar.contentKey),
          matching: find.byIcon(Icons.shuffle),
        ),
        findsNothing,
        reason: 'The desktop-only mode controls must not leak into mobile',
      );
    },
  );

  testWidgets(
    'tablet\'s bottom bar keeps its now-playing info block (regression '
    'guard)',
    (tester) async {
      tester.view.devicePixelRatio = 1.0;
      tester.view.physicalSize = const Size(800, 900);
      addTearDown(tester.view.reset);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      expect(
        find.descendant(
          of: find.byKey(MiniPlayerBar.contentKey),
          matching: find.byType(MiniPlayerArtwork),
        ),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'the desktop shell does not overflow at the 1200dp breakpoint floor',
    (tester) async {
      tester.view.devicePixelRatio = 1.0;
      tester.view.physicalSize = const Size(1200, 800);
      addTearDown(tester.view.reset);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
    },
  );

  // -------------------------------------------------------------------------
  // v5.30.5 — macOS traffic-light gutter handoff. The gutter used to live on
  // DesktopAppBar (in the content column, x >= 236 — nowhere near the
  // lights); it now lives on the sidebar itself, which actually occupies the
  // window's top-left corner. debugDefaultTargetPlatformOverride is reset
  // synchronously at the end of the test body, same as desktop_app_bar_test.
  // dart — the binding's end-of-test invariant check runs before
  // addTearDown/tearDown would fire.
  // -------------------------------------------------------------------------

  testWidgets(
    'macOS sidebar reserves extra top padding for the traffic lights; other '
    'platforms keep the original inset',
    (tester) async {
      _useDesktopWindow(tester);

      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();
      final defaultPanelTopLeft = tester.getTopLeft(find.byType(GlassPanel));
      final defaultMarkTop = tester.getTopLeft(find.byType(InoriMark)).dy;
      expect(
        defaultMarkTop - defaultPanelTopLeft.dy,
        closeTo(24, 2),
        reason: 'Non-macOS keeps the pre-v5.30.5 24px title-row inset',
      );
      expect(
        defaultPanelTopLeft,
        const Offset(8, 8),
        reason: 'Non-macOS keeps floating 8px in from every edge',
      );

      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();
      final macPanelTopLeft = tester.getTopLeft(find.byType(GlassPanel));
      final macMarkTop = tester.getTopLeft(find.byType(InoriMark)).dy;
      final macPanel = tester.widget<GlassPanel>(find.byType(GlassPanel));
      debugDefaultTargetPlatformOverride = null;

      // v5.30.7 field report: the lights sat *outside* the panel's rounded
      // top-left corner, not merely close to it — the fix is the panel
      // sitting flush against that corner (see _DesktopLayout's own doc
      // comment on why more margin can never fix a fixed-position overlay),
      // not a bigger inset within the old floating shape.
      expect(
        macPanelTopLeft,
        Offset.zero,
        reason:
            'The panel must sit flush against the window\'s top-left corner '
            'so the traffic lights land inside it, not on the margin gap '
            'around it',
      );
      expect(
        macMarkTop - macPanelTopLeft.dy,
        closeTo(38, 2),
        reason:
            'macOS needs the enlarged inset to clear the traffic lights '
            '(see _DesktopSidebar._macTitleRowTopInset — re-derived for the '
            'flush-corner panel origin above, no longer the old 30 that '
            'assumed an 8px-inset panel)',
      );
      expect(
        macPanel.borderRadiusOverride,
        const BorderRadius.only(
          topLeft: Radius.circular(12),
          topRight: Radius.circular(16),
          bottomLeft: Radius.circular(16),
          bottomRight: Radius.circular(16),
        ),
        reason:
            'The flush corner should match the macOS window\'s own radius '
            '(12) rather than the panel\'s usual rounding (16), so it reads '
            'as part of the window frame instead of a second competing curve',
      );
    },
  );

  testWidgets(
    'non-macOS keeps the panel\'s uniform corner radius (no override)',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final panel = tester.widget<GlassPanel>(find.byType(GlassPanel));
      expect(panel.borderRadiusOverride, isNull);
    },
  );
}
