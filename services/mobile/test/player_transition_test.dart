// player_transition_test.dart
//
// Covers the v5.32.0 player-page transition: the slide-up/scrim entrance
// (Apple Music / EchoMusic's shape, replacing the platform default
// FadeUpwardsPageTransitionsBuilder — a 25%-offset-plus-fade that read as
// cheap for a page this central to the app) and the drag-to-dismiss gesture
// layered on top of it. See player_transition.dart for the actual mechanics
// and the derivation behind each threshold/curve.
//
// Exercises the real playerPageTransitionsBuilder — the function router.dart
// wires into AppRoutes.player — through a plain PageRouteBuilder rather than
// the full go_router/auth stack: no existing test constructs the real
// production router (routerProvider is only ever used by main.dart), and
// this function renders identically regardless of which Navigator pushes it.
//
// Cannot use pumpAndSettle() anywhere here: once the player's own backdrop
// settles it may mount CoverFluidBackground, whose 60s/150s repeat()
// controllers never reach AnimationStatus.completed (see this repo's other
// cover-backdrop tests for the same guard) — every settle below is a bounded
// tester.pump(duration) instead.
import 'package:audio_service/audio_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/artwork_provider.dart';
import 'package:inori_music/src/catalog/cover_palette_provider.dart';
import 'package:inori_music/src/lyrics/lyric_line.dart';
import 'package:inori_music/src/lyrics/lyrics_provider.dart';
import 'package:inori_music/src/player/full_player_screen.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/player/player_transition.dart';
import 'package:inori_music/src/playback/playback_engine_provider.dart';

import 'support/fake_playback_engine.dart';

class _StubLyricsNotifier extends LyricsNotifier {
  @override
  Future<List<LyricLine>?> build(String trackId) async => null;
}

class _StubPlayerNotifier extends PlayerNotifier {
  _StubPlayerNotifier(this._state);
  final pstate.PlayerState _state;

  @override
  pstate.PlayerState build() => _state;
}

class _StubArtworkNotifier extends ArtworkUrlNotifier {
  _StubArtworkNotifier(this._url);
  final String? _url;

  @override
  Future<String?> build(String albumId) async => _url;
}

pstate.PlayerState _playing() {
  const item = MediaItem(id: 'track-1', title: 'Track', artist: 'Artist');
  return pstate.PlayerState(
    queue: [item],
    currentIndex: 0,
    mediaItem: item,
    playbackState: PlaybackState(playing: true),
    duration: const Duration(minutes: 3),
  );
}

/// The *route's own* transition duration for these tests — deliberately
/// short and independent of [playerTransitionDuration]/
/// [playerTransitionReverseDuration] (both real, unshortened production
/// values, used below to pump the local drag-settle animations they
/// actually drive). Mirrors exactly how router.dart wires
/// playerPageTransitionsBuilder: only the *route's* duration is a call-site
/// choice; the drag mechanics' own durations are constants inside
/// player_transition.dart, not something a caller can shorten.
const _routeTransitionDuration = Duration(milliseconds: 100);

/// Pumps a MaterialApp whose home screen pushes FullPlayerScreen through the
/// exact playerPageTransitionsBuilder router.dart uses, then settles the
/// entrance transition fully open.
Future<void> _pumpAndOpen(WidgetTester tester) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        playerProvider.overrideWith(() => _StubPlayerNotifier(_playing())),
        playbackEngineProvider.overrideWithValue(FakePlaybackEngine()),
        lyricsProvider.overrideWith(_StubLyricsNotifier.new),
        artworkUrlProvider.overrideWith(() => _StubArtworkNotifier(null)),
        coverPaletteProvider.overrideWith((ref, source) async => null),
      ],
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Builder(
          builder: (context) => Scaffold(
            body: Center(
              child: ElevatedButton(
                onPressed: () => Navigator.of(context).push(
                  PageRouteBuilder<void>(
                    transitionDuration: _routeTransitionDuration,
                    reverseTransitionDuration: _routeTransitionDuration,
                    pageBuilder: (context, animation, secondaryAnimation) =>
                        const SizedBox(),
                    transitionsBuilder: playerPageTransitionsBuilder,
                  ),
                ),
                child: const Text('home'),
              ),
            ),
          ),
        ),
      ),
    ),
  );

  await tester.tap(find.text('home'));
  await tester.pump();
  await tester.pump(_routeTransitionDuration);
  await tester.pump();
}

