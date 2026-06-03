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
