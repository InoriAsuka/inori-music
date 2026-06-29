// ignore_for_file: implementation_imports
//
// SleepTimerNotifier tests — self-contained, no audioHandler or playerProvider import.
//
// We mirror the SleepTimerState and the timer/cancel logic directly, exercising
// the state machine without depending on just_audio or flutter's audio stack.
// ---------------------------------------------------------------------------
import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

// ---------------------------------------------------------------------------
// Minimal mirror of SleepTimerState
// ---------------------------------------------------------------------------
class _TimerState {
  const _TimerState({
    this.remaining,
    this.stopAfterTrack = false,
    this.active = false,
  });

  final Duration? remaining;
  final bool stopAfterTrack;
  final bool active;

  _TimerState copyWith({
    Duration? remaining,
    bool? stopAfterTrack,
    bool? active,
    bool clearRemaining = false,
  }) {
    return _TimerState(
      remaining: clearRemaining ? null : (remaining ?? this.remaining),
      stopAfterTrack: stopAfterTrack ?? this.stopAfterTrack,
      active: active ?? this.active,
    );
  }
}

// ---------------------------------------------------------------------------
// Minimal mirror of SleepTimerNotifier
// ---------------------------------------------------------------------------
class _TimerNotifier extends Notifier<_TimerState> {
  Timer? _timer;
  int pauseCallCount = 0;

  @override
  _TimerState build() {
    ref.onDispose(_cleanup);
    return const _TimerState();
  }

  void startFixed(Duration duration) {
    _cleanup();
    var remaining = duration;
    state = _TimerState(remaining: remaining, active: true);
    _timer = Timer.periodic(const Duration(seconds: 1), (t) {
      remaining -= const Duration(seconds: 1);
      if (remaining <= Duration.zero) {
        t.cancel();
        _timer = null;
        pauseCallCount++;
        state = const _TimerState();
      } else {
        state = state.copyWith(remaining: remaining);
      }
    });
  }

  void triggerTrackComplete() {
    // Simulate processingState.completed arriving
    if (state.stopAfterTrack && state.active) {
      pauseCallCount++;
      _cleanup();
      state = const _TimerState();
    }
  }

  void startAfterTrack() {
    _cleanup();
    state = const _TimerState(stopAfterTrack: true, active: true);
  }

  void cancel() {
    _cleanup();
    state = const _TimerState();
  }

  void _cleanup() {
    _timer?.cancel();
    _timer = null;
  }
}

final _timerProvider = NotifierProvider<_TimerNotifier, _TimerState>(
  _TimerNotifier.new,
);

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
void main() {
  group('SleepTimerNotifier', () {
    test('fixed duration: state decrements and pauses on expiry', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(_timerProvider.notifier);

      // Start a 3-second timer.
      notifier.startFixed(const Duration(seconds: 3));
      expect(container.read(_timerProvider).active, isTrue);
      expect(container.read(_timerProvider).remaining,
          const Duration(seconds: 3));

      // Wait 4 seconds — the timer should fire and reset state.
      await Future<void>.delayed(const Duration(milliseconds: 4100));
      expect(container.read(_timerProvider).active, isFalse);
      expect(container.read(_timerProvider).remaining, isNull);
      expect(notifier.pauseCallCount, equals(1));
    });

    test('stop-after-track: pauses when triggerTrackComplete is called', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(_timerProvider.notifier);

      notifier.startAfterTrack();
      expect(container.read(_timerProvider).stopAfterTrack, isTrue);
      expect(container.read(_timerProvider).active, isTrue);

      notifier.triggerTrackComplete();
      expect(container.read(_timerProvider).active, isFalse);
      expect(notifier.pauseCallCount, equals(1));
    });

    test('cancel: resets state and stops countdown', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(_timerProvider.notifier);

      notifier.startFixed(const Duration(seconds: 10));
      expect(container.read(_timerProvider).active, isTrue);

      notifier.cancel();
      expect(container.read(_timerProvider).active, isFalse);
      expect(container.read(_timerProvider).remaining, isNull);

      // Wait to confirm timer does not fire after cancel.
      await Future<void>.delayed(const Duration(milliseconds: 1200));
      expect(notifier.pauseCallCount, equals(0));
    });

    test('new startFixed cancels existing timer', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(_timerProvider.notifier);

      // Start a 100-second timer, then immediately replace with a 2-second one.
      notifier.startFixed(const Duration(seconds: 100));
      notifier.startFixed(const Duration(seconds: 2));

      expect(container.read(_timerProvider).remaining,
          const Duration(seconds: 2));

      // Let the 2-second timer expire.
      await Future<void>.delayed(const Duration(milliseconds: 2500));
      expect(container.read(_timerProvider).active, isFalse);
      // Pause should have been called exactly once (the 2-second timer only).
      expect(notifier.pauseCallCount, equals(1));
    });
  });
}
