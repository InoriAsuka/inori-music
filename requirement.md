# Requirement

## Project

Inori Music is a full-platform centralized music playback system that supports browser/server and client/server usage models across Web, Android, iOS, and desktop platforms.

## Current Scope

The current 0.x line focuses on architecture bootstrapping, requirements traceability, and the first runnable service foundations. The first phase defines the server-managed media storage architecture before feature implementation begins.

The phase-2 development scope starts the Go API service scaffold and implements the storage administration domain as testable server-side code.

The phase-3 development scope exposes the storage administration domain through a versioned HTTP JSON API for server-side backend management.

The phase-4 development scope protects storage administration routes with explicit administrator Bearer Token authentication while keeping health checks public.

The phase-5 development scope adds safe, real filesystem probe checks and persists the latest backend health state in the server-managed repository.

The phase-6 development scope adds safe, real S3-compatible object probes using short-lived server-owned probe objects and secret references resolved from environment variables.

The phase-7 development scope adds batch health refresh, filesystem capacity reporting, and an optional background refresh scheduler.

The phase-8 development scope adds a versioned OpenAPI 3.1 contract for the storage administration HTTP API and contract tests that keep the route surface documented.

The phase-9 development scope adds an optional durable file-backed repository for storage backend configuration so development and self-hosted servers can retain backend state across restarts before PostgreSQL persistence lands.

The phase-10 development scope adds a media object registry scaffold that validates and records binary asset references against enabled storage backends without storing media bytes in the API service.

The phase-11 development scope exposes authenticated media object registry HTTP endpoints so administrator and import clients can register, fetch, and filter media object metadata through the API scaffold.

The phase-12 development scope adds optional durable file-backed persistence for media object metadata so development and self-hosted API servers can retain media object references across restarts before PostgreSQL persistence lands.

The phase-13 development scope adds media object integrity verification for metadata references, beginning with read-only filesystem verification for LocalSystem, NFS, SMB, and mounted-filesystem distributed backends.

The phase-14 development scope adds batch media object integrity verification by backend ID or content hash so administrators and import workflows can validate groups of metadata references while continuing after individual object failures.

The phase-15 development scope persists the latest media object verification result in media object metadata so operators can inspect recent integrity state after single or batch verification runs.

## Storage Requirements

### Media Storage Scope

The system must store and manage media-related binary assets through a stable storage abstraction instead of binding business logic to a single storage vendor or protocol.

Media assets include:

- Original audio files.
- Transcoded audio files.
- Album artwork and artist images.
- Lyrics files.
- Waveform and audio-analysis artifacts.
- Import packages and backup artifacts.

### Supported Backend Families

The server backend must manage media storage configurations and support multiple backend families:

- `local`: local filesystem storage for development, single-node self-hosting, and simple NAS-mounted paths.
- `nfs`: NFS-mounted filesystem storage for LAN and homelab deployments.
- `smb`: SMB/CIFS-mounted filesystem storage for Windows shares, NAS appliances, and mixed-platform networks.
- `s3`: S3-compatible object storage for cloud deployments and object-storage products.
- `distributed`: distributed storage backends such as Ceph, Garage, SeaweedFS, or similar systems exposed through either filesystem mounts, S3-compatible APIs, or dedicated adapters.

### Server-Managed Configuration

The server backend is the source of truth for storage backend configuration. It must provide administrative capabilities to:

- Register storage backends.
- Enable or disable storage backends.
- Validate backend connectivity and permissions.
- Select a default backend for new media assets.
- Track backend health and capacity metadata.
- Prevent accidental deletion of referenced media assets.
- Support migration planning between backends.

### Storage Safety Requirements

The system must not store large audio files directly inside the relational database. The database stores metadata, references, checksums, ownership, storage keys, and lifecycle state.

The storage abstraction must support:

- Object identity and immutable content references.
- Range reads for streaming playback.
- Atomic metadata updates at the application level.
- Content hashing and deduplication planning.
- Backend capability detection.
- Pluggable URL generation, including direct server streaming and presigned URLs where supported.

## Database Requirements

The server-side primary database for 0.x should be PostgreSQL-first. The client-side local database should be SQLite for offline queues, cache indexes, and local search. Multi-primary-database compatibility is not a 0.x delivery requirement, but server code should avoid unnecessary coupling between domain logic and SQL implementation details.

## Search Requirements

