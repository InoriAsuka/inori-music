import 'package:audio_service/audio_service.dart';

// ---------------------------------------------------------------------------
// Repeat / shuffle mode enums
// ---------------------------------------------------------------------------

enum RepeatMode { none, all, one }

/// A single playback failure occurrence, surfaced once (e.g. as a SnackBar)
/// by whichever screen is listening. Deliberately does not override `==` —
/// every instance is its own identity, so two failures with identical
/// [message] text (e.g. retrying the same broken file) still compare unequal
/// and a `ref.listen`/`.select` re-fires instead of being deduped away.
class PlaybackFailure {
  PlaybackFailure(this.message);
  final String message;
}

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
    this.playbackError,
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

  /// Most recent playback failure, if any — set by [PlayerNotifier.playTrack]
  /// when the audio source can't be resolved/loaded, so a UI shell can show
  /// it instead of the tap silently doing nothing. See [PlaybackFailure] for
  /// why this isn't just a `String?`.
  final PlaybackFailure? playbackError;

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
    PlaybackFailure? playbackError,
    bool clearMediaItem = false,
    bool clearPlaybackError = false,
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
      playbackError: clearPlaybackError
          ? null
          : (playbackError ?? this.playbackError),
    );
  }
}
