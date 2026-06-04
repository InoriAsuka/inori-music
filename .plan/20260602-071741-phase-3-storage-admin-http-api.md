# Plan: Phase 3 Storage Admin HTTP API

## Requirement Version

v0.3.0

## Goals

- Turn the server scaffold into a runnable HTTP service.
- Expose versioned administrative JSON endpoints for storage backend validation and lifecycle management.
- Apply strict request handling and stable JSON response conventions.
- Preserve the storage domain boundary so future persistence and authentication work can be added without rewriting handlers.

## Phase 1: Requirement Update

- [x] Append `v0.3.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and the README baseline to `0.3.0`.

## Phase 2: Domain Hardening

- [x] Add explicit JSON names to storage domain models.
- [x] Reject storage backend payloads containing zero or multiple family-specific configuration branches.
- [x] Add a service method for non-persisting configuration validation.

## Phase 3: HTTP API

- [x] Add `GET /healthz`.
- [x] Add `GET /api/v1/admin/storage/backends`.
- [x] Add `POST /api/v1/admin/storage/backends`.
- [x] Add `POST /api/v1/admin/storage/backends/validate`.
- [x] Add `POST /api/v1/admin/storage/backends/{id}/default`.
- [x] Add `POST /api/v1/admin/storage/backends/{id}/disable`.
- [x] Add bounded, strict JSON decoding, JSON content-type enforcement, writable request DTOs, and consistent error envelopes.
- [x] Replace the placeholder entry point with a runnable HTTP server that binds to loopback by default.
- [x] Document endpoints, request rules, error envelopes, and the next security step.

## Phase 4: Verification

- [x] Add handler tests for health, validation, registration, listing, default selection, disabling, duplicate registration, malformed JSON, unknown JSON fields, and not-found responses.
- [x] Extend domain tests for exactly-one-config enforcement.
- [x] Run `gofmt`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.

## Future Implementation Tasks

- [x] Add bootstrap administrator authentication middleware.
- [ ] Add role-based administrator authorization middleware.
- [ ] Add PostgreSQL persistence and migrations.
- [ ] Add encrypted backend configuration persistence.
- [x] Add OpenAPI contracts and contract validation.
- [x] Add real local and mounted-filesystem backend probes.
- [x] Add S3-compatible backend probes.
- [x] Add scheduled storage health refresh jobs and filesystem capacity reporting. On-demand health endpoints were added in phase 5 and scheduled refresh in phase 7.

## Completion Notes

This phase exposes static storage configuration management over HTTP. At phase-3 completion it intentionally did not claim that a backend was reachable. Phase 5 later added explicit on-demand probes for local and mounted-filesystem semantics; phase 6 later added S3-compatible object probes.
