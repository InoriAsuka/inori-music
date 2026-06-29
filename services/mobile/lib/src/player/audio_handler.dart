// ignore_for_file: implementation_imports
import 'package:audio_service/audio_service.dart';
import 'package:flutter/material.dart' show Color;
import 'package:just_audio/just_audio.dart';

/// [InoriAudioHandler] bridges just_audio with audio_service so that:
/// - Android shows a media notification with playback controls.
/// - iOS / macOS obey AVAudioSession (background audio, lock screen).
/// - Desktop MediaSession (Windows / Linux) controls work.
///
/// The actual playback state is managed by [PlayerNotifier]; this handler
/// acts as the bridge between the OS / notification layer and Riverpod.
class InoriAudioHandler extends BaseAudioHandler with QueueHandler, SeekHandler {
  final AudioPlayer _player;

  InoriAudioHandler(this._player) {
    _forwardPlayerState();
    queue.add([]);
  }

  // ---- BaseAudioHandler overrides ----

  @override
  Future<void> play() async => _player.play();

  @override
  Future<void> pause() async => _player.pause();

  @override
  Future<void> stop() async {
    await _player.stop();
    await super.stop();
  }

  @override
  Future<void> seek(Duration position) async => _player.seek(position);

  @override
  Future<void> skipToNext() async {
    // Signal to PlayerNotifier via the customEvent stream.
    customEvent.add({'action': 'next'});
  }

  @override
  Future<void> skipToPrevious() async {
    customEvent.add({'action': 'previous'});
  }

  // ---- Forwarding helpers ----

  void _forwardPlayerState() {
    _player.playerStateStream.listen((ps) {
      final playing = ps.playing;
      final processingState = _toAudioProcessingState(ps.processingState);
      playbackState.add(PlaybackState(
        playing: playing,
        processingState: processingState,
        controls: [
          MediaControl.skipToPrevious,
          if (playing) MediaControl.pause else MediaControl.play,
          MediaControl.skipToNext,
        ],
        systemActions: const {
          MediaAction.seek,
          MediaAction.skipToNext,
          MediaAction.skipToPrevious,
        },
        updatePosition: _player.position,
        bufferedPosition: _player.bufferedPosition,
        speed: _player.speed,
      ));
    });

    _player.positionStream.listen((pos) {
      final current = playbackState.valueOrNull;
      if (current != null) {
        playbackState.add(current.copyWith(updatePosition: pos));
      }
    });
  }

  /// Expose the player so PlayerNotifier can share the same instance.
  AudioPlayer get audioPlayer => _player;

  // ---- Gapless playback ----

  late ConcatenatingAudioSource _concatSource;

  /// Rebuild the concatenating source with the given URLs and start playback.
  Future<void> updateConcatQueue(List<String> urls) async {
    final sources = urls
        .map((u) => ProgressiveAudioSource(Uri.parse(u)))
        .toList();
    _concatSource = ConcatenatingAudioSource(children: sources);
    await _player.setAudioSource(_concatSource);
  }

  // ---- Equalizer ----

  /// Apply per-band gain values (in dB) to the underlying player.
  void applyEqualizerBands(List<double> bands) {
    try {
      // ignore: avoid_dynamic_calls
      (_player as dynamic).setBands(bands);
    } catch (_) {}
  }

  /// Reset all EQ bands to 0 dB.
  void resetEqualizer() {
    try {
      // ignore: avoid_dynamic_calls
      (_player as dynamic).setBands(List<double>.filled(10, 0.0));
    } catch (_) {}
  }

  static AudioProcessingState _toAudioProcessingState(ProcessingState ps) {
    switch (ps) {
      case ProcessingState.idle:
        return AudioProcessingState.idle;
      case ProcessingState.loading:
        return AudioProcessingState.loading;
      case ProcessingState.buffering:
        return AudioProcessingState.buffering;
      case ProcessingState.ready:
        return AudioProcessingState.ready;
      case ProcessingState.completed:
        return AudioProcessingState.completed;
    }
  }

  static Future<InoriAudioHandler> create() async {
    final player = AudioPlayer();
    return AudioService.init(
      builder: () => InoriAudioHandler(player),
      config: AudioServiceConfig(
        androidNotificationChannelId: 'com.inori.music.channel.audio',
        androidNotificationChannelName: 'Inori Music',
        androidNotificationOngoing: true,
        androidStopForegroundOnPause: true,
        notificationColor: const Color(0xFF9B5CFF), // NeonShrine primary violet
      ),
    );
  }
}
