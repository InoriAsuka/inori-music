import 'dart:async';

import 'package:audio_service/audio_service.dart';
import 'package:flutter/material.dart' show Color;
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:inori_music/src/playback/playback_engine.dart';

/// Bridges the OS media session to the app: notification and lock-screen
/// controls in, playback state out.
///
/// **Only that.** Until v5.27.0 this class also owned the `AudioPlayer`, the
/// gapless queue, the crossfade envelope and the Android equalizer, and four
/// unrelated notifiers reached it through a global in `main.dart` to get at
/// them. All of that now lives behind [PlaybackEngine]; what is left here is
/// the part that genuinely belongs to `audio_service`.
///
/// Playback state is still owned by `PlayerNotifier` — this forwards, it does
/// not decide.
class InoriAudioHandler extends BaseAudioHandler
    with QueueHandler, SeekHandler {
  InoriAudioHandler(this._engine) {
    _forwardEngineState();
    queue.add([]);
  }

  final PlaybackEngine _engine;
  StreamSubscription<EnginePlaybackState>? _stateSub;

  // ---- OS controls in ----

  @override
  Future<void> play() => _engine.play();

  @override
  Future<void> pause() => _engine.pause();

  @override
  Future<void> stop() async {
    await _engine.stop();
    await super.stop();
  }

  @override
  Future<void> seek(Duration position) => _engine.seek(position);

  // Queue navigation is PlayerNotifier's call, not the engine's: it has to
  // resolve the next track's URL and metadata first. Signalled rather than
  // acted on.
  @override
  Future<void> skipToNext() async => customEvent.add({'action': 'next'});

  @override
  Future<void> skipToPrevious() async =>
      customEvent.add({'action': 'previous'});

  // ---- playback state out ----

  void _forwardEngineState() {
    _stateSub = _engine.stateStream.listen((engineState) {
      final playing = engineState == EnginePlaybackState.playing;
      playbackState.add(
        PlaybackState(
          controls: [
            MediaControl.skipToPrevious,
            if (playing) MediaControl.pause else MediaControl.play,
            MediaControl.skipToNext,
          ],
          systemActions: const {MediaAction.seek},
          androidCompactActionIndices: const [0, 1, 2],
          processingState: switch (engineState) {
            EnginePlaybackState.idle => AudioProcessingState.idle,
            EnginePlaybackState.loading => AudioProcessingState.loading,
            EnginePlaybackState.buffering => AudioProcessingState.buffering,
            EnginePlaybackState.playing ||
            EnginePlaybackState.paused => AudioProcessingState.ready,
            EnginePlaybackState.completed => AudioProcessingState.completed,
          },
          playing: playing,
          updatePosition: _engine.position,
        ),
      );
    });
  }

  Future<void> dispose() async {
    await _stateSub?.cancel();
  }

  /// Boots `audio_service` around [engine].
  ///
  /// On Windows and Linux `audio_service`'s platform interface substitutes a
  /// no-op implementation, so this succeeds there and the app starts — which
  /// is why a missing *playback* backend on those platforms shows up at the
  /// first play() rather than at launch.
  static Future<InoriAudioHandler> create(PlaybackEngine engine) {
    return AudioService.init(
      builder: () => InoriAudioHandler(engine),
      config: const AudioServiceConfig(
        androidNotificationChannelId: 'com.inori.music.channel.audio',
        androidNotificationChannelName: 'Inori Music',
        androidNotificationOngoing: true,
        androidStopForegroundOnPause: true,
        notificationColor: Color(0xFFD42062), // SakuraDusk primary pink
      ),
    );
  }
}

/// The OS media session bridge.
///
/// Separate from [playbackEngineProvider] because they answer different
/// questions: the engine is "what makes sound", this is "what the lock screen
/// shows". Merging them is what let unrelated notifiers reach the audio
/// player through the media session in the first place.
final mediaSessionProvider = Provider<InoriAudioHandler>((ref) {
  throw UnimplementedError(
    'mediaSessionProvider must be overridden — see main().',
  );
});
