import 'dart:async';
import 'dart:io' show Platform;

import 'package:just_audio/just_audio.dart';

import 'package:inori_music/src/playback/playback_engine.dart';

/// [PlaybackEngine] backed by `just_audio`.
///
/// Owns everything that used to be spread across `InoriAudioHandler` and
/// `PlayerNotifier`: the player instance, the gapless concatenating source,
/// the crossfade volume envelope, and the Android equalizer. Those pieces all
/// manipulate the same underlying player, so keeping them in one place is
/// what stops a second caller from fighting the first — the crossfade ramp
/// and `setVolume` did exactly that before, via a shared mutable field on the
/// audio handler.
///
/// **Platform note:** `just_audio` 0.9.x registers plugin implementations for
/// android, ios, macos and web only, and its platform interface defaults to a
/// method channel with no fallback. On Windows and Linux every call therefore
/// throws `MissingPluginException`. [capabilities] reports that honestly
/// rather than pretending, so the UI can say so instead of failing silently.
class JustAudioEngine implements PlaybackEngine {
  JustAudioEngine._(this._player, this._equalizer) {
    _stateController = StreamController<EnginePlaybackState>.broadcast();
    _stateSub = _player.playerStateStream.listen((s) {
      _state = _toEngineState(s);
      _stateController.add(_state);
    });
    _fadePosSub = _player.positionStream.listen(_maybeFadeOut);
    _fadeIdxSub = _player.currentIndexStream.listen((idx) {
      if (idx == null) return;
      _fadeOutDone = false;
      if (_crossfadeSeconds > 0) _fadeIn();
    });
  }

  factory JustAudioEngine.create() {
    // Must stay null off Android: just_audio activates every effect in the
    // pipeline whenever the player goes active, and AndroidEqualizer's
    // activation calls an Android-only platform method. Constructing one
    // anywhere else made *every* playback attempt throw
    // MissingPluginException — the root cause found in v5.20.2.
    final equalizer = Platform.isAndroid ? AndroidEqualizer() : null;
    final player = AudioPlayer(
      audioPipeline: equalizer != null
          ? AudioPipeline(androidAudioEffects: [equalizer])
          : null,
    );
    return JustAudioEngine._(player, equalizer);
  }

  final AudioPlayer _player;
  final AndroidEqualizer? _equalizer;

  late final StreamController<EnginePlaybackState> _stateController;
  StreamSubscription<PlayerState>? _stateSub;
  EnginePlaybackState _state = EnginePlaybackState.idle;

  ConcatenatingAudioSource? _queueSource;

  /// The player instance, for the audio_service bridge and the Android
  /// equalizer only. Not for app logic — everything else goes through
  /// [PlaybackEngine].
  AudioPlayer get rawPlayer => _player;

  @override
  EngineEqualizer? get equalizer {
    final eq = _equalizer;
    return eq == null ? null : _AndroidEngineEqualizer(eq);
  }

  @override
  PlaybackCapabilities get capabilities => PlaybackCapabilities(
    equalizer: _equalizer != null,
    speedControl: true,
    gapless: true,
    crossfade: true,
    // just_audio exposes none of the output chain: no device enumeration, no
    // share-mode control, no sample-rate or bit-depth control. Reporting
    // these as false is what keeps the settings UI from offering controls
    // that would do nothing.
    outputDeviceSelection: false,
    exclusiveOutput: false,
    outputFormatControl: false,
  );

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
  Future<void> seekToIndex(int index) =>
      _player.seek(Duration.zero, index: index);

  // ---- queue ----

  @override
  Future<void> setQueue(List<String> urls, {int initialIndex = 0}) async {
    final source = ConcatenatingAudioSource(
      children: urls.map((u) => ProgressiveAudioSource(Uri.parse(u))).toList(),
    );
    _queueSource = source;
    await _player.setAudioSource(source, initialIndex: initialIndex);
  }

  @override
  Future<void> appendToQueue(String url) async =>
      _queueSource?.add(ProgressiveAudioSource(Uri.parse(url)));

  @override
  Future<void> insertIntoQueue(int index, String url) async =>
      _queueSource?.insert(index, ProgressiveAudioSource(Uri.parse(url)));

  @override
  Future<void> removeFromQueue(int index) async =>
      _queueSource?.removeAt(index);

  @override
  Future<void> moveInQueue(int oldIndex, int newIndex) async =>
      _queueSource?.move(oldIndex, newIndex);

  // ---- parameters ----

  @override
  Future<void> setVolume(double volume) async {
    // Remembered as the ceiling the fade ramps toward, so a crossfade in
    // flight can't strand playback at whatever partial level it was at.
    _targetVolume = volume;
    await _player.setVolume(volume);
  }

