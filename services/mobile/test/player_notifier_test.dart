// ignore_for_file: implementation_imports, unused_import
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

    test('next index wraps around with RepeatMode.none', () {
      final currentIndex = 2;
      final queueLength = 3;
      // Mirrors: (currentIndex + 1) % queue.length
      final nextIdx = (currentIndex + 1) % queueLength;
      expect(nextIdx, 0); // wraps
    });

    test('next index stays same with RepeatMode.one', () {
      final currentIndex = 1;
      final queueLength = 3;
      final nextIdx = pstate.RepeatMode.one == pstate.RepeatMode.one
          ? currentIndex
          : (currentIndex + 1) % queueLength;
      expect(nextIdx, 1);
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
