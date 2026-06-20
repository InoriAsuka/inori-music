# inori-music

A centralized music platform targeting Web, Android, iOS, and desktop clients while supporting
both browser/server and client/server deployment styles.

## Version

Current architecture baseline version: `1.24.0`

## Documentation Policy

Repository Markdown documentation is maintained in English. Historical phase plans, requirements,
ADRs, and architecture documents are also kept in English so implementation records remain
consistent across releases.

## Architecture Direction

- Cross-platform clients: Flutter-first (`packages/app/`).
- Web player: React 19 + Vite 8 + Tailwind 4 + shadcn/ui (`packages/web/`).
- Admin console: React 19 + TanStack Table (`packages/admin/`).
- Server: Go modular monolith, with later service extraction by domain boundaries when justified.
- Server metadata database: PostgreSQL-first.
- Client local persistence: SQLite for offline queues, cache indexes, and local search.
- Search: PostgreSQL full-text search first, with optional external search engines later.
- Media storage: server-managed multi-backend; large media bytes stay outside the relational
  database.

## Completed Phases

### Phase 1: Storage Architecture

Establish server-managed multi-backend media storage covering local, NFS, SMB, S3-compatible,
and distributed backends.

### Phase 2: Storage Domain Scaffold

Create the Go API scaffold and storage domain with validation, capability inference, default
backend handling, and in-memory repositories.

### Phase 3: Storage Admin HTTP API

Expose storage administration through versioned HTTP endpoints for validation, registration,
listing, default selection, and disabling.

### Phase 4: Admin Authentication

Protect administrator routes with bootstrap bearer-token authentication while keeping /healthz
public.

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

### Phase 22–23: Documentation Policy

Restore Markdown documentation to English as the repository documentation policy.

### Phase 24: Media Object List Sorting

Support deterministic media object list sorting before pagination.

### Phase 25: Media Object Duplicate Detection

Add metadata-only duplicate content-hash report for deduplication planning.

### Phase 26: Bulk Media Object Lifecycle Updates

Add metadata-only bulk lifecycle updates by exactly one media-object selection filter.

### Phase 27: Bulk Lifecycle Dry Runs

Add dry-run previews for bulk lifecycle updates.

### Phase 28: Latest Lifecycle Change Metadata

Persist the latest committed lifecycle change source and transition metadata.

### Phase 29: Media Object Metadata Timeline

Expose a read-only per-object metadata timeline derived from registration, verification, and
lifecycle transition metadata.

### Phase 30: CI Release and Container Automation

Add GitHub Actions workflows for Go API validation, tagged releases, and Docker image publishing.

### Phase 31: Runtime Version Metadata

Expose public build metadata via `/versionz`; inject version into release binaries and images.

### Phase 32: Runtime Readiness Diagnostics

Expose public readiness diagnostics via `/readyz`; add Docker liveness healthcheck.

### Phase 33: Runtime Metrics Endpoint

Expose public Prometheus-compatible runtime metrics for readiness gauges and build information.

### Phase 34: HTTP Request Metrics

Add low-cardinality HTTP request counters and duration-sum metrics by method, route, and status.

### Phase 35: PostgreSQL Persistence Layer

Add PostgreSQL-backed repository implementations for storage backends and media objects with
automatic schema migration and shared connection pool.

### Phase 36–37: Auth Domain and Session Login

Add user domain with PostgreSQL persistence; bcrypt password hashing; `auth.Service` with
Login/Logout/ValidateToken/CreateUser/DisableUser; POST /api/v1/auth/login and /logout endpoints;
session-token-based `requireAdminAuth` upgrade.

### Phase 38: Admin User Management API

Expose list, create, disable, and delete user endpoints under `/api/v1/admin/users/`.

### Phase 39–67: Catalog Domain

Artist/Album/Track/Playlist CRUD; PostgreSQL full-text search; viewer catalog browse routes;
batch import; patch metadata; playlist track management; stats and breakdown endpoints; recently
added/updated timelines; track playback descriptor with S3 presigned URL; list pagination and
sorting; nested browse routes; PostgreSQL sort/pagination pushdown and SQL aggregate stats.

### Phase 68–84: Playback History Domain

