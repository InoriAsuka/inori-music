import 'dart:async';

import 'package:inori_music/src/playback/playback_engine.dart';

/// In-memory [PlaybackEngine] for tests: records what it was asked to do and
/// makes no platform calls.
///
/// Before v5.27.0 nothing above the player could be tested without booting
/// the real `just_audio` stack through a global in `main.dart`, so tests
/// either avoided those paths or leaned on `Platform.isAndroid` being false
/// on the test host — which tested the host, not the code.
class FakePlaybackEngine implements PlaybackEngine {
  FakePlaybackEngine({
    this.capabilities = PlaybackCapabilities.none,
    this.equalizer,
  });

  @override
  final PlaybackCapabilities capabilities;

  @override
  final FakeEngineEqualizer? equalizer;

  // ---- recorded calls ----
  final calls = <String>[];
  List<String>? lastQueue;
  int? lastInitialIndex;
  double? lastVolume;
  double? lastSpeed;
  EngineRepeatMode? lastRepeatMode;
  bool? lastShuffleEnabled;
  int crossfadeSecondsSet = 0;

  final _stateController = StreamController<EnginePlaybackState>.broadcast();
  final _positionController = StreamController<Duration>.broadcast();
  final _durationController = StreamController<Duration?>.broadcast();
  final _indexController = StreamController<int?>.broadcast();

  /// Drives [stateStream] as if the engine had changed state.
  void emitState(EnginePlaybackState next) {
    _state = next;
    _stateController.add(next);
  }

  EnginePlaybackState _state = EnginePlaybackState.idle;

  @override
  Future<void> play() async => calls.add('play');

  @override
  Future<void> pause() async => calls.add('pause');

  @override
  Future<void> stop() async => calls.add('stop');

  @override
  Future<void> seek(Duration position) async => calls.add('seek:$position');

  @override
  Future<void> seekToIndex(int index) async => calls.add('seekToIndex:$index');

  @override
  Future<void> setQueue(List<String> urls, {int initialIndex = 0}) async {
    calls.add('setQueue');
    lastQueue = List.of(urls);
    lastInitialIndex = initialIndex;
  }

  @override
  Future<void> appendToQueue(String url) async => calls.add('append:$url');

  @override
  Future<void> insertIntoQueue(int index, String url) async =>
      calls.add('insert:$index:$url');

  @override
  Future<void> removeFromQueue(int index) async => calls.add('remove:$index');

  @override
  Future<void> moveInQueue(int oldIndex, int newIndex) async =>
      calls.add('move:$oldIndex:$newIndex');

  @override
  Future<void> setVolume(double volume) async => lastVolume = volume;

  @override
  Future<void> setSpeed(double speed) async => lastSpeed = speed;

  @override
  Future<void> setRepeatMode(EngineRepeatMode mode) async =>
      lastRepeatMode = mode;

  @override
  Future<void> setShuffleEnabled(bool enabled) async =>
      lastShuffleEnabled = enabled;

  @override
  Future<void> shuffleQueue() async => calls.add('shuffleQueue');

  @override
  set crossfadeSeconds(int seconds) => crossfadeSecondsSet = seconds;

  @override
  Stream<Duration> get positionStream => _positionController.stream;

  @override
  Stream<Duration?> get durationStream => _durationController.stream;

  @override
  Stream<EnginePlaybackState> get stateStream => _stateController.stream;

  @override
  Stream<int?> get currentIndexStream => _indexController.stream;

  @override
  Duration get position => Duration.zero;

  @override
  Duration? get duration => null;

  @override
  EnginePlaybackState get state => _state;

  @override
  Future<void> dispose() async {
    await _stateController.close();
    await _positionController.close();
    await _durationController.close();
    await _indexController.close();
  }
}

/// Equalizer that records gains instead of touching a platform channel.
class FakeEngineEqualizer implements EngineEqualizer {
  FakeEngineEqualizer({this.bands = 5, this.min = -15, this.max = 15});

  /// Bands the "device" exposes — deliberately not the ten the UI shows, so
  /// tests exercise the mapping rather than a 1:1 pass-through.
  final int bands;
  final double min;
  final double max;

  bool? enabled;
  final gains = <int, double>{};

  @override
  Future<void> setEnabled(bool value) async => enabled = value;

  @override
  Future<int> bandCount() async => bands;

  @override
  Future<({double min, double max})> gainRange() async => (min: min, max: max);

  @override
  Future<void> setBandGain(int index, double gainDb) async =>
      gains[index] = gainDb;
}
