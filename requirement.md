# inori-music Requirements

## Current Version

`0.95.0`

## Product Goal

Build a cross-platform music playback system for Web, Android, iOS, and desktop clients while supporting both browser/server and client/server architectures. The server owns media storage configuration, metadata registration, health checks, integrity verification, and administrative APIs. Large media bytes are stored in external storage backends rather than the relational database.

## Technical Requirements

- Flutter-first client direction.
- Go modular monolith first for the server.
- PostgreSQL-first server metadata database.
- SQLite for client-side local persistence.
- PostgreSQL full-text search first for 0.x, with external search engines left as future extensions.
- Media storage must support local filesystems, NFS, SMB, S3-compatible object storage, and distributed storage adapters.
- Repository automation must validate builds and tests, publish tagged release binaries, and publish Docker images for deployable API artifacts.
- Runtime API artifacts must expose non-sensitive build metadata for deployment diagnostics.
- Runtime API artifacts must expose public readiness diagnostics for storage, media registry, and admin-auth configuration.
- Runtime API artifacts must expose non-sensitive Prometheus-compatible metrics for deployment monitoring.
- Runtime HTTP metrics must avoid high-cardinality labels by using route patterns instead of raw URLs.

## Storage Requirements

- Do not store large audio, image, or derived media files in the relational database.
- Store object IDs, backend IDs, object keys, hashes, lifecycle state, asset kind, verification state, and references as metadata.
- Probes and verification must use server-owned temporary objects or read-only checks to avoid damaging user media.
- Admin APIs must expose a read-only per-object metadata timeline derived from retained registration, latest verification, and latest lifecycle transition state.

## Documentation Requirements

- Markdown documentation is maintained in English.
- Phase work must be recorded under `.plan/` with requirements, task checklists, non-goals, and follow-up candidates.
- README, requirements, ADRs, and architecture notes must stay aligned with the current version baseline.
- Media object list APIs must support deterministic sort controls before pagination so admin clients can build predictable tables.
- Media object administration must expose metadata-only duplicate content-hash detection for deduplication planning without reading media bytes.
- Media object administration must support metadata-only bulk lifecycle updates scoped by exactly one safe selection filter.
- Bulk lifecycle updates must support dry-run previews that do not persist metadata changes.
- Committed lifecycle updates must record latest transition metadata for audit preparation.

## Requirement History

### v0.1.0 - 2026-06-02

- Establish server-managed multi-backend media storage covering local, NFS, SMB, S3-compatible, and distributed backends.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.2.0 - 2026-06-02

- Create the Go API scaffold and storage domain with validation, capability inference, default backend handling, and in-memory repositories.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.3.0 - 2026-06-02

- Expose storage administration through versioned HTTP endpoints for validation, registration, listing, default selection, and disabling.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.4.0 - 2026-06-02

- Protect administrator routes with bootstrap bearer-token authentication while keeping /healthz public.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.5.0 - 2026-06-02

- Add safe filesystem probes for local, NFS, SMB, and mounted distributed backends.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.6.0 - 2026-06-02

- Add conservative S3-compatible object probes with environment-referenced credentials.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.7.0 - 2026-06-02

- Add batch refresh, optional background refresh, and filesystem capacity reporting.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.8.0 - 2026-06-03

- Publish and test the OpenAPI 3.1 contract for the admin API.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.9.0 - 2026-06-03

- Add optional JSON file-backed persistence for storage backend state.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.10.0 - 2026-06-03

- Add the media object registry domain for metadata-only binary asset references.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.11.0 - 2026-06-03

- Expose authenticated media object registration, fetch, and filter endpoints.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.12.0 - 2026-06-03

- Add optional JSON file-backed persistence for media object metadata.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.13.0 - 2026-06-03

- Add read-only filesystem integrity verification for media object references.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.14.0 - 2026-06-03

- Add batch media object verification by backend ID or content hash.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.15.0 - 2026-06-04

- Persist the latest media object verification result in metadata.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.16.0 - 2026-06-04

- Support filtering media objects by latest verification status.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.17.0 - 2026-06-04

- Add limit/offset pagination and pagination metadata to media object lists.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.18.0 - 2026-06-04

- Add metadata-only media object statistics for dashboard-style summaries.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.19.0 - 2026-06-04

- Add metadata-only media object lifecycle updates with terminal deleted semantics.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.20.0 - 2026-06-04

- Support filtering media object lists by lifecycle state.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.21.0 - 2026-06-04

- Support filtering media object lists by asset kind.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.22.0 - 2026-06-04

- Split README content and localize documentation in the previous phase.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.23.0 - 2026-06-04

- Restore Markdown documentation to English as the repository documentation policy.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.24.0 - 2026-06-04

- Support media object list sorting by `backend_object_key`, `created_at`, `updated_at`, `size_bytes`, `object_key`, or `id`, with `asc` or `desc` order before pagination.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.25.0 - 2026-06-04

- Add metadata-only media object duplicate detection by content hash, with optional backend scoping and configurable minimum copy counts.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.26.0 - 2026-06-05

- Add metadata-only bulk media object lifecycle updates selected by exactly one filter, preserving terminal deleted semantics and never deleting media bytes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.27.0 - 2026-06-05

- Add dry-run previews for metadata-only bulk media object lifecycle updates, reporting would-update outcomes without persisting changes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.28.0 - 2026-06-05

- Persist latest committed media object lifecycle change metadata, including previous state, new state, change time, and single/bulk source.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.29.0 - 2026-06-05

- Add a read-only media object metadata timeline endpoint for registration, latest verification, and latest lifecycle transition summaries.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.30.0 - 2026-06-05

- Add metadata-only media object statistics backend scoping.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.31.0 - 2026-06-05

- Add a public `/versionz` endpoint exposing API name, version, commit, and build time.
- Inject version metadata into release binaries and Docker images through build flags and Docker build arguments.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.32.0 - 2026-06-05

- Add a public `/readyz` endpoint with storage service, media registry, and admin authentication readiness checks.
- Add a Docker liveness healthcheck that uses `/healthz` for container runtime monitoring.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.33.0 - 2026-06-06

- Add a public `/metrics` endpoint using Prometheus text exposition for readiness gauges and API build information.
- Keep metrics non-sensitive and aligned with `/readyz` readiness checks and `/versionz` build metadata.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.34.0 - 2026-06-06

- Add low-cardinality HTTP request counters and cumulative duration metrics labeled by method, route pattern, and status.
- Reuse the public `/metrics` endpoint while avoiding raw URL labels and secret-bearing request data.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.35.0 - 2026-06-10

- Add PostgreSQL-backed repository implementations for storage backends and media objects with automatic schema migration and shared connection pool.
- File and in-memory repositories remain available when INORI_DATABASE_URL is not set.
- Integration tests use testcontainers-go with a real PostgreSQL container under the integration build tag.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.36.0 - 2026-06-11

- Add user domain with PostgreSQL persistence: User and Session types, UserRepository and SessionRepository interfaces and PostgreSQL implementations.
- Add bcrypt password hashing (cost=12) and SHA-256 session token storage; plaintext token returned once at login, never stored.
- Add auth.Service: CreateUser, Login, Logout, ValidateToken, DisableUser, DeleteUser, EnsureInitialAdmin.
- Add INORI_SESSION_TTL env var (default 24h) and INORI_INITIAL_ADMIN_USER/PASSWORD bootstrap env vars.
- Add migrations 003_users and 004_sessions to shared PostgreSQL migration runner.
- Add 13 unit tests covering all service paths, race-clean.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.37.0 - 2026-06-11

- Add POST /api/v1/auth/login and POST /api/v1/auth/logout endpoints for session-based authentication.
- Upgrade requireAdminAuth middleware: validate session token via auth.Service first, fall back to INORI_ADMIN_TOKEN bootstrap token.
- Return 503 when neither auth service nor admin token is configured; return 401 on bad/missing credentials.
- Add 8 HTTP-layer tests covering login, logout, session token access, and revoked token denial.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.38.0 - 2026-06-11

- Add authenticated user management APIs for administrators: list users, create users, disable users, and delete users.
- Restrict user management routes to session-authenticated admin users while preserving bootstrap-token fallback behavior.
- Add HTTP-layer tests covering the full user management workflow, validation, conflicts, authorization, and missing auth service handling.
- Extend the OpenAPI contract with auth login/logout and user management schemas and paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.39.0 - 2026-06-12

- Add the music catalog domain foundation with Artist, Album, and Track metadata entities and repository interfaces.
- Add catalog service validation for required names, artist ownership, album membership, media object references, and non-negative numeric metadata.
- Add PostgreSQL-backed catalog repository implementations for artists, albums, and tracks.
- Add migration 005_catalog to the shared PostgreSQL migration runner with catalog tables and lookup indexes.
- Add race-clean catalog service tests and integration-build coverage for the PostgreSQL repository.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.40.0 - 2026-06-12

- Expose authenticated catalog administration endpoints for artists, albums, and tracks under `/api/v1/admin/catalog/`.
- Add list, create, get, and delete operations for all three entity types with `artistId` and `albumId` filter parameters on list endpoints.
- Add `MemoryRepository` to the catalog package for use in HTTP handler tests without external dependencies.
- Update the OpenAPI contract with catalog paths, `UserId`, and `CatalogId` path parameter components.
- Add 11 HTTP-layer catalog tests covering workflows, not-found errors, validation errors, and unconfigured service handling.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.41.0 - 2026-06-13

- Add `POST /api/v1/admin/catalog/import` endpoint that converts a verified media object into a catalog track record.
- Import validates that the media object exists, has `original_audio` or `transcoded_audio` asset kind, and is in `active` lifecycle state.
- Track title falls back to the media object ID when not supplied.
- Artist inherits from the album when only `albumId` is provided.
- Add `MediaObjectReader` interface and `ImportTrackRequest` to the catalog package; `GetMediaObjectInfoForImport` helper on `MediaObjectService`.
- Wire media object service → catalog service via `mediaObjectReaderAdapter` in the HTTP handler layer (no import cycle).
- Add `WithMediaObjectReader` method to `catalog.Service`.
- Add `ErrImportRejected` sentinel error mapped to HTTP 422.
- Add 7 `ImportTrack` unit tests in the catalog package and 7 HTTP-layer tests.
- Update OpenAPI contract with import route and `CatalogImportRequest` schema.


