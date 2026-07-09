// ignore_for_file: implementation_imports
//
// search_screen_test.dart
//
// exercises the real SearchScreen against stub providers — verifies that the
// search history list renders, that tapping a history entry populates the
// search field, and that the per-entry close / clear-all buttons drive the
// SearchHistoryNotifier.remove/clear paths.
//
// Strategy: stub _searchNotifierProvider (an AutoDisposeNotifierProvider of
// _SearchState) with one that returns an empty state (no debounce), and
// override searchHistoryProvider with a _StubHistoryNotifier that returns a
// fixed history list without touching SharedPreferences.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/l10n/app_localizations.dart';
import 'package:inori_music/src/catalog/search_history_provider.dart';
import 'package:inori_music/src/catalog/search_screen.dart';

// ---------------------------------------------------------------------------
// Stub history notifier: in-memory state, no SharedPreferences.
// ---------------------------------------------------------------------------
class _StubHistoryNotifier extends SearchHistoryNotifier {
  _StubHistoryNotifier(this._items);
  List<String> _items;

  @override
  List<String> build() => List<String>.unmodifiable(_items);

  @override
  Future<void> add(String query) async {}

  @override
  Future<void> remove(String query) async {
    _items = _items.where((q) => q != query).toList();
    state = List<String>.unmodifiable(_items);
  }

  @override
  Future<void> clear() async {
    _items = [];
    state = const <String>[];
  }
}

Widget _buildApp(_StubHistoryNotifier stub) {
  return ProviderScope(
    overrides: [
      searchHistoryProvider.overrideWith(() => stub),
    ],
    child: MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: const Scaffold(body: SearchScreen()),
    ),
  );
}

void main() {
  group('SearchScreen history overlay', () {
    testWidgets('renders history entries when history is populated',
        (tester) async {
      final stub = _StubHistoryNotifier(['hello', 'world']);
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      // Each history query should be findable as a ListTile title.
      expect(find.text('hello'), findsOneWidget);
      expect(find.text('world'), findsOneWidget);
    });

    testWidgets('hides history when history is empty', (tester) async {
      final stub = _StubHistoryNotifier(<String>[]);
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      // No history ListTiles — querying for a non-existent title is enough.
      expect(find.text('hello'), findsNothing);
    });

    testWidgets('tapping a history entry populates the search field',
        (tester) async {
      final stub = _StubHistoryNotifier(['hello', 'world']);
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      await tester.tap(find.text('hello'));
      await tester.pump();

      // history overlay disappears, field populated => text in TextField.
      final textField = tester.widget<TextField>(find.byType(TextField));
      expect(textField.controller?.text, 'hello');
    });

    testWidgets('close button on an entry removes that entry', (tester) async {
      final stub = _StubHistoryNotifier(['hello', 'world']);
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      expect(find.text('hello'), findsOneWidget);

      // The trailing IconButton (Icons.close) per entry — tap the first one
      // (associated with the 'hello' entry).
      final closeButtons = find.byIcon(Icons.close);
      expect(closeButtons, findsNWidgets(2));
      await tester.tap(closeButtons.first);
      await tester.pump();

      expect(find.text('hello'), findsNothing);
      expect(find.text('world'), findsOneWidget);
    });

    testWidgets('clear-all button removes all entries', (tester) async {
      final stub = _StubHistoryNotifier(['hello', 'world']);
      await tester.pumpWidget(_buildApp(stub));
      await tester.pump();

      expect(find.text('hello'), findsOneWidget);
      expect(find.text('world'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.delete_outline));
      await tester.pump();

      expect(find.text('hello'), findsNothing);
      expect(find.text('world'), findsNothing);
    });
  });
}
