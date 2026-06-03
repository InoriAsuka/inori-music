# Plan: Phase 2 Storage Domain Scaffold

## Requirement Version

v0.2.0

## Goals

- Create the first Go API service scaffold under `services/api`.
- Implement the storage administration domain model for server-managed backend configuration.
- Add validation and capability inference for LocalSystem, NFS, SMB, S3-compatible, and distributed backend families.
- Add unit tests so the storage domain can evolve safely before database migrations and HTTP APIs are added.

## Phase 1: Requirement Update

- [x] Append `v0.2.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Keep `.plan/` version tracked for implementation history.

## Phase 2: Service Scaffold

- [x] Create `services/api/go.mod`.
- [x] Create a minimal server entry point under `services/api/cmd/server`.
- [x] Keep the initial scaffold dependency-free and standard-library only.

## Phase 3: Storage Domain Implementation

- [x] Define storage backend families and strongly typed backend configuration.
- [x] Define capability and health status models.
- [x] Define media object references that point to backend IDs and object keys.
- [x] Implement backend validation and capability inference.
- [x] Implement an in-memory repository for early domain tests.
- [x] Implement a service for registration, listing, disabling, and default backend selection.

## Phase 4: Verification

- [x] Add unit tests for backend validation.
- [x] Add unit tests for default backend selection and disabling behavior.
- [x] Run `gofmt`.
- [x] Run `go test ./...`.

## Future Implementation Tasks

- [ ] Add PostgreSQL migrations for storage backends and media objects.
- [ ] Add encrypted configuration persistence.
- [x] Add HTTP admin endpoints.
- [ ] Add OpenAPI contracts.
- [x] Add real filesystem probe checks for local, NFS, and SMB mount paths.
- [x] Add S3-compatible probe checks with temporary test objects.
- [x] Add scheduled health checks and capacity metadata refresh. On-demand probes were added in phase 5 and scheduled refresh in phase 7.

## Completion Notes

This phase turns the phase-1 storage architecture into executable domain code while intentionally postponing database persistence, network probes, and HTTP handlers.