### v0.42.0 - 2026-06-13

- Add PostgreSQL full-text search over catalog metadata via `GET /api/v1/admin/catalog/search?q=`.
- Add migration `006_catalog_fts` with generated `tsvector` columns (weighted `A`/`B` for name vs sort-name) and GIN indexes on `artists`, `albums`, and `tracks`.
- Add `SearchCatalog(ctx, query)` to `catalog.Repository` interface and `catalog.Service`; empty query rejected with validation error.
- Implement `SearchCatalog` on `catalogpg.Repository` using `plainto_tsquery('simple', ...)` with `ts_rank` ordering; results grouped artists → albums → tracks.
- Add `MemoryRepository.SearchCatalog` with case-insensitive substring fallback for unit-test environments.
- Add `CatalogSearchResult`, `SearchResultItem`, `SearchResultKind` types.
- Add 5 `SearchCatalog` service unit tests and 4 HTTP-layer tests.
- Add `TestRepositorySearchCatalog` integration test (build tag: integration).
- Update OpenAPI contract with `/api/v1/admin/catalog/search` path, `CatalogSearchResult`, `SearchResultItem`, `SearchResultKind` schemas, and new error codes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.43.0 - 2026-06-13

- Add read-only catalog browse endpoints for session-authenticated viewers and admins under `/api/v1/catalog/`: list/get artists, list/get albums (`?artistId=`), list/get tracks (`?albumId=`/`?artistId=`), and full-text search (`?q=`).
- Add `requireViewerAuth` middleware that accepts any valid session token (admin or viewer role) but rejects the static bootstrap admin token; returns 503 when no auth service is configured, 401 for missing or invalid tokens.
- Reuse existing `listArtists`, `getArtist`, `listAlbums`, `getAlbum`, `listTracks`, `getTrack`, and `searchCatalog` handlers without modification.
- Fix missing 405 method-not-allowed fallback for `/api/v1/admin/catalog/search`.
- Add `newViewerTestHandler` helper and 11 HTTP-layer tests covering viewer/admin session, 401 unauthorized, 503 for no-auth-service, not-found, missing query, seeded search, and 405 guards.
- Update OpenAPI contract with 7 new viewer catalog paths; bump `info.version` to `0.43.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.44.0 - 2026-06-13

- Add `POST /api/v1/admin/catalog/batch-import` endpoint that accepts a list of `CatalogImportRequest` items and processes each independently.
- Return HTTP 200 on full success, HTTP 207 Multi-Status on partial success, HTTP 422 when all items fail.
- Each result item carries `index`, `mediaObjectId`, the created `track` on success, or `error`/`errorCode` on failure.
- Add `BatchImportTracks(ctx, items)` method to `catalog.Service`; individual item failures do not abort subsequent items.
- Add `BatchImportResult`, `BatchImportResultItem` types to the catalog package.
- Add 5 `BatchImportTracks` unit tests and 6 HTTP-layer tests covering full-success, partial-success, all-fail, empty batch, no-catalog-service, and 405 guard.
- Update OpenAPI contract with `/api/v1/admin/catalog/batch-import` path and `CatalogBatchImportRequest`, `CatalogBatchImportResult`, `CatalogBatchImportResultItem` schemas; bump `info.version` to `0.44.0`.
- Fix pre-existing flaky `TestMediaObjectServiceUpdatesLifecycleState` by injecting a stepping clock that guarantees distinct timestamps across `RegisterMediaObject` and `SetMediaObjectLifecycleState` calls.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.45.0 - 2026-06-14

- Add `PATCH /api/v1/admin/catalog/artists/{id}`, `PATCH /api/v1/admin/catalog/albums/{id}`, and `PATCH /api/v1/admin/catalog/tracks/{id}` endpoints for partial metadata updates.
- Pointer-typed request fields (`*string`, `*int`) distinguish "not provided" from "explicitly empty", enabling clients to clear optional fields without touching unmentioned fields.
- Validation mirrors create: name, title, and artistId may not be set to an empty string; numeric fields must be non-negative; artist ownership of the referenced album is enforced when updating a track's albumId.
- Add `UpdateArtistRequest`, `UpdateAlbumRequest`, `UpdateTrackRequest` types to the catalog package.
- Add `WithClock` setter on `catalog.Service` for deterministic timestamp injection in tests.
- Implement `UpdateArtist`, `UpdateAlbum`, `UpdateTrack` on `catalog.Service`; each reads the current record, applies non-nil fields, bumps `UpdatedAt`, and saves.
- Add 11 `catalog.Service` unit tests and 7 HTTP-layer tests covering field changes, nil-field passthrough, empty-name rejection, not-found, and unconfigured-service guard.
- Update OpenAPI contract with `CatalogUpdateArtistRequest`, `CatalogUpdateAlbumRequest`, `CatalogUpdateTrackRequest` schemas and `patch` operations on artist, album, and track `/{id}` paths; bump `info.version` to `0.45.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.46.0 - 2026-06-14

- Add `Playlist` entity to the catalog domain: ordered collection of tracks with name, optional description, and an ordered `TrackIDs` list.
- Add `ErrInvalidPlaylist` and `ErrPlaylistNotFound` sentinel errors to the catalog package.
- Extend `catalog.Repository` interface with `SavePlaylist`, `GetPlaylist`, `ListPlaylists`, `DeletePlaylist`.
- Implement playlist methods on `catalog.MemoryRepository` (defensive slice copies) and `catalogpg.Repository` (transactional upsert + playlist_tracks replace).
- Add `CreatePlaylist`, `ListPlaylists`, `GetPlaylist`, `DeletePlaylist`, `UpdatePlaylist`, `AddTrackToPlaylist`, `RemoveTrackFromPlaylist` methods to `catalog.Service`; validate name non-empty, enforce track existence on add.
- Add migration `007_playlists` with `playlists` and `playlist_tracks` tables; `playlist_tracks` uses `ON DELETE CASCADE` for both foreign keys and a `(playlist_id, position)` primary key for ordering.
- Expose admin playlist endpoints under `/api/v1/admin/catalog/playlists/`: list, create, get, PATCH metadata, delete, `POST /{id}/tracks` (append), `DELETE /{id}/tracks/{trackId}` (remove first occurrence).
- Expose viewer-only read endpoints under `/api/v1/catalog/playlists/`: list and get.
- Add `ErrInvalidPlaylist` and `ErrPlaylistNotFound` to the `writeError` switch in the HTTP handler.
- Add 9 `catalog.Service` unit tests and 7 HTTP-layer tests covering CRUD, add/remove track, viewer access, not-found, empty-name rejection, and 405 guard.
- Update OpenAPI contract with `Playlist`, `CreatePlaylistRequest`, `UpdatePlaylistRequest`, `AddPlaylistTrackRequest` schemas and all 8 new paths; bump `info.version` to `0.46.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.47.0 - 2026-06-14

- Add `PUT /api/v1/admin/catalog/playlists/{id}/tracks` endpoint that atomically replaces the entire ordered track list of a playlist.
- Every supplied track ID must exist; an unknown ID returns 404. An empty `trackIds` array is valid and clears the playlist. Duplicate entries are preserved.
- Add `SetPlaylistTracks(ctx, playlistID, trackIDs)` to `catalog.Service`; validates track existence via `repo.GetTrack` for each ID, then calls `repo.SavePlaylist` which already performs a transactional full-replace of `playlist_tracks` rows in the PostgreSQL backend.
- No `catalog.Repository` interface change required — `SavePlaylist` already handles atomic replacement.
- Add `setPlaylistTracksRequest` struct and `setPlaylistTracks` handler to the HTTP handler layer.
- Add 5 `catalog.Service` unit tests and 7 HTTP-layer tests covering reorder, clear, duplicate preservation, unknown track, unknown playlist, missing `trackIds` field, and no-catalog-service 503.
- Add `SetPlaylistTracksRequest` schema and `put` operation on `/api/v1/admin/catalog/playlists/{id}/tracks` to the OpenAPI contract; bump `info.version` to `0.47.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.48.0 - 2026-06-14

