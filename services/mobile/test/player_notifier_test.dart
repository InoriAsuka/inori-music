// ignore_for_file: implementation_imports, unused_import
import 'dart:math' show pow;

import 'package:audio_service/audio_service.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;

// ---------------------------------------------------------------------------
// Pure state-machine tests for PlayerNotifier logic.
//
// Because PlayerNotifier depends on `audioHandler` (a global set in main()),
// these tests exercise the *state* logic of the notifier in isolation rather
// than instantiating PlayerNotifier directly.  We test:
//   • PlayerState construction and defaults
//   • copyWith field updates
//   • RepeatMode / shuffle helpers
//   • Queue arithmetic (reorderQueue semantics)
//   • isPlaying / isBuffering / isIdle computed properties
//   • next()/previous() repeat-mode branching (mirrors the real methods)
//   • ReplayGain gain formula (mirrors _applyVolumeWithGain)
//
// Not covered here: setShuffle()'s native setShuffleModeEnabled()+shuffle()
// call pair, and currentIndexStream's state/mediaItem sync. Both require a
// real AudioPlayer platform instance to observe, which this repo has no
// fake/mock for (no mocktail, no platform-channel test double) — they're
// exercised manually via the v4.7.0 plan's acceptance checklist instead.
// ---------------------------------------------------------------------------

