# inori-music

A centralized music platform targeting Web, Android, iOS, and desktop clients while supporting both browser/server and client/server deployment styles.

## Version

Current architecture baseline version: `0.26.0`.

## Documentation Policy

Repository Markdown documentation is maintained in English. Historical phase plans, requirements, ADRs, and architecture documents are also kept in English so implementation records remain consistent across releases.

## 0.x Architecture Direction

- Cross-platform clients: Flutter-first.
- Server: Go modular monolith first, with later service extraction by domain boundaries when justified.
- Server metadata database: PostgreSQL-first.
- Client local persistence: SQLite for offline queues, cache indexes, and local search.
- Search: PostgreSQL full-text search first, with optional external search engines later.
- Media storage: server-managed multi-backend storage; large media bytes stay outside the relational database.

## Completed Phases

### Phase 1: Storage Architecture

Establish server-managed multi-backend media storage covering local, NFS, SMB, S3-compatible, and distributed backends.

### Phase 2: Storage Domain Scaffold

Create the Go API scaffold and storage domain with validation, capability inference, default backend handling, and in-memory repositories.

### Phase 3: Storage Admin HTTP API

Expose storage administration through versioned HTTP endpoints for validation, registration, listing, default selection, and disabling.

### Phase 4: Admin Authentication

Protect administrator routes with bootstrap bearer-token authentication while keeping /healthz public.

### Phase 5: Filesystem Health Probes

Add safe filesystem probes for local, NFS, SMB, and mounted distributed backends.

### Phase 6: S3-Compatible Object Probes

Add conservative S3-compatible object probes with environment-referenced credentials.

### Phase 7: Health Refresh and Capacity Reporting

Add batch refresh, optional background refresh, and filesystem capacity reporting.

### Phase 8: OpenAPI Contract

Publish and test the OpenAPI 3.1 contract for the admin API.

### Phase 9: Durable File Repository

Add optional JSON file-backed persistence for storage backend state.

### Phase 10: Media Object Registry Scaffold

Add the media object registry domain for metadata-only binary asset references.

### Phase 11: Media Object Admin HTTP API

Expose authenticated media object registration, fetch, and filter endpoints.

### Phase 12: Durable Media Object Repository

Add optional JSON file-backed persistence for media object metadata.

### Phase 13: Media Object Integrity Verification

Add read-only filesystem integrity verification for media object references.

### Phase 14: Batch Media Object Verification

Add batch media object verification by backend ID or content hash.

### Phase 15: Latest Verification State

Persist the latest media object verification result in metadata.

### Phase 16: Verification Status Filter

Support filtering media objects by latest verification status.

### Phase 17: Media Object List Pagination

Add limit/offset pagination and pagination metadata to media object lists.

### Phase 18: Media Object Metadata Statistics

Add metadata-only media object statistics for dashboard-style summaries.

### Phase 19: Media Object Lifecycle Administration

Add metadata-only media object lifecycle updates with terminal deleted semantics.

### Phase 20: Media Object Lifecycle Filter

Support filtering media object lists by lifecycle state.

### Phase 21: Media Object Asset Kind Filter

Support filtering media object lists by asset kind.

### Phase 22: Chinese Documentation Split

Split README content and localize documentation in the previous phase.

### Phase 23: English Documentation Policy

Restore Markdown documentation to English as the repository documentation policy.

### Phase 24: Media Object List Sorting

Support deterministic media object list sorting by backend/object key, created time, updated time, size, object key, or ID before pagination.

### Phase 25: Media Object Duplicate Detection

Add a metadata-only duplicate content-hash report for admin deduplication and storage cleanup planning.

### Phase 26: Bulk Media Object Lifecycle Updates

Add metadata-only bulk lifecycle updates by exactly one media-object selection filter.

## Run the API Scaffold

```bash
INORI_ADMIN_TOKEN=change-me-development-token INORI_STORAGE_REPOSITORY_FILE=./var/storage-backends.json INORI_MEDIA_OBJECT_REPOSITORY_FILE=./var/media-objects.json INORI_STORAGE_REFRESH_INTERVAL=15m go run ./services/api/cmd/server
```

The server listens on `127.0.0.1:8080` by default. Admin endpoints require `Authorization: Bearer <INORI_ADMIN_TOKEN>`. If repository file environment variables are omitted, the API uses in-memory development repositories.

## Project Documents

- [`requirement.md`](requirement.md): versioned requirements and history.
- [`.plan/`](.plan/): phase plans and completed task checklists.
- [`docs/architecture/`](docs/architecture/): architecture notes.
- [`docs/adr/`](docs/adr/): architecture decision records.
- [`packages/api-contract/openapi/storage-admin.v1.json`](packages/api-contract/openapi/storage-admin.v1.json): OpenAPI 3.1 admin API contract.

## Future Outlook

- Introduce PostgreSQL migrations and indexes to replace development JSON repositories.
- Add import jobs, audit events, bulk lifecycle updates, and admin UI workflows.
- Expand player-side streaming, cache management, offline queues, and cross-device sync.