- Add `POST /api/v1/admin/catalog/tracks/{id}/relink` endpoint that replaces the media object reference on an existing track.
- The new media object must exist, have `assetKind` of `original_audio` or `transcoded_audio`, and have `lifecycleState` of `active`; otherwise `422 relink_rejected` is returned.
- Add `ErrRelinkRejected` sentinel error and `RelinkTrackRequest` type to the catalog package.
- Add `RelinkTrack(ctx, id, req)` to `catalog.Service`; validates media object via `MediaObjectReader` before overwriting `mediaObjectId` and bumping `UpdatedAt`.
- Add `relinkTrack` handler and route to the HTTP handler layer; register 405 fallback for the sub-path.
- Add 7 `catalog.Service` unit tests and 6 HTTP-layer tests covering success, wrong asset kind, not-active lifecycle, media not found, track not found, no reader configured, and empty `mediaObjectId`.
- Add `CatalogRelinkTrackRequest` schema and `post` operation on `/api/v1/admin/catalog/tracks/{id}/relink` to the OpenAPI contract; bump `info.version` to `0.48.0` (corrected in v0.49.0).
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.49.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/playlists/{id}/tracks` and `GET /api/v1/catalog/playlists/{id}/tracks` endpoints that return the full ordered `Track` object list for a playlist in a single request, eliminating the need for N separate per-track fetches.
- An empty playlist returns an empty `tracks` array. Duplicate track entries (same ID appearing multiple times) are expanded once per occurrence in order.
- Add `GetPlaylistTracks(ctx, playlistID)` to `catalog.Service`; resolves each `trackID` in the playlist's `TrackIDs` slice via `repo.GetTrack` and returns them in order. Returns `ErrPlaylistNotFound` for unknown playlist IDs.
- Add `getPlaylistTracks` handler shared by both the admin and viewer routes; response shape is `{"tracks": [...Track...]}`.
- Register `GET /api/v1/admin/catalog/playlists/{id}/tracks` (admin-auth) and `GET /api/v1/catalog/playlists/{id}/tracks` (viewer-auth) routes; register 405 fallback for the viewer sub-path.
- Add 4 `catalog.Service` unit tests (ordered, empty, not-found, duplicate expansion) and 6 HTTP-layer tests (admin happy path, empty playlist, 404, viewer access, no-catalog-service 503, method-not-allowed 405).
- Add `PlaylistTracksResult` schema and `get` operations on both tracks sub-paths to the OpenAPI contract; bump `info.version` to `0.49.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert all eight playlist paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.50.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/stats` endpoint returning metadata-only aggregate entity counts for artists, albums, tracks, and playlists as a `CatalogStats` object.
- Add `CatalogStats` struct to the catalog package with `artists`, `albums`, `tracks`, `playlists` integer fields.
- Add `GetCatalogStats(ctx)` to `catalog.Service`; delegates to `repo.ListArtists`, `repo.ListAlbums`, `repo.ListTracks`, `repo.ListPlaylists` and returns counts. No new `Repository` interface methods required.
- Add `getCatalogStats` handler to the HTTP handler layer; returns 503 when no catalog service is configured.
- Register `GET /api/v1/admin/catalog/stats` (admin-auth) and 405 fallback for the path.
- Add 3 `catalog.Service` unit tests (empty catalog, populated counts, no-error baseline) and 4 HTTP-layer tests (empty response shape, populated counts, no-catalog-service 503, method-not-allowed 405).
- Add `CatalogStats` schema and `get` operation on `/api/v1/admin/catalog/stats` to the OpenAPI contract; bump `info.version` to `0.50.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.51.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/stats/artists` endpoint returning per-artist album and track counts as a `CatalogArtistStatsBreakdown` object, eliminating the need for N separate list calls.
- Add `GET /api/v1/admin/catalog/stats/albums` endpoint returning per-album track counts as a `CatalogAlbumStatsBreakdown` object.
- Add `ArtistStatItem`, `ArtistStatsBreakdown`, `AlbumStatItem`, `AlbumStatsBreakdown` types to the catalog package.
- Add `GetArtistStatsBreakdown(ctx)` and `GetAlbumStatsBreakdown(ctx)` to `catalog.Service`; counts are derived from existing `ListAlbumsByArtist`, `ListTracksByArtist`, `ListTracksByAlbum` calls. No new `Repository` interface methods required.
- Add `getArtistStatsBreakdown` and `getAlbumStatsBreakdown` handlers to the HTTP handler layer; each returns 503 when no catalog service is configured.
- Register `GET /api/v1/admin/catalog/stats/artists` and `GET /api/v1/admin/catalog/stats/albums` (admin-auth) with 405 fallbacks.
- Add 4 `catalog.Service` unit tests (empty/populated for each breakdown) and 8 HTTP-layer tests (empty shape, populated counts, 503, 405 for each endpoint).
- Add `CatalogArtistStatItem`, `CatalogArtistStatsBreakdown`, `CatalogAlbumStatItem`, `CatalogAlbumStatsBreakdown` schemas and `get` operations on both new paths to the OpenAPI contract; bump `info.version` to `0.51.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.52.0 - 2026-06-14

- Add `GET /api/v1/admin/catalog/stats/playlists` endpoint returning per-playlist track counts as a `CatalogPlaylistStatsBreakdown` object.
- Add `PlaylistStatItem`, `PlaylistStatsBreakdown` types to the catalog package; each item carries `playlistId`, `name`, and `trackCount` (duplicate track entries counted separately).
- Add `GetPlaylistStatsBreakdown(ctx)` to `catalog.Service`; counts are derived from each playlist's `TrackIDs` slice length. No new `Repository` interface methods required.
- Add `getPlaylistStatsBreakdown` handler to the HTTP handler layer; returns 503 when no catalog service is configured.
- Register `GET /api/v1/admin/catalog/stats/playlists` (admin-auth) and 405 fallback for the path.
- Add 2 `catalog.Service` unit tests (empty, populated with duplicate-track counting) and 4 HTTP-layer tests (empty shape, populated counts, no-catalog-service 503, method-not-allowed 405).
- Add `CatalogPlaylistStatItem`, `CatalogPlaylistStatsBreakdown` schemas and `get` operation on `/api/v1/admin/catalog/stats/playlists` to the OpenAPI contract; bump `info.version` to `0.52.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the new playlists stats path.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.


### v0.53.0 - 2026-06-15

- Add `GET /api/v1/admin/catalog/recently-added` endpoint returning a newest-first unified timeline of recently created artists, albums, and tracks.
- Add `RecentItemKind`, `RecentCatalogItem`, and `RecentCatalogResult` types to the catalog package. Each timeline item includes `kind`, one entity payload, and `addedAt` copied from the entity's `CreatedAt` timestamp.
- Add `GetRecentlyAdded(ctx, kind, limit)` to `catalog.Service`; it derives results from existing `ListArtists`, `ListAlbums`, and `ListTracks` repository methods, supports `kind=artist|album|track`, defaults `limit` to 20, and clamps values above 100. No new `Repository` interface methods required.
- Add `getRecentlyAdded` handler to the HTTP handler layer; returns 400 for invalid `limit` or `kind`, 503 when no catalog service is configured, and registers a 405 fallback for the path.
- Register `GET /api/v1/admin/catalog/recently-added` (admin-auth).
- Add 5 `catalog.Service` unit tests and 8 HTTP-layer tests covering empty response shape, populated timeline payload, kind filter, invalid kind, invalid limit, limit handling, no-catalog-service 503, and method-not-allowed 405.
- Add `RecentItemKind`, `RecentCatalogItem`, and `RecentCatalogResult` schemas plus the `get` operation on `/api/v1/admin/catalog/recently-added` to the OpenAPI contract; bump `info.version` to `0.53.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the recently-added path.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.54.0 - 2026-06-15

- Add `GET /api/v1/admin/catalog/recently-updated` endpoint returning a newest-first unified timeline of recently updated artists, albums, and tracks.
- Add `UpdatedCatalogItem` and `UpdatedCatalogResult` types to the catalog package. Each timeline item includes `kind`, one entity payload, and `updatedAt` copied from the entity's `UpdatedAt` timestamp.
- Add `GetRecentlyUpdated(ctx, kind, limit)` to `catalog.Service`; it derives results from existing `ListArtists`, `ListAlbums`, and `ListTracks` repository methods, supports `kind=artist|album|track`, defaults `limit` to 20, and clamps values above 100. No new `Repository` interface methods required.
- Add `getRecentlyUpdated` handler to the HTTP handler layer; returns 400 for invalid `limit` or `kind`, 503 when no catalog service is configured, and registers a 405 fallback for the path.
- Register `GET /api/v1/admin/catalog/recently-updated` (admin-auth).
- Add `catalog.Service` unit tests and HTTP-layer tests covering empty response shape, updated timestamp ordering, kind filter, invalid kind, invalid limit, limit handling, no-catalog-service 503, and method-not-allowed 405.
- Add `UpdatedCatalogItem` and `UpdatedCatalogResult` schemas plus the `get` operation on `/api/v1/admin/catalog/recently-updated` to the OpenAPI contract; bump `info.version` to `0.54.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the recently-updated path.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.55.0 - 2026-06-15

- Add viewer-facing `GET /api/v1/catalog/recently-added` and `GET /api/v1/catalog/recently-updated` endpoints requiring session authentication (`requireViewerAuth`).
- Reuse existing `getRecentlyAdded` and `getRecentlyUpdated` handlers, wrapping them with `requireViewerAuth` middleware instead of `requireAdminAuth`.
- Register 405 fallbacks for both viewer paths.
- Add 16 HTTP-layer tests covering viewer auth success, admin session acceptance, static bootstrap token rejection, unauthorized requests, invalid kind/limit, and method-not-allowed.
- Add viewer path operations to the OpenAPI contract under the "Catalog" tag; bump `info.version` to `0.55.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the viewer recently-added and recently-updated paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.56.0 - 2026-06-15

- Add playlist entries to the recently-added and recently-updated unified catalog timelines.
- Extend `RecentItemKind` enum with `playlist`, and add `Playlist` fields to `RecentCatalogItem` and `UpdatedCatalogItem` types.
- Extend `GetRecentlyAdded` and `GetRecentlyUpdated` to iterate over playlists when `kind` is empty or `playlist`.
- Update `validateRecentItemKind` to accept `playlist` as a valid kind.
- Update OpenAPI contract: `RecentItemKind` enum, schemas, and endpoint descriptions; bump `info.version` to `0.56.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.57.0 - 2026-06-16

- Repair playlist participation in recently-added and recently-updated catalog timelines: the v0.56.0 implementation was incomplete — `RecentItemPlaylist` constant, `Playlist` payload fields on `RecentCatalogItem` and `UpdatedCatalogItem`, and `ListPlaylists` iterations in `GetRecentlyAdded` / `GetRecentlyUpdated` were missing.
- Update `validateRecentItemKind` to accept `"playlist"` and update the validation message to name all four valid kinds.
- Update OpenAPI contract: add `relink_rejected`, `validation_error`, and `invalid_limit` to the error code enum; bump `info.version` to `0.57.0`.
- Correct `services/api/internal/storage/capacity.go`: remove duplicate `FilesystemCapacityProvider` body now superseded by the build-tagged `capacity_unix.go` and `capacity_unsupported.go` files pulled in with the upstream update.
- Strengthen `openapi_contract_test.go`: assert `patch` on artist/album `{id}` paths, assert all three new error codes, and add `TestStorageAdminOpenAPIContractRecentTimelineSchemas` asserting the `RecentItemKind` enum includes `"playlist"` and both recent timeline item schemas carry a `playlist` payload ref.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.58.0 - 2026-06-17

- Add `GET /api/v1/catalog/tracks/{id}/playback` viewer-only endpoint returning a metadata-only `TrackPlaybackDescriptor` with `trackId`, `mediaObjectId`, `mimeType`, `durationMs`, `backendId`, `backendType`, and `objectKey`.
- Validate that the linked media object has `lifecycleState = active` and `assetKind ∈ {original_audio, transcoded_audio}`; return 422 `playback_unavailable` otherwise.
- Add `ErrPlaybackUnavailable` sentinel to the storage package and `playback_unavailable` to the `writeError` switch.
- Add `TrackPlaybackDescriptor` schema, `GET /api/v1/catalog/tracks/{id}/playback` path, and `playback_unavailable` error code to the OpenAPI contract; bump `info.version` to `0.58.0`.
- Add 8 HTTP-layer tests covering success, admin-session access, track-not-found, media-object-not-found, non-active lifecycle, wrong asset kind, no-catalog-service, and method-not-allowed.
- Extend `openapi_contract_test.go` with the new path, schema, error code, and `TestStorageAdminOpenAPIContractTrackPlaybackDescriptor`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.59.0 - 2026-06-17

- Expose viewer-accessible catalog stats endpoints: `GET /api/v1/catalog/stats`, `GET /api/v1/catalog/stats/artists`, `GET /api/v1/catalog/stats/albums`, and `GET /api/v1/catalog/stats/playlists`.
- Reuse existing `getCatalogStats`, `getArtistStatsBreakdown`, `getAlbumStatsBreakdown`, and `getPlaylistStatsBreakdown` handler functions under `requireViewerAuth`; no new domain logic required.
- Add 405 fallbacks for all four new viewer stats paths.
- Add 14 HTTP-layer tests covering empty stats, populated counts, admin session acceptance, no-catalog-service 503, and method-not-allowed for all four endpoints.
- Add four viewer stats paths to the OpenAPI contract; bump `info.version` to `0.59.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the four new viewer stats paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.60.0 - 2026-06-17