/// FullPlayerScreen's own top-left corner in global coordinates — Offset.zero
/// once fully open, moving toward the screen height as it closes.
Offset _playerOffset(WidgetTester tester) =>
    tester.getTopLeft(find.byType(FullPlayerScreen));

/// A point safely inside the drag handle (the top bar's title area) — clear
/// of the close button on the left and the queue/karaoke/etc. icon buttons
/// on the right, and well above where the seek bar / transport controls sit
/// lower in the screen. See FullPlayerScreen._dragHandle's own doc comment
/// for why the handle is scoped to just this region.
///
/// y: 30, not the top bar's own vertical midpoint (24) — measured against
/// the actual rendered GestureDetector, whose hit region is y:[22, 42] (the
/// *text's* rendered band, vertically centred within the taller 48px-tall
/// icon row beside it, not the full row height) — 30 sits comfortably
/// inside that with margin either way, where an estimate off the row's own
/// height would have landed just outside it (y: 20, 2px above the region's
/// own top edge, which is exactly what the first version of this test got
/// wrong).
Offset _dragHandlePoint(WidgetTester tester) =>
    _playerOffset(tester) + const Offset(200, 30);

/// Pumps enough simulated time for whichever local settle animation a drag
/// release just started (bounce-back uses [playerTransitionDuration],
/// commit-to-dismiss uses [playerTransitionReverseDuration] — both equal in
/// practice, see that constant's own doc comment) to finish, *plus* the
/// route's own (short, test-configured) reverse transition in case the
/// settle triggered a real pop — see _dismiss's doc comment on why that pop
/// is deferred until after the local animation completes.
///
/// The leading bare pump() (no duration) matters and isn't just tidiness:
/// starting an AnimationController.animateTo() from a raw gesture callback
/// (_handleDragEnd, invoked here from outside any widget's build/frame
/// callback) leaves its Ticker uncalibrated against the test binding's fake
/// clock until the *next* frame actually runs — verified empirically by
/// probing a bare AnimationController the same way. Without this pump()
/// first, the following duration-pump's time budget goes partly toward that
/// calibration instead of toward the animation itself, and the settle
/// doesn't finish within the budget below.
///
/// Durations are deliberately generous (2x the nominal ones) rather than
/// exact: the same probe showed a controller settle completing roughly one
/// polling step past its nominal duration even once correctly calibrated —
/// this pumps well past that margin instead of chasing an exact minimum.
Future<void> _settleAfterRelease(WidgetTester tester) async {
  await tester.pump();
  await tester.pump(playerTransitionReverseDuration * 2);
  await tester.pump(_routeTransitionDuration * 2);
  await tester.pump();
}

