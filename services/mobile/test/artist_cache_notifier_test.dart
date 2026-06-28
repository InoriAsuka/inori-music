// ignore_for_file: implementation_imports
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:inori_api/src/api/catalog_api.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_album.dart';

import 'package:inori_music/src/catalog/catalog_repository.dart';
import 'package:inori_music/src/catalog/catalog_cache_providers.dart';

// ---------------------------------------------------------------------------
// Stub repositories
// ---------------------------------------------------------------------------

CatalogApi _unreachableApi() =>
    CatalogApi(Dio(BaseOptions(baseUrl: 'http://localhost')));

class _StubCatalogRepository extends CatalogRepository {
  _StubCatalogRepository({this.failArtist = false})
      : super(_unreachableApi());

  final bool failArtist;

  @override
  Future<CatalogArtist> getArtist(String id) async {
    if (failArtist) throw Exception('artist not found');
    return CatalogArtist(
      id: id,
      name: 'Yoasobi',
      createdAt: DateTime(2024),
      updatedAt: DateTime(2024),
    );
  }

  @override
  Future<CatalogAlbum> getAlbum(String id) async {
    return CatalogAlbum(
      id: id,
      title: 'THE BOOK',
      artistId: 'artist-001',
      createdAt: DateTime(2024),
      updatedAt: DateTime(2024),
    );
  }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('artistNameProvider', () {
    test('returns artist.name on success', () async {
      final container = ProviderContainer(
        overrides: [
          catalogRepositoryProvider.overrideWithValue(
            _StubCatalogRepository(),
          ),
        ],
      );
      addTearDown(container.dispose);

      final name = await container.read(
        artistNameProvider('artist-001').future,
      );
      expect(name, equals('Yoasobi'));
    });

    test('hasError is true when repository throws', () async {
      final container = ProviderContainer(
        overrides: [
          catalogRepositoryProvider.overrideWithValue(
            _StubCatalogRepository(failArtist: true),
          ),
        ],
      );
      addTearDown(container.dispose);

      // Await the future to let the notifier settle into error state.
      try {
        await container.read(artistNameProvider('artist-fail').future);
      } catch (_) {
        // expected
      }

      final asyncVal = container.read(artistNameProvider('artist-fail'));
      expect(asyncVal.hasError, isTrue,
          reason: 'Provider should be in error state when repository throws');
    });
  });

  group('albumTitleProvider', () {
    test('returns album.title on success', () async {
      final container = ProviderContainer(
        overrides: [
          catalogRepositoryProvider.overrideWithValue(
            _StubCatalogRepository(),
          ),
        ],
      );
      addTearDown(container.dispose);

      final title = await container.read(
        albumTitleProvider('album-001').future,
      );
      expect(title, equals('THE BOOK'));
    });
  });
}