- Add `presignS3URL` to the storage package: generates an AWS Signature Version 4 presigned GET URL using query-parameter signing, reusing existing `s3ObjectURL`, `s3SigningKey`, and `hmacSHA256` helpers from `s3_probe.go`.
- Add `storage.Service.GetBackend(ctx, id)` for direct single-backend lookup; `storage.DefaultPresignedURLTTL = 15 * time.Minute` constant; `storage.Service.GeneratePresignedURL(ctx, backendID, objectKey, ttl)` orchestrating capability check, credential resolution via env var refs, and presigned URL generation.
- Extend `GET /api/v1/catalog/tracks/{id}/playback` response: populate optional `presignedUrl` field when the backend has `PresignedURLs` capability and credentials are configured; presign failures are non-fatal.
- Replace the backend full-list scan in `getTrackPlayback` with a single `GetBackend` call.
- Add `presignedUrl` optional property to `TrackPlaybackDescriptor` OpenAPI schema; bump `info.version` to `0.60.0`.
- Add 4 `presignS3URL` unit tests, 4 `GetBackend`/`GeneratePresignedURL` service tests, and 1 HTTP-layer presigned URL handler test.
- Extend `TestStorageAdminOpenAPIContractTrackPlaybackDescriptor` to assert `presignedUrl` is present in properties but absent from `required`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.61.0 - 2026-06-17

- Add `limit`/`offset` pagination to all four catalog list endpoints: artists, albums, tracks, and playlists (both admin and viewer routes).
- Add `parseCatalogPage` and `paginateCatalog[T]` helpers to the HTTP handler layer; no Repository or Service interface changes required.
- Add `CatalogPaginationMeta` type with `limit`, `offset`, `total`, and `hasMore` fields; all four list responses now include a `pagination` envelope alongside the existing entity array.
- `limit` defaults to 50 (max 500); `offset` defaults to 0; invalid values return 400 `invalid_limit` / `invalid_offset`.
- Existing `artistId` and `albumId` filter params continue to work and are applied before pagination.
- Add `CatalogPaginationMeta` schema, `limit`/`offset` params, and `pagination` response property to all 8 catalog list paths in the OpenAPI contract; bump `info.version` to `0.61.0`.
- Add `invalid_offset` to the OpenAPI error code enum and contract test assertion.
- Add 5 HTTP-layer tests covering limit, offset, hasMore, invalid params, and viewer-session access.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.62.0 - 2026-06-17

- Add `sortBy` and `sortOrder` query parameters to all four catalog list endpoints (artists, albums, tracks, playlists) on both admin and viewer routes.
- Artists: `sortBy` accepts `name` (default), `sortName`, `createdAt`, `updatedAt`.
- Albums: `sortBy` accepts `title` (default), `sortTitle`, `releaseYear`, `createdAt`, `updatedAt`.
- Tracks: `sortBy` accepts `title` (default), `sortTitle`, `trackNumber`, `discNumber`, `durationMs`, `createdAt`, `updatedAt`.
- Playlists: `sortBy` accepts `name` (default), `createdAt`, `updatedAt`.
- `sortOrder` must be `asc` (default) or `desc`; any other value returns `400 invalid_sort_order`.
- Sort is applied before pagination so `limit`/`offset` windows remain stable.
- Add sort-field constants to `catalog/types.go`; add `normalizeSortOrder` and per-entity sort functions to the HTTP handler layer; no Repository or Service interface changes.
- Add `sortBy`/`sortOrder` params and sort descriptions to all 8 catalog list paths in the OpenAPI contract; add `invalid_sort_order` to error enum; bump `info.version` to `0.62.0`.
- Add `TestStorageAdminOpenAPIContractCatalogListSortParams` asserting sort params on all 8 list paths.
- Add 6 HTTP-layer tests covering per-entity sort directions, invalid `sortOrder`, and viewer-session sort access.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.63.0 - 2026-06-17

- Add nested catalog browse routes: `GET /api/v1/catalog/artists/{id}/albums`, `GET /api/v1/catalog/artists/{id}/tracks`, and `GET /api/v1/catalog/albums/{id}/tracks` under both admin (`/api/v1/admin/catalog/…`) and viewer (`/api/v1/catalog/…`) paths.
- Each handler validates the parent entity (artist or album) before listing sub-entities; unknown IDs return 404.
- All six new routes support the same `limit`, `offset`, `sortBy`, and `sortOrder` parameters established in Phases 61–62.
- Add 4 HTTP-layer tests covering pagination, sort, 404 on unknown parent, 405 method-not-allowed, and viewer-session access to all three nested routes.
- Add 6 new paths to the OpenAPI contract with typed response schemas (albums/tracks with pagination) and full pagination+sort parameter declarations; bump `info.version` to `0.63.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` and `TestStorageAdminOpenAPIContractCatalogListSortParams` for all 6 new paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.64.0 - 2026-06-17

- Add `limit`/`offset` pagination to `GET /api/v1/catalog/playlists/{id}/tracks` and `GET /api/v1/admin/catalog/playlists/{id}/tracks`.
- Playlist track order is user-curated and is preserved exactly within each page; `sortBy`/`sortOrder` are intentionally not exposed.
- Response now includes a `pagination` envelope (`limit`, `offset`, `total`, `hasMore`) alongside the `tracks` array.
- Add `limit`/`offset` query params and `pagination` response property to both playlist tracks paths in the OpenAPI contract; bump `info.version` to `0.64.0`.
- Add `TestStorageAdminOpenAPIContractPlaylistTracksPagination` asserting `limit`/`offset` present and `sortBy`/`sortOrder` absent.
- Add 2 HTTP-layer tests covering order preservation, limit, offset, `hasMore`, invalid params, and viewer-session access.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.65.0 - 2026-06-17

- Add `ListQuery` and `ListPage[T]` types and 7 new `ListXxxPage` methods to the `catalog.Repository` interface: `ListArtistsPage`, `ListAlbumsPage`, `ListAlbumsByArtistPage`, `ListTracksPage`, `ListTracksByAlbumPage`, `ListTracksByArtistPage`, `ListPlaylistsPage`.
- Implement the 7 page methods on `catalog.MemoryRepository` (Go in-memory sort + slice) and add a new `catalog/postgres/repository_page.go` with SQL `ORDER BY … LIMIT $1 OFFSET $2` and `COUNT(*) OVER ()` window function for accurate total counts without a separate query.
- PostgreSQL ORDER BY uses `lower()` wrapping for text fields (consistent with existing list queries) and an `id` tiebreak for stable pagination across pages.
- Update all 7 catalog list HTTP handlers (`listArtists`, `listAlbums`, `listTracks`, `listPlaylists`, `listAlbumsByArtist`, `listTracksByArtist`, `listTracksByAlbum`) to call the new Page methods and remove the previous in-handler sort+paginate logic.
- Add 4 `ListXxxPage` catalog service unit tests (artist sort/paginate/offset-past-end, albums-by-artist sort, tracks paginate, playlists desc).
- Add 2 PostgreSQL integration tests under the `integration` build tag (`TestRepositoryListArtistsPage`, `TestRepositoryListAlbumsPageByArtist`).
- No change to HTTP API shape — client-facing behavior is identical to Phases 61–62.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.66.0 - 2026-06-17

- Add 4 aggregate stats methods to the `catalog.Repository` interface: `CountEntities`, `ArtistAlbumTrackCounts`, `AlbumTrackCounts`, `PlaylistTrackCounts`.
- Implement on `catalog.MemoryRepository` (in-memory counting) and `catalogpg.Repository` with SQL `COUNT(*)`/`GROUP BY` aggregate queries (single query per stats method, no N+1 iteration).
- Replace `GetCatalogStats`, `GetArtistStatsBreakdown`, `GetAlbumStatsBreakdown`, and `GetPlaylistStatsBreakdown` in `catalog.Service` with single-aggregate-call implementations.
- Add 2 PostgreSQL integration tests (`TestRepositoryCountEntities`, `TestRepositoryArtistAlbumTrackCounts`) under the `integration` build tag.
- Bump OpenAPI `info.version` to `0.66.0`. No HTTP API shape change.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.67.0 - 2026-06-17

- Add `RecentlyAdded(ctx, kind, limit)` and `RecentlyUpdated(ctx, kind, limit)` to the `catalog.Repository` interface.
- Implement on `catalog.MemoryRepository` (in-memory sort + slice) and `catalogpg.Repository` (single `UNION ALL … ORDER BY … LIMIT` query per method).
- Replace the 4-branch list-and-merge logic in `GetRecentlyAdded` and `GetRecentlyUpdated` (`catalog.Service`) with single delegate calls to the new repo methods; remove unused `sort` import from `service.go`.
- Add `TestRepositoryRecentlyAdded` PostgreSQL integration test (unified + kind filter + limit).
- Bump OpenAPI `info.version` to `0.67.0`. No HTTP API shape change.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.68.0 - 2026-06-17