The 0.x server-side search should begin with PostgreSQL full-text search, normalized fields, aliases, and ranking rules. External search engines are optional future integrations when scale, language quality, or typo-tolerance requirements exceed PostgreSQL capabilities.

## Requirement History

### v0.16.0 - 2026-06-04

- Required `GET /api/v1/admin/media/objects` to support exactly one of `backendId`, `contentHash`, or `verificationStatus`.
- Required `verificationStatus` to accept `verified`, `failed`, and `unknown`; `unknown` means no `lastVerification` result is present.
- Required filtering to use persisted media object metadata only and never read media bytes.
- Required memory and file-backed media object repositories to provide stable object-key ordering for status-filtered results.
- Required HTTP, repository, OpenAPI, and validation tests for valid and invalid filter combinations.

### v0.15.0 - 2026-06-04

- Required media object metadata to retain the latest verification result after single-object verification.
- Required batch verification to persist each object's latest verification result while still continuing after failures.
- Required persisted verification metadata to include status, verification time, content hash, size, and failure message when present.
- Required file-backed media object repositories to preserve latest verification metadata across restarts.
- Required tests for latest verification persistence on success, failure, batch verification, and file repository reopening.

### v0.14.0 - 2026-06-03

- Required authenticated batch media object verification by `backendId` or `contentHash` filters.
- Required batch verification to continue after individual object failures and return per-object outcomes.
- Required batch verification to use the same read-only verification semantics as single-object verification.
- Required the HTTP API and OpenAPI contract to document the batch verification route, response schema, and filter rules.
- Required tests for successful mixed batch results, filter validation, unsupported object outcomes, and route authentication.

### v0.13.0 - 2026-06-03

- Required authenticated media object integrity verification by media object ID.
- Required read-only verification for filesystem-backed media objects by checking existence, regular-file shape, byte size, and `sha256` content hash.
- Required verification to reject absolute or escaping object paths and to avoid mutating media bytes.
- Required explicit unsupported responses for non-filesystem or unsupported hash algorithms.
- Required OpenAPI and handler tests for successful verification, hash mismatch, disabled backend rejection, unsupported S3 verification, and missing media objects.

### v0.12.0 - 2026-06-03

- Required an optional durable repository implementation for media object metadata without adding external database dependencies.
- Required `INORI_MEDIA_OBJECT_REPOSITORY_FILE` to switch the API server from in-memory media object storage to an atomic JSON file repository.
- Required persisted media object state to retain server-owned timestamps and metadata references across API restarts.
- Required file writes to create parent directories and use temp-file, sync, close, and atomic rename semantics.
- Required tests for media object persistence across repository reopening, stable filtered listings, malformed repository files, unsupported schema versions, and server repository selection.

### v0.11.0 - 2026-06-03

- Required authenticated HTTP endpoints for media object registration, lookup by ID, listing by backend ID, and listing by content hash.
- Required strict JSON decoding for media object registration with server-owned timestamps excluded from request bodies.
- Required media object API errors to use a distinct `invalid_media_object` code for invalid metadata.
- Required the OpenAPI contract to document media object endpoints, schemas, query filters, path parameters, and error code additions.
- Required handler tests for registration, lookup, backend/content-hash filtering, invalid object input, disabled backend rejection, and authentication.

### v0.10.0 - 2026-06-03

- Required a media object registry scaffold for binary asset references stored outside the relational database.
- Required media object validation for IDs, enabled backend references, relative object keys, content hashes, non-negative sizes, MIME types, asset kinds, and lifecycle states.
- Required an in-memory media object repository for early domain tests before PostgreSQL media metadata persistence is introduced.
- Required service-level registration to reject disabled or missing storage backends.
- Required tests for successful registration, invalid object keys, disabled backend rejection, stable backend listing, and content-hash lookup.

### v0.9.0 - 2026-06-03

- Required an optional durable repository implementation for storage backend configuration without adding external database dependencies.
- Required `INORI_STORAGE_REPOSITORY_FILE` to switch the API server from in-memory storage to an atomic JSON file repository.
- Required persisted backend state to include server-owned health and capacity metadata so probe and refresh results survive process restarts.
- Required repository writes to be atomic at the file level and create parent directories when needed.
- Required tests for persistence across repository reopening, default-backend clearing, malformed repository files, and server repository selection.

### v0.8.0 - 2026-06-03

