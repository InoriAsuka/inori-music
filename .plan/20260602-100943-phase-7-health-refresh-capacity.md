# Plan: Phase 7 Health Refresh and Capacity Reporting

## Requirement Version

v0.7.0

## Goals

- Refresh all enabled storage backends in one administrator operation without one failed backend blocking the rest.
- Report filesystem capacity for LocalSystem, NFS, SMB, and mounted-filesystem distributed backends.
- Run optional periodic refresh from the server process when configured.
- Preserve explicit unsupported outcomes for backends without capacity providers.

## Phase 1: Requirement Update

- [x] Append `v0.7.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.7.0`.

## Phase 2: Capacity Domain

- [x] Add capacity report models and provider interface.
- [x] Add filesystem capacity provider using mounted filesystem statistics.
- [x] Return explicit unsupported errors for S3-compatible capacity reporting.
- [x] Add service method for reading and recording backend capacity.

## Phase 3: Batch Refresh and Scheduler

- [x] Add batch refresh that skips disabled backends and isolates per-backend failures.
- [x] Add optional background scheduler configured by duration.
- [x] Stop background refresh when the scheduler context is canceled.
- [x] Add server bootstrap support for `INORI_STORAGE_REFRESH_INTERVAL`.

## Phase 4: Admin HTTP API

- [x] Add `POST /api/v1/admin/storage/backends/refresh`.
- [x] Add `GET /api/v1/admin/storage/backends/{id}/capacity`.
- [x] Protect both endpoints with administrator Bearer Token authentication.
- [x] Map unsupported capacity providers to a stable HTTP error envelope.

## Phase 5: Verification

- [x] Add capacity provider and service tests.
- [x] Add scheduler lifecycle tests.
- [x] Add handler tests for refresh and capacity endpoints.
- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add S3-compatible provider capacity or quota integrations where providers expose reliable APIs.
- [ ] Persist latest health, capacity, and refresh history in PostgreSQL.
- [ ] Add configurable probe timeouts, worker limits, and retry policy.
- [ ] Add refresh audit logging and metrics.
- [x] Add graceful HTTP shutdown handling for interrupt and SIGTERM.

## Completion Notes

Filesystem capacity is derived from the configured mounted path. S3-compatible providers do not expose a uniform bucket-capacity API, so they intentionally return `capacity_unsupported` until provider-specific integrations are designed.
