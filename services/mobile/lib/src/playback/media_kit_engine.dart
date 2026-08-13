import 'dart:async';

import 'package:media_kit/media_kit.dart';

import 'package:inori_music/src/playback/playback_engine.dart';

/// media_kit's capability declaration. A top-level constant rather than
/// something computed inside [MediaKitEngine] so it can be asserted on in a
/// test without constructing a real [Player] — that would reach for libmpv
/// itself, which this repo deliberately does not bundle outside Windows
/// (see pubspec.yaml: only `media_kit_libs_windows_audio`), so a real
/// [MediaKitEngine] cannot be built on the machines this test suite
/// actually runs on.
const PlaybackCapabilities mediaKitCapabilities = PlaybackCapabilities(
  equalizer: false,
  speedControl: true,
  gapless: true,
  // Not implemented on purpose. Crossfade has no libmpv primitive either —
  // JustAudioEngine's own crossfade is a hand-rolled volume ramp timed
  // against position, not something just_audio provides natively — and it
  // is the piece of this engine most likely to be thrown away once the
  // in-house engine lands. PlaybackCapabilities exists precisely so an
  // engine can decline a feature honestly instead of faking it.
  crossfade: false,
  // False here is an *interface* gap, not something libmpv can't do.
  // libmpv genuinely supports all three: output device enumeration and
  // selection (`audio-device` / `audio-device-list`), exclusive/hog mode
  // (`--audio-exclusive`), and sample-rate/format control
  // (`--audio-format` / `--audio-samplerate`) are all real mpv options.
  // What doesn't exist yet is a way to *say* any of this through
  // PlaybackEngine — the interface declares these three capability flags
  // but has no enumerate-devices/select-device method for an engine to
  // implement them against. shell_scaffold.dart's sidebar renders an
  // "output device" entry the moment outputDeviceSelection is true;
  // flipping it before the API and a real picker UI exist would produce a
  // sidebar entry that leads nowhere — the exact dead-entry shape this
  // codebase has spent the last several versions fixing elsewhere. Left
  // false on purpose. Device selection is a separate follow-up phase.
  outputDeviceSelection: false,
  exclusiveOutput: false,
  outputFormatControl: false,
);

/// [PlaybackEngine] backed by `media_kit` (libmpv), for Windows only.
///
/// **Why this exists**: `just_audio` registers no Windows implementation at
/// all — Windows has never been able to play audio in this app, and never
/// could without a second engine. This is a deliberately transitional fix,
/// not the final answer: an in-house Rust/Symphonia engine is planned
/// later.
///
/// **Why it matters beyond Windows**: an interface with exactly one
/// implementation has never actually been forced to prove it is engine-
/// agnostic. Writing this class is what tests whether `PlaybackEngine` is
/// shaped around playback in general or around `just_audio` specifically.
/// Where it turned out to be the latter, the gap is called out inline below
/// (volume range, duration nullability, seekToIndex's play-state contract)
/// rather than quietly worked around with no trace left behind.
///
/// **Scope**: transport, queue, volume, speed, repeat, shuffle and the
/// observation streams — the same surface `JustAudioEngine` covers minus
/// crossfade and the equalizer. Both are reported as unsupported through
/// [capabilities] rather than faked; see [mediaKitCapabilities].
class MediaKitEngine implements PlaybackEngine {
  MediaKitEngine._(this._player) {
    _stateController = StreamController<EnginePlaybackState>.broadcast();
    void emitState() => _stateController.add(_toEngineState(_player.state));
    // media_kit splits what just_audio reports as one PlayerState stream
    // into several independent ones. All four together are what "playback
    // state changed" means here, so each is wired to recompute and re-emit
    // the same coarse EnginePlaybackState the rest of the app already
    // branches on — see _toEngineState below.
    _playingSub = _player.stream.playing.listen((_) => emitState());
    _completedSub = _player.stream.completed.listen((_) => emitState());
    _bufferingSub = _player.stream.buffering.listen((_) => emitState());
    _playlistSub = _player.stream.playlist.listen((_) => emitState());
  }

