// full_player_transition_test.dart
//
// Covers the v5.32.0 fix for "点击封面展开播放页，会卡顿后才弹出播放页" — while
// FullPlayerScreen's own page-route transition (see player_transition.dart)
// is still sliding into view, it must render a cheap stand-in backdrop
// instead of the real CoverFluidBackground, which composites four separate
// image draws through a colour-matrix filter and then a 64-sigma
// BackdropFilter blur over the full screen (verified by reading
// cover_fluid_background.dart directly, not assumed from the field report
// alone) — real GPU work that landed on exactly the frames a slide-up
// transition needed to stay smooth.
//
// This does not go through go_router/player_transition.dart's real
// CustomTransitionPage wiring — a plain PageRouteBuilder with a short,
// fully-controlled transitionDuration is enough to drive
// FullPlayerScreen.transitionProgress exactly like the real router does
// (same Animation<double>, same push/pop lifecycle), without needing the
// rest of the app's routing/auth setup along for the ride.
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
import 'package:inori_music/src/playback/playback_engine_provider.dart';
import 'package:inori_music/src/shared/widgets/cover_fluid_background.dart';

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

/// Same shape as full_player_layout_test.dart's own stub — avoids a real
/// network round trip for the cover URL lookup.
class _StubArtworkNotifier extends ArtworkUrlNotifier {
  _StubArtworkNotifier(this._url);
  final String? _url;

  @override
  Future<String?> build(String albumId) async => _url;
}

pstate.PlayerState _playingWithArtwork(String albumId) {
  final item = MediaItem(
    id: 'track-1',
    title: 'Track',
    artist: 'Artist',
    extras: {'albumId': albumId},
  );
  return pstate.PlayerState(
    queue: [item],
    currentIndex: 0,
    mediaItem: item,
    playbackState: PlaybackState(playing: true),
    duration: const Duration(minutes: 3),
  );
}

void main() {
  testWidgets(
    'CoverFluidBackground is absent while the entrance transition is in '
    'flight, and present once it settles',
    (tester) async {
      const transitionDuration = Duration(milliseconds: 100);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            playerProvider.overrideWith(
              () => _StubPlayerNotifier(_playingWithArtwork('album-x')),
            ),
            playbackEngineProvider.overrideWithValue(FakePlaybackEngine()),
            lyricsProvider.overrideWith(_StubLyricsNotifier.new),
            artworkUrlProvider.overrideWith(
              () => _StubArtworkNotifier('https://example/a.jpg'),
            ),
            // A real, non-null palette so the cheap stand-in's accent branch
            // is exercised too (not just its no-palette fallback) — this is
            // also what MiniPlayerBar's pre-warming watch (the other half of
            // the v5.32.0 stutter fix) is meant to guarantee is already
            // resolved by the time a real user could ever open this screen.
            coverPaletteProvider.overrideWith(
              (ref, source) async =>
                  const CoverPalette(dominant: Color(0xFF224466)),
            ),
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
                        transitionDuration: transitionDuration,
                        reverseTransitionDuration: transitionDuration,
                        pageBuilder: (context, animation, secondaryAnimation) =>
                            FullPlayerScreen(
                              transition: PlayerTransition(
                                progress: animation,
                                onDragStart: () {},
                                onDragUpdate: (_) {},
                                onDragEnd: (_) {},
                              ),
                            ),
                      ),
                    ),
                    child: const Text('open'),
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('open'));
      // One frame with no simulated time elapsed: the push has started but
      // transitionDuration hasn't run yet, so the entrance animation is
      // still at (or essentially at) its starting value.
      await tester.pump();

      expect(
        find.byType(CoverFluidBackground),
        findsNothing,
        reason:
            'The real fluid backdrop must not build while the page is '
            'still sliding into view',
      );

      // Advance simulated time past the transition's own duration. A bare
      // pump(duration) rather than pumpAndSettle(): once CoverFluidBackground
      // *does* mount below, it starts its own 60s/150s repeat() controllers,
      // which never reach AnimationStatus.completed — pumpAndSettle would
      // hang waiting for them (project-wide testing note, see this repo's
      // other cover-backdrop tests for the same guard).
      await tester.pump(transitionDuration);
      await tester.pump();

      expect(
        find.byType(CoverFluidBackground),
        findsOneWidget,
        reason:
            'Once the entrance transition completes, the real backdrop must '
            'take over from the cheap stand-in',
      );
    },
  );
}
