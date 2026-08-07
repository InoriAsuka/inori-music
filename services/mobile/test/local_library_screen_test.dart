// local_library_screen_test.dart
//
// Covers the v5.23.0 local-library interaction rework: the play-all/shuffle
// toolbar, the in-library filter, hover-to-play on artwork, and long-press
// multi-select with batch removal.
//
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:inori_music/src/local_library/local_library_db.dart';
import 'package:inori_music/src/local_library/local_library_notifier.dart';
import 'package:inori_music/src/local_library/local_library_screen.dart';
import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;

// ---------------------------------------------------------------------------
// Stubs — no sqflite, no file system, no audio stack.
// ---------------------------------------------------------------------------
class _StubLocalLibraryNotifier extends LocalLibraryNotifier {
  _StubLocalLibraryNotifier(this._tracks, {this.importOutcome});
  final List<LocalLibraryTrack> _tracks;

  /// What [importFiles]/[importFolder] should report back.
  final ImportOutcome? importOutcome;

  List<String>? lastRemovedBatch;
  final removedSingles = <String>[];

  @override
  Future<List<LocalLibraryTrack>> build() async => _tracks;

  @override
  Future<void> remove(String id) async => removedSingles.add(id);

  @override
  Future<void> removeAll(Iterable<String> ids) async =>
      lastRemovedBatch = ids.toList();

  @override
  Future<ImportOutcome> importFiles() async =>
      importOutcome ?? const ImportOutcome.cancelled();

  @override
  Future<ImportOutcome> importFolder() async =>
      importOutcome ?? const ImportOutcome.cancelled();
}

class _StubPlayerNotifier extends PlayerNotifier {
  List<String>? lastQueue;
  int? lastInitialIndex;

  @override
  pstate.PlayerState build() => pstate.PlayerState();

  @override
  Future<void> playQueue(List<String> trackIds, {int initialIndex = 0}) async {
    lastQueue = List.of(trackIds);
    lastInitialIndex = initialIndex;
  }
}

LocalLibraryTrack _track(String id, String title, {String artist = ''}) =>
    LocalLibraryTrack(
      id: id,
      title: title,
      artistName: artist,
      albumTitle: '',
      localPath: '/tmp/$id.flac',
      durationMs: 210000,
      sizeBytes: 1024,
      importedAt: DateTime(2026),
    );

Widget _buildApp(
  _StubLocalLibraryNotifier library,
  _StubPlayerNotifier player,
) => ProviderScope(
  overrides: [
    localLibraryProvider.overrideWith(() => library),
    playerProvider.overrideWith(() => player),
  ],
  child: const MaterialApp(home: LocalLibraryScreen()),
);

/// Taps the empty-state "导入文件" button and settles the SnackBar.
Future<void> _tapEmptyStateImport(WidgetTester tester) async {
  await tester.tap(find.widgetWithText(FilledButton, '导入文件'));
  await tester.pumpAndSettle();
}

