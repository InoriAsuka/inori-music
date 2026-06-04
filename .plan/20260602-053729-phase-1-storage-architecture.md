# Plan: Phase 1 Storage Architecture Bootstrap

## Requirement Version

v0.1.0

## Goals

- Record the first-phase requirement baseline before implementation work.
- Define a storage architecture that supports LocalSystem, NFS, SMB, S3-compatible, and distributed storage backends.
- Make server-managed storage configuration an explicit product requirement.
- Prepare the repository for future phased implementation by tracking `.plan/` files in Git.

## Phase 1: Requirement and Architecture Baseline

- [x] Create `requirement.md` with the first 0.x requirement baseline.
- [x] Define media asset scope and storage safety constraints.
- [x] Define supported backend families.
- [x] Define server-managed configuration requirements.
- [x] Define database and search direction for 0.x.

## Phase 2: Storage Backend Design

- [x] Create an architecture document for the media storage layer.
- [x] Define storage backend types and expected deployment modes.
- [x] Define server-side administrative responsibilities.
- [x] Define a capability model for different backend types.
- [x] Define a configuration model and validation workflow.

## Phase 3: Architecture Decision Records

- [x] Add ADR for server-managed multi-backend media storage.
- [x] Add ADR for PostgreSQL-first database and search direction.

## Phase 4: Repository Communication

- [x] Update `README.md` with the phase-1 architecture summary.
- [x] Document where requirements, plans, architecture notes, and ADRs live.

## Future Implementation Tasks

- [x] Scaffold the Go API service and storage domain module.
- [ ] Implement `StorageBackend` and `MediaObject` database models.
- [x] Implement local filesystem backend validation and safe real probe checks.
- [x] Implement mounted filesystem backend validation and safe real probe checks for NFS and SMB paths.
- [x] Implement S3-compatible backend validation and safe real probe checks.
- [x] Add scheduled storage health check jobs. On-demand probes were added in phase 5 and scheduled refresh in phase 7.
- [x] Add administrative API endpoints for storage backend management.
- [ ] Add OpenAPI contracts for storage administration.
- [ ] Add integration tests for local and S3-compatible storage adapters. Domain unit tests were added in phase 2; real adapter integration tests remain pending.

## Completion Notes

Phase 1 establishes the design baseline only. It intentionally avoids adding runtime code before the repository has a service scaffold, database migration strategy, and API contract workflow.
