# Plan: Phase 9 Durable File Repository

## Requirement Version

v0.9.0

## Goals

- Add an optional durable repository for storage backend configuration before introducing PostgreSQL migrations.
- Preserve the in-memory repository as the default development behavior.
- Allow self-hosted and local deployments to retain backend health and capacity state across process restarts.
- Keep the implementation dependency-free and safe for bootstrap environments.

## Phase 1: Requirement Update

- [x] Append `v0.9.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.9.0`.

## Phase 2: Repository Implementation

- [x] Add a file-backed `storage.Repository` implementation.
- [x] Load existing JSON repository state during startup.
- [x] Persist repository changes with temp-file write, sync, close, and atomic rename.
- [x] Create repository parent directories when needed.
- [x] Keep repository records sorted by priority and ID for stable API responses.

## Phase 3: Server Wiring

- [x] Add `INORI_STORAGE_REPOSITORY_FILE` configuration.
- [x] Keep `MemoryRepository` as the default when no file path is configured.
- [x] Fail startup when the configured repository file cannot be loaded.

## Phase 4: Validation

- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add PostgreSQL migrations and a PostgreSQL-backed repository once the database layer is introduced.
- [ ] Add migration tooling from the JSON file repository into PostgreSQL.
- [ ] Add encrypted-at-rest secret provider integrations for environments that cannot rely on environment-variable secret references.
- [ ] Add repository-level optimistic concurrency once multiple admin users are supported.

## Completion Notes

This phase adds durable bootstrap persistence only. The JSON file repository is intended for development, single-node self-hosting, and migration staging; production metadata remains PostgreSQL-first in the architecture direction.
