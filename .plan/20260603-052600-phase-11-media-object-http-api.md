# Plan: Phase 11 Media Object Admin HTTP API

## Requirement Version

v0.11.0

## Goals

- Expose the media object registry scaffold through authenticated HTTP routes.
- Keep registration request bodies free of server-owned timestamps.
- Document the new route surface in the OpenAPI contract.
- Preserve strict JSON decoding and stable error envelopes.

## Phase 1: Requirement Update

- [x] Append `v0.11.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.11.0`.

## Phase 2: HTTP Implementation

- [x] Add handler wiring for `MediaObjectService`.
- [x] Add `POST /api/v1/admin/media/objects` for media object registration.
- [x] Add `GET /api/v1/admin/media/objects/{id}` for lookup.
- [x] Add `GET /api/v1/admin/media/objects?backendId=...` and `?contentHash=...` filters.
- [x] Add `invalid_media_object` error mapping.

## Phase 3: Contract and Validation

- [x] Update OpenAPI paths, schemas, parameters, and error enum.
- [x] Add handler tests for media object workflows and auth behavior.
- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Persist media object metadata in PostgreSQL.
- [ ] Add upload/import endpoints that create media objects only after content hashing and backend writes succeed.
- [ ] Add streaming URL generation and access-control decisions for playback clients.
- [ ] Add pagination and cursor contracts once large libraries are modeled.

## Completion Notes

This phase exposes metadata registration and lookup only. It does not upload, stream, delete, or move media bytes.
