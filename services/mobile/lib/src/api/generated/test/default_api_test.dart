import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for DefaultApi
void main() {
  final instance = InoriApi().getDefaultApi();

  group(DefaultApi, () {
    // List albums
    //
    // List albums. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogAlbumsGet200Response> apiV1AdminCatalogAlbumsGet({ String artistId, int limit, int offset, String sortBy, String sortOrder, int releaseYearMin, int releaseYearMax }) async
    test('test apiV1AdminCatalogAlbumsGet', () async {
      // TODO
    });

    // Delete album
    //
    //Future apiV1AdminCatalogAlbumsIdDelete(String id) async
    test('test apiV1AdminCatalogAlbumsIdDelete', () async {
      // TODO
    });

    // Get album
    //
    //Future<CatalogAlbum> apiV1AdminCatalogAlbumsIdGet(String id) async
    test('test apiV1AdminCatalogAlbumsIdGet', () async {
      // TODO
    });

    // Update album metadata
    //
    //Future<CatalogAlbum> apiV1AdminCatalogAlbumsIdPatch(String id, CatalogUpdateAlbumRequest catalogUpdateAlbumRequest) async
    test('test apiV1AdminCatalogAlbumsIdPatch', () async {
      // TODO
    });

    // Create album
    //
    //Future<CatalogAlbum> apiV1AdminCatalogAlbumsPost() async
    test('test apiV1AdminCatalogAlbumsPost', () async {
      // TODO
    });

    // List artists
    //
    // List artists. sortBy: name (default), sortName, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogArtistsGet200Response> apiV1AdminCatalogArtistsGet({ int limit, int offset, String sortBy, String sortOrder }) async
    test('test apiV1AdminCatalogArtistsGet', () async {
      // TODO
    });

    // Delete artist
    //
    //Future apiV1AdminCatalogArtistsIdDelete(String id) async
    test('test apiV1AdminCatalogArtistsIdDelete', () async {
      // TODO
    });

    // Get artist
    //
    //Future<CatalogArtist> apiV1AdminCatalogArtistsIdGet(String id) async
    test('test apiV1AdminCatalogArtistsIdGet', () async {
      // TODO
    });

    // Update artist metadata
    //
    //Future<CatalogArtist> apiV1AdminCatalogArtistsIdPatch(String id, CatalogUpdateArtistRequest catalogUpdateArtistRequest) async
    test('test apiV1AdminCatalogArtistsIdPatch', () async {
      // TODO
    });

    // Create artist
    //
    //Future<CatalogArtist> apiV1AdminCatalogArtistsPost() async
    test('test apiV1AdminCatalogArtistsPost', () async {
      // TODO
    });

    // Batch import media objects as catalog tracks
    //
    //Future<CatalogBatchImportResult> apiV1AdminCatalogBatchImportPost(CatalogBatchImportRequest catalogBatchImportRequest) async
    test('test apiV1AdminCatalogBatchImportPost', () async {
      // TODO
    });

    // Import media object as a catalog track
    //
    //Future<CatalogTrack> apiV1AdminCatalogImportPost(CatalogImportRequest catalogImportRequest) async
    test('test apiV1AdminCatalogImportPost', () async {
      // TODO
    });

    // List tracks
    //
    // List tracks. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.
    //
    //Future<ApiV1AdminCatalogTracksGet200Response> apiV1AdminCatalogTracksGet({ String artistId, String albumId, int limit, int offset, String sortBy, String sortOrder, String genre }) async
    test('test apiV1AdminCatalogTracksGet', () async {
      // TODO
    });

    // Delete track
    //
    //Future apiV1AdminCatalogTracksIdDelete(String id) async
    test('test apiV1AdminCatalogTracksIdDelete', () async {
      // TODO
    });

    // Get track
    //
    //Future<CatalogTrack> apiV1AdminCatalogTracksIdGet(String id) async
    test('test apiV1AdminCatalogTracksIdGet', () async {
      // TODO
    });

    // Update track metadata
    //
    //Future<CatalogTrack> apiV1AdminCatalogTracksIdPatch(String id, CatalogUpdateTrackRequest catalogUpdateTrackRequest) async
    test('test apiV1AdminCatalogTracksIdPatch', () async {
      // TODO
    });

    // Relink track to a different media object
    //
    //Future<CatalogTrack> apiV1AdminCatalogTracksIdRelinkPost(String id, CatalogRelinkTrackRequest catalogRelinkTrackRequest) async
    test('test apiV1AdminCatalogTracksIdRelinkPost', () async {
      // TODO
    });

    // Create track
    //
    //Future<CatalogTrack> apiV1AdminCatalogTracksPost() async
    test('test apiV1AdminCatalogTracksPost', () async {
      // TODO
    });

    // Find duplicate media objects by content hash.
    //
    // Returns metadata-only groups of media objects that share the same content hash, optionally limited to one backend. Does not read media bytes.
    //
    //Future<MediaObjectDuplicateReport> apiV1AdminMediaObjectsDuplicatesGet({ String backendId, int minCopies }) async
    test('test apiV1AdminMediaObjectsDuplicatesGet', () async {
      // TODO
    });

    // List media objects by metadata filter with offset pagination.
    //
    // Returns a bounded media-object metadata page filtered by exactly one of backendId, contentHash, verificationStatus, lifecycleState, or assetKind.
    //
    //Future<ApiV1AdminMediaObjectsGet200Response> apiV1AdminMediaObjectsGet({ String backendId, String contentHash, String verificationStatus, String lifecycleState, String assetKind, String sortBy, String sortOrder, int limit, int offset }) async
    test('test apiV1AdminMediaObjectsGet', () async {
      // TODO
    });

    // Get a media object by ID.
    //
    //Future<MediaObject> apiV1AdminMediaObjectsIdGet(String id) async
    test('test apiV1AdminMediaObjectsIdGet', () async {
      // TODO
    });

    // Update a media object lifecycle state.
    //
    // Updates only metadata lifecycle state and updatedAt. It does not delete media bytes or mutate storage backends.
    //
    //Future<MediaObject> apiV1AdminMediaObjectsIdLifecyclePost(String id, MediaObjectLifecycleRequest mediaObjectLifecycleRequest) async
    test('test apiV1AdminMediaObjectsIdLifecyclePost', () async {
      // TODO
    });

    // Get the retained metadata timeline for a media object.
    //
    //Future<MediaObjectTimeline> apiV1AdminMediaObjectsIdTimelineGet(String id) async
    test('test apiV1AdminMediaObjectsIdTimelineGet', () async {
      // TODO
    });

    // Verify a media object metadata reference against stored bytes.
    //
    //Future<MediaObjectVerificationResult> apiV1AdminMediaObjectsIdVerifyPost(String id) async
    test('test apiV1AdminMediaObjectsIdVerifyPost', () async {
      // TODO
    });

    // Bulk update media object lifecycle metadata.
    //
    // Updates only media object lifecycle metadata selected by exactly one metadata filter. Set dryRun to true to preview matched objects and outcomes without persisting changes. Does not delete media bytes or mutate storage backends.
    //
    //Future<MediaObjectLifecycleUpdateReport> apiV1AdminMediaObjectsLifecyclePost(MediaObjectBulkLifecycleRequest mediaObjectBulkLifecycleRequest) async
    test('test apiV1AdminMediaObjectsLifecyclePost', () async {
      // TODO
    });

    // Register a media object metadata reference.
    //
    //Future<MediaObject> apiV1AdminMediaObjectsPost(MediaObjectRequest mediaObjectRequest) async
    test('test apiV1AdminMediaObjectsPost', () async {
      // TODO
    });

    // Get media object metadata statistics.
    //
    // Returns metadata-only media object counts for all backends or one backend without reading media bytes.
    //
    //Future<MediaObjectStats> apiV1AdminMediaObjectsStatsGet({ String backendId }) async
    test('test apiV1AdminMediaObjectsStatsGet', () async {
      // TODO
    });

    // Batch verify media object metadata references by backend ID or content hash.
    //
    //Future<MediaObjectVerificationReport> apiV1AdminMediaObjectsVerifyPost({ String backendId, String contentHash }) async
    test('test apiV1AdminMediaObjectsVerifyPost', () async {
      // TODO
    });

    // List configured storage backends.
    //
    //Future<ApiV1AdminStorageBackendsGet200Response> apiV1AdminStorageBackendsGet() async
    test('test apiV1AdminStorageBackendsGet', () async {
      // TODO
    });

    // Read and record backend capacity where supported.
    //
    //Future<CapacityReport> apiV1AdminStorageBackendsIdCapacityGet(String id) async
    test('test apiV1AdminStorageBackendsIdCapacityGet', () async {
      // TODO
    });

    // Select a backend as default.
    //
    //Future<StorageBackend> apiV1AdminStorageBackendsIdDefaultPost(String id) async
    test('test apiV1AdminStorageBackendsIdDefaultPost', () async {
      // TODO
    });

    // Disable a non-default backend.
    //
    //Future<StorageBackend> apiV1AdminStorageBackendsIdDisablePost(String id) async
    test('test apiV1AdminStorageBackendsIdDisablePost', () async {
      // TODO
    });

    // Read latest recorded backend health.
    //
    //Future<ProbeResult> apiV1AdminStorageBackendsIdHealthGet(String id) async
    test('test apiV1AdminStorageBackendsIdHealthGet', () async {
      // TODO
    });

    // Run an on-demand backend probe.
    //
    //Future<ProbeResult> apiV1AdminStorageBackendsIdProbePost(String id) async
    test('test apiV1AdminStorageBackendsIdProbePost', () async {
      // TODO
    });

    // Validate and register a storage backend.
    //
    //Future<StorageBackend> apiV1AdminStorageBackendsPost(StorageBackendRequest storageBackendRequest) async
    test('test apiV1AdminStorageBackendsPost', () async {
      // TODO
    });

    // Refresh all enabled backend health and supported capacity state.
    //
    //Future<RefreshReport> apiV1AdminStorageBackendsRefreshPost() async
    test('test apiV1AdminStorageBackendsRefreshPost', () async {
      // TODO
    });

    // Validate a backend candidate without persisting it.
    //
    //Future<StorageBackend> apiV1AdminStorageBackendsValidatePost(StorageBackendRequest storageBackendRequest) async
    test('test apiV1AdminStorageBackendsValidatePost', () async {
      // TODO
    });

    // List users.
    //
    //Future<ApiV1AdminUsersGet200Response> apiV1AdminUsersGet({ int limit, int offset, String sortBy, String sortOrder, String username, String role, bool enabled }) async
    test('test apiV1AdminUsersGet', () async {
      // TODO
    });

    // Force-reset a user's password without requiring the current password.
    //
    //Future apiV1AdminUsersIdChangePasswordPost(String id, ForceChangePasswordRequest forceChangePasswordRequest) async
    test('test apiV1AdminUsersIdChangePasswordPost', () async {
      // TODO
    });

    // Delete a user.
    //
    //Future apiV1AdminUsersIdDelete(String id, String id2) async
    test('test apiV1AdminUsersIdDelete', () async {
      // TODO
    });

    // Disable a user.
    //
    //Future<UserView> apiV1AdminUsersIdDisablePost(String id, String id2) async
    test('test apiV1AdminUsersIdDisablePost', () async {
      // TODO
    });

    // Enable a previously disabled user.
    //
    //Future<UserView> apiV1AdminUsersIdEnablePost(String id, String id2) async
    test('test apiV1AdminUsersIdEnablePost', () async {
      // TODO
    });

    // Get a user by ID.
    //
    //Future<UserView> apiV1AdminUsersIdGet(String id) async
    test('test apiV1AdminUsersIdGet', () async {
      // TODO
    });

    // Partially update a user's role or username.
    //
    //Future<UserView> apiV1AdminUsersIdPatch(String id, PatchUserRequest patchUserRequest) async
    test('test apiV1AdminUsersIdPatch', () async {
      // TODO
    });

    // Create a user.
    //
    //Future<UserView> apiV1AdminUsersPost(CreateUserRequest createUserRequest) async
    test('test apiV1AdminUsersPost', () async {
      // TODO
    });

    // Create a session token from username and password.
    //
    //Future<LoginResponse> apiV1AuthLoginPost(LoginRequest loginRequest) async
    test('test apiV1AuthLoginPost', () async {
      // TODO
    });

    // Revoke the current session token.
    //
    //Future apiV1AuthLogoutPost() async
    test('test apiV1AuthLogoutPost', () async {
      // TODO
    });

    // Change the authenticated viewer's password.
    //
    //Future apiV1MeChangePasswordPost(ChangePasswordRequest changePasswordRequest) async
    test('test apiV1MeChangePasswordPost', () async {
      // TODO
    });

    // Get the authenticated viewer's profile.
    //
    //Future<UserView> apiV1MeGet() async
    test('test apiV1MeGet', () async {
      // TODO
    });

    // Revoke all sessions for the authenticated user, including the current session.
    //
    //Future<DeleteAdminUserSessions200Response> apiV1MeSessionsRevokeAllDevicesPost() async
    test('test apiV1MeSessionsRevokeAllDevicesPost', () async {
      // TODO
    });

    // Delete a storage backend
    //
    //Future deleteStorageBackend(String id) async
    test('test deleteStorageBackend', () async {
      // TODO
    });

    // Enable a storage backend
    //
    //Future<StorageBackend> enableStorageBackend(String id) async
    test('test enableStorageBackend', () async {
      // TODO
    });

    // Get a storage backend by ID
    //
    //Future<StorageBackend> getStorageBackend(String id) async
    test('test getStorageBackend', () async {
      // TODO
    });

    // Return process health.
    //
    //Future<HealthzGet200Response> healthzGet() async
    test('test healthzGet', () async {
      // TODO
    });

    // Get public Prometheus-compatible process metrics.
    //
    //Future<String> metricsGet() async
    test('test metricsGet', () async {
      // TODO
    });

    // Update correctable metadata on a media object (assetKind, mimeType)
    //
    //Future<MediaObject> patchMediaObject(String id, PatchMediaObjectRequest patchMediaObjectRequest) async
    test('test patchMediaObject', () async {
      // TODO
    });

    // Update a storage backend's display name or priority
    //
    //Future<StorageBackend> patchStorageBackend(String id, PatchStorageBackendRequest patchStorageBackendRequest) async
    test('test patchStorageBackend', () async {
      // TODO
    });

    // Get public readiness status for this API process.
    //
    //Future<ReadinessReport> readyzGet() async
    test('test readyzGet', () async {
      // TODO
    });

    // Get public build metadata for this API process.
    //
    //Future<ServiceInfo> versionzGet() async
    test('test versionzGet', () async {
      // TODO
    });

  });
}
