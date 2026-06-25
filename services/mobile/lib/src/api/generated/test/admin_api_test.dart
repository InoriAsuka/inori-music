import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for AdminApi
void main() {
  final instance = InoriApi().getAdminApi();

  group(AdminApi, () {
    // Clear all favorites for a user (admin)
    //
    //Future adminClearUserFavorites(String userId) async
    test('test adminClearUserFavorites', () async {
      // TODO
    });

    // List a user's favorites (admin)
    //
    //Future<AdminListUserFavorites200Response> adminListUserFavorites(String userId, { int limit, int offset }) async
    test('test adminListUserFavorites', () async {
      // TODO
    });

    // Remove a single user-track favorite (admin)
    //
    //Future adminRemoveUserFavoriteTrack(String userId, String trackId) async
    test('test adminRemoveUserFavoriteTrack', () async {
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

    // Batch-delete play events by IDs (admin)
    //
    //Future<BatchDeleteResult> apiV1AdminHistoryBatchDeletePost(BatchDeleteRequest batchDeleteRequest) async
    test('test apiV1AdminHistoryBatchDeletePost', () async {
      // TODO
    });

    // Bulk-delete play events within a time window (admin)
    //
    //Future apiV1AdminHistoryDelete({ DateTime since, DateTime until }) async
    test('test apiV1AdminHistoryDelete', () async {
      // TODO
    });

    // Delete a single play event by ID (admin)
    //
    //Future apiV1AdminHistoryEventIdDelete(String eventId) async
    test('test apiV1AdminHistoryEventIdDelete', () async {
      // TODO
    });

    // Get a single play event by ID (admin)
    //
    //Future<PlayEvent> apiV1AdminHistoryEventIdGet(String eventId) async
    test('test apiV1AdminHistoryEventIdGet', () async {
      // TODO
    });

    // Update a play event's playedAt timestamp (admin)
    //
    //Future<PlayEvent> apiV1AdminHistoryEventIdPatch(String eventId, UpdatePlayEventRequest updatePlayEventRequest) async
    test('test apiV1AdminHistoryEventIdPatch', () async {
      // TODO
    });

    // List all play events (admin)
    //
    // Returns a paginated list of all play events across every user and track. Supports optional filtering by userId, trackId, since, and until.
    //
    //Future<PlayEventList> apiV1AdminHistoryGet({ String userId, String trackId, DateTime since, DateTime until, int limit, int offset, String order }) async
    test('test apiV1AdminHistoryGet', () async {
      // TODO
    });

    // Delete all play events for a specific track across all users (admin)
    //
    //Future apiV1AdminHistoryTracksTrackIdDelete(String trackId) async
    test('test apiV1AdminHistoryTracksTrackIdDelete', () async {
      // TODO
    });

    // Get play history for a specific track across all users (admin)
    //
    //Future<PlayEventList> apiV1AdminHistoryTracksTrackIdGet(String trackId, { String userId, int limit, int offset, String order, DateTime since, DateTime until }) async
    test('test apiV1AdminHistoryTracksTrackIdGet', () async {
      // TODO
    });

    // Delete all play events for a specific user (admin)
    //
    //Future apiV1AdminHistoryUsersUserIdDelete(String userId) async
    test('test apiV1AdminHistoryUsersUserIdDelete', () async {
      // TODO
    });

    // Get play history for a specific user (admin)
    //
    //Future<PlayEventList> apiV1AdminHistoryUsersUserIdGet(String userId, { String trackId, int limit, int offset, String order, DateTime since, DateTime until }) async
    test('test apiV1AdminHistoryUsersUserIdGet', () async {
      // TODO
    });

    // Get global history summary
    //
    //Future<GlobalHistorySummary> getAdminHistorySummary({ DateTime since, DateTime until, int limit }) async
    test('test getAdminHistorySummary', () async {
      // TODO
    });

  });
}
