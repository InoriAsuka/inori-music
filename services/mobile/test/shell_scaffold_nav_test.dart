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
// v5.31.0 drops the floating-GlassPanel sidebar (margin, rounded corners,
// its own macOS-only traffic-light gutter math) for an EchoMusic-style flush
// column: no margin, a single right-edge hairline, a flat blank drag strip
// (48px macOS / 24px other desktop platforms) standing in for the old
// geometry-derived `_macTitleRowTopInset`. The macOS-gutter tests below are
// replaced accordingly rather than merely retuned — the mechanism they were
// asserting against no longer exists. This file also gains coverage for the
// new cross-column accent gradient (`_LayoutAccentGradient`) that papers
// over the seam between the flush sidebar and the content column.
//
import 'package:audio_service/audio_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart' show LogicalKeyboardKey;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/auth/auth_notifier.dart';
import 'package:inori_music/src/catalog/playlists_screen.dart'
    show catalogPlaylistsProvider;
import 'package:inori_music/src/playback/playback_engine.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';
import 'package:inori_music/src/player/mini_player_bar.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/player/player_transition.dart';
import 'package:inori_music/src/player/queue_list.dart';
import 'package:inori_music/src/shared/router.dart';
import 'package:inori_music/src/shared/widgets/inori_mark.dart';
import 'package:inori_music/src/shared/widgets/shell_scaffold.dart';
import 'package:inori_music/src/user_playlist/user_playlist_notifier.dart';

import 'support/fake_playback_engine.dart';

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

