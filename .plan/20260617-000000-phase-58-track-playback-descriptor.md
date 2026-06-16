# Phase 58 — Track playback descriptor for viewer clients

## Goal

Viewer clients need to discover the binary location of a track's linked audio
file before initiating playback. This phase adds a metadata-only endpoint that
resolves track → media object → backend and returns the fields a client needs,
without the server streaming bytes or exposing storage credentials.

## Requirements

- Add `GET /api/v1/catalog/tracks/{id}/playback` (viewer-auth, session only).
- Return `TrackPlaybackDescriptor` with: `trackId`, `mediaObjectId`, `mimeType`,
  `durationMs`, `backendId`, `backendType` (omitted when backend not found),
  `objectKey`.
- Validate that the linked media object has `lifecycleState = active` and
  `assetKind ∈ {original_audio, transcoded_audio}`; otherwise return 422
  `playback_unavailable`.
- Return 404 when track not found or media object not found.
- Return 503 `catalog_not_configured` when catalog service is absent, and 503
  `media_registry_not_configured` when media object service is absent.
- Add 405 fallback for the new path.
- No URL signing, no byte serving, no credential exposure.
- Add `ErrPlaybackUnavailable` sentinel to the storage package.
- Add `playback_unavailable` to `writeError`, OpenAPI error enum, and
  contract test.

## Non-goals

- No S3 presigned URL generation.
- No range-request proxying or byte streaming.
- No changes to the catalog or storage domain interfaces.
- No admin-facing counterpart.

## Tasks

- [x] Add `ErrPlaybackUnavailable` to `services/api/internal/storage/media_object.go`.
- [x] Add `trackPlaybackDescriptor` struct and `getTrackPlayback` handler to `services/api/internal/httpapi/handler.go`.
- [x] Register `GET /api/v1/catalog/tracks/{id}/playback` (viewer-auth) and 405 fallback.
- [x] Add `playback_unavailable` case to `writeError` switch.
- [x] Add `newViewerWithMediaHandler` helper and 8 HTTP-layer tests in `handler_test.go`.
- [x] Add `TrackPlaybackDescriptor` schema to OpenAPI contract.
- [x] Add `GET /api/v1/catalog/tracks/{id}/playback` path to OpenAPI contract.
- [x] Add `playback_unavailable` to OpenAPI error code enum.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes`, `TestStorageAdminOpenAPIContractSchemasAndErrors`, and add `TestStorageAdminOpenAPIContractTrackPlaybackDescriptor`.
- [x] Bump OpenAPI `info.version` → `0.58.0`.
- [x] Bump `VERSION` → `0.58.0`.
- [x] Update `requirement.md` current version and append v0.58.0 history entry.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.58.0`, push.

## Follow-up candidates

- Phase 59: viewer-facing catalog stats (`GET /api/v1/catalog/stats`).
- S3 presigned URL generation for backends with `PresignedURLs` capability.
- Repository-level ordered/limited timeline queries for large catalogs.
- Playback history / listening activity domain.
