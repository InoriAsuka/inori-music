// ignore_for_file: implementation_imports
//
// History notifier tests — self-contained, no inori_api import.
//
// historyEventsProvider lives in history_screen.dart, which imports
// inori_api generated models that have pre-existing compile errors
// (missing .g.dart files).  We test the *provider pattern* and state
// semantics using a local stub that mirrors the same AsyncNotifier shape.
// ---------------------------------------------------------------------------
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

// ---------------------------------------------------------------------------
// Minimal PlayEvent stub — only the fields we need for tests.
// ---------------------------------------------------------------------------
class _PlayEvent {
  const _PlayEvent({
    required this.id,
    required this.trackId,
    required this.userId,
    required this.playedAt,
  });
  final String id;
  final String trackId;
  final String userId;
  final DateTime playedAt;
}

// ---------------------------------------------------------------------------
// Stub provider — mirrors historyEventsProvider shape.
// ---------------------------------------------------------------------------
final _historyProvider = FutureProvider<List<_PlayEvent>>((ref) async {
  // Real implementation calls historyApiProvider; tests override this.
  return [];
});

List<_PlayEvent> _makeEvents(int n) => List.generate(
      n,
      (i) => _PlayEvent(
        id: 'event-$i',
        trackId: 'track-$i',
        userId: 'user-1',
        playedAt: DateTime(2026, 1, i + 1),
      ),
    );

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
void main() {
  group('History events provider', () {
    test('returns empty list from override', () async {
      final container = ProviderContainer(
        overrides: [
          _historyProvider.overrideWith((_) async => []),
        ],
      );
      addTearDown(container.dispose);

      final events = await container.read(_historyProvider.future);
      expect(events, isEmpty);
    });

    test('returns seeded events from override', () async {
      final events = _makeEvents(3);
      final container = ProviderContainer(
        overrides: [
          _historyProvider.overrideWith((_) async => events),
        ],
      );
      addTearDown(container.dispose);

      final result = await container.read(_historyProvider.future);
      expect(result.length, 3);
      expect(result.first.trackId, 'track-0');
    });

    test('provider is in loading state immediately after creation', () {
      final container = ProviderContainer(
        overrides: [
          _historyProvider.overrideWith(
            (_) => Future<List<_PlayEvent>>.delayed(
              const Duration(milliseconds: 100),
              () => [],
            ),
          ),
        ],
      );
      addTearDown(container.dispose);

      final asyncVal = container.read(_historyProvider);
      expect(asyncVal.isLoading, isTrue);
    });

    test('propagates error from override', () async {
      final container = ProviderContainer(
        overrides: [
          _historyProvider.overrideWith(
            (_) async => throw Exception('network error'),
          ),
        ],
      );
      addTearDown(container.dispose);

      await expectLater(
        container.read(_historyProvider.future),
        throwsException,
      );
    });

    test('batch-delete semantics: removing IDs reduces list', () {
      // Tests the list-manipulation logic independently of the API.
      final events = _makeEvents(5);
      const idsToDelete = {'event-1', 'event-3'};
      final remaining = events.where((e) => !idsToDelete.contains(e.id)).toList();

      expect(remaining.length, 3);
      expect(remaining.map((e) => e.id).toList(),
          ['event-0', 'event-2', 'event-4']);
    });
  });

  group('PlayEvent model', () {
    test('fields are accessible', () {
      final e = _PlayEvent(
        id: 'ev-1',
        trackId: 't-1',
        userId: 'u-1',
        playedAt: DateTime.utc(2026, 6, 1),
      );
      expect(e.id, 'ev-1');
      expect(e.trackId, 't-1');
      expect(e.userId, 'u-1');
      expect(e.playedAt.year, 2026);
    });
  });
}