- Add playback history domain: new `history` package with `PlayEvent` type, `Repository` interface, `Service` with `RecordPlay`/`ListPlays`/`ClearHistory` methods, in-memory repository, and PostgreSQL repository.
- Add migration `008_play_events` (id, user_id, track_id, played_at, created_at; FK cascades, two indexes).
- Extend `requireViewerAuth` and `requireAdminAuth` to inject the authenticated `auth.User` into the request context for downstream handler use.
- Add viewer-only `POST/GET/DELETE /api/v1/me/history` endpoints (user-scoped, session-auth required).
- Add `PlayEvent`, `PlayEventList` schemas and `/api/v1/me/history` path to OpenAPI contract; add `history_not_configured` error code; bump `info.version` to `0.68.0`.
- Add 5 history service unit tests and 5 HTTP-layer tests (record, list w/ pagination, clear, 503 not-configured, 405).
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.69.0 - 2026-06-17

- Extend the `history.Repository` interface with 3 aggregate stats methods: `HistoryStats(ctx)`, `TopTracks(ctx, limit)`, `TopUsers(ctx, limit)`.
- Implement on `history.MemoryRepository` (in-memory counting + sort) and `historypg.Repository` (single SQL `COUNT`/`GROUP BY` aggregate queries per method).
- Add `GetHistoryStats`, `GetTopTracks`, `GetTopUsers` to `history.Service`; limit clamped to 100, default 10.
- Add 3 admin-only routes: `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, `GET /api/v1/admin/history/top-users`.
- Add `HistoryStats`, `TrackPlayCount`, `UserPlayCount`, `TopTracksResult`, `TopUsersResult` schemas and the 3 new admin paths to the OpenAPI contract; bump `info.version` to `0.69.0`.
- Add 3 history service unit tests (GetHistoryStats, GetTopTracks, GetTopUsers) and 5 HTTP-layer tests (stats, top-tracks with limit, top-users, not-configured 503, 405).
- Add `TestStorageAdminOpenAPIContractAdminHistoryPaths` asserting new schemas and path operations.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.70.0 - 2026-06-18

- Add optional `?since=<RFC3339>` query parameter to `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, and `GET /api/v1/admin/history/top-users`; omitting it returns all-time data.
- Add `StatsFilter{Since time.Time}` type to `history/types.go`; update `Repository` interface so all three aggregate methods accept `StatsFilter`.
- Implement `since` filtering on `history.MemoryRepository` (in-memory `played_at >= since` guard) and `historypg.Repository` (`WHERE played_at >= $N` SQL clause).
- Thread `StatsFilter` through `history.Service` methods `GetHistoryStats`, `GetTopTracks`, `GetTopUsers`.
- Add `parseHistoryAdminFilter` helper in the HTTP handler layer to parse and validate the `since` param; invalid timestamps return `400 invalid_since`.
- Add 3 service unit tests (windowed stats, top-tracks, top-users) and 2 HTTP-layer tests (`TestAdminHistorySinceFilter`, `TestAdminHistorySinceInvalid`).
- Add `since` query param (string/date-time, optional) to all three admin history GET paths in the OpenAPI contract; add `invalid_since` to error code enum; bump `info.version` to `0.70.0`.
- Add `TestStorageAdminOpenAPIContractAdminHistorySinceParam` asserting the `since` param is present and optional on all three paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.71.0 - 2026-06-18

- Add optional `?until=<RFC3339>` (exclusive upper bound) query parameter to `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, and `GET /api/v1/admin/history/top-users`; composes with `?since`.
- Add `Until time.Time` to `StatsFilter` in `history/types.go`.
- Implement `until` guard (`played_at < until`, exclusive) on `history.MemoryRepository` aggregate methods.
- Replace four-branch since/no-since logic in `historypg.Repository` with a shared `statsWhere(f)` helper that builds `WHERE played_at >= $N AND played_at < $M` dynamically for any combination of bounds.
- Extend `parseHistoryAdminFilter` in `handler.go` to parse `?until` (returns `400 invalid_until` if unparseable) and validate `since < until` when both are present (returns `400 invalid_time_range`).
- Add 3 service unit tests (until-stats, since+until window on top-tracks, until combined) and 4 HTTP-layer tests (until filter, invalid until, invalid time range, updated since test).
- Add `until` query param (string/date-time, optional) to all three admin history GET paths in OpenAPI; add `invalid_until` and `invalid_time_range` to error code enum; bump `info.version` to `0.71.0`.
- Add `TestStorageAdminOpenAPIContractAdminHistoryUntilParam`; extend `TestStorageAdminOpenAPIContractSchemasAndErrors` for new error codes.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.72.0 - 2026-06-18

- Add `AdminPlayEventFilter{TrackID, UserID, Limit, Offset}` to `history/types.go`; add `ListPlayEventsByTrack(ctx, AdminPlayEventFilter)` to the `Repository` interface.
- Implement `ListPlayEventsByTrack` on `history.MemoryRepository` (in-memory filter + sort + slice) and `historypg.Repository` (`WHERE track_id = $3 [AND user_id = $4] … COUNT(*) OVER()`).
- Add `GetUserHistory` (admin-facing reuse of `ListPlayEvents` without user-scope restriction) and `GetTrackHistory` to `history.Service`.
- Add 2 admin routes: `GET /api/v1/admin/history/users/{userId}` (paginated events for any user, optional `?trackId` filter) and `GET /api/v1/admin/history/tracks/{trackId}` (paginated events for any track, optional `?userId` filter); add `methodNotAllowed` fallbacks for both.
- Add `getAdminUserHistory`, `getAdminTrackHistory` handler functions and `parseHistoryAdminPagination` helper; response shape is `{events, pagination}` identical to `GET /api/v1/me/history`.
- Add 2 service unit tests (`TestGetUserHistory`, `TestGetTrackHistory`) and 4 HTTP-layer tests (user history with pagination, track history, 405, 503 not-configured).
- Add 2 new paths to the OpenAPI contract with `PlayEventList` response schema ref; bump `info.version` to `0.72.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractAdminHistoryDetailPaths` asserting path params, query filters, pagination params, and response schema ref.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.73.0 - 2026-06-18

- Add `DeletePlayEventsByUserAdmin(ctx, userID)`, `DeletePlayEventsByTrack(ctx, trackID)`, and `DeletePlayEventsInWindow(ctx, StatsFilter)` to the `history.Repository` interface.
- Implement all three on `history.MemoryRepository` (in-memory guard under lock) and `historypg.Repository` (`DELETE FROM play_events WHERE …`; `DeletePlayEventsInWindow` reuses `statsWhere` helper).
- Add `AdminDeleteUserHistory`, `AdminDeleteTrackHistory`, and `AdminDeleteHistoryWindow` to `history.Service`; `AdminDeleteHistoryWindow` validates that at least one time bound is set.
- Add `DELETE /api/v1/admin/history/users/{userId}` and `DELETE /api/v1/admin/history/tracks/{trackId}` to existing paths; add new `DELETE /api/v1/admin/history` path with optional `?since`/`?until` time-window filter (at least one required at runtime).
- Return 400 `missing_time_filter` when neither `since` nor `until` is supplied to the window endpoint.
- Add `methodNotAllowed` fallback for `/api/v1/admin/history`.
- Add 3 `history.Service` unit tests (`TestAdminDeleteUserHistory`, `TestAdminDeleteTrackHistory`, `TestAdminDeleteHistoryWindow`).
- Add 5 HTTP-layer tests (`TestAdminDeleteUserHistory`, `TestAdminDeleteTrackHistory`, `TestAdminDeleteHistoryWindow`, `TestAdminDeleteHistoryWindowMissingFilter`, `TestAdminBulkDeleteHistoryNotConfigured`).
- Add `delete` operations to both detail paths and new window path in OpenAPI contract; add `missing_time_filter` to error code enum; bump `info.version` to `0.73.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `delete` on detail paths and new window path; add `TestStorageAdminOpenAPIContractAdminHistoryBulkDelete`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.74.0 - 2026-06-18