void main() {
  testWidgets('the player is fully on-screen once its entrance settles', (
    tester,
  ) async {
    await _pumpAndOpen(tester);
    expect(_playerOffset(tester), Offset.zero);
  });

  testWidgets(
    'the player is still off-screen (below the fold) partway through the '
    'entrance transition',
    (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            playerProvider.overrideWith(() => _StubPlayerNotifier(_playing())),
            playbackEngineProvider.overrideWithValue(FakePlaybackEngine()),
            lyricsProvider.overrideWith(_StubLyricsNotifier.new),
            artworkUrlProvider.overrideWith(() => _StubArtworkNotifier(null)),
            coverPaletteProvider.overrideWith((ref, source) async => null),
          ],
          child: MaterialApp(
            localizationsDelegates: AppLocalizations.localizationsDelegates,
            supportedLocales: AppLocalizations.supportedLocales,
            home: Builder(
              builder: (context) => Scaffold(
                body: Center(
                  child: ElevatedButton(
                    onPressed: () => Navigator.of(context).push(
                      PageRouteBuilder<void>(
                        transitionDuration: _routeTransitionDuration,
                        reverseTransitionDuration: _routeTransitionDuration,
                        pageBuilder: (context, animation, secondaryAnimation) =>
                            const SizedBox(),
                        transitionsBuilder: playerPageTransitionsBuilder,
                      ),
                    ),
                    child: const Text('home'),
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('home'));
      // Two bare pump()s with no simulated time elapsed: Navigator.push()'s
      // route installation needs a frame beyond the one the tap itself
      // lands in before the new route's OverlayEntry actually materializes
      // (verified empirically — find.byType(FullPlayerScreen) is still
      // empty after only one pump()) — *neither* pump advances the fake
      // clock, so the entrance is still nowhere near _routeTransitionDuration
      // once both have run.
      await tester.pump();
      await tester.pump();

      final screenHeight = tester.getSize(find.byType(FullPlayerScreen)).height;
      expect(
        _playerOffset(tester).dy,
        greaterThan(screenHeight * 0.5),
        reason:
            'A decelerating (easeOutCubic) entrance still has most of its '
            'distance left this early in the transition',
      );
    },
  );

  testWidgets(
    'dragging the top bar down tracks the finger 1:1 before release',
    (tester) async {
      await _pumpAndOpen(tester);

      final gesture = await tester.startGesture(_dragHandlePoint(tester));
      await gesture.moveBy(const Offset(0, 80));
      await tester.pump();
      expect(_playerOffset(tester).dy, closeTo(80, 1));

      await gesture.moveBy(const Offset(0, 40));
      await tester.pump();
      expect(_playerOffset(tester).dy, closeTo(120, 1));

      // Release without building any deliberate velocity (startGesture's
      // default timeStamp is Duration.zero on every event, which — per
      // dragFrom's own implementation, the mechanism WidgetTester.drag()
      // itself is built on — yields an effectively-zero computed release
      // velocity) so this exercises the *extent* threshold alone: 120px on
      // a 600px-tall default test surface is 20%, under
      // _dismissExtentThreshold's 30%.
      await gesture.up();
      await _settleAfterRelease(tester);

      expect(
        _playerOffset(tester),
        Offset.zero,
        reason: 'Under the extent threshold, a release must bounce back open',
      );
      expect(find.byType(FullPlayerScreen), findsOneWidget);
    },
  );

  testWidgets('releasing past the extent threshold dismisses the player', (
    tester,
  ) async {
    await _pumpAndOpen(tester);
    final screenHeight = tester.getSize(find.byType(FullPlayerScreen)).height;

    final gesture = await tester.startGesture(_dragHandlePoint(tester));
    // 40% of the screen height — past _dismissExtentThreshold's 30%.
    await gesture.moveBy(Offset(0, screenHeight * 0.4));
    await tester.pump();
    await gesture.up();
    await _settleAfterRelease(tester);

    expect(
      find.byType(FullPlayerScreen),
      findsNothing,
      reason: 'Past the extent threshold, release must dismiss the route',
    );
    expect(find.text('home'), findsOneWidget);
  });

  testWidgets(
    'a fast downward fling dismisses even without crossing the extent '
    'threshold',
    (tester) async {
      await _pumpAndOpen(tester);
      final screenHeight = tester.getSize(find.byType(FullPlayerScreen)).height;
      final start = _dragHandlePoint(tester);

      await tester.flingFrom(
        start,
        // 8% of the screen height — comfortably under the 30% extent
        // threshold, so only the velocity below can be what dismisses this.
        Offset(0, screenHeight * 0.08),
        // px/s, well above _flingVelocityThreshold's 700 (Flutter's own
        // Dismissible _kMinFlingVelocity, reused for this same decision).
        1500,
      );
      await _settleAfterRelease(tester);

      expect(
        find.byType(FullPlayerScreen),
        findsNothing,
        reason:
            'A flick fast enough must dismiss regardless of how little '
            'distance it covered',
      );
      expect(find.text('home'), findsOneWidget);
    },
  );

  testWidgets('the scrim darkens as the player is dragged toward closed', (
    tester,
  ) async {
    await _pumpAndOpen(tester);
    final screenHeight = tester.getSize(find.byType(FullPlayerScreen)).height;

    // The scrim is the one Opacity this subtree adds wrapping a plain black
    // ColoredBox — precise enough to not accidentally match some unrelated
    // Opacity/AnimatedOpacity elsewhere in Material's own widget internals.
    double scrimOpacity() {
      final match = tester.widgetList<Opacity>(find.byType(Opacity)).where((w) {
        final child = w.child;
        return child is ColoredBox && child.color == Colors.black;
      });
      expect(
        match,
        hasLength(1),
        reason: 'Expected to find exactly the player transition\'s own scrim',
      );
      return match.single.opacity;
    }

    final restOpacity = scrimOpacity();

    final gesture = await tester.startGesture(_dragHandlePoint(tester));
    await gesture.moveBy(Offset(0, screenHeight * 0.2));
    await tester.pump();

    expect(
      scrimOpacity(),
      greaterThan(restOpacity),
      reason: 'Dragging the player toward closed must darken the scrim less',
    );

    // Release under threshold so this test tears down cleanly (bounces back
    // rather than leaving a dangling pop).
    await gesture.up();
    await _settleAfterRelease(tester);
  });
}
