# inori-music Requirements

## Current Version

`0.53.0`

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