- Add `UserStatsFilter{UserID, Since, Until}` and `UserHistoryStats{TotalEvents, UniqueTracks}` types to `history/types.go`.
- Add `UserTopTracks(ctx, UserStatsFilter, limit)` and `UserHistoryStats(ctx, UserStatsFilter)` methods to the `Repository` interface.
- Implement both on `history.MemoryRepository` (user-scoped in-memory filter + sort) and `historypg.Repository` (new `userStatsWhere` helper that mandates `user_id = $1`).
- Add `GetMyStats` and `GetMyTopTracks` to `history.Service`; validate `UserID != ""`.
- Add viewer-only `GET /api/v1/me/history/stats` and `GET /api/v1/me/history/top-tracks` endpoints; both accept optional `?since`, `?until`; top-tracks also accepts `?limit` (default 10, max 100).
- Reuse `parseHistoryAdminFilter` and `parseHistoryAdminLimit` in the new handlers; inject `UserID` from auth context.
- Add `methodNotAllowed` fallbacks for both new viewer paths.
- Add 3 `history.Service` unit tests (`TestGetMyStats`, `TestGetMyTopTracks`, `TestGetMyTopTracksTimeWindow`).
- Add 4 HTTP-layer tests (`TestGetMyHistoryStats`, `TestGetMyTopTracks`, `TestGetMyHistoryStatsTimeWindow`, `TestGetMyHistoryStatsNotConfigured`).
- Add `UserHistoryStats` schema and 2 new viewer paths to OpenAPI contract; bump `info.version` to `0.74.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractViewerHistoryStatsPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.75.0 - 2026-06-18

- Add `GlobalPlayEventFilter{UserID, TrackID, Since, Until, Limit, Offset}` to `history/types.go`.
- Add `ListAllPlayEvents(ctx, GlobalPlayEventFilter)` to the `Repository` interface.
- Implement `ListAllPlayEvents` on `history.MemoryRepository` (in-memory multi-filter + sort + slice) and `historypg.Repository` (dynamic `WHERE` clause construction with `LIMIT`/`OFFSET` + `COUNT(*) OVER()`).
- Add `GetAllHistory` to `history.Service`; limit clamped to 500, default 50.
- Add admin route `GET /api/v1/admin/history` (paginated global event list, optional `?userId`, `?trackId`, `?since`, `?until`, `?limit`, `?offset` filters); handler `getAdminAllHistory` reuses `parseHistoryAdminFilter` and `parseHistoryAdminPagination`.
- Add 3 `history.Service` unit tests (`TestGetAllHistory`, `TestGetAllHistoryUserFilter`, `TestGetAllHistoryTimeWindow`).
- Add 4 HTTP-layer tests (`TestAdminGetAllHistory`, `TestAdminGetAllHistoryTrackFilter`, `TestAdminGetAllHistoryNotConfigured`, `TestAdminGetAllHistoryMethodNotAllowed`).
- Add `get` operation to `/api/v1/admin/history` in OpenAPI contract with 6 query params and `PlayEventList` response schema ref; bump `info.version` to `0.75.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history`; add `TestStorageAdminOpenAPIContractAdminHistoryGlobalList`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.76.0 - 2026-06-18

- Add `Asc bool` field to `PlayEventFilter`, `AdminPlayEventFilter`, and `GlobalPlayEventFilter` in `history/types.go`; `false` (default) → `played_at DESC`, `true` → `played_at ASC`.
- Update sort comparators in `history.MemoryRepository` for `ListPlayEvents`, `ListPlayEventsByTrack`, and `ListAllPlayEvents` to respect `f.Asc`.
- Add `eventOrder(asc bool) string` helper to `historypg.Repository`; replace hard-coded `ORDER BY played_at DESC, id DESC` with `eventOrder(f.Asc)` in `ListPlayEvents`, `ListPlayEventsByTrack`, and `ListAllPlayEvents`.
- Add `parseHistoryOrder` helper to `httpapi/handler.go`; parses `?order=asc|desc` (default `desc`); returns `400 invalid_order` for any other value.
- Thread `Asc` through `listPlayEvents`, `getAdminUserHistory`, `getAdminTrackHistory`, and `getAdminAllHistory` handlers.
- Add 2 `history.Service` unit tests (`TestListPlaysAscOrder`, `TestGetAllHistoryAscOrder`).
- Add 4 HTTP-layer tests (`TestListPlayEventsAscOrder`, `TestListPlayEventsInvalidOrder`, `TestAdminGetAllHistoryAscOrder`, `TestAdminGetAllHistoryInvalidOrder`).
- Add `order` query param (string enum `["asc","desc"]`, optional, default `"desc"`) to `GET /api/v1/me/history`, `GET /api/v1/admin/history/users/{userId}`, `GET /api/v1/admin/history/tracks/{trackId}`, and `GET /api/v1/admin/history` in OpenAPI contract; add `invalid_order` to error code enum; bump `info.version` to `0.76.0`.
- Add `TestStorageAdminOpenAPIContractHistoryOrderParam` asserting `order` param on all four paths and `invalid_order` in the error enum.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.77.0 - 2026-06-18

- Add `ErrEventNotFound` and `ErrEventForbidden` sentinel errors to `history/types.go`.
- Add `GetPlayEventByID(ctx, id)` and `DeletePlayEventByID(ctx, id)` to the `Repository` interface.
- Implement both on `history.MemoryRepository` (map lookup; `ErrEventNotFound` on miss) and `historypg.Repository` (`SELECT`/`DELETE` by primary key; `ErrEventNotFound` on `ErrNoRows` or zero `RowsAffected`).
- Add 4 `history.Service` methods: `GetEventByID` (admin), `DeleteEventByID` (admin), `GetMyEvent` (viewer, ownership-checked), `DeleteMyEvent` (viewer, ownership-checked).
- Add admin routes `GET /api/v1/admin/history/{eventId}` and `DELETE /api/v1/admin/history/{eventId}`.
- Add viewer routes `GET /api/v1/me/history/{eventId}` and `DELETE /api/v1/me/history/{eventId}`; ownership check returns `403 event_forbidden` when the authenticated user does not own the event.
- Map `ErrEventNotFound` → `404 not_found` and `ErrEventForbidden` → `403 event_forbidden` in `writeError`.
- Add 5 `history.Service` unit tests (`TestGetEventByID`, `TestGetEventByIDNotFound`, `TestDeleteEventByID`, `TestGetMyEvent`, `TestDeleteMyEvent`).
- Add 7 HTTP-layer tests (`TestAdminGetEvent`, `TestAdminGetEventNotFound`, `TestAdminDeleteEvent`, `TestViewerGetEvent`, `TestViewerGetEventNotOwned`, `TestViewerDeleteEvent`, `TestPerEventHistoryNotConfigured`).
- Add `GET`/`DELETE` operations to `/api/v1/admin/history/{eventId}` and `/api/v1/me/history/{eventId}` in OpenAPI; add `event_forbidden` to error code enum; `PlayEvent` schema ref as 200 response; bump `info.version` to `0.77.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both new paths; add `TestStorageAdminOpenAPIContractPerEventPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.78.0 - 2026-06-19

- Add `UpdatePlayEventByID(ctx, id, playedAt)` to the `Repository` interface.
- Implement on `history.MemoryRepository` (lock + map update; `ErrEventNotFound` on miss) and `historypg.Repository` (`UPDATE … SET played_at = $2 … RETURNING …`; `ErrEventNotFound` on `ErrNoRows`).
- Add `UpdateEventByID(ctx, id, playedAt)` to `history.Service` (admin; validates non-zero `playedAt`).
- Add `UpdateMyEvent(ctx, userID, id, playedAt)` to `history.Service` (viewer; ownership-checked; returns `ErrEventForbidden` for non-owners).
- Add admin route `PATCH /api/v1/admin/history/{eventId}` → `patchAdminEvent`; decodes `{"playedAt": "<RFC3339>"}` request body; returns `400 invalid_played_at` for missing or unparseable timestamp.
- Add viewer route `PATCH /api/v1/me/history/{eventId}` → `patchMyEvent`; same validation; returns `403 event_forbidden` for non-owners.
- Add `UpdatePlayEventRequest` schema (`{playedAt: string/date-time}`) to OpenAPI components.
- Add `patch` operation to `/api/v1/admin/history/{eventId}` and `/api/v1/me/history/{eventId}` in OpenAPI contract; 200 response refs `PlayEvent`; `requestBody` refs `UpdatePlayEventRequest`; add `invalid_played_at` to error code enum; bump `info.version` to `0.78.0`.
- Add 3 `history.Service` unit tests (`TestUpdateEventByID`, `TestUpdateEventByIDNotFound`, `TestUpdateMyEvent`).
- Add 7 HTTP-layer tests (`TestAdminPatchEvent`, `TestAdminPatchEventNotFound`, `TestAdminPatchEventInvalidPlayedAt`, `TestViewerPatchEvent`, `TestViewerPatchEventInvalidPlayedAt`, `TestViewerPatchEventMissingPlayedAt`, `TestPatchEventHistoryNotConfigured`).
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `patch` on both paths; extend `TestStorageAdminOpenAPIContractPerEventPaths` to assert `UpdatePlayEventRequest` requestBody ref and `invalid_played_at` error code.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.79.0 - 2026-06-19

- Add `DeletePlayEventsByIDs(ctx, ids)` and `DeletePlayEventsByIDsForUser(ctx, userID, ids)` to the `Repository` interface.
- Implement both on `history.MemoryRepository` (lock + range-delete; unknown IDs silently skipped) and `historypg.Repository` (`DELETE … WHERE id = ANY($1)` and `DELETE … WHERE id = ANY($1) AND user_id = $2`).
- Add `MaxBatchDeleteIDs = 100` constant; add `BatchDeleteEvents(ctx, ids)` (admin) and `BatchDeleteMyEvents(ctx, userID, ids)` (viewer) to `history.Service`; validate non-empty ids and size ≤ 100.
- Add admin route `POST /api/v1/admin/history/batch-delete` → `batchDeleteAdminEvents`; decodes `{"ids":[…]}`; returns `{"deleted": N}`.
- Add viewer route `POST /api/v1/me/history/batch-delete` → `batchDeleteMyEvents`; same shape; silently skips IDs not owned by the viewer.
- Both routes return `400 invalid_ids` for empty or oversized `ids` array.
- Add `BatchDeleteRequest` schema (`{ids: string[], minItems:1, maxItems:100}`) and `BatchDeleteResult` schema (`{deleted: integer}`) to OpenAPI components; add both new paths; add `invalid_ids` to error code enum; bump `info.version` to `0.79.0`.
- Add 4 `history.Service` unit tests (`TestBatchDeleteEvents`, `TestBatchDeleteEventsUnknownIDsIgnored`, `TestBatchDeleteMyEvents`, `TestBatchDeleteEventsEmpty`).
- Add 5 HTTP-layer tests (`TestAdminBatchDeleteEvents`, `TestAdminBatchDeleteEventsEmptyBody`, `TestViewerBatchDeleteMyEvents`, `TestViewerBatchDeleteSkipsOtherUsersEvents`, `TestBatchDeleteHistoryNotConfigured`).
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both batch-delete paths; add `TestStorageAdminOpenAPIContractBatchDelete`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.80.0 - 2026-06-19

- Add `Since time.Time` and `Until time.Time` fields to `PlayEventFilter` and `AdminPlayEventFilter` in `history/types.go`.
- Update `history.MemoryRepository.ListPlayEvents` and `ListPlayEventsByTrack` to apply `Since`/`Until` guards matching the existing pattern in `ListAllPlayEvents`.
- Replace two-branch (TrackID/no-TrackID) SQL logic in `historypg.Repository.ListPlayEvents` and `ListPlayEventsByTrack` with a unified dynamic `WHERE` clause builder that handles all combinations of `user_id`, `track_id`, `since`, and `until` in a single query path.
- Thread `Since`/`Until` from `parseHistoryAdminFilter` into `listPlayEvents`, `getAdminUserHistory`, and `getAdminTrackHistory` handlers.
- Add `since` and `until` query params (string/date-time, optional) to `GET /api/v1/me/history`, `GET /api/v1/admin/history/users/{userId}`, and `GET /api/v1/admin/history/tracks/{trackId}` in OpenAPI; bump `info.version` to `0.80.0`.
- Add 3 `history.Service` unit tests (`TestListPlaysSinceFilter`, `TestListPlaysUntilFilter`, `TestGetUserHistorySinceFilter`).
- Add 4 HTTP-layer tests (`TestListPlayEventsSinceFilter`, `TestListPlayEventsUntilFilter`, `TestAdminUserHistorySinceUntilFilter`, `TestAdminTrackHistorySinceFilter`).
- Add `TestStorageAdminOpenAPIContractListSinceUntilParams` asserting `since`/`until` on all three paths.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.81.0 - 2026-06-19

- Add `GetAdminUserStats(ctx, UserStatsFilter)` and `GetAdminUserTopTracks(ctx, UserStatsFilter, limit)` to `history.Service` (admin-facing; delegate to `UserHistoryStats` and `UserTopTracks` on the repository; require non-empty `UserID`).
- Add admin routes `GET /api/v1/admin/history/users/{userId}/stats` → `getAdminUserStats` and `GET /api/v1/admin/history/users/{userId}/top-tracks` → `getAdminUserTopTracks`; reuse `parseHistoryAdminFilter` and `parseHistoryAdminLimit`; respond with the same shapes as their `/me/history/stats` and `/me/history/top-tracks` counterparts.
- Add `methodNotAllowed` fallbacks for `GET /api/v1/admin/history/users/{userId}/stats` and `GET /api/v1/admin/history/users/{userId}/top-tracks`.
- Add 3 `history.Service` unit tests (`TestGetAdminUserStats`, `TestGetAdminUserTopTracks`, `TestGetAdminUserTopTracksTimeWindow`).
- Add 4 HTTP-layer tests (`TestAdminGetUserStats`, `TestAdminGetUserTopTracks`, `TestAdminGetUserStatsNotConfigured`, `TestAdminGetUserTopTracksTimeWindow`).
- Add `get` operation to `/api/v1/admin/history/users/{userId}/stats` and `/api/v1/admin/history/users/{userId}/top-tracks` in OpenAPI contract; both accept optional `?since`, `?until`; top-tracks also accepts `?limit`; stats refs `UserHistoryStats` schema; top-tracks refs `TrackPlayCountList` schema (matching the viewer path); bump `info.version` to `0.81.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both new paths; add `TestStorageAdminOpenAPIContractAdminUserStatsPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.82.0 - 2026-06-19

- Add `TrackHistoryStats{TotalEvents int, UniqueListeners int}` and `TrackStatsFilter{TrackID string, Since time.Time, Until time.Time}` types to `history/types.go`.
- Add `TrackHistoryStats(ctx, TrackStatsFilter)` and `TrackTopListeners(ctx, TrackStatsFilter, limit)` methods to the `Repository` interface.
- Implement both on `history.MemoryRepository` (track-scoped in-memory filter + sort) and `historypg.Repository` (new `trackStatsWhere` helper that mandates `track_id = $1`; `TrackTopListeners` returns `UserPlayCount` rows ordered by `play_count DESC, user_id ASC`).
- Add `GetTrackStats(ctx, TrackStatsFilter)` and `GetTrackTopListeners(ctx, TrackStatsFilter, limit)` to `history.Service` (admin; validate non-empty `TrackID`; limit clamp 1–100 default 10).
- Add admin routes `GET /api/v1/admin/history/tracks/{trackId}/stats` → `getAdminTrackStats` and `GET /api/v1/admin/history/tracks/{trackId}/top-listeners` → `getAdminTrackTopListeners`; reuse `parseHistoryAdminFilter` and `parseHistoryAdminLimit`; extract `trackId` from path.
- Add `methodNotAllowed` fallbacks for both new sub-paths.
- Add 3 `history.Service` unit tests (`TestGetTrackStats`, `TestGetTrackTopListeners`, `TestGetTrackTopListenersTimeWindow`).
- Add 4 HTTP-layer tests (`TestAdminGetTrackStats`, `TestAdminGetTrackTopListeners`, `TestAdminGetTrackTopListenersTimeWindow`, `TestAdminGetTrackStatsNotConfigured`).
- Add `TrackHistoryStats` schema to OpenAPI components; add `get` operation to `/api/v1/admin/history/tracks/{trackId}/stats` (refs `TrackHistoryStats`) and `/api/v1/admin/history/tracks/{trackId}/top-listeners` (refs `TopUsersResult`); both accept optional `?since`, `?until`; top-listeners also accepts `?limit`; bump `info.version` to `0.82.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with both new paths; add `TestStorageAdminOpenAPIContractAdminTrackStatsPaths`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.83.0 - 2026-06-19

