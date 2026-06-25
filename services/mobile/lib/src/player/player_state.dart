import 'package:audio_service/audio_service.dart';

// ---------------------------------------------------------------------------
// Repeat / shuffle mode enums
// ---------------------------------------------------------------------------

enum RepeatMode { none, all, one }

// ---------------------------------------------------------------------------
// Player state
// ---------------------------------------------------------------------------

class PlayerState {
  PlayerState({
    this.queue = const [],
    this.currentIndex = -1,
    PlaybackState? playbackState,
    this.mediaItem,
    this.position = Duration.zero,
    this.duration = Duration.zero,
    this.volume = 1.0,
    this.shuffle = false,
    this.repeat = RepeatMode.none,
  }) : playbackState = playbackState ?? PlaybackState();

  /// Ordered playback queue.
  final List<MediaItem> queue;

  /// Index of the currently playing (or paused) item; -1 if empty.
  final int currentIndex;

  /// Raw audio_service playback state (contains playing/paused/buffering etc.)
  final PlaybackState playbackState;

  /// Currently playing MediaItem (mirrors queue[currentIndex]).
  final MediaItem? mediaItem;

  /// Current playback position.
  final Duration position;

  /// Duration of the current track.
  final Duration duration;

  /// Volume [0.0 – 1.0].
  final double volume;

  final bool shuffle;
  final RepeatMode repeat;

  // Convenience
  bool get isPlaying => playbackState.playing;
  bool get isBuffering =>
      playbackState.processingState == AudioProcessingState.buffering ||
      playbackState.processingState == AudioProcessingState.loading;
  bool get isIdle => queue.isEmpty || currentIndex < 0;

  PlayerState copyWith({
    List<MediaItem>? queue,
    int? currentIndex,
    PlaybackState? playbackState,
    MediaItem? mediaItem,
    Duration? position,
    Duration? duration,
    double? volume,
    bool? shuffle,
    RepeatMode? repeat,
    bool clearMediaItem = false,
  }) {
    return PlayerState(
      queue: queue ?? this.queue,
      currentIndex: currentIndex ?? this.currentIndex,
      playbackState: playbackState ?? this.playbackState,
      mediaItem: clearMediaItem ? null : (mediaItem ?? this.mediaItem),
      position: position ?? this.position,
      duration: duration ?? this.duration,
      volume: volume ?? this.volume,
      shuffle: shuffle ?? this.shuffle,
      repeat: repeat ?? this.repeat,
    );
  }
}
