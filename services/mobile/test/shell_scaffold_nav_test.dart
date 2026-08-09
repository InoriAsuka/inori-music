// shell_scaffold_nav_test.dart
//
// Covers the v5.22.0 navigation rework: the desktop sidebar's grouped
// sections and account block, and the fact that every destination — including
// the previously entry-less catalog Playlists route — is reachable at each
// breakpoint.
//
// v5.30.5 adds: the four-region desktop layout (full-height sidebar, player
// bar scoped to the content column), the sidebar's own SidebarNowPlaying
// foot block versus the desktop player bar's now-controls-only shape, the
// mobile/tablet regression guard for the bar keeping its now-playing info,
// and the macOS traffic-light gutter handoff from DesktopAppBar to the
// sidebar itself.
//
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
  @override
  pstate.PlayerState build() => pstate.PlayerState();
}

/// The routes the shell navigates between. Bodies are placeholders — this
/// exercises the shell chrome, not the screens.
///
/// `/player` is registered as a top-level sibling of the `ShellRoute`, not
/// nested inside it — matching production's actual router.dart, where
/// FullPlayerScreen is a full-bleed route with no shell chrome behind it.
/// It needs to exist here at all only because v5.30.5 gave both
/// SidebarNowPlaying and the desktop player bar's queue button a real
/// `context.push(AppRoutes.player)` call to land somewhere.
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

Widget _buildApp(GoRouter router, {AuthState auth = _signedIn}) =>
    ProviderScope(
      overrides: [
        authProvider.overrideWith(() => _StubAuthNotifier(auth)),
        playerProvider.overrideWith(_StubPlayerNotifier.new),
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
    'the sidebar foot carries now-playing info; the desktop player bar '
    'carries mode controls instead',
    (tester) async {
      _useDesktopWindow(tester);
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();

      expect(find.byType(SidebarNowPlaying), findsOneWidget);
      expect(
        find.descendant(
          of: find.byKey(MiniPlayerBar.contentKey),
          matching: find.byType(MiniPlayerArtwork),
        ),
        findsNothing,
        reason:
            'Now-playing info moved to the sidebar; the bar must not '
            'duplicate it',
      );
      // The width that artwork/title used to occupy now carries shuffle and
      // repeat instead (see MiniPlayerBar.showNowPlaying's doc comment).
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

  testWidgets('tapping the sidebar now-playing tile opens the full player', (
    tester,
  ) async {
    _useDesktopWindow(tester);
    await tester.pumpWidget(_buildApp(_router()));
    await tester.pumpAndSettle();

    await tester.tap(find.byType(SidebarNowPlaying));
    await tester.pumpAndSettle();

    expect(find.text('body:${AppRoutes.player}'), findsOneWidget);
  });

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
      final defaultPanelTop = tester.getTopLeft(find.byType(GlassPanel)).dy;
      final defaultMarkTop = tester.getTopLeft(find.byType(InoriMark)).dy;
      expect(
        defaultMarkTop - defaultPanelTop,
        closeTo(24, 2),
        reason: 'Non-macOS keeps the pre-v5.30.5 24px title-row inset',
      );

      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
      await tester.pumpWidget(_buildApp(_router()));
      await tester.pumpAndSettle();
      final macPanelTop = tester.getTopLeft(find.byType(GlassPanel)).dy;
      final macMarkTop = tester.getTopLeft(find.byType(InoriMark)).dy;
      debugDefaultTargetPlatformOverride = null;

      expect(
        macMarkTop - macPanelTop,
        closeTo(30, 2),
        reason:
            'macOS needs the enlarged inset to clear the traffic lights '
            '(see _DesktopSidebar._macTitleRowTopInset)',
      );
    },
  );
}