- Add `TimelineGranularity` string type with constants `GranularityDay`, `GranularityWeek`, `GranularityMonth` to `history/types.go`.
- Add `TimelineFilter{Since time.Time, Until time.Time, Granularity TimelineGranularity, UserID string, TrackID string}` and `TimelineBucket{BucketStart time.Time (json:"bucketStart"), EventCount int (json:"eventCount")}` to `history/types.go`.
- Add `HistoryTimeline(ctx, TimelineFilter) ([]TimelineBucket, error)` to the `Repository` interface.
- Implement on `history.MemoryRepository`: iterate events, apply `UserID`/`TrackID`/`Since`/`Until` guards, truncate `played_at` to the bucket boundary (day=UTC day, week=Monday-anchored UTC week, month=UTC month), accumulate counts into a `map[time.Time]int`, then emit sorted `[]TimelineBucket` (empty bucket list is `[]TimelineBucket{}`).
- Implement on `historypg.Repository`: use `DATE_TRUNC($granularity, played_at AT TIME ZONE 'UTC')` in a dynamic `WHERE` clause built from `timelineWhere` helper (extends `statsWhere` with optional `user_id` and `track_id`); `GROUP BY bucket` order by `bucket ASC`; return `[]TimelineBucket` (empty → `[]TimelineBucket{}`).
- Add `GetHistoryTimeline(ctx, TimelineFilter)` to `history.Service`; validate `Since` and `Until` are both non-zero and `Since` < `Until`; validate `Granularity` is one of `day`/`week`/`month` (default `day`); return `ErrInvalidTimeRange` sentinel on bad range.
- Add `ErrInvalidTimeRange` sentinel to `history/types.go`.
- Add admin route `GET /api/v1/admin/history/timeline` → `getAdminHistoryTimeline`; parses `?since`, `?until` (both required, `400 missing_time_bounds` if absent), `?granularity` (optional, default `day`, `400 invalid_granularity` for other values), optional `?userId` and `?trackId`; returns `{"buckets":[{"bucketStart":"...","eventCount":N},...]}`; add `methodNotAllowed` fallback.
- Add 4 `history.Service` unit tests (`TestGetHistoryTimelineDay`, `TestGetHistoryTimelineWeek`, `TestGetHistoryTimelineUserFilter`, `TestGetHistoryTimelineInvalidRange`).
- Add 4 HTTP-layer tests (`TestAdminGetHistoryTimeline`, `TestAdminGetHistoryTimelineMissingSince`, `TestAdminGetHistoryTimelineInvalidGranularity`, `TestAdminGetHistoryTimelineNotConfigured`).
- Add `TimelineBucket` schema and `TimelineResult` schema (`{buckets: [TimelineBucket]}`) to OpenAPI components; add `get` operation to `/api/v1/admin/history/timeline` with `since`(required), `until`(required), `granularity`(enum day/week/month, default day), `userId`(optional), `trackId`(optional) params; 200 refs `TimelineResult`; add `missing_time_bounds` and `invalid_granularity` to error code enum; bump `info.version` to `0.83.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/timeline`; add `TestStorageAdminOpenAPIContractHistoryTimeline`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.84.0 - 2026-06-19

- Add `GetMyTimeline(ctx, TimelineFilter)` to `history.Service` (viewer-facing; validate non-empty `UserID` in the filter, then delegate to `repo.HistoryTimeline`; reuse the same `ErrInvalidTimeRange` validation as `GetHistoryTimeline`).
- Add viewer route `GET /api/v1/me/history/timeline` → `getMyHistoryTimeline`; parses `?since`, `?until` (both required, `400 missing_time_bounds`), `?granularity` (optional, default `day`, `400 invalid_granularity`), optional `?trackId`; injects `UserID` from auth context; returns `{"buckets":[...]}`.
- Add 3 `history.Service` unit tests (`TestGetMyTimelineDay`, `TestGetMyTimelineTrackFilter`, `TestGetMyTimelineInvalidRange`).
- Add 4 HTTP-layer tests (`TestViewerGetHistoryTimeline`, `TestViewerGetHistoryTimelineMissingSince`, `TestViewerGetHistoryTimelineInvalidGranularity`, `TestViewerGetHistoryTimelineNotConfigured`).
- Add `get` operation to `/api/v1/me/history/timeline` in OpenAPI contract; `since`(required), `until`(required), `granularity`(enum day/week/month, default day), `trackId`(optional) params; 200 refs `TimelineResult`; bump `info.version` to `0.84.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/timeline`; add `TestStorageAdminOpenAPIContractViewerHistoryTimeline`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.85.0 - 2026-06-19

- Add `GetUser(ctx, id)` to `auth.Service` (delegates to `users.GetUser`; wraps result with `toView`).
- Add `getMe` handler: reads the authenticated user from `userFromContext` and writes `UserView` at `GET /api/v1/me`.
- Add `GET /api/v1/me` route (viewer-auth); add `/api/v1/me` methodNotAllowed catch-all.
- Add 2 `auth.Service` unit tests (`TestGetUser`, `TestGetUser_NotFound`).
- Add 3 HTTP-layer tests (`TestGetMe`, `TestGetMeUnauthenticated`, `TestGetMeNotConfigured`).
- Add `GET /api/v1/me` to OpenAPI contract; 200 refs `UserView`; bump `info.version` to `0.85.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me`; add `TestStorageAdminOpenAPIContractGetMe`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.86.0 - 2026-06-19

- Add `getAdminUser` handler: reads path value `id`, calls `authService.GetUser(ctx, id)`, writes `UserView` at `GET /api/v1/admin/users/{id}`.
- Register `GET /api/v1/admin/users/{id}` route (admin-auth); add `/api/v1/admin/users/{id}` to `TestStorageAdminOpenAPIContractCoversRoutes` expected map alongside `admin/users` and `admin/users/{id}/disable`.
- Add 3 HTTP-layer tests (`TestAdminGetUser`, `TestAdminGetUserNotFound`, `TestAdminGetUserNotConfigured`).
- Add `get` operation to `/api/v1/admin/users/{id}` in OpenAPI contract; 200 refs `UserView`, 404 ErrorEnvelope; bump `info.version` to `0.86.0`.
- Add `TestStorageAdminOpenAPIContractAdminGetUser`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.87.0 - 2026-06-19

