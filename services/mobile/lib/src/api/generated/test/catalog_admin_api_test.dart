import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for CatalogAdminApi
void main() {
  final instance = InoriApi().getCatalogAdminApi();

  group(CatalogAdminApi, () {
    // Per-album catalog stats breakdown
    //
    // Returns per-album track counts derived from catalog metadata.
    //
    //Future<CatalogAlbumStatsBreakdown> getAlbumStatsBreakdown() async
    test('test getAlbumStatsBreakdown', () async {
      // TODO
    });

    // Per-artist catalog stats breakdown
    //
    // Returns per-artist album and track counts derived from catalog metadata.
    //
    //Future<CatalogArtistStatsBreakdown> getArtistStatsBreakdown() async
    test('test getArtistStatsBreakdown', () async {
      // TODO
    });

    // Catalog entity count statistics
    //
    // Returns metadata-only aggregate counts for artists, albums, tracks, and playlists.
    //
    //Future<CatalogStats> getCatalogStats() async
    test('test getCatalogStats', () async {
      // TODO
    });

    // Per-playlist catalog stats breakdown
    //
    // Returns per-playlist track counts derived from catalog metadata.
    //
    //Future<CatalogPlaylistStatsBreakdown> getPlaylistStatsBreakdown() async
    test('test getPlaylistStatsBreakdown', () async {
      // TODO
    });

    // List recently added catalog items
    //
    // Returns a newest-first unified timeline of recently created artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).
    //
    //Future<RecentCatalogResult> getRecentlyAddedCatalogItems({ RecentItemKind kind, int limit }) async
    test('test getRecentlyAddedCatalogItems', () async {
      // TODO
    });

    // List recently updated catalog items
    //
    // Returns a newest-first unified timeline of recently updated artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).
    //
    //Future<UpdatedCatalogResult> getRecentlyUpdatedCatalogItems({ RecentItemKind kind, int limit }) async
    test('test getRecentlyUpdatedCatalogItems', () async {
      // TODO
    });

  });
}
