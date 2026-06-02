# Plan: Phase 5 Filesystem Health Probes

## Requirement Version

v0.5.0

## Goals

- Add real but narrowly scoped health probes for filesystem-backed media storage.
- Update server-managed health state after each probe attempt.
- Keep probes safe by creating and deleting only a short-lived application-owned probe file.
- Expose authenticated probe and health-read endpoints for administrator clients.

## Phase 1: Requirement Update

- [x] Append `v0.5.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.5.0`.

## Phase 2: Probe Domain

- [x] Add explicit probe result and probe error models.
- [x] Add filesystem probe root selection for LocalSystem, NFS, SMB, and mounted-filesystem distributed backends.
- [x] Add safe write, full-read, range-read, and cleanup checks using a short-lived probe file.
- [x] Reject disabled backends and unsupported probe adapters explicitly.
- [x] Persist healthy or unhealthy state and the latest check timestamp after probe attempts.

## Phase 3: Admin HTTP API

- [x] Add `POST /api/v1/admin/storage/backends/{id}/probe`.
- [x] Add `GET /api/v1/admin/storage/backends/{id}/health`.
- [x] Protect both endpoints with the existing administrator Bearer Token boundary.
- [x] Map unsupported probes to a stable HTTP error envelope.

## Phase 4: Verification

- [x] Add domain tests for successful probe, cleanup, missing path, unsupported S3, and disabled backend behavior.
- [x] Add handler tests for authenticated probe and health workflows.
- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add S3-compatible object probes with short-lived objects.
- [ ] Add dedicated distributed adapter probes.
- [ ] Add scheduled health refresh jobs and capacity reporting.
- [ ] Add PostgreSQL persistence for probe history and latest health state.
- [ ] Add probe timeout and cancellation policies for slow network mounts.

## Completion Notes

This phase probes mounted filesystem semantics only. NFS and SMB mounts remain host-level operational responsibilities; the server verifies the configured mount path behaves as writable storage but does not mount or unmount remote filesystems itself.