  @override
  Future<void> setSpeed(double speed) => _player.setSpeed(speed);

  @override
  Future<void> setRepeatMode(EngineRepeatMode mode) =>
      _player.setLoopMode(switch (mode) {
        EngineRepeatMode.off => LoopMode.off,
        EngineRepeatMode.one => LoopMode.one,
        EngineRepeatMode.all => LoopMode.all,
      });

  @override
  Future<void> setShuffleEnabled(bool enabled) =>
      _player.setShuffleModeEnabled(enabled);

  @override
  Future<void> shuffleQueue() => _player.shuffle();

  // ---- observation ----

  @override
  Stream<Duration> get positionStream => _player.positionStream;

  @override
  Stream<Duration?> get durationStream => _player.durationStream;

  @override
  Stream<EnginePlaybackState> get stateStream => _stateController.stream;

  @override
  Stream<int?> get currentIndexStream => _player.currentIndexStream;

  @override
  Duration get position => _player.position;

  @override
  Duration? get duration => _player.duration;

  @override
  EnginePlaybackState get state => _state;

  static EnginePlaybackState _toEngineState(PlayerState s) =>
      switch (s.processingState) {
        ProcessingState.idle => EnginePlaybackState.idle,
        ProcessingState.loading => EnginePlaybackState.loading,
        ProcessingState.buffering => EnginePlaybackState.buffering,
        ProcessingState.completed => EnginePlaybackState.completed,
        ProcessingState.ready =>
          s.playing ? EnginePlaybackState.playing : EnginePlaybackState.paused,
      };

  // ---- crossfade ----

  int _crossfadeSeconds = 0;
  double _targetVolume = 1.0;
  bool _fading = false;
  bool _fadeOutDone = false;
  StreamSubscription<Duration>? _fadePosSub;
  StreamSubscription<int?>? _fadeIdxSub;

  @override
  set crossfadeSeconds(int seconds) => _crossfadeSeconds = seconds;

  Future<void> _maybeFadeOut(Duration position) async {
    if (_crossfadeSeconds <= 0 || _fading || _fadeOutDone) return;
    final dur = _player.duration;
    if (dur == null || dur == Duration.zero) return;
    final remaining = dur - position;
    if (remaining.inMilliseconds > _crossfadeSeconds * 1000 ||
        remaining <= Duration.zero) {
      return;
    }
    _fading = true;
    _fadeOutDone = true;
    final steps = (remaining.inMilliseconds ~/ 100).clamp(
      1,
      _crossfadeSeconds * 10,
    );
    for (var i = steps; i >= 0; i--) {
      if (!_player.playing) break;
      await _player.setVolume(_targetVolume * i / steps);
      await Future<void>.delayed(const Duration(milliseconds: 100));
    }
    _fading = false;
  }

  Future<void> _fadeIn() async {
    if (_fading) return;
    _fading = true;
    final steps = _crossfadeSeconds * 10;
    for (var i = 0; i <= steps; i++) {
      if (!_player.playing) break;
      await _player.setVolume(_targetVolume * i / steps);
      await Future<void>.delayed(const Duration(milliseconds: 100));
    }
    // Restore the ceiling explicitly: breaking out of the loop early would
    // otherwise leave playback stuck at part volume.
    await _player.setVolume(_targetVolume);
    _fading = false;
  }

  @override
  Future<void> dispose() async {
    await _fadePosSub?.cancel();
    await _fadeIdxSub?.cancel();
    await _stateSub?.cancel();
    await _stateController.close();
    await _player.dispose();
  }
}

/// Adapts just_audio's [AndroidEqualizer] to [EngineEqualizer].
///
/// Every call is wrapped: `parameters` reaches the native effect, and a
/// device that reports an effect but then refuses to describe it would
/// otherwise take down the settings screen. An equalizer that quietly does
/// nothing is a far better outcome than a crash.
class _AndroidEngineEqualizer implements EngineEqualizer {
  _AndroidEngineEqualizer(this._eq);
  final AndroidEqualizer _eq;

  @override
  Future<void> setEnabled(bool enabled) => _eq.setEnabled(enabled);

  @override
  Future<int> bandCount() async {
    try {
      return (await _eq.parameters).bands.length;
    } catch (_) {
      return 0;
    }
  }

  @override
  Future<({double min, double max})> gainRange() async {
    try {
      final p = await _eq.parameters;
      return (min: p.minDecibels, max: p.maxDecibels);
    } catch (_) {
      return (min: 0.0, max: 0.0);
    }
  }

  @override
  Future<void> setBandGain(int index, double gainDb) async {
    try {
      final bands = (await _eq.parameters).bands;
      if (index < 0 || index >= bands.length) return;
      await bands[index].setGain(gainDb);
    } catch (_) {}
  }
}