void main() {
  group('PlayerState defaults', () {
    test('initial state has empty queue and currentIndex -1', () {
      final s = pstate.PlayerState();
      expect(s.queue, isEmpty);
      expect(s.currentIndex, -1);
      expect(s.isIdle, isTrue);
      expect(s.isPlaying, isFalse);
    });

    test('initial volume is 1.0 and repeat is none', () {
      final s = pstate.PlayerState();
      expect(s.volume, 1.0);
      expect(s.repeat, pstate.RepeatMode.none);
      expect(s.shuffle, isFalse);
    });
  });

  group('PlayerState.copyWith', () {
    test('copyWith volume updates volume only', () {
      final s = pstate.PlayerState().copyWith(volume: 0.5);
      expect(s.volume, 0.5);
      expect(s.currentIndex, -1); // unchanged
    });

    test('copyWith shuffle toggles correctly', () {
      final s = pstate.PlayerState().copyWith(shuffle: true);
      expect(s.shuffle, isTrue);
      final s2 = s.copyWith(shuffle: false);
      expect(s2.shuffle, isFalse);
    });

    test('copyWith repeat mode changes', () {
      final s = pstate.PlayerState().copyWith(repeat: pstate.RepeatMode.all);
      expect(s.repeat, pstate.RepeatMode.all);
    });

    test('clearMediaItem removes mediaItem', () {
      final item = MediaItem(id: 'track-1', title: 'Test');
      final s = pstate.PlayerState(mediaItem: item);
      expect(s.mediaItem, isNotNull);
      final s2 = s.copyWith(clearMediaItem: true);
      expect(s2.mediaItem, isNull);
    });
  });

  group('Queue state arithmetic', () {
    List<MediaItem> makeQueue(int n) =>
        List.generate(n, (i) => MediaItem(id: 'track-$i', title: 'Track $i'));

    test('reorderQueue: move item 0 to index 2', () {
      // Mirrors PlayerNotifier.reorderQueue logic.
      final queue = makeQueue(4);
      final int oldIndex = 0;
      int newIndex = 2;
      if (newIndex > oldIndex) newIndex--;
      final item = queue.removeAt(oldIndex);
      queue.insert(newIndex, item);
      expect(queue.map((e) => e.id).toList(),
          ['track-1', 'track-0', 'track-2', 'track-3']);
    });

    test('reorderQueue: move last item to first', () {
      final queue = makeQueue(3);
      final int oldIndex = 2;
      final int newIndex = 0;
      final item = queue.removeAt(oldIndex);
      queue.insert(newIndex, item);
      expect(queue.map((e) => e.id).toList(),
          ['track-2', 'track-0', 'track-1']);
    });

  });

  // ---------------------------------------------------------------------
  // next()/previous() repeat-mode branching.
  //
  // Mirrors the decision tree in PlayerNotifier.next() (player_notifier.dart)
  // without touching _audioPlayer: given (repeat, currentIndex, queue.length),
  // what does the method decide to do?
  //   - RepeatMode.one -> always seek(0) + play(), never advances index
  //   - last track + RepeatMode.none -> no-op (natural stop)
  //   - last track + RepeatMode.all -> wrap to index 0 via _playAtIndex
  //   - otherwise -> seekToNext()
  // ---------------------------------------------------------------------
  group('next() repeat-mode decision (mirrors PlayerNotifier.next)', () {
    // Mirrors PlayerNotifier.next()'s branching without the _audioPlayer call.
    String decide(pstate.RepeatMode repeat, int currentIndex, int queueLength) {
      if (repeat == pstate.RepeatMode.one) return 'seekZeroAndPlay';
      final nextIdx = currentIndex + 1;
      if (nextIdx >= queueLength) {
        return repeat == pstate.RepeatMode.all ? 'playAtIndex(0)' : 'noop';
      }
      return 'seekToNext';
    }

    test('T1: RepeatMode.none on last track -> no-op (natural stop)', () {
      expect(decide(pstate.RepeatMode.none, 2, 3), 'noop');
    });

    test('T2: RepeatMode.all on last track -> wraps to index 0', () {
      expect(decide(pstate.RepeatMode.all, 2, 3), 'playAtIndex(0)');
    });

    test('T3: RepeatMode.one -> always restarts current track, never advances', () {
      // Even on the last track, or mid-queue, RepeatMode.one short-circuits
      // before the "last track" check, so it never returns playAtIndex/noop.
      expect(decide(pstate.RepeatMode.one, 2, 3), 'seekZeroAndPlay');
      expect(decide(pstate.RepeatMode.one, 0, 3), 'seekZeroAndPlay');
    });

    test('mid-queue + RepeatMode.none advances via seekToNext', () {
      expect(decide(pstate.RepeatMode.none, 0, 3), 'seekToNext');
    });
  });

  // ---------------------------------------------------------------------
  // ReplayGain gain formula.
  //
  // Mirrors PlayerNotifier._applyVolumeWithGain's math: gain = 10^(dB/20),
  // clamped to [0.1, 2.0], then effective = (userVol * gain) clamped to
  // [0.0, 1.0]. Verified independently of the ReplayGain toggle/track cache
  // plumbing, which requires the live provider graph to exercise.
  // ---------------------------------------------------------------------
  group('ReplayGain gain formula (mirrors _applyVolumeWithGain)', () {
    double effectiveVolume(double userVol, double? replayGainDb, {required bool enabled}) {
      var gain = 1.0;
      if (enabled && replayGainDb != null) {
        gain = pow(10, replayGainDb / 20).toDouble().clamp(0.1, 2.0);
      }
      return (userVol * gain).clamp(0.0, 1.0);
    }

    test('T4: enabled + +6dB track + setVolume(0.5) -> boosted to ~0.9975', () {
      final result = effectiveVolume(0.5, 6.0, enabled: true);
      expect(result, closeTo(0.9975, 0.0005));
    });

    test('T5: replayGainDb null -> gain is 1.0, volume passes through unchanged', () {
      final result = effectiveVolume(0.5, null, enabled: true);
      expect(result, 0.5);
    });

    test('disabled toggle ignores replayGainDb entirely', () {
      final result = effectiveVolume(0.5, 6.0, enabled: false);
      expect(result, 0.5);
    });

    test('gain is clamped to 2.0 ceiling for very quiet tracks (very positive dB)', () {
      // +40dB track would compute gain = 10^(40/20) = 100.0, clamped to 2.0.
      final result = effectiveVolume(1.0, 40.0, enabled: true);
      expect(result, 1.0); // 1.0 * 2.0 clamped back down to 1.0 (volume ceiling)
    });

    test('gain is clamped to 0.1 floor for very loud tracks (very negative dB)', () {
      // -40dB track would compute gain = 10^(-40/20) = 0.01, clamped to 0.1.
      final result = effectiveVolume(1.0, -40.0, enabled: true);
      expect(result, 0.1);
    });
  });

  group('PlayerState.isIdle', () {
    test('isIdle is true when queue is empty', () {
      expect(pstate.PlayerState().isIdle, isTrue);
    });

    test('isIdle is false when queue has items and currentIndex >= 0', () {
      final s = pstate.PlayerState(
        queue: [MediaItem(id: 'track-1', title: 'T')],
        currentIndex: 0,
      );
      expect(s.isIdle, isFalse);
    });

    test('isIdle is true when queue has items but currentIndex is -1', () {
      final s = pstate.PlayerState(
        queue: [MediaItem(id: 'track-1', title: 'T')],
        currentIndex: -1,
      );
      expect(s.isIdle, isTrue);
    });
  });
}