  /// Builds the engine. Only ever called on Windows — see
  /// `engine_selection.dart` — so libmpv's own initialization (locating and
  /// loading the native library `media_kit_libs_windows_audio` bundles) is
  /// unconditional here, the same way `JustAudioEngine.create()`
  /// unconditionally constructs an `AudioPlayer`: each factory owns the
  /// setup that only makes sense on the platform it is actually invoked on.
  factory MediaKitEngine.create() {
    MediaKit.ensureInitialized();
    return MediaKitEngine._(Player());
  }

  final Player _player;

  late final StreamController<EnginePlaybackState> _stateController;
  StreamSubscription<bool>? _playingSub;
  StreamSubscription<bool>? _completedSub;
  StreamSubscription<bool>? _bufferingSub;
  StreamSubscription<Playlist>? _playlistSub;

  @override
  PlaybackCapabilities get capabilities => mediaKitCapabilities;

  @override
  EngineEqualizer? get equalizer => null;

  // ---- transport ----

  @override
  Future<void> play() => _player.play();

  @override
  Future<void> pause() => _player.pause();

  @override
  Future<void> stop() => _player.stop();

  @override
  Future<void> seek(Duration position) => _player.seek(position);

  @override
  Future<void> seekToIndex(int index) async {
    // Player.jump() calls play() before moving the playlist position (see
    // media_kit's native player implementation), unlike just_audio's
    // seek(index: ...), which only changes position/index and leaves
    // play/pause state untouched. Left alone, skipping tracks while paused
    // would silently resume playback on Windows only —
    // PlayerNotifier.next()/previous() don't ask for that, and just_audio
    // doesn't do it. seekToIndex's own doc comment says nothing about play
    // state either way, because only one engine existed when it was
    // written. Restoring the prior state here is what makes both engines
    // actually agree on a contract the interface only ever implied.
    final wasPlaying = _player.state.playing;
    await _player.jump(index);
    if (!wasPlaying) await _player.pause();
  }

  // ---- queue ----

  @override
  Future<void> setQueue(List<String> urls, {int initialIndex = 0}) =>
      _player.open(
        Playlist(urls.map(Media.new).toList(), index: initialIndex),
        // just_audio's setAudioSource() only prepares the source; playback
        // starts on an explicit play() call afterwards (every caller in
        // PlayerNotifier makes that call itself right after setQueue).
        // open()'s own default is to start playing immediately — matching
        // that default here would make this engine start audio a step
        // earlier than just_audio does for the exact same call sequence.
        play: false,
      );

  @override
  Future<void> appendToQueue(String url) => _player.add(Media(url));

  @override
  Future<void> insertIntoQueue(int index, String url) async {
    // add() only appends — like just_audio's own
    // ConcatenatingAudioSource, libmpv has no single "insert at position"
    // primitive — so this is append-then-move: two native calls instead of
    // one, but the same end state.
    await _player.add(Media(url));
    final lastIndex = _player.state.playlist.medias.length - 1;
    if (lastIndex != index) await _player.move(lastIndex, index);
  }

  @override
  Future<void> removeFromQueue(int index) => _player.remove(index);

  @override
  Future<void> moveInQueue(int oldIndex, int newIndex) =>
      _player.move(oldIndex, newIndex);

  // ---- parameters ----

  @override
  Future<void> setVolume(double volume) =>
      // PlaybackEngine.setVolume's 0.0-1.0 range is just_audio's own native
      // range, carried into the interface unexamined — media_kit/libmpv's
      // native range is 0-100 (Player.setVolume's own doc comment:
      // "Defaults to 100.0"). Passing the app's 0.0-1.0 value straight
      // through would not error, it would just make every volume level
      // sound like the player was almost muted. Scaled here, at the one
      // seam that has to know both conventions.
      _player.setVolume(volume.clamp(0.0, 1.0) * 100);

