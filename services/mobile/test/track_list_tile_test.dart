// ignore_for_file: implementation_imports
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:inori_api/src/api/catalog_api.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_track.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/shared/widgets/track_list_tile.dart';

// ---------------------------------------------------------------------------
// Helper — build a minimal CatalogTrack without optional fields.
// ---------------------------------------------------------------------------
CatalogTrack _makeTrack({String artistId = 'artist-uuid-001'}) => CatalogTrack(
      id: 'track-001',
      title: 'Idol',
      artistId: artistId,
      albumId: null,
      durationMs: 210000,
      mediaObjectId: 'mo-001',
      trackNumber: null,
      genre: null,
      createdAt: DateTime(2024),
      updatedAt: DateTime(2024),
    );

// CatalogApi is injected by CatalogRepository but never called through in the
// stub — pass a no-op instance pointing at localhost.
CatalogApi _unreachableApi() =>
    CatalogApi(Dio(BaseOptions(baseUrl: 'http://localhost')));

// ---------------------------------------------------------------------------
// Stub repository — resolves artist name synchronously without HTTP.
// ---------------------------------------------------------------------------
class _StubCatalogRepository extends CatalogRepository {
  _StubCatalogRepository(this._artistName, {this.failArtist = false})
      : super(_unreachableApi());

  final String _artistName;
  final bool failArtist;

  @override
  Future<CatalogArtist> getArtist(String id) async {
    if (failArtist) throw Exception('network error');
    return CatalogArtist(
      id: id,
      name: _artistName,
      createdAt: DateTime(2024),
      updatedAt: DateTime(2024),
    );
  }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  const artistId = 'artist-uuid-001';
  const resolvedName = 'Yoasobi';

  testWidgets('TrackListTile shows resolved artist name', (tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          catalogRepositoryProvider.overrideWithValue(
            _StubCatalogRepository(resolvedName),
          ),
        ],
        child: MaterialApp(
          home: Scaffold(body: TrackListTile(track: _makeTrack())),
        ),
      ),
    );

    // Pump once to trigger the async provider, then again to settle the build.
    await tester.pump();
    await tester.pump();

    expect(find.text(resolvedName), findsOneWidget,
        reason: 'Resolved name should be visible in the subtitle');
    expect(find.text(artistId), findsNothing,
        reason: 'UUID should not be shown when resolved name is available');
  });

  testWidgets('TrackListTile falls back to artistId on provider error',
      (tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          catalogRepositoryProvider.overrideWithValue(
            _StubCatalogRepository('', failArtist: true),
          ),
        ],
        child: MaterialApp(
          home: Scaffold(body: TrackListTile(track: _makeTrack())),
        ),
      ),
    );

    await tester.pump();
    await tester.pump();

    expect(find.text(artistId), findsOneWidget,
        reason: 'artistId is shown as fallback on error');
  });

  testWidgets('TrackListTile shows no subtitle when artistId is empty',
      (tester) async {
    // When track.artistId is empty, subtitle should be null (no Text rendered).
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          catalogRepositoryProvider.overrideWithValue(
            _StubCatalogRepository(resolvedName),
          ),
        ],
        child: MaterialApp(
          home: Scaffold(
            body: TrackListTile(track: _makeTrack(artistId: '')),
          ),
        ),
      ),
    );

    await tester.pump();
    await tester.pump();

    // No subtitle text — neither a resolved name nor the empty artist id.
    expect(find.text(resolvedName), findsNothing,
        reason: 'No subtitle should be rendered when artistId is empty');
    // The title must still be present.
    expect(find.text('Idol'), findsOneWidget,
        reason: 'Track title is always visible');
  });
}
