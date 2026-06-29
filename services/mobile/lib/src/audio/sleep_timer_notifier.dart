import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:just_audio/just_audio.dart';

import 'package:inori_music/main.dart' show audioHandler;
import 'package:inori_music/src/player/player_notifier.dart';

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

class SleepTimerState {
  const SleepTimerState({
    this.remaining,
    this.stopAfterTrack = false,
    this.active = false,
  });

  final Duration? remaining;
  final bool stopAfterTrack;
  final bool active;

  SleepTimerState copyWith({
    Duration? remaining,
    bool? stopAfterTrack,
    bool? active,
    bool clearRemaining = false,
  }) {
    return SleepTimerState(
      remaining: clearRemaining ? null : (remaining ?? this.remaining),
      stopAfterTrack: stopAfterTrack ?? this.stopAfterTrack,
      active: active ?? this.active,
    );
  }
}

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

final sleepTimerProvider =
    NotifierProvider<SleepTimerNotifier, SleepTimerState>(
  SleepTimerNotifier.new,
);

// ---------------------------------------------------------------------------
// Notifier
// ---------------------------------------------------------------------------

class SleepTimerNotifier extends Notifier<SleepTimerState> {
  Timer? _timer;
  StreamSubscription<ProcessingState>? _trackSub;

  @override
  SleepTimerState build() {
    ref.onDispose(_cleanup);
    return const SleepTimerState();
  }

  /// Start a fixed-duration countdown. Pauses the player on expiry.
  void startFixed(Duration duration) {
    _cleanup();
    var remaining = duration;
    state = SleepTimerState(remaining: remaining, active: true);
    _timer = Timer.periodic(const Duration(seconds: 1), (t) {
      remaining -= const Duration(seconds: 1);
      if (remaining <= Duration.zero) {
        t.cancel();
        _timer = null;
        ref.read(playerProvider.notifier).pause();
        state = const SleepTimerState();
      } else {
        state = state.copyWith(remaining: remaining);
      }
    });
  }

  /// Pause after the current track completes.
  void startAfterTrack() {
    _cleanup();
    state = const SleepTimerState(stopAfterTrack: true, active: true);
    _trackSub =
        audioHandler.audioPlayer.processingStateStream.listen((ps) {
      if (ps == ProcessingState.completed) {
        ref.read(playerProvider.notifier).pause();
        _cleanup();
        state = const SleepTimerState();
      }
    });
  }

  /// Cancel the active timer / after-track listener and reset state.
  void cancel() {
    _cleanup();
    state = const SleepTimerState();
  }

  // ---- private ----

  void _cleanup() {
    _timer?.cancel();
    _timer = null;
    _trackSub?.cancel();
    _trackSub = null;
  }
}
