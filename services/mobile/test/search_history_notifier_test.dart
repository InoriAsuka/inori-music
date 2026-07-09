// ignore_for_file: implementation_imports
//
// SearchHistoryNotifier tests — exercises the real production notifier against
// SharedPreferences.setMockInitialValues({}). The restore path is verified by
// seeding prefs in a second container after the first one persists.

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:inori_music/src/catalog/search_history_provider.dart';

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  group('SearchHistoryNotifier', () {
    test('initial state is empty list', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      expect(container.read(searchHistoryProvider), isEmpty);
    });

    test('add() appends new queries, most-recent first', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(searchHistoryProvider.notifier);

      await notifier.add('query1');
      await notifier.add('query2');

      final state = container.read(searchHistoryProvider);
      expect(state, ['query2', 'query1']);
    });

    test('add() trims whitespace and ignores empty queries', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(searchHistoryProvider.notifier);

      await notifier.add('  spaced query  ');
      await notifier.add('   ');
      await notifier.add('');

      final state = container.read(searchHistoryProvider);
      expect(state, ['spaced query']);
    });

    test('add() dedupes by moving the existing query to front', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(searchHistoryProvider.notifier);

      await notifier.add('a');
      await notifier.add('b');
      await notifier.add('a'); // move to front, no duplicate

      expect(container.read(searchHistoryProvider), ['a', 'b']);
    });

    test('add() caps the list at 20 entries', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(searchHistoryProvider.notifier);

      for (var i = 0; i < 25; i++) {
        await notifier.add('q$i');
      }

      final state = container.read(searchHistoryProvider);
      expect(state.length, 20);
      expect(state.first, 'q24');
      expect(state.last, 'q5');
    });

    test('remove() filters out the given query', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(searchHistoryProvider.notifier);

      await notifier.add('a');
      await notifier.add('b');
      await notifier.add('c'); // most recent first → ['c', 'b', 'a']
      await notifier.remove('b');

      expect(container.read(searchHistoryProvider), ['c', 'a']);
    });

    test('clear() empties state and removes the prefs key', () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final notifier = container.read(searchHistoryProvider.notifier);

      await notifier.add('a');
      await notifier.clear();

      expect(container.read(searchHistoryProvider), isEmpty);
      final prefs = await SharedPreferences.getInstance();
      expect(prefs.containsKey('search.history'), isFalse);
    });

    test('state persists and restores from SharedPreferences', () async {
      final container1 = ProviderContainer();
      final notifier1 = container1.read(searchHistoryProvider.notifier);
      await notifier1.add('saved-2');
      await notifier1.add('saved-1'); // saved-1 is most recent
      container1.dispose();

      // New container: build() runs _restore() asynchronously, so we give
      // the event loop a turn before reading.
      final container2 = ProviderContainer();
      addTearDown(container2.dispose);
      container2.read(searchHistoryProvider); // trigger build + _restore
      await Future<void>.delayed(Duration.zero);
      await Future<void>.delayed(Duration.zero);

      expect(container2.read(searchHistoryProvider), ['saved-1', 'saved-2']);
    });
  });
}
