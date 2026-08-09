// ignore_for_file: implementation_imports
//
// play_actions_row_test.dart
//
// Widget tests for PlayActionsRow — the collection-level "Play all"/"Shuffle"
// header introduced in v5.22.0 so album/artist/playlist detail pages can start
// playback without tapping individual rows.
//
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/player/player_notifier.dart';
import 'package:inori_music/src/player/player_state.dart' as pstate;
import 'package:inori_music/src/shared/theme/skin_definition.dart';
import 'package:inori_music/src/shared/widgets/play_actions_row.dart';

// ---------------------------------------------------------------------------
// Stub PlayerNotifier — records the queue it was handed instead of touching
// audio_service / just_audio.
// ---------------------------------------------------------------------------
class _StubPlayerNotifier extends PlayerNotifier {
  List<String>? lastQueue;

  @override
  pstate.PlayerState build() => pstate.PlayerState();

  @override
  Future<void> playQueue(List<String> trackIds, {int initialIndex = 0}) async {
    lastQueue = List.of(trackIds);
  }
}

CatalogTrack _track(String id) => CatalogTrack(
  id: id,
  title: 'Track $id',
  artistId: 'artist-001',
  albumId: 'album-001',
  durationMs: 210000,
  mediaObjectId: 'mo-$id',
  createdAt: DateTime(2024),
  updatedAt: DateTime(2024),
);

Widget _buildApp(
  _StubPlayerNotifier stub,
  AsyncValue<List<CatalogTrack>> tracksState,
) {
  return ProviderScope(
    overrides: [playerProvider.overrideWith(() => stub)],
    child: MaterialApp(
      // The real skin theme, not Flutter's own MaterialApp default — this
      // widget's FilledButton is a plain Row sibling with no Expanded
      // around it (see the Row in play_actions_row.dart), which is exactly
      // the shape v5.30.7 found actually crashes under the app's
      // then-global `FilledButtonTheme.minimumSize: Size(double.infinity,
      // 48)` (Flutter hands a non-flex Row child unbounded main-axis
      // constraints, so "infinite minWidth" there trips a hard layout
      // assertion rather than merely being clamped). Every test below ran
      // green through v5.30.6 despite that, purely because Flutter's own
      // default theme was never wide enough to expose it — this file was
      // accidentally never actually exercising the button under production
      // styling. Using the real theme here closes that gap: any future
      // regression of the fix (skin_definition_test.dart's own dedicated
      // guard) would now also fail every test in this file, not silently
      // pass them.
      theme: buildThemeFromSkin(SkinDefinition.sakuraDusk),
      home: Scaffold(body: PlayActionsRow(tracksState: tracksState)),
    ),
  );
}

void main() {
  testWidgets('both actions are disabled while tracks are loading', (
    tester,
  ) async {
    final stub = _StubPlayerNotifier();
    await tester.pumpWidget(
      _buildApp(stub, const AsyncValue<List<CatalogTrack>>.loading()),
    );
    await tester.pump();

    expect(
      tester.widget<FilledButton>(find.byType(FilledButton)).onPressed,
      isNull,
      reason: 'Play must not be tappable before the track list resolves',
    );
    expect(
      tester.widget<IconButton>(find.byType(IconButton)).onPressed,
      isNull,
      reason: 'Shuffle must not be tappable before the track list resolves',
    );
  });

  testWidgets('both actions are disabled for an empty collection', (
    tester,
  ) async {
    final stub = _StubPlayerNotifier();
    await tester.pumpWidget(
      _buildApp(stub, const AsyncValue<List<CatalogTrack>>.data([])),
    );
    await tester.pump();

    expect(
      tester.widget<FilledButton>(find.byType(FilledButton)).onPressed,
      isNull,
    );
    expect(
      tester.widget<IconButton>(find.byType(IconButton)).onPressed,
      isNull,
    );
  });

  testWidgets('Play enqueues every track in list order', (tester) async {
    final stub = _StubPlayerNotifier();
    await tester.pumpWidget(
      _buildApp(stub, AsyncValue.data([_track('a'), _track('b'), _track('c')])),
    );
    await tester.pump();

    await tester.tap(find.byType(FilledButton));
    await tester.pump();

    expect(stub.lastQueue, ['a', 'b', 'c']);
  });

  testWidgets('Shuffle enqueues the same tracks, order unconstrained', (
    tester,
  ) async {
    final stub = _StubPlayerNotifier();
    await tester.pumpWidget(
      _buildApp(stub, AsyncValue.data([_track('a'), _track('b'), _track('c')])),
    );
    await tester.pump();

    await tester.tap(find.byType(IconButton));
    await tester.pump();

    // Order is randomised, so assert on membership rather than sequence —
    // a shuffle that happens to land on the identity permutation is valid.
    expect(stub.lastQueue, hasLength(3));
    expect(stub.lastQueue, containsAll(['a', 'b', 'c']));
  });
}
