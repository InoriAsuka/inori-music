// ignore_for_file: implementation_imports
//
// mini_player_bar_test.dart
//
// Widget tests for MiniPlayerBar.
//
// Strategy: override playerProvider with a stub that subclasses PlayerNotifier
// and overrides build() to return a fixed PlayerState without touching
// audio_service or just_audio.
//
import 'package:audio_service/audio_service.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/player/mini_player_bar.dart';

// ---------------------------------------------------------------------------
// Stub PlayerNotifier — subclasses the real one but overrides build() to
// return a pre-built state and never initialise the audio subsystem.
// ---------------------------------------------------------------------------
class _StubPlayerNotifier extends PlayerNotifier {
  _StubPlayerNotifier(this._fixedState);
  final pstate.PlayerState _fixedState;

  int toggleCount = 0;

  @override
  pstate.PlayerState build() => _fixedState;

  @override
  Future<void> togglePlayPause() async {
    toggleCount++;
    state = state.copyWith(
      playbackState: PlaybackState(playing: !state.isPlaying),
    );
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

Widget _buildApp(_StubPlayerNotifier stub) {
  return ProviderScope(
    overrides: [playerProvider.overrideWith(() => stub)],
    child: MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: const Scaffold(body: MiniPlayerBar()),
    ),
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  testWidgets('MiniPlayerBar shows nothingPlaying text when no media item', (
    tester,
  ) async {
    final stub = _StubPlayerNotifier(pstate.PlayerState());
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    // No track data should appear when there is no media item.
    expect(find.text('Idol'), findsNothing);
    expect(find.text('Yoasobi'), findsNothing);
    // Controls are always rendered.
    expect(
      find.byIcon(Icons.play_arrow_rounded),
      findsOneWidget,
      reason: 'Play icon shown when not playing and no media item',
    );
  });

  testWidgets('MiniPlayerBar shows track title when media item is set', (
    tester,
  ) async {
    const trackTitle = 'Idol';
    const artistName = 'Yoasobi';
    final mediaItem = MediaItem(
      id: 'track-001',
      title: trackTitle,
      artist: artistName,
    );
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );

    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    expect(
      find.text(trackTitle),
      findsOneWidget,
      reason: 'Track title should be displayed in the mini player bar',
    );
    expect(
      find.text(artistName),
      findsOneWidget,
      reason: 'Artist name should be displayed below the title',
    );
  });

  testWidgets('Play button calls togglePlayPause on the notifier', (
    tester,
  ) async {
    final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );

    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    final playButton = find.byTooltip('Play');
    expect(
      playButton,
      findsOneWidget,
      reason: 'Play button should be present when not playing',
    );

    await tester.tap(playButton);
    await tester.pump();

    expect(
      stub.toggleCount,
      equals(1),
      reason: 'togglePlayPause should have been called once',
    );
  });

  // ---------------------------------------------------------------------
  // v5.30.0 — EchoMusic scale (requirement.md v5.30.0 / plan Phase v5.30.0)
  // ---------------------------------------------------------------------

  testWidgets('the content row is 84px tall, EchoMusic scale', (tester) async {
    final stub = _StubPlayerNotifier(pstate.PlayerState());
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    final size = tester.getSize(find.byKey(MiniPlayerBar.contentKey));
    expect(
      size.height,
      84,
      reason:
          'EchoMusic\'s PlayerBar.vue footer (h-21) is a fixed 84px; Apple '
          "Music's thinner bar was explicitly rejected as too small once "
          'the player page itself was reworked to match Apple Music.',
    );
  });

  testWidgets('the artwork is 56px, EchoMusic scale (was 44px)', (
    tester,
  ) async {
    final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    // No albumId on this mediaItem, so the artwork renders its fallback
    // music-note icon inside the cover box — the box itself is what v5.30.0
    // actually specifies, regardless of what's drawn inside it.
    final box = tester.widget<Container>(
      find.ancestor(
        of: find.byIcon(Icons.music_note),
        matching: find.byType(Container),
      ),
    );
    expect(box.constraints?.maxWidth, 56);
    expect(box.constraints?.maxHeight, 56);
  });

  testWidgets('transport icons render at EchoMusic\'s 22px, not the old 24px', (
    tester,
  ) async {
    final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
    final stub = _StubPlayerNotifier(
      pstate.PlayerState(
        queue: [mediaItem],
        currentIndex: 0,
        mediaItem: mediaItem,
        playbackState: PlaybackState(playing: false),
      ),
    );
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    final prevIcon = tester.widget<Icon>(find.byIcon(Icons.skip_previous));
    final nextIcon = tester.widget<Icon>(find.byIcon(Icons.skip_next));
    expect(prevIcon.size, 22);
    expect(nextIcon.size, 22);
  });

  testWidgets(
    'the transport trio stays centred regardless of which side has the '
    'sleep timer button',
    (tester) async {
      // Before v5.30.0 the sleep timer button sat directly after "next" with
      // nothing to balance it on the title side, so the trio was not
      // actually centred — this is the same three-section centring
      // full_player_screen.dart's own transport row already relies on.
      final mediaItem = MediaItem(
        id: 'track-001',
        title: 'A very very very long track title that eats up space',
        artist: 'Some Artist',
      );
      final stub = _StubPlayerNotifier(
        pstate.PlayerState(
          queue: [mediaItem],
          currentIndex: 0,
          mediaItem: mediaItem,
          playbackState: PlaybackState(playing: false),
        ),
      );
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      final barCentre = tester
          .getCenter(find.byKey(MiniPlayerBar.contentKey))
          .dx;
      final trioCentre =
          (tester.getCenter(find.byIcon(Icons.skip_previous)).dx +
              tester.getCenter(find.byIcon(Icons.skip_next)).dx) /
          2;
      expect((trioCentre - barCentre).abs(), lessThan(4));
    },
  );
}
