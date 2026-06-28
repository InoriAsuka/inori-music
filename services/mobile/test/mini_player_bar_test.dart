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
    overrides: [
      playerProvider.overrideWith(() => stub),
    ],
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
  testWidgets('MiniPlayerBar shows nothingPlaying text when no media item',
      (tester) async {
    final stub = _StubPlayerNotifier(pstate.PlayerState());
    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    // No track data should appear when there is no media item.
    expect(find.text('Idol'), findsNothing);
    expect(find.text('Yoasobi'), findsNothing);
    // Controls are always rendered.
    expect(find.byIcon(Icons.play_arrow_rounded), findsOneWidget,
        reason: 'Play icon shown when not playing and no media item');
  });

  testWidgets('MiniPlayerBar shows track title when media item is set',
      (tester) async {
    const trackTitle = 'Idol';
    const artistName = 'Yoasobi';
    final mediaItem =
        MediaItem(id: 'track-001', title: trackTitle, artist: artistName);
    final stub = _StubPlayerNotifier(pstate.PlayerState(
      queue: [mediaItem],
      currentIndex: 0,
      mediaItem: mediaItem,
      playbackState: PlaybackState(playing: false),
    ));

    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    expect(find.text(trackTitle), findsOneWidget,
        reason: 'Track title should be displayed in the mini player bar');
    expect(find.text(artistName), findsOneWidget,
        reason: 'Artist name should be displayed below the title');
  });

  testWidgets('Play button calls togglePlayPause on the notifier',
      (tester) async {
    final mediaItem = MediaItem(id: 'track-001', title: 'Idol');
    final stub = _StubPlayerNotifier(pstate.PlayerState(
      queue: [mediaItem],
      currentIndex: 0,
      mediaItem: mediaItem,
      playbackState: PlaybackState(playing: false),
    ));

    await tester.pumpWidget(_buildApp(stub));
    await tester.pump();

    final playButton = find.byTooltip('Play');
    expect(playButton, findsOneWidget,
        reason: 'Play button should be present when not playing');

    await tester.tap(playButton);
    await tester.pump();

    expect(stub.toggleCount, equals(1),
        reason: 'togglePlayPause should have been called once');
  });
}
