# inori-music Requirements

## Current Version

`0.23.0`

## Product Goal

Build a cross-platform music playback system for Web, Android, iOS, and desktop clients while supporting both browser/server and client/server architectures. The server owns media storage configuration, metadata registration, health checks, integrity verification, and administrative APIs. Large media bytes are stored in external storage backends rather than the relational database.

## Technical Requirements

- Flutter-first client direction.
- Go modular monolith first for the server.
- PostgreSQL-first server metadata database.
- SQLite for client-side local persistence.
- PostgreSQL full-text search first for 0.x, with external search engines left as future extensions.
- Media storage must support local filesystems, NFS, SMB, S3-compatible object storage, and distributed storage adapters.

## Storage Requirements

- Do not store large audio, image, or derived media files in the relational database.
- Store object IDs, backend IDs, object keys, hashes, lifecycle state, asset kind, verification state, and references as metadata.
- Probes and verification must use server-owned temporary objects or read-only checks to avoid damaging user media.

## Documentation Requirements

- Markdown documentation is maintained in English.
- Phase work must be recorded under `.plan/` with requirements, task checklists, non-goals, and follow-up candidates.
- README, requirements, ADRs, and architecture notes must stay aligned with the current version baseline.

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