- Add `ChangePassword(ctx, userID, currentPassword, newPassword)` to `auth.Service`: verifies current password via `CheckPassword`, enforces 8-character minimum on new password, hashes and saves the updated credential.
- Add `changePassword` handler: `POST /api/v1/me/change-password`; decodes `{currentPassword, newPassword}`; returns `204 No Content` on success; `400 invalid_user` for weak/missing new password, `401 unauthorized` for wrong current password.
- Register `POST /api/v1/me/change-password` (viewer-auth); add `/api/v1/me/change-password` methodNotAllowed catch-all.
- Add `ChangePasswordRequest` schema to OpenAPI components.
- Add 3 `auth.Service` unit tests (`TestChangePassword`, `TestChangePassword_WrongCurrent`, `TestChangePassword_WeakNew`).
- Add 4 HTTP-layer tests (`TestChangePassword`, `TestChangePasswordWrongCurrent`, `TestChangePasswordUnauthenticated`, `TestChangePasswordNotConfigured`).
- Add `POST /api/v1/me/change-password` to OpenAPI contract; bump `info.version` to `0.87.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractChangePassword`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.88.0 - 2026-06-19

- Add `EnableUser(ctx, id string) (UserView, error)` to `auth.Service`: mirrors `DisableUser`; sets `Enabled = true`, updates `UpdatedAt`, saves.
- Add `enableUser` handler: `POST /api/v1/admin/users/{id}/enable`; returns `UserView` of the re-enabled user.
- Register `POST /api/v1/admin/users/{id}/enable` (admin-auth); add `/api/v1/admin/users/{id}/enable` methodNotAllowed catch-all.
- Add 2 `auth.Service` unit tests (`TestEnableUser`, `TestEnableUser_NotFound`).
- Add 3 HTTP-layer tests (`TestEnableUser`, `TestEnableUserNotFound`, `TestEnableUserNotConfigured`).
- Add `POST /api/v1/admin/users/{id}/enable` to OpenAPI contract (with `UserId` path parameter); bump `info.version` to `0.88.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractEnableUser`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.89.0 - 2026-06-19

- Add `PatchUser(ctx, id string, role *Role, username *string) (UserView, error)` to `auth.Service`: validates and applies non-nil role/username fields; checks username uniqueness on change; updates `UpdatedAt`; returns `UserView`.
- Add `patchAdminUser` handler: `PATCH /api/v1/admin/users/{id}`; decodes `{role?, username?}`; rejects empty patch (`400 invalid_user`); returns `UserView` on success; propagates `ErrInvalidUser` (400) and `ErrUserConflict` (409).
- Register `PATCH /api/v1/admin/users/{id}` (admin-auth).
- Add `PatchUserRequest` schema to OpenAPI components (`role?: enum[admin,viewer], username?: string`).
- Add 3 `auth.Service` unit tests (`TestPatchUserRole`, `TestPatchUserUsername`, `TestPatchUserConflict`).
- Add 4 HTTP-layer tests (`TestAdminPatchUserRole`, `TestAdminPatchUserUsernameConflict`, `TestAdminPatchUserEmpty`, `TestAdminPatchUserNotConfigured`).
- Add `patch` operation to `/api/v1/admin/users/{id}` in OpenAPI contract; bump `info.version` to `0.89.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `patch` on `/api/v1/admin/users/{id}`; add `TestStorageAdminOpenAPIContractAdminPatchUser`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.90.0 - 2026-06-19

- Add `SessionView{UserID, ExpiresAt, CreatedAt}` to `auth` package as a safe public projection of a session (no token hash).
- Extend `SessionRepository` interface with `ListSessionsByUser(ctx, userID string) ([]Session, error)` and `RevokeAllSessionsByUser(ctx, userID string, revokedAt time.Time) (int, error)`.
- Implement both methods in `authpg.SessionRepository` (SQL: `SELECT … WHERE user_id = $1` and `UPDATE … WHERE user_id = $2 AND revoked_at IS NULL AND expires_at > $1`).
- Implement both methods in the in-memory test stubs (`memSessionRepo` in `service_test.go` and `memAuthSessionRepo` in `handler_test.go`).
- Add `ListActiveSessions(ctx, userID string) ([]SessionView, error)` to `auth.Service`: verifies user exists, calls `ListSessionsByUser`, filters out revoked and expired sessions.
- Add `RevokeAllSessionsForUser(ctx, userID string) (int, error)` to `auth.Service`: verifies user exists, delegates to `RevokeAllSessionsByUser`.
- Add 3 `auth.Service` unit tests: `TestListActiveSessionsEmpty`, `TestListActiveSessionsFiltersRevoked`, `TestListActiveSessionsFiltersExpired`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.91.0 - 2026-06-19

- Add `getAdminUserSessions` handler: `GET /api/v1/admin/users/{id}/sessions`; requires admin auth; calls `auth.Service.ListActiveSessions`; returns `{sessions: [SessionView], count: N}`; propagates `ErrUserNotFound` (404) and auth not configured (503).
- Register `GET /api/v1/admin/users/{id}/sessions` (admin-auth) and its `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/admin/users/{id}/sessions` in OpenAPI contract; add `SessionView` schema to components; bump `info.version` to `0.91.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/users/{id}/sessions`.
- Add 4 HTTP-layer tests: `TestAdminGetUserSessionsEmpty`, `TestAdminGetUserSessionsActive`, `TestAdminGetUserSessionsNotFound`, `TestAdminGetUserSessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.92.0 - 2026-06-19

- Add `deleteAdminUserSessions` handler: `DELETE /api/v1/admin/users/{id}/sessions`; requires admin auth; calls `auth.Service.RevokeAllSessionsForUser`; returns `{"revoked": N}`; propagates `ErrUserNotFound` (404) and auth not configured (503).
- Register `DELETE /api/v1/admin/users/{id}/sessions` (admin-auth).
- Add `delete` operation to `/api/v1/admin/users/{id}/sessions` in OpenAPI contract; bump `info.version` to `0.92.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `delete` on `/api/v1/admin/users/{id}/sessions`.
- Add 3 HTTP-layer tests: `TestAdminDeleteUserSessionsRevokeActive`, `TestAdminDeleteUserSessionsNotFound`, `TestAdminDeleteUserSessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.93.0 - 2026-06-19

- Add `getMyActiveSessions` handler: `GET /api/v1/me/sessions`; requires viewer auth; calls `auth.Service.ListActiveSessions` with the authenticated user's ID; returns `{sessions: [SessionView], count: N}`; 503 when auth not configured.
- Register `GET /api/v1/me/sessions` (viewer-auth) and its `methodNotAllowed` fallback.
- Add `get` operation to `/api/v1/me/sessions` in OpenAPI contract; bump `info.version` to `0.93.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/sessions`.
- Add 3 HTTP-layer tests: `TestViewerGetMySessionsFiltersRevoked`, `TestViewerGetMySessionsActive`, `TestViewerGetMySessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.94.0 - 2026-06-19

- Add `RevokeAllExcept(ctx, userID, exceptTokenHash string) (int, error)` to `auth.Service`: lists all sessions for the user via `ListSessionsByUser`, skips the session matching `exceptTokenHash`, revokes every other active non-expired session via `RevokeSession`; returns revoked count.
- Add `revokeMyOtherSessions` handler: `POST /api/v1/me/sessions/revoke-all`; requires viewer auth; extracts current bearer token hash with `auth.HashToken`; calls `RevokeAllExcept`; returns `{"revoked": N}`; 503 when auth not configured.
- Register `POST /api/v1/me/sessions/revoke-all` (viewer-auth) and its `methodNotAllowed` fallback.
- Add `post` operation to `/api/v1/me/sessions/revoke-all` in OpenAPI contract; bump `info.version` to `0.94.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `post` on `/api/v1/me/sessions/revoke-all`.
- Add 3 HTTP-layer tests: `TestViewerRevokeMyOtherSessions`, `TestViewerRevokeMyOtherSessionsNoneOther`, `TestViewerRevokeMyOtherSessionsNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.95.0 - 2026-06-19

- Add `?limit`, `?offset`, `?sortBy` (username/role/createdAt/updatedAt), and `?sortOrder` (asc/desc) query parameters to `GET /api/v1/admin/users`.
- Sort and paginate over the full `[]UserView` slice in the `listUsers` handler; no new repository interface methods required.
- `limit=0` (absent) returns all users from `offset`; `limit > 0` pages the result; `hasMore` reflects whether more items follow.
- Response is `{"users":[...],"pagination":{"limit":N,"offset":N,"total":N,"hasMore":bool}}`.
- Invalid `sortOrder` returns `400 invalid_sort_order`; invalid `limit` or `offset` returns `400`.
- Add 5 HTTP-layer tests: `TestAdminListUsersPagination`, `TestAdminListUsersSortByUsername`, `TestAdminListUsersSortDesc`, `TestAdminListUsersInvalidSortOrder`, `TestAdminListUsersInvalidLimit`.
- Update `GET /api/v1/admin/users` in OpenAPI contract with `limit`, `offset`, `sortBy`, `sortOrder` query params and updated 200 response schema including `pagination`; bump `info.version` to `0.95.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

### v0.96.0 - 2026-06-19

- Add `?username` (exact match), `?role` (admin/viewer), and `?enabled` (true/false) filter query parameters to `GET /api/v1/admin/users`; filters are applied before sort and pagination.
- Invalid `?role` values return `400 invalid_role`; invalid `?enabled` values return `400 invalid_enabled`.
- Add 5 HTTP-layer tests: `TestAdminListUsersFilterByRole`, `TestAdminListUsersFilterByEnabled`, `TestAdminListUsersFilterByUsername`, `TestAdminListUsersFilterInvalidRole`, `TestAdminListUsersFilterInvalidEnabled`.
- Extend `GET /api/v1/admin/users` in OpenAPI contract with `username`, `role`, and `enabled` query params; bump `info.version` to `0.96.0`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.
