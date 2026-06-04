# Plan: Phase 13 Media Object Integrity Verification

## Requirement Version

v0.13.0

## Goals

- Add read-only media object integrity verification for registered metadata references.
- Start with LocalSystem, NFS, SMB, and mounted-filesystem distributed backends.
- Verify object existence, regular-file shape, byte size, and `sha256` content hashes.
- Expose verification through the authenticated admin HTTP API and OpenAPI contract.

## Phase 1: Requirement Update

- [x] Append `v0.13.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.13.0`.

## Phase 2: Domain Implementation

- [x] Add media object verification result and errors.
- [x] Add read-only filesystem verification for registered media objects.
- [x] Reject disabled backends, unsupported backend families, unsupported hash algorithms, path traversal, size mismatch, and hash mismatch.

## Phase 3: HTTP and Contract

- [x] Add `POST /api/v1/admin/media/objects/{id}/verify`.
- [x] Update OpenAPI paths, schemas, error enum, and route coverage tests.
- [x] Add handler and domain tests for success and failure cases.

## Phase 4: Validation

- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add S3-compatible object verification with range-safe reads and configured credentials.
- [ ] Persist last verification status and timestamps in media object metadata.
- [ ] Add batch verification for import and repair workflows.
- [ ] Add additional hash algorithms after importer support is defined.

## Completion Notes

This phase verifies existing object bytes only. It does not upload, rewrite, delete, move, or repair media files.
