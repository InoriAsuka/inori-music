// ignore_for_file: implementation_imports
import 'dart:async';

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

  /// Append a single track URL to the end of the gapless queue.
  Future<void> addSource(String url) async {
    await _concatSource.add(ProgressiveAudioSource(Uri.parse(url)));
  }

  /// Insert a single track URL at [index] in the gapless queue.
  Future<void> insertSource(int index, String url) async {
    await _concatSource.insert(index, ProgressiveAudioSource(Uri.parse(url)));
  }

  /// Remove the track at [index] from the gapless queue.
  Future<void> removeSourceAt(int index) async {
    await _concatSource.removeAt(index);
  }

  /// Move a track from [oldIndex] to [newIndex] in the gapless queue.
  Future<void> moveSource(int oldIndex, int newIndex) async {
    await _concatSource.move(oldIndex, newIndex);
  }

  // ---- Equalizer ----

  late final AndroidEqualizer _equalizer;

  /// Native Android equalizer effect. Only meaningful on Android — callers
  /// must guard with `Platform.isAndroid` before use.
  AndroidEqualizer get androidEqualizer => _equalizer;

  // ---- Crossfade support (fade-out at track end, fade-in on track change) ----

  /// Crossfade duration in seconds (0 = disabled).
  int crossfadeSeconds = 0;

  /// Volume basis for fades — the user's intended volume × ReplayGain,
  /// kept in sync by [PlayerNotifier._applyVolumeWithGain].
  double targetVolume = 1.0;

  bool _fading = false;
  bool _fadeOutDone = false;
  StreamSubscription<Duration>? _fadePosSub;
  StreamSubscription<int?>? _fadeIdxSub;

  /// Wire up crossfade listeners. Call once after [create].
  void initCrossfade() {
    _fadePosSub?.cancel();
    _fadeIdxSub?.cancel();
    _fadePosSub = _player.positionStream.listen(_maybeFadeOut);
    _fadeIdxSub = _player.currentIndexStream.listen((idx) {
      if (idx == null) return;
      _fadeOutDone = false;
      if (crossfadeSeconds > 0) _fadeIn();
    });
  }

  Future<void> _maybeFadeOut(Duration position) async {
    if (crossfadeSeconds <= 0 || _fading || _fadeOutDone) return;
    final dur = _player.duration;
    if (dur == null || dur == Duration.zero) return;
    final remaining = dur - position;
    if (remaining.inMilliseconds > crossfadeSeconds * 1000 || remaining <= Duration.zero) return;
    _fading = true;
    _fadeOutDone = true;
    final steps = (remaining.inMilliseconds ~/ 100).clamp(1, crossfadeSeconds * 10);
    for (int i = steps; i >= 0; i--) {
      if (!_player.playing) break;
      await _player.setVolume(targetVolume * i / steps);
      await Future<void>.delayed(const Duration(milliseconds: 100));
    }
    _fading = false;
  }

  Future<void> _fadeIn() async {
    if (_fading) return;
    _fading = true;
    final steps = crossfadeSeconds * 10;
    for (int i = 0; i <= steps; i++) {
      if (!_player.playing) break;
      await _player.setVolume(targetVolume * i / steps);
      await Future<void>.delayed(const Duration(milliseconds: 100));
    }
    await _player.setVolume(targetVolume); // 兜底恢复基准，避免停在半音量
    _fading = false;
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
    final equalizer = AndroidEqualizer();
    final player = AudioPlayer(
      audioPipeline: AudioPipeline(androidAudioEffects: [equalizer]),
    );
    final handler = InoriAudioHandler(player);
    handler._equalizer = equalizer;
    return AudioService.init(
      builder: () => handler,
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
