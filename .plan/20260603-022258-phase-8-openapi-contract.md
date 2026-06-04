# Plan: Phase 8 OpenAPI Contract

## Requirement Version

v0.8.0

## Goals

- Publish a versioned OpenAPI 3.1 contract for the storage administration HTTP API.
- Capture the current route surface, request bodies, response envelopes, and Bearer authentication requirements.
- Add automated contract tests so future handler changes must update the API contract.
- Keep the contract dependency-free and parseable with the Go standard library.

## Phase 1: Requirement Update

- [x] Append `v0.8.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.8.0`.

## Phase 2: OpenAPI Contract

- [x] Add `packages/api-contract/openapi/storage-admin.v1.json`.
- [x] Document `/healthz` as public.
- [x] Document authenticated `/api/v1/admin/storage/backends*` routes.
- [x] Add schemas for storage backends, backend config families, capabilities, probe results, capacity reports, refresh reports, and error envelopes.
- [x] Add reusable Bearer authentication security scheme.

## Phase 3: Contract Verification

- [x] Add standard-library tests that parse the OpenAPI document as JSON.
- [x] Verify all implemented storage admin routes are present with the expected HTTP methods.
- [x] Verify admin routes require Bearer authentication while `/healthz` remains public.
- [x] Verify core schemas and error codes are represented.

## Phase 4: Validation

- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Generate client/server code from the OpenAPI contract when dependencies are available.
- [ ] Add request/response golden tests against the OpenAPI schemas.
- [ ] Publish rendered API documentation in the website app once the web project exists.
- [ ] Add OpenAPI security examples for token rotation and future role-based authorization.

## Completion Notes

This phase documents and tests the current API surface. It does not introduce code generation because the current environment blocks external dependency downloads.