`PlayEvent` recording and listing; admin aggregate stats (top tracks, top users); time-windowed
filters; admin user/track history detail views; bulk and window deletes; per-event fetch/delete/
patch; batch delete; since/until list filters; viewer stats, top-tracks, timeline, summary;
admin user/track stats and timelines; global history list and summary.

### Phase 85–99: Auth Extensions

Viewer self-profile (`GET /api/v1/me`); admin get-user; change password; enable user; patch user;
user list pagination/sort/filter; force-change-password; delete-user session cascade; revoke all
sessions; revoke all devices.

### Phase 100–119: History Detail Views and OpenAPI Cleanup

Per-track viewer history and stats; admin track/user timelines; viewer track timeline; per-entity
summary endpoints; time-filter coverage for all history aggregate endpoints; PostgreSQL integration
tests; full OpenAPI contract audit (115 operations, 100% coverage).

### Phase 120: Catalog and History Service Wiring

Wire `catalog.Service` and `history.Service` into `main.go`; all catalog and history routes now
respond correctly instead of returning 503. Added `catalogRepository` and `historyRepository`
helpers following the existing PostgreSQL-or-memory pattern.

### Phase 121: Readiness Check Coverage

Extend `/readyz` with `catalog_service` and `history_service` checks; `ready` becomes `false`
when either service is nil. Added `newNoCatalogTestHandler` and `newNoHistoryTestHandler` test
helpers; updated all `NoCatalogService` tests to use the dedicated no-catalog handler.

### Phase 122: CORS Middleware

Add `corsMiddleware` with origin allowlist; OPTIONS preflight returns `204 No Content`;
permissive mode when `INORI_CORS_ORIGINS` is unset; `WithCORSOrigins` handler option.

### Phase 123: Request-ID Middleware

Add `requestIDMiddleware` that reads or generates `X-Request-ID` and echoes it on every
response; injects the ID into the request context; chains outermost in `Routes()`.

### Phase 124: README and Documentation Sync

Bring README to v1.24.0 baseline; enumerate all completed phases; update run command; add
frontend client constraint document reference; update Future Outlook.

---

## Run the API Server

```bash
INORI_ADMIN_TOKEN=change-me \
INORI_INITIAL_ADMIN_USER=admin \
INORI_INITIAL_ADMIN_PASSWORD=changeme123 \
INORI_CORS_ORIGINS=http://localhost:5173,http://localhost:5174 \
INORI_DATABASE_URL=postgres://user:pass@localhost:5432/inori \
go run ./services/api/cmd/server
```

Without `INORI_DATABASE_URL`, all repositories fall back to in-memory mode (no persistence).
Without `INORI_CORS_ORIGINS`, the CORS middleware reflects any origin (permissive dev mode).

The server listens on `127.0.0.1:8080` by default (`INORI_HTTP_ADDR` overrides this).

---

## Project Documents

- [`requirement.md`](requirement.md): versioned requirements and phase history.
- [`.plan/`](.plan/): per-phase plans and completed task checklists.
- [`docs/architecture/`](docs/architecture/): architecture notes.
  - [`docs/architecture/frontend-client-constraints.md`](docs/architecture/frontend-client-constraints.md): tech-stack, design language, and boundary constraints for inori-web, inori-admin, and the Flutter client.
- [`docs/adr/`](docs/adr/): architecture decision records.
- [`docs/operations/release-and-container.md`](docs/operations/release-and-container.md): GitHub Actions release and container publishing notes.
- [`packages/api-contract/openapi/storage-admin.v1.json`](packages/api-contract/openapi/storage-admin.v1.json): OpenAPI 3.1 contract (115 operations, v1.24.0).

---

## Future Outlook

- **inori-web** (`packages/web/`): React 19 / Vite 8 / Tailwind 4 / shadcn/ui music player for
  viewer-role users — catalog browse, playback with queue management, personal history and stats.
- **inori-admin** (`packages/admin/`): React 19 management console for admin-role users — user
  management, catalog CRUD, media object administration, storage backend configuration, history
  analysis.
- **inori-app** (`packages/app/`): Flutter 3.22+ cross-platform client (Android / iOS / macOS /
  Windows) with Riverpod state management, `just_audio` + `audio_service` background playback,
  and Drift SQLite offline cache.
- Shared TypeScript API client (`packages/api-client/`) auto-generated from the OpenAPI spec.
- Shared shadcn/ui component library (`packages/ui/`) consumed by both web products.