- Required a versioned OpenAPI 3.1 contract for the storage administration HTTP API.
- Required OpenAPI documentation for health, backend lifecycle, validation, refresh, default, disable, probe, health-state, and capacity endpoints.
- Required a Bearer authentication security scheme in the OpenAPI contract for `/api/v1/admin/*` routes while leaving `/healthz` public.
- Required schema coverage for backend configuration families, capabilities, probe results, capacity reports, refresh reports, and error envelopes.
- Required contract tests to verify every implemented storage admin route is represented in the OpenAPI document with the expected method and security posture.

### v0.7.0 - 2026-06-02

- Required authenticated batch refresh for all enabled storage backends.
- Required batch refresh to continue after individual backend failures and return per-backend outcomes.
- Required filesystem capacity reports for LocalSystem, NFS, SMB, and mounted-filesystem distributed backends.
- Required explicit unsupported responses for capacity providers that are not implemented yet, including S3-compatible backends.
- Required an optional background refresh scheduler configured through `INORI_STORAGE_REFRESH_INTERVAL`.
- Required scheduler shutdown when the server context is canceled.
- Required tests for batch refresh isolation, disabled backend skipping, filesystem capacity reporting, unsupported capacity, and scheduler lifecycle.

### v0.6.0 - 2026-06-02

- Required safe real S3-compatible object probes for `s3` and distributed `s3-compatible` storage backends.
- Required S3 probes to resolve credentials from `accessKeySecretRef` and `secretKeySecretRef` environment variable names without logging secret values.
- Required S3 probes to put, full-read, range-read, and delete only a short-lived server-owned probe object.
- Required S3 probe cleanup to run even when read or validation steps fail after object creation.
- Required S3 probe tests using a local fake S3-compatible HTTP server instead of external cloud services.
- Required probe docs to distinguish supported filesystem and S3-compatible probes from future dedicated distributed adapters.

### v0.5.0 - 2026-06-02

- Required safe real probe checks for LocalSystem, NFS, SMB, and mounted-filesystem distributed storage backends.
- Required probes to create, write, read, range-read, and remove only a short-lived server-owned probe file inside the configured storage root.
- Required probe failures to update backend health state without deleting or modifying unrelated user media.
- Required disabled backends and unsupported probe adapters to return explicit domain errors.
- Required authenticated HTTP operations to probe a registered backend and inspect its latest health state.
- Required probe tests for successful local filesystems, cleanup, missing directories, unsupported S3 probes, disabled backends, and HTTP health workflows.

### v0.4.0 - 2026-06-02

- Required administrator authentication for all `/api/v1/admin/*` routes before further storage management work.
- Required `/healthz` to remain public for local process and container health checks.
- Required `INORI_ADMIN_TOKEN` as the initial bootstrap credential source for the Go API server.
- Required admin routes to fail closed with `503 admin_auth_not_configured` when no token is configured.
- Required malformed, missing, and invalid credentials to use stable JSON error envelopes.
- Required constant-time token comparison and tests for public health, missing auth, invalid auth, configured auth, and disabled auth states.

### v0.3.0 - 2026-06-02

- Required a runnable HTTP server with a health endpoint and versioned storage administration routes.
- Required HTTP JSON operations to validate, register, list, disable, and select the default storage backend.
- Required consistent JSON error envelopes and HTTP status mapping for invalid input, missing resources, conflicts, and unsupported methods.
- Required strict JSON decoding with unknown-field rejection and bounded request bodies.
- Required JSON field names for storage domain resources and omission of absent backend-family configuration branches.
- Required backend configuration to contain exactly one family-specific configuration branch.
- Required HTTP handler tests for successful workflows and error responses.

### v0.2.0 - 2026-06-02

- Started the server API scaffold for the storage administration domain.
- Required a typed storage domain model for backend families, capability detection, health status, and media object references.
- Required a validation service that rejects invalid backend configuration before a backend can be registered or selected as default.
- Required an in-memory repository implementation for early domain tests before PostgreSQL migrations are introduced.
- Required unit tests for local, NFS, SMB, S3-compatible, distributed backend validation, default backend selection, and unsafe configuration rejection.

### v0.1.0 - 2026-06-02

- Established the first-phase architecture scope for server-managed media storage.
- Defined storage backend families: LocalSystem, NFS, SMB, S3-compatible, and distributed storage.
- Required the server backend to own storage configuration, validation, health tracking, and default-backend selection.
- Confirmed PostgreSQL-first server metadata storage and SQLite client-side local storage direction.
- Confirmed PostgreSQL-first search for 0.x with future external search extensibility.
