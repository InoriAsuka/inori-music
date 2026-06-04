# Plan: Phase 10 Media Object Registry Scaffold

## Requirement Version

v0.10.0

## Goals

- Add a domain-level registry for media object metadata and storage references.
- Keep large media bytes outside the API service and relational database.
- Validate object keys, content hashes, asset kinds, lifecycle states, and backend availability before registration.
- Prepare the storage domain for future import, streaming, deduplication, and PostgreSQL persistence work.

## Phase 1: Requirement Update

- [x] Append `v0.10.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.10.0`.

## Phase 2: Domain Implementation

- [x] Add media object asset-kind and lifecycle constants.
- [x] Add static media object validation.
- [x] Add an in-memory media object repository.
- [x] Add a media object registry service that verifies referenced backends exist and are enabled.

## Phase 3: Validation

- [x] Add unit tests for registration, backend rejection, object key validation, backend listing, and content-hash lookup.
- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add PostgreSQL persistence for media object metadata.
- [ ] Add import workflows that create media objects after upload/probe/hash verification.
- [ ] Add streaming URL generation and range-read integration.
- [ ] Add deduplication indexes based on content hash and asset kind.
- [ ] Expose media object registry endpoints once authorization and library models are defined.

## Completion Notes

This phase is a domain scaffold only. It records and validates media object metadata but does not upload, delete, move, stream, or mutate binary media files.
