# inori-music Requirements

## Current Version

`0.41.0`

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
