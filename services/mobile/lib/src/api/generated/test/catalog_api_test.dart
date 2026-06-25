import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for CatalogApi
void main() {
  final instance = InoriApi().getCatalogApi();

  group(CatalogApi, () {
    // List albums by artist (admin)
    //
    // List albums belonging to an artist. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogAlbumsGet200Response> adminListAlbumsByArtist(String id, { int limit, int offset, String sortBy, String sortOrder, int releaseYearMin, int releaseYearMax }) async
    test('test adminListAlbumsByArtist', () async {
      // TODO
    });

    // List tracks by album (admin)
    //
    // List tracks belonging to an album. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogTracksGet200Response> adminListTracksByAlbum(String id, { int limit, int offset, String sortBy, String sortOrder, String genre }) async
    test('test adminListTracksByAlbum', () async {
      // TODO
    });

    // List tracks by artist (admin)
    //
    // List tracks belonging to an artist. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogTracksGet200Response> adminListTracksByArtist(String id, { int limit, int offset, String sortBy, String sortOrder, String genre }) async
    test('test adminListTracksByArtist', () async {
      // TODO
    });

    // List all playlists
    //
    // List playlists. sortBy: name (default), createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogPlaylistsGet200Response> apiV1AdminCatalogPlaylistsGet({ int limit, int offset, String sortBy, String sortOrder }) async
    test('test apiV1AdminCatalogPlaylistsGet', () async {
      // TODO
    });

    // Delete a playlist
    //
    //Future apiV1AdminCatalogPlaylistsIdDelete(String id) async
    test('test apiV1AdminCatalogPlaylistsIdDelete', () async {
      // TODO
    });

    // Get a playlist by ID
    //
    //Future<Playlist> apiV1AdminCatalogPlaylistsIdGet(String id) async
    test('test apiV1AdminCatalogPlaylistsIdGet', () async {
      // TODO
    });

    // Update playlist metadata
    //
    //Future<Playlist> apiV1AdminCatalogPlaylistsIdPatch(String id, UpdatePlaylistRequest updatePlaylistRequest) async
    test('test apiV1AdminCatalogPlaylistsIdPatch', () async {
      // TODO
    });

    // List the tracks of a playlist in playlist order
    //
    //Future<ApiV1AdminCatalogPlaylistsIdTracksGet200Response> apiV1AdminCatalogPlaylistsIdTracksGet(String id, { int limit, int offset }) async
    test('test apiV1AdminCatalogPlaylistsIdTracksGet', () async {
      // TODO
    });

    // Append a track to a playlist
    //
    //Future<Playlist> apiV1AdminCatalogPlaylistsIdTracksPost(String id, AddPlaylistTrackRequest addPlaylistTrackRequest) async
    test('test apiV1AdminCatalogPlaylistsIdTracksPost', () async {
      // TODO
    });

    // Replace the ordered track list of a playlist
    //
    //Future<Playlist> apiV1AdminCatalogPlaylistsIdTracksPut(String id, SetPlaylistTracksRequest setPlaylistTracksRequest) async
    test('test apiV1AdminCatalogPlaylistsIdTracksPut', () async {
      // TODO
    });

    // Remove first occurrence of a track from a playlist
    //
    //Future<Playlist> apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete(String trackId, String id) async
    test('test apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete', () async {
      // TODO
    });

    // Create a playlist
    //
    //Future<Playlist> apiV1AdminCatalogPlaylistsPost(CreatePlaylistRequest createPlaylistRequest) async
    test('test apiV1AdminCatalogPlaylistsPost', () async {
      // TODO
    });

    // List playlists (viewer)
    //
    // List playlists. sortBy: name (default), createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogPlaylistsGet200Response> apiV1CatalogPlaylistsGet({ int limit, int offset, String sortBy, String sortOrder }) async
    test('test apiV1CatalogPlaylistsGet', () async {
      // TODO
    });

    // Get a playlist by ID (viewer)
    //
    //Future<Playlist> apiV1CatalogPlaylistsIdGet(String id) async
    test('test apiV1CatalogPlaylistsIdGet', () async {
      // TODO
    });

    // List the tracks of a playlist in playlist order (viewer)
    //
    //Future<ApiV1AdminCatalogPlaylistsIdTracksGet200Response> apiV1CatalogPlaylistsIdTracksGet(String id, { int limit, int offset }) async
    test('test apiV1CatalogPlaylistsIdTracksGet', () async {
      // TODO
    });

    // Stream track audio
    //
    // Proxies audio bytes with HTTP 206 Range support for filesystem-based storage backends. Authenticates via Bearer token in Authorization header or ?token= query parameter.
    //
    //Future apiV1CatalogTracksIdStreamGet(String id, { String token }) async
    test('test apiV1CatalogTracksIdStreamGet', () async {
      // TODO
    });

    // Get album
    //
    //Future<CatalogAlbum> getCatalogAlbum(String id) async
    test('test getCatalogAlbum', () async {
      // TODO
    });

    // Get artist
    //
    //Future<CatalogArtist> getCatalogArtist(String id) async
    test('test getCatalogArtist', () async {
      // TODO
    });

    // Get track
    //
    //Future<CatalogTrack> getCatalogTrack(String id) async
    test('test getCatalogTrack', () async {
      // TODO
    });

    // Get track playback descriptor
    //
    // Returns a metadata-only playback descriptor for the specified track. The linked media object must be active and have an audio asset kind (original_audio or transcoded_audio); otherwise 422 is returned.
    //
    //Future<TrackPlaybackDescriptor> getTrackPlaybackDescriptor(String id) async
    test('test getTrackPlaybackDescriptor', () async {
      // TODO
    });

    // Get per-album stats breakdown
    //
    //Future<CatalogAlbumStatsBreakdown> getViewerCatalogAlbumStats() async
    test('test getViewerCatalogAlbumStats', () async {
      // TODO
    });

    // Get per-artist stats breakdown
    //
    //Future<CatalogArtistStatsBreakdown> getViewerCatalogArtistStats() async
    test('test getViewerCatalogArtistStats', () async {
      // TODO
    });

    // Get per-playlist stats breakdown
    //
    //Future<CatalogPlaylistStatsBreakdown> getViewerCatalogPlaylistStats() async
    test('test getViewerCatalogPlaylistStats', () async {
      // TODO
    });

    // Get catalog stats
    //
    //Future<CatalogStats> getViewerCatalogStats() async
    test('test getViewerCatalogStats', () async {
      // TODO
    });

    // List albums by artist
    //
    // List albums belonging to an artist. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogAlbumsGet200Response> listAlbumsByArtist(String id, { int limit, int offset, String sortBy, String sortOrder, int releaseYearMin, int releaseYearMax }) async
    test('test listAlbumsByArtist', () async {
      // TODO
    });

    // List albums
    //
    // List albums. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogAlbumsGet200Response> listCatalogAlbums({ String artistId, int limit, int offset, String sortBy, String sortOrder, int releaseYearMin, int releaseYearMax }) async
    test('test listCatalogAlbums', () async {
      // TODO
    });

    // List artists
    //
    // List artists. sortBy: name (default), sortName, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogArtistsGet200Response> listCatalogArtists({ int limit, int offset, String sortBy, String sortOrder }) async
    test('test listCatalogArtists', () async {
      // TODO
    });

    // Search catalog
    //
    //Future<CatalogSearchResult> listCatalogSearch(String q, { String types }) async
    test('test listCatalogSearch', () async {
      // TODO
    });

    // List tracks
    //
    // List tracks. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogTracksGet200Response> listCatalogTracks({ String albumId, String artistId, int limit, int offset, String sortBy, String sortOrder, String genre }) async
    test('test listCatalogTracks', () async {
      // TODO
    });

    // List recently added catalog items
    //
    // Returns a newest-first unified timeline of recently created artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).
    //
    //Future<RecentCatalogResult> listRecentlyAddedCatalogItems({ RecentItemKind kind, int limit }) async
    test('test listRecentlyAddedCatalogItems', () async {
      // TODO
    });

    // List recently updated catalog items
    //
    // Returns a newest-first unified timeline of recently updated artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).
    //
    //Future<UpdatedCatalogResult> listRecentlyUpdatedCatalogItems({ RecentItemKind kind, int limit }) async
    test('test listRecentlyUpdatedCatalogItems', () async {
      // TODO
    });

    // List tracks by album
    //
    // List tracks belonging to an album. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogTracksGet200Response> listTracksByAlbum(String id, { int limit, int offset, String sortBy, String sortOrder, String genre }) async
    test('test listTracksByAlbum', () async {
      // TODO
    });

    // List tracks by artist
    //
    // List tracks belonging to an artist. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogTracksGet200Response> listTracksByArtist(String id, { int limit, int offset, String sortBy, String sortOrder, String genre }) async
    test('test listTracksByArtist', () async {
      // TODO
    });

    // Search catalog
    //
    // Full-text search across artists, albums, and tracks. Returns ordered results with artist hits first, then albums, then tracks.
    //
    //Future<CatalogSearchResult> searchCatalog(String q, { String types }) async
    test('test searchCatalog', () async {
      // TODO
    });

  });
}
