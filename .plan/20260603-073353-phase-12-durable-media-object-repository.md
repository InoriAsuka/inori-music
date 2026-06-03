# Plan: Phase 12 Durable Media Object Repository

## Requirement Version

v0.12.0

## Goals

- Add optional durable persistence for media object metadata before PostgreSQL is introduced.
- Preserve `MemoryMediaObjectRepository` as the default for tests and ephemeral development.
- Keep metadata persistence dependency-free and suitable for single-node self-hosting.
- Use the same conservative atomic JSON file write pattern as the storage backend repository.

## Phase 1: Requirement Update

- [x] Append `v0.12.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.12.0`.

## Phase 2: Repository Implementation

- [x] Add a file-backed `MediaObjectRepository` implementation.
- [x] Load existing JSON media object metadata during startup.
- [x] Persist repository changes with temp-file write, sync, close, and atomic rename.
- [x] Create repository parent directories when needed.
- [x] Preserve stable filtered ordering by backend/object key and content hash/object key.

## Phase 3: Server Wiring

- [x] Add `INORI_MEDIA_OBJECT_REPOSITORY_FILE` configuration.
- [x] Keep `MemoryMediaObjectRepository` as the default when no media repository path is configured.
- [x] Fail startup when the configured media object repository file cannot be loaded.

## Phase 4: Validation

- [x] Add unit tests for persistence, filtered listings, malformed files, unsupported schema versions, and empty IDs.
- [x] Add command tests for repository selection.
- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add PostgreSQL media object repository and migrations.
- [ ] Add JSON-to-PostgreSQL migration tooling.
- [ ] Add pagination and indexes for large media object sets.
- [ ] Add audit metadata once authenticated user identity is modeled.

## Completion Notes

This phase adds bootstrap metadata persistence only. It does not upload, stream, delete, move, or verify binary media files.