void main() {
  final tracks = [
    _track('local:1', 'Idol', artist: 'Yoasobi'),
    _track('local:2', 'Racing Into The Night', artist: 'Yoasobi'),
    _track('local:3', 'Lemon', artist: 'Kenshi Yonezu'),
  ];

  testWidgets('Play all enqueues the whole library', (tester) async {
    final library = _StubLocalLibraryNotifier(tracks);
    final player = _StubPlayerNotifier();
    await tester.pumpWidget(_buildApp(library, player));
    await tester.pumpAndSettle();

    await tester.tap(find.text('播放全部'));
    await tester.pump();

    expect(player.lastQueue, ['local:1', 'local:2', 'local:3']);
  });

  testWidgets('Shuffle enqueues the same tracks in unconstrained order', (
    tester,
  ) async {
    final library = _StubLocalLibraryNotifier(tracks);
    final player = _StubPlayerNotifier();
    await tester.pumpWidget(_buildApp(library, player));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.shuffle));
    await tester.pump();

    expect(player.lastQueue, hasLength(3));
    expect(player.lastQueue, containsAll(['local:1', 'local:2', 'local:3']));
  });

  testWidgets('in-library filter matches title and artist', (tester) async {
    final library = _StubLocalLibraryNotifier(tracks);
    await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField), 'lemon');
    await tester.pumpAndSettle();

    expect(find.text('Lemon'), findsOneWidget);
    expect(find.text('Idol'), findsNothing);

    // Artist match, case-insensitive, keeps both of that artist's tracks.
    await tester.enterText(find.byType(TextField), 'YOASOBI');
    await tester.pumpAndSettle();

    expect(find.text('Idol'), findsOneWidget);
    expect(find.text('Racing Into The Night'), findsOneWidget);
    expect(find.text('Lemon'), findsNothing);
  });

  testWidgets('Play all respects the active filter', (tester) async {
    // The toolbar acts on what's on screen, not on the whole table —
    // otherwise filtering then hitting play would silently ignore the filter.
    final library = _StubLocalLibraryNotifier(tracks);
    final player = _StubPlayerNotifier();
    await tester.pumpWidget(_buildApp(library, player));
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField), 'yoasobi');
    await tester.pumpAndSettle();
    await tester.tap(find.text('播放全部'));
    await tester.pump();

    expect(player.lastQueue, ['local:1', 'local:2']);
  });

  testWidgets('tapping a row plays from that row within the filtered list', (
    tester,
  ) async {
    final library = _StubLocalLibraryNotifier(tracks);
    final player = _StubPlayerNotifier();
    await tester.pumpWidget(_buildApp(library, player));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Lemon'));
    await tester.pump();

    expect(player.lastInitialIndex, 2);
  });

  testWidgets('long-press enters selection mode and batch-removes', (
    tester,
  ) async {
    final library = _StubLocalLibraryNotifier(tracks);
    await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
    await tester.pumpAndSettle();

    await tester.longPress(find.text('Idol'));
    await tester.pumpAndSettle();

    expect(find.text('已选择 1 项'), findsOneWidget);
    expect(find.byType(Checkbox), findsNWidgets(3));

    // A plain tap extends the selection rather than starting playback.
    await tester.tap(find.text('Lemon'));
    await tester.pumpAndSettle();
    expect(find.text('已选择 2 项'), findsOneWidget);

    await tester.tap(find.byIcon(Icons.delete_outline));
    await tester.pumpAndSettle();
    await tester.tap(find.text('移除'));
    await tester.pumpAndSettle();

    expect(library.lastRemovedBatch, hasLength(2));
    expect(library.lastRemovedBatch, containsAll(['local:1', 'local:3']));
    // Selection mode ends once the batch is handed off.
    expect(find.byType(Checkbox), findsNothing);
  });

  testWidgets('closing selection mode leaves the library untouched', (
    tester,
  ) async {
    final library = _StubLocalLibraryNotifier(tracks);
    await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
    await tester.pumpAndSettle();

    await tester.longPress(find.text('Idol'));
    await tester.pumpAndSettle();
    await tester.tap(find.byIcon(Icons.close));
    await tester.pumpAndSettle();

    expect(find.byType(Checkbox), findsNothing);
    expect(library.lastRemovedBatch, isNull);
    expect(library.removedSingles, isEmpty);
  });

  testWidgets('hovering a row reveals a play glyph over its artwork', (
    tester,
  ) async {
    final library = _StubLocalLibraryNotifier(tracks);
    await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
    await tester.pumpAndSettle();

    // Scoped to the rows: the toolbar's "play all" button carries the same
    // icon and is always on screen.
    final rowPlayGlyph = find.descendant(
      of: find.byType(ListTile),
      matching: find.byIcon(Icons.play_arrow_rounded),
    );
    expect(rowPlayGlyph, findsNothing);

    final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
    await mouse.addPointer(location: Offset.zero);
    addTearDown(mouse.removePointer);
    await mouse.moveTo(tester.getCenter(find.text('Idol')));
    await tester.pump();

    expect(rowPlayGlyph, findsOneWidget);

    await mouse.moveTo(const Offset(-100, -100));
    await tester.pump();
    expect(rowPlayGlyph, findsNothing);
  });

  testWidgets('right-click toggles selection (desktop has no long-press)', (
    tester,
  ) async {
    final library = _StubLocalLibraryNotifier(tracks);
    await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
    await tester.pumpAndSettle();

    final mouse = await tester.createGesture(
      kind: PointerDeviceKind.mouse,
      buttons: kSecondaryMouseButton,
    );
    await mouse.down(tester.getCenter(find.text('Idol')));
    await mouse.up();
    await tester.pumpAndSettle();

    expect(find.text('已选择 1 项'), findsOneWidget);
  });

  // -------------------------------------------------------------------------
  // v5.25.1 — an import must never fail silently
  // -------------------------------------------------------------------------

  group('import outcome reporting', () {
    testWidgets('a failed import says why', (tester) async {
      // The Windows case: the picker (or the copy, or the DB write) throws and
      // the library stays empty. Before this, that looked identical to the
      // user simply cancelling — which is how a completely broken Windows
      // import went unnoticed through several releases.
      final library = _StubLocalLibraryNotifier(
        const [],
        importOutcome: ImportOutcome.failed(
          Exception('MissingPluginException'),
        ),
      );
      await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
      await tester.pumpAndSettle();

      await _tapEmptyStateImport(tester);

      expect(find.byType(SnackBar), findsOneWidget);
      expect(
        find.textContaining('导入失败'),
        findsOneWidget,
        reason: 'The failure, not silence',
      );
      expect(find.textContaining('MissingPluginException'), findsOneWidget);
    });

    testWidgets('cancelling the picker stays quiet', (tester) async {
      final library = _StubLocalLibraryNotifier(
        const [],
        importOutcome: const ImportOutcome.cancelled(),
      );
      await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
      await tester.pumpAndSettle();

      await _tapEmptyStateImport(tester);

      expect(
        find.byType(SnackBar),
        findsNothing,
        reason: 'Backing out of the picker is not an error',
      );
    });

    testWidgets('a successful import reports the count', (tester) async {
      final library = _StubLocalLibraryNotifier(
        const [],
        importOutcome: const ImportOutcome(imported: 3, failed: 0),
      );
      await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
      await tester.pumpAndSettle();

      await _tapEmptyStateImport(tester);

      expect(find.text('已导入 3 首'), findsOneWidget);
    });

    testWidgets('a partial import reports both counts and the first error', (
      tester,
    ) async {
      final library = _StubLocalLibraryNotifier(
        const [],
        importOutcome: ImportOutcome(
          imported: 2,
          failed: 1,
          firstError: Exception('corrupt tag'),
        ),
      );
      await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
      await tester.pumpAndSettle();

      await _tapEmptyStateImport(tester);

      expect(find.textContaining('已导入 2 首'), findsOneWidget);
      expect(find.textContaining('1 首失败'), findsOneWidget);
      expect(find.textContaining('corrupt tag'), findsOneWidget);
    });

    testWidgets('picking nothing usable is distinguished from failing', (
      tester,
    ) async {
      final library = _StubLocalLibraryNotifier(
        const [],
        importOutcome: const ImportOutcome(imported: 0, failed: 0),
      );
      await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
      await tester.pumpAndSettle();

      await _tapEmptyStateImport(tester);

      expect(find.text('没有找到可导入的音频文件'), findsOneWidget);
    });
  });

  testWidgets('a filter matching nothing reports it instead of a blank list', (
    tester,
  ) async {
    final library = _StubLocalLibraryNotifier(tracks);
    await tester.pumpWidget(_buildApp(library, _StubPlayerNotifier()));
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField), 'zzzz');
    await tester.pumpAndSettle();

    expect(find.text('没有匹配「zzzz」的曲目'), findsOneWidget);
    expect(
      tester.widget<FilledButton>(find.byType(FilledButton)).onPressed,
      isNull,
      reason: 'Play all must be disabled when nothing matches',
    );
  });
}
