import 'dart:async';

/// Throttles cross-device player-state reporting (v5.4.0).
///
/// Semantics (mirrors the web client):
///   * While playback is active, [onReport] fires once per [interval] (30s).
///   * [reportNow] fires an immediate report — used on track change, pause, and
///     app background — and *resets the periodic phase* so the next periodic
///     report lands a full [interval] later, avoiding a double PUT right after
///     an immediate one.
///   * [setPlaying] gates the periodic timer: it runs only while playing so a
///     paused/idle player does not keep PUTting an unchanging snapshot.
///
/// The class owns a single [Timer]; callers must [dispose] it. It is
/// deliberately free of Riverpod/Dio so the throttle logic can be unit-tested
/// with `fake_async`.
class PlayerStateReporter {
  PlayerStateReporter({
    required Future<void> Function() onReport,
    this.interval = const Duration(seconds: 30),
  }) : _onReport = onReport;

  final Future<void> Function() _onReport;
  final Duration interval;

  Timer? _timer;
  bool _playing = false;
  bool _disposed = false;

  /// Number of times a report has been dispatched. Exposed for tests.
  int reportCount = 0;

  /// True while the periodic timer is armed.
  bool get isPeriodicActive => _timer != null;

  /// Update whether playback is active. Starts the periodic timer on the
  /// rising edge and cancels it on the falling edge. Idempotent.
  void setPlaying(bool playing) {
    if (_disposed || playing == _playing) return;
    _playing = playing;
    if (playing) {
      _arm();
    } else {
      _timer?.cancel();
      _timer = null;
    }
  }

  /// Immediately dispatch a report and reset the periodic phase (if armed).
  void reportNow() {
    if (_disposed) return;
    _fire();
    if (_playing) _arm(); // reset phase: next periodic is a full interval away
  }

  void _arm() {
    _timer?.cancel();
    _timer = Timer.periodic(interval, (_) => _fire());
  }

  void _fire() {
    if (_disposed) return;
    reportCount++;
    // Fire-and-forget: the supplied callback is responsible for swallowing its
    // own network errors so a failed PUT never breaks the timer loop.
    unawaited(_onReport());
  }

  void dispose() {
    _disposed = true;
    _timer?.cancel();
    _timer = null;
  }
}
