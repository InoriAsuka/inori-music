import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for PlaylistsApi
void main() {
  final instance = InoriApi().getPlaylistsApi();

  group(PlaylistsApi, () {
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

  });
}