  @override
  Future<void> setSpeed(double speed) => _player.setRate(speed);

  @override
  Future<void> setRepeatMode(EngineRepeatMode mode) =>
      _player.setPlaylistMode(switch (mode) {
        EngineRepeatMode.off => PlaylistMode.none,
        EngineRepeatMode.one => PlaylistMode.single,
        EngineRepeatMode.all => PlaylistMode.loop,
      });

  @override
  Future<void> setShuffleEnabled(bool enabled) => _player.setShuffle(enabled);

  @override
  Future<void> shuffleQueue() async {
    // "Only meaningful while shuffle is on" (see the interface doc
    // comment) — and media_kit's setShuffle(bool) is a no-op whenever the
    // flag isn't actually changing, so calling setShuffle(true) again
    // would silently do nothing on a second "reshuffle" request. Toggling
    // off then back on forces a genuinely fresh shuffle either way.
    if (!_player.state.shuffle) return;
    await _player.setShuffle(false);
    await _player.setShuffle(true);
  }

  @override
  set crossfadeSeconds(int seconds) {
    // Accepted, not thrown. CrossfadeNotifier calls this setter
    // unconditionally on every launch and every settings change with no
    // capability check of its own — unlike the equalizer path, there is no
    // "null object" an engine can hand back to decline crossfade, because
    // it is a bare setter on PlaybackEngine itself rather than a queryable
    // sub-interface. capabilities.crossfade is false; this no-op is what
    // keeps that flag honest instead of the setter throwing on every app
    // start on Windows.
  }

  // ---- observation ----

  @override
  Stream<Duration> get positionStream => _player.stream.position;

  @override
  Stream<Duration?> get durationStream =>
      _player.stream.duration.map(_nullIfZero);

  @override
  Stream<EnginePlaybackState> get stateStream => _stateController.stream;

  @override
  Stream<int?> get currentIndexStream =>
      _player.stream.playlist.map((p) => p.medias.isEmpty ? null : p.index);

  @override
  Duration get position => _player.state.position;

  @override
  Duration? get duration => _nullIfZero(_player.state.duration);

  @override
  EnginePlaybackState get state => _toEngineState(_player.state);

  // PlayerState.duration defaults to Duration.zero and stays there until
  // real metadata loads — media_kit has no separate "unknown" signal the
  // way just_audio's nullable duration does. PlayerNotifier's own listener
  // only updates its state on a non-null duration, so passing Duration.zero
  // straight through would flash "0:00" as the total time on every track
  // change until the real value arrives a moment later. No real track is
  // actually zero-length, so folding zero into null here restores the
  // "unknown" meaning the interface's nullable Duration is meant to carry.
  static Duration? _nullIfZero(Duration d) => d == Duration.zero ? null : d;

  // media_kit collapses just_audio's separate "loading" (opening a new
  // source) and "buffering" (stalled mid-playback) states into a single
  // `buffering` flag — both fire from the same underlying mpv event.
  // Rather than guess which one a given buffering=true means, this engine
  // only ever reports .buffering and never .loading:
  // EnginePlaybackState.loading is simply unreachable from this engine.
  // That is a narrower distinction than the interface offers, reported
  // honestly rather than papered over with a guess.
  static EnginePlaybackState _toEngineState(PlayerState s) {
    if (s.playlist.medias.isEmpty) return EnginePlaybackState.idle;
    if (s.completed) return EnginePlaybackState.completed;
    if (s.buffering) return EnginePlaybackState.buffering;
    return s.playing ? EnginePlaybackState.playing : EnginePlaybackState.paused;
  }

  @override
  Future<void> dispose() async {
    await _playingSub?.cancel();
    await _completedSub?.cancel();
    await _bufferingSub?.cancel();
    await _playlistSub?.cancel();
    await _stateController.close();
    await _player.dispose();
  }
}
