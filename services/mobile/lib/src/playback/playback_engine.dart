import 'dart:async';

/// What the app is allowed to ask of whatever is actually producing sound.
///
/// Everything above this line — the Riverpod notifiers, the screens — talks
/// only to this interface. Nothing above it imports `just_audio`, and nothing
/// above it reaches for a global singleton. That is the whole point: the
/// engine underneath is replaceable without touching a single screen.
///
/// The immediate reason this exists is that Windows has no `just_audio`
/// implementation at all, so the app cannot play there and never could. The
/// fix is a different engine, not a patch to this one — and a swap is only
/// cheap if the seam is already in place.
///
/// The shape follows foobar2000's split (decode/DSP/output as separate
/// interfaces rather than one monolithic player). We can only take the
/// coarse half of it today: both `just_audio` and libmpv are monolithic —
/// they decode *and* output, with no seam in between to hook. Splitting
/// decode from output for real needs an engine we own end to end (Rust +
/// Symphonia + miniaudio). [PlaybackCapabilities] is what lets the UI cope
/// with that difference instead of hard-coding what today's engine can do.
abstract class PlaybackEngine {
  // ---- transport ----

  Future<void> play();
  Future<void> pause();
  Future<void> stop();
  Future<void> seek(Duration position);

  /// Jump to a queue position. Replaces separate next/previous calls so
  /// queue navigation has exactly one entry point.
  Future<void> seekToIndex(int index);

  // ---- queue ----

  /// Replaces the whole queue and starts at [initialIndex].
  Future<void> setQueue(List<String> urls, {int initialIndex = 0});

  Future<void> appendToQueue(String url);
  Future<void> insertIntoQueue(int index, String url);
  Future<void> removeFromQueue(int index);
  Future<void> moveInQueue(int oldIndex, int newIndex);

  // ---- parameters ----

  Future<void> setVolume(double volume);
  Future<void> setSpeed(double speed);
  Future<void> setRepeatMode(EngineRepeatMode mode);
  Future<void> setShuffleEnabled(bool enabled);

  /// Reshuffles the queue's play order (only meaningful while shuffle is on).
  Future<void> shuffleQueue();

  /// Seconds of fade at each track boundary; 0 disables it.
  set crossfadeSeconds(int seconds);

  // ---- observation ----

  Stream<Duration> get positionStream;
  Stream<Duration?> get durationStream;
  Stream<EnginePlaybackState> get stateStream;
  Stream<int?> get currentIndexStream;

  Duration get position;
  Duration? get duration;
  EnginePlaybackState get state;

  // ---- introspection ----

  /// What this engine can actually do on the current platform.
  ///
  /// The rule this enables, straight out of the cross-platform output-settings
  /// checklist: **a control the engine cannot honour must not be shown.** The
  /// alternative — what this codebase did before — is `Platform.isAndroid`
  /// checks sprinkled through the UI, which is both a guess about the engine
  /// and wrong the moment the engine changes.
  PlaybackCapabilities get capabilities;

  /// The engine's equalizer, or null when it has none. Always null when
  /// [PlaybackCapabilities.equalizer] is false.
  EngineEqualizer? get equalizer;

  Future<void> dispose();
}

/// Minimal equalizer contract.
///
/// Band count is a *query*, not a constant: the app's UI shows ten fixed
/// bands, while the device underneath commonly exposes five. Anything that
/// hard-codes the count silently maps the wrong gains onto the wrong
/// frequencies.
abstract class EngineEqualizer {
  Future<void> setEnabled(bool enabled);

  /// How many bands the effect actually exposes.
  Future<int> bandCount();

  /// Gain limits, in dB.
  Future<({double min, double max})> gainRange();

  Future<void> setBandGain(int index, double gainDb);
}

/// Loop behaviour, named independently of any engine's own enum so switching
/// engines doesn't ripple a type change through the app.
enum EngineRepeatMode { off, one, all }

/// Coarse playback state. Deliberately smaller than `just_audio`'s
/// `ProcessingState` × `playing` matrix: the app only ever branches on these
/// five, and a narrower contract is easier for a new engine to satisfy
/// honestly.
enum EnginePlaybackState {
  idle,
  loading,
  buffering,
  playing,
  paused,
  completed,
}

/// What an engine supports. Every field is a question the UI would otherwise
/// answer by guessing from `Platform`.
class PlaybackCapabilities {
  const PlaybackCapabilities({
    this.equalizer = false,
    this.speedControl = false,
    this.gapless = false,
    this.crossfade = false,
    this.outputDeviceSelection = false,
    this.exclusiveOutput = false,
    this.outputFormatControl = false,
  });

  /// A built-in equalizer the app can drive. False does **not** mean the user
  /// has no EQ — it may exist at the OS level (Spotube, for instance, hands
  /// Android a session id and lets the system equalizer attach). It means
  /// *this app's* EQ screen has nothing to talk to.
  final bool equalizer;

  final bool speedControl;

  /// Track-to-track playback with no silence between.
  final bool gapless;

  /// Overlapping fade at track boundaries. Mutually exclusive with [gapless]
  /// in practice — a crossfade by definition is not gapless.
  final bool crossfade;

  /// Enumerating and choosing an output device.
  final bool outputDeviceSelection;

  /// WASAPI exclusive / CoreAudio hog mode — bypassing the system mixer.
  final bool exclusiveOutput;

  /// Setting output sample rate and bit depth rather than accepting whatever
  /// the system mixer negotiates.
  final bool outputFormatControl;

  /// Nothing beyond the basics. The honest default for an engine that hasn't
  /// declared otherwise.
  static const none = PlaybackCapabilities();
}