/// v5.33.0: the desktop sidebar's own playlist tabs section
/// (`SidebarPlaylistsSection`) watches `userPlaylistProvider` for every
/// signed-in desktop test in this file now, not just the ones specifically
/// about playlists — without this stub it would hit the real notifier's
/// `_fetchPlaylists()`, a live Dio call this test harness has no server
/// behind (the same reason `catalogPlaylistsProvider` below is stubbed
/// too).
class _EmptyUserPlaylistNotifier extends UserPlaylistNotifier {
  @override
  Future<List<UserPlaylist>> build() async => const [];
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
          AppRoutes.myPlaylists,
          AppRoutes.forYou,
          AppRoutes.explore,
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
  // Defaults to PlaybackCapabilities.none, matching the real just_audio
  // engine's own reported capabilities (see just_audio_engine.dart) — every
  // pre-v5.33.0 test here implicitly assumed "no output-device entry",
  // which this default preserves without touching any of them.
  PlaybackCapabilities capabilities = PlaybackCapabilities.none,
}) => ProviderScope(
  overrides: [
    authProvider.overrideWith(() => _StubAuthNotifier(auth)),
    playerProvider.overrideWith(() => _StubPlayerNotifier(playerState)),
    playbackEngineProvider.overrideWithValue(
      FakePlaybackEngine(capabilities: capabilities),
    ),
    userPlaylistProvider.overrideWith(_EmptyUserPlaylistNotifier.new),
    catalogPlaylistsProvider.overrideWith((ref) async => const []),
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
  // v5.33.0: SidebarGroupCollapseNotifier persists to SharedPreferences,
  // whose test-time mock backing store is a plain static and does *not*
  // reset itself between individual `testWidgets` cases in the same file —
  // without this, a collapse left behind by one test (e.g. "collapsing a
  // group hides its rows...", below) would leak into whichever test runs
  // next and change what it starts from. Matches
  // search_history_notifier_test.dart's own setUp for the same class of
  // provider.
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  testWidgets('desktop sidebar groups destinations under section headers', (
    tester,
  ) async {
    // v5.33.0: the desktop sidebar's own two groups no longer share their
    // content with the mobile/tablet flat nav (_navItems) — see
    // shell_scaffold.dart's _desktopDiscoverItems/_desktopLibraryItems doc
    // comments. DISCOVER now holds the two EchoMusic-skeleton destinations
    // (For You/Explore) instead of catalog browsing, and catalog browsing
    // (Artists/Albums/Search) moved into LIBRARY alongside Favorites/
    // History — Playlists is deliberately absent from both: it now lives in
    // the sidebar's own tabbed playlist section instead of the flat groups
    // (see the "Playlists is reachable" test below for its new path).
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    expect(find.text('DISCOVER'), findsOneWidget);
    expect(find.text('LIBRARY'), findsOneWidget);
    for (final label in [
      'For You',
      'Explore',
      'Favorites',
      'History',
      'Artists',
      'Albums',
      'Search',
    ]) {
      expect(
        find.text(label),
        findsOneWidget,
        reason: '$label must have exactly one sidebar entry',
      );
    }
    expect(
      find.text('Playlists'),
      findsNothing,
      reason:
          'Playlists is no longer a flat nav item — it lives in the '
          'tabbed playlist section now',
    );
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

  testWidgets(
    'catalog playlists are reachable from the sidebar\'s Collected tab',
    (tester) async {
      // Regression guard, updated for v5.33.0: /playlists and
      // PlaylistsScreen both existed before v5.22.0 but nothing in the app
      // linked to them; that stayed true through v5.32.0 via a flat
      // "Playlists" nav item. v5.33.0 moves catalog playlists into the
      // sidebar's own tabbed playlist section (SidebarPlaylistsSection)
      // instead — the *destination* being reachable is what this guards,
      // not any particular nav-item shape, so this follows the section's
      // new path (Collected tab -> View All) rather than asserting the old
      // one is still there.
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Collected'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('View All'));
      await tester.pumpAndSettle();

      expect(find.text('body:${AppRoutes.playlists}'), findsOneWidget);
    },
  );

  testWidgets('this account\'s own playlists are reachable from the sidebar\'s '
      'Created tab', (tester) async {
    // The other half of the same guard — user_playlist_notifier.dart's
    // own playlists, previously unreachable from the desktop sidebar at
    // all (AppRoutes.myPlaylists existed but nothing linked to it either).
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    // Created is already the section's default tab — no tap needed to
    // select it, unlike the Collected case above.
    await tester.tap(find.text('View All'));
    await tester.pumpAndSettle();

    expect(find.text('body:${AppRoutes.myPlaylists}'), findsOneWidget);
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

  testWidgets(
    'the desktop sidebar sits flush against the window\'s top-left corner '
    '(v5.31.0 EchoMusic-style flush column — no floating panel, no margin)',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      expect(
        tester.getTopLeft(find.byKey(desktopSidebarKey)),
        Offset.zero,
        reason:
            'EchoMusic\'s Sidebar.vue is `h-full` with no margin — a '
            'floating panel (the pre-v5.31.0 shape) would start at a '
            'positive offset on both axes',
      );
    },
  );

  testWidgets(
    'the desktop sidebar paints only a right-edge hairline — no left/top/'
    'bottom border, no rounding, no shadow',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final box = tester.widget<DecoratedBox>(find.byKey(desktopSidebarKey));
      final decoration = box.decoration as BoxDecoration;
      final border = decoration.border! as Border;
      expect(border.right.width, greaterThan(0));
      expect(
        border.top,
        BorderSide.none,
        reason:
            'EchoMusic\'s sidebar has no top border — it is flush, not '
            'boxed',
      );
      expect(border.left, BorderSide.none);
      expect(border.bottom, BorderSide.none);
      expect(
        decoration.borderRadius,
        isNull,
        reason: 'A flush column has no corners to round',
      );
      expect(
        decoration.boxShadow,
        anyOf(isNull, isEmpty),
        reason: 'A flush column casts no shadow — it is not floating',
      );
    },
  );

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

      final sidebarHeight = tester
          .getSize(find.byKey(desktopSidebarKey))
          .height;
      // Window height (1000), full stop — v5.31.0 drops the 8px top/bottom
      // margins the floating-panel shape used to reserve, and critically
      // this is still *not* also reduced by the player bar's own height,
      // which is what the pre-v5.30.5 nested-Column shape used to do.
      expect(
        sidebarHeight,
        closeTo(1000, 1),
        reason:
            'A sidebar still sharing height with the player bar below it '
            'would be noticeably shorter than the full window height',
      );
    },
  );

  testWidgets('the player bar no longer runs full-width beneath the sidebar', (
    tester,
  ) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    final sidebarRight = tester.getTopRight(find.byKey(desktopSidebarKey)).dx;
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
      // Nowhere in the sidebar itself (list, account block, title row)
      // duplicates it.
      expect(
        find.descendant(
          of: find.byKey(desktopSidebarKey),
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
    'the sidebar\'s selected pill sits inset from the sidebar edge rather '
    'than spanning it edge-to-edge',
    (tester) async {
      // v5.30.7 field report: the selected pill's rounded background used
      // to run flush against the sidebar's own hairline border, reading as
      // if it were spilling out of it. Still true after v5.31.0 flushed the
      // sidebar itself against the window edge — the pill's own inset is
      // independent of whether the sidebar around it floats or not.
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final panelLeft = tester.getTopLeft(find.byKey(desktopSidebarKey)).dx;
      final panelWidth = tester.getSize(find.byKey(desktopSidebarKey)).width;
      final selectedTile = find
          .byWidgetPredicate((w) => w is ListTile && w.selected)
          .first;
      final tileRect = tester.getRect(selectedTile);

      expect(
        tileRect.left,
        greaterThan(panelLeft),
        reason: 'The pill must not start flush at the sidebar\'s own left edge',
      );
      expect(
        tileRect.right,
        lessThan(panelLeft + panelWidth),
        reason:
            'The pill must not extend all the way to the sidebar\'s right '
            'edge',
      );
    },
  );

  // -------------------------------------------------------------------------
  // v5.33.0 — the desktop queue drawer. Through v5.32.0 this same button
  // pushed AppRoutes.player outright (a deliberate v5.30.5 stopgap: "没有从
  // 主界面打开队列的通路就先跳播放页，不要为此新造队列 UI") — now that the desktop
  // shell has a spacious, stationary content column to dock a panel over,
  // that stopgap is replaced with an actual drawer instead of navigating
  // away. QueueList (queue_list.dart) is a public class specifically so this
  // drawer and FullPlayerScreen's own docked panel/bottom sheet share one
  // implementation — see full_player_layout_test.dart's own assertion that
  // the *same* type renders there.
  // -------------------------------------------------------------------------

  Widget appWithQueue() {
    final items = [
      MediaItem(id: 'q-1', title: 'First Track', artist: 'Artist A'),
      MediaItem(id: 'q-2', title: 'Second Track', artist: 'Artist B'),
    ];
    return _buildApp(
      _router(),
      playerState: pstate.PlayerState(
        queue: items,
        currentIndex: 0,
        mediaItem: items.first,
        playbackState: PlaybackState(playing: false),
      ),
    );
  }

  testWidgets(
    'the desktop player bar\'s queue button opens a docked drawer without '
    'leaving the current route',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(appWithQueue());
      await tester.pumpAndSettle();

      expect(find.byType(QueueList), findsNothing);

      await tester.tap(find.byIcon(Icons.queue_music));
      // Never pumpAndSettle a controller-driven slide (playerTransitionDuration
      // reuses the same non-repeating AnimationController this codebase's
      // other transitions use, so pumpAndSettle *would* eventually return
      // here — but every other test in this codebase advances explicitly by
      // the known duration instead, and doing the same here keeps this test
      // from being the odd one out if the drawer's animation is ever swapped
      // for something that does repeat).
      await tester.pump();
      await tester.pump(playerTransitionDuration);

      expect(
        find.byType(QueueList),
        findsOneWidget,
        reason: 'The drawer must actually render the shared queue list',
      );
      expect(
        find.text('body:${AppRoutes.artists}'),
        findsOneWidget,
        reason:
            'A drawer overlays the current page — it must not navigate '
            'away from it',
      );
      expect(find.text('body:${AppRoutes.player}'), findsNothing);
    },
  );

  testWidgets('tapping outside the open drawer closes it', (tester) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(appWithQueue());
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.queue_music));
    await tester.pump();
    await tester.pump(playerTransitionDuration);
    expect(find.byType(QueueList), findsOneWidget);

    // Inside the content column (sidebar is 220px wide) but well clear of
    // the drawer's own 360px-wide panel anchored to the content column's
    // right edge (window width 1400) — this point can only be hit by the
    // drawer's barrier, never by the panel itself.
    await tester.tapAt(const Offset(400, 300));
    await tester.pump();
    await tester.pump(playerTransitionReverseDuration);

    expect(find.byType(QueueList), findsNothing);
  });

  testWidgets('pressing Escape closes the open drawer', (tester) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(appWithQueue());
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.queue_music));
    await tester.pump();
    await tester.pump(playerTransitionDuration);
    expect(find.byType(QueueList), findsOneWidget);

    await tester.sendKeyEvent(LogicalKeyboardKey.escape);
    await tester.pump();
    await tester.pump(playerTransitionReverseDuration);

    expect(find.byType(QueueList), findsNothing);
  });

  testWidgets(
    'the closed drawer does not block clicks on the content underneath it',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(appWithQueue());
      await tester.pumpAndSettle();

      // Same point the "tapping outside" case above uses as its barrier
      // target — with the drawer closed this must reach whatever is
      // actually there instead, which a stray IgnorePointer(ignoring:
      // false) bug would prevent silently (no exception, just a dead tap).
      await tester.tapAt(const Offset(400, 300));
      await tester.pump();

      expect(tester.takeException(), isNull);
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
  // v5.31.0 — the sidebar's own blank drag strip replaces the geometry-
  // derived macOS-only top inset (`_macTitleRowTopInset`, deleted this
  // phase) with EchoMusic's flat `isMac ? 48 : 24` band. The sidebar is
  // flush against the window's top-left corner on *every* desktop platform
  // now — there is no floating-panel margin left for the traffic lights to
  // straddle — so unlike the v5.30.5-v5.30.7 tests this replaces, there is
  // no separate "does the panel itself move" question left to ask.
  // debugDefaultTargetPlatformOverride is reset synchronously at the end of
  // the test body, same as desktop_app_bar_test.dart — the binding's
  // end-of-test invariant check runs before addTearDown/tearDown would fire.
  // -------------------------------------------------------------------------

  testWidgets(
    'the sidebar\'s drag strip is 24px on non-macOS desktop platforms, with '
    'the wordmark row at a fixed inset below it',
    (tester) async {
      _useDesktopWindow(tester);
      // The drag strip only renders under DesktopIntegration.isDesktop (see
      // _DesktopSidebar._usesCustomChrome) — the ambient test-runner platform
      // is not itself one of macOS/windows/linux, so this needs an explicit
      // override to exercise the "real desktop, not macOS" branch at all.
      // Same convention desktop_app_bar_test.dart already uses throughout.
      debugDefaultTargetPlatformOverride = TargetPlatform.windows;
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final dragStripHeight = tester
          .getSize(find.byKey(desktopSidebarDragStripKey))
          .height;
      final markTop = tester.getTopLeft(find.byType(InoriMark)).dy;
      debugDefaultTargetPlatformOverride = null;

      expect(dragStripHeight, 24);
      expect(
        markTop,
        closeTo(24 + 24, 3),
        reason:
            'Sidebar origin (y=0, flush) + 24px drag strip + the wordmark '
            'row\'s own fixed 24px top padding',
      );
    },
  );

  testWidgets(
    'the sidebar\'s drag strip is 48px on macOS, with the wordmark row '
    'pushed further down to match',
    (tester) async {
      _useDesktopWindow(tester);
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final dragStripHeight = tester
          .getSize(find.byKey(desktopSidebarDragStripKey))
          .height;
      final markTop = tester.getTopLeft(find.byType(InoriMark)).dy;
      final sidebarTopLeft = tester.getTopLeft(find.byKey(desktopSidebarKey));
      debugDefaultTargetPlatformOverride = null;

      expect(dragStripHeight, 48);
      expect(
        sidebarTopLeft,
        Offset.zero,
        reason:
            'The sidebar is flush on every desktop platform now — macOS '
            'gets a taller drag strip, not a different margin',
      );
      expect(
        markTop,
        closeTo(48 + 24, 3),
        reason:
            'Sidebar origin (y=0, flush) + 48px drag strip (comfortably '
            'clears the traffic lights\' lowest point at window-y=26) + the '
            'wordmark row\'s own fixed 24px top padding',
      );
    },
  );

  // -------------------------------------------------------------------------
  // v5.31.0 — the sidebar dropped GlassPanel (which supplied its own
  // Material internally) for a plain DecoratedBox + Material(type:
  // transparency). This regressed twice before (v5.22.0, and again the
  // GlassPanel-based sidebar avoided it through v5.30.7) whenever something
  // opaque ended up between a ListTile and its nearest Material ancestor, so
  // it gets an explicit guard rather than relying on incidental coverage
  // from the tests above.
  // -------------------------------------------------------------------------

  testWidgets(
    'sidebar tiles still paint their selected background and ink after '
    'losing GlassPanel\'s built-in Material (regression guard)',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      // A held-down gesture (rather than a plain tap) forces Flutter to
      // actually paint an ink-splash frame while the pointer is down — a
      // missing Material ancestor throws during that paint, not merely on
      // tap-up, so a bare tap()+settle() would not exercise the failure
      // mode this guards against.
      final gesture = await tester.startGesture(
        tester.getCenter(find.text('History')),
      );
      await tester.pump(const Duration(milliseconds: 100));
      expect(tester.takeException(), isNull);
      await gesture.up();
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);

      final tile = tester
          .widgetList<ListTile>(find.byType(ListTile))
          .firstWhere((t) => (t.title! as Text).data == 'History');
      expect(tile.selected, isTrue);
    },
  );

  // -------------------------------------------------------------------------
  // v5.31.0 — the cross-column accent gradient that papers over the seam
  // between the flush sidebar and the content column (EchoMusic's own
  // `layout-accent-gradient`, see _DesktopLayout's doc comment).
  // -------------------------------------------------------------------------

  testWidgets(
    'the accent gradient spans both the sidebar and the content column, '
    'anchored to the window\'s top edge',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      final gradientRect = tester.getRect(find.byKey(layoutAccentGradientKey));
      expect(gradientRect.topLeft, Offset.zero);
      expect(
        gradientRect.width,
        closeTo(1400, 1),
        reason:
            'It must cover the full window width — both the 220px sidebar '
            'and the content column beside it — not just one of them',
      );
    },
  );

  testWidgets(
    'the accent gradient does not intercept clicks meant for what sits '
    'beneath it — a sidebar nav item under the wash is still selectable',
    (tester) async {
      _useDesktopWindow(tester);
      // macOS, deliberately: it has the taller (48px) drag strip, so its
      // header stack is the worst case for how far down the gradient needs
      // to reach — see _layoutAccentGradientHeight's doc comment for the
      // measured numbers this test is calibrated against. If this passes on
      // macOS it passes on every other desktop platform too. Reset
      // synchronously at the end of the test body below (not via
      // addTearDown) — same reason as the drag-strip tests above and
      // desktop_app_bar_test.dart: the binding's end-of-test invariant check
      // runs before addTearDown/tearDown callbacks would fire.
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
      // Artists is the default `_router()` initialLocation, which would make
      // it pre-selected — tapping an already-selected tile and observing it
      // still selected afterwards would not prove the tap actually landed.
      // Starting on Favorites instead (v5.33.0: the desktop sidebar's own
      // DISCOVER group no longer contains Artists at all — see
      // shell_scaffold.dart's _desktopDiscoverItems/_desktopLibraryItems —
      // so Favorites, LIBRARY's own first item, is what leaves the whole
      // DISCOVER group free to be this test's real, observable state
      // change now) leaves every item in the first (DISCOVER) group free.
      await tester.pumpWidget(
        _buildApp(_router(initialLocation: AppRoutes.favorites)),
      );
      await tester.pumpAndSettle();

      // Found dynamically rather than assumed by name/index: exactly which
      // row's on-screen *centre* (what tester.tap actually targets) falls
      // under the gradient's band depends on live measurements this test
      // should not have to hardcode (dense ListTile's rendered height,
      // Divider's default height, the account block's height, …). Skipping
      // already-selected tiles matters too — tapping a tile that was
      // selected before the tap proves nothing about whether the tap itself
      // landed.
      final gradientBottom = tester
          .getRect(find.byKey(layoutAccentGradientKey))
          .bottom;
      final tileFinder = find.byType(ListTile);
      int? targetIndex;
      for (var i = 0; i < tileFinder.evaluate().length; i++) {
        final candidate = tileFinder.at(i);
        final tile = tester.widget<ListTile>(candidate);
        if (tile.selected) continue;
        if (tester.getCenter(candidate).dy < gradientBottom) {
          targetIndex = i;
          break;
        }
      }
      expect(
        targetIndex,
        isNotNull,
        reason:
            'This test only proves something if some unselected nav tile '
            'actually has its tap-target centre under the gradient\'s band '
            '— if this fails, _layoutAccentGradientHeight no longer reaches '
            'far enough to be testing anything real',
      );
      final targetFinder = tileFinder.at(targetIndex!);
      final label = (tester.widget<ListTile>(targetFinder).title! as Text).data;

      await tester.tap(targetFinder);
      await tester.pumpAndSettle();

      final selected = tester
          .widgetList<ListTile>(find.byType(ListTile))
          .where((t) => t.selected)
          .toList();
      debugDefaultTargetPlatformOverride = null;
      expect(selected, hasLength(1));
      expect((selected.single.title! as Text).data, label);
    },
  );

  // -------------------------------------------------------------------------
  // v5.33.0 — collapsible groups + the flat-index trap. This codebase has
  // been bitten by the "closure captures the loop's final value, every tile
  // reports the same index" bug before (see the `final index = flatIndex;`
  // snapshot in _DesktopSidebar.build), and collapsing a group is exactly
  // the kind of change that could plausibly reintroduce it (or a sibling
  // bug: renumbering the *surviving* rows instead of just hiding some).
  // -------------------------------------------------------------------------

  testWidgets('collapsing a group hides its rows without renumbering any other '
      'group\'s tiles — a tap on a later item still dispatches to the '
      'correct route', (tester) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    expect(find.text('For You'), findsOneWidget);
    expect(find.text('Explore'), findsOneWidget);

    // The whole header row (label + chevron) shares one InkWell/onTap —
    // tapping the label text itself is enough to toggle it.
    await tester.tap(find.text('DISCOVER'));
    await tester.pumpAndSettle();

    expect(
      find.text('For You'),
      findsNothing,
      reason: 'Collapsing DISCOVER must hide its own rows',
    );
    expect(find.text('Explore'), findsNothing);
    // LIBRARY's own rows are unaffected by DISCOVER collapsing — this is
    // what proves flatIndex kept advancing through the hidden group
    // instead of skipping it.
    expect(find.text('Search'), findsOneWidget);

    await tester.tap(find.text('Search'));
    await tester.pumpAndSettle();

    expect(
      find.text('body:${AppRoutes.search}'),
      findsOneWidget,
      reason:
          'Search is the *last* item in the flat desktop list — if '
          'collapsing DISCOVER had renumbered anything, this is the tile '
          'most likely to land on the wrong route',
    );
  });

  testWidgets('a collapsed group can be expanded again, restoring its rows', (
    tester,
  ) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('DISCOVER'));
    await tester.pumpAndSettle();
    expect(find.text('For You'), findsNothing);

    await tester.tap(find.text('DISCOVER'));
    await tester.pumpAndSettle();
    expect(find.text('For You'), findsOneWidget);
    expect(find.text('Explore'), findsOneWidget);
  });

  // -------------------------------------------------------------------------
  // v5.33.0 — the "装置" (output device) footer entry is capability-driven:
  // PlaybackCapabilities.outputDeviceSelection is false under the current
  // just_audio engine (see just_audio_engine.dart), so this entry must not
  // render at all today — not a disabled/greyed-out state, an absence.
  // -------------------------------------------------------------------------

  testWidgets(
    'the sidebar\'s Output Device entry is absent when the engine reports '
    'no output-device-selection capability (today\'s default)',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      expect(
        find.text('Downloads'),
        findsOneWidget,
        reason: 'Downloads has no capability gate — it must still render',
      );
      expect(find.text('Output Device'), findsNothing);
    },
  );

  testWidgets(
    'the sidebar\'s Output Device entry appears once the engine reports '
    'outputDeviceSelection: true — proving this is capability-driven, not '
    'a hardcoded hide',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(
        _buildApp(
          _router(),
          capabilities: const PlaybackCapabilities(outputDeviceSelection: true),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Output Device'), findsOneWidget);
    },
  );

  // -------------------------------------------------------------------------
  // v5.33.0 — the guest-mode desktop sidebar must stay the simple shape
  // ("范围限制": guest gets the two nav items + account card + download, not
  // the full EchoMusic skeleton) even though the signed-in sidebar grew a
  // playlist section and a footer this phase.
  // -------------------------------------------------------------------------

  testWidgets(
    'guest mode\'s desktop sidebar has no playlist section and no Output '
    'Device entry — the EchoMusic skeleton addition is signed-in only',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(
        _buildApp(
          _router(initialLocation: AppRoutes.localLibrary),
          auth: _guest,
          capabilities: const PlaybackCapabilities(outputDeviceSelection: true),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Created'), findsNothing);
      expect(find.text('Collected'), findsNothing);
      expect(
        find.text('Output Device'),
        findsNothing,
        reason:
            'Guest mode omits the whole footer even when the capability '
            'would otherwise show it — the gate is isGuest, not just the '
            'capability',
      );
    },
  );

  testWidgets(
    'guest mode\'s desktop sidebar omits Downloads too — the whole footer '
    'is signed-in only, not just the capability-gated half of it',
    (tester) async {
      // This test documents the actual v5.33.0 decision rather than just
      // asserting it: the task brief left "download 对游客有无意义" an open
      // question ("如果下载对游客无意义就不显示"). download_notifier.dart's own
      // downloadProvider has no guest/signed-in branch of its own, but
      // settings_screen.dart's _OfflineLibrarySection — the screen this
      // entry actually opens — is itself `if (!isGuest)`-gated ("downloads
      // require an account"). Keeping the whole footer, Downloads included,
      // out of the guest shell (rather than rendering a Downloads entry
      // that opens a Settings screen with no Offline Library section left
      // to show) is the consistent reading of that existing gate.
      _useDesktopWindow(tester);
      await tester.pumpWidget(
        _buildApp(
          _router(initialLocation: AppRoutes.localLibrary),
          auth: _guest,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Downloads'), findsNothing);
    },
  );
}
