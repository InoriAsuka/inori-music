# Plan: Phase 14 Batch Media Object Verification

## Requirement Version

v0.14.0

## Goals

- Add batch media object integrity verification for filtered object sets.
- Support `backendId` and `contentHash` filters without introducing broad unbounded scans.
- Continue verification after individual media object failures.
- Expose per-object outcomes through the authenticated admin HTTP API and OpenAPI contract.

## Phase 1: Requirement Update

- [x] Append `v0.14.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.14.0`.

## Phase 2: Domain Implementation

- [x] Add media object verification report type.
- [x] Add service methods for verification by backend ID and content hash.
- [x] Continue after per-object failures and include failure messages in report results.

## Phase 3: HTTP and Contract

- [x] Add `POST /api/v1/admin/media/objects/verify?backendId=...`.
- [x] Add `POST /api/v1/admin/media/objects/verify?contentHash=...`.
- [x] Update OpenAPI paths, schemas, route coverage tests, and docs.
- [x] Add handler tests for success, mixed failures, filter validation, and authentication.

## Phase 4: Validation

- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add pagination/cursors for large verification sets.
- [ ] Add background verification jobs and resumable repair reports.
- [ ] Persist latest verification status per media object.
- [ ] Add S3-compatible batch verification after single-object S3 verification exists.

## Completion Notes

This phase runs synchronous metadata-filtered verification only. It does not upload, delete, move, repair, or mutate media files.
