# Phase 41 - Catalog Import Workflow

## Requirement

Convert verified media objects into track metadata via an import API.

## Tasks

- [x] Add `ErrImportRejected` sentinel error.
- [x] Add `MediaObjectInfo`, `MediaObjectReader`, `ImportTrackRequest` to `catalog/types.go`.
- [x] Implement `Service.ImportTrack` with validation (asset kind, lifecycle state, artist/album consistency).
- [x] Add `WithMediaObjectReader` method to `catalog.Service`.
- [x] Add `GetMediaObjectInfoForImport` on `storage.MediaObjectService`.
- [x] Add `mediaObjectReaderAdapter` in httpapi to bridge the two without import cycle.
- [x] Register `POST /api/v1/admin/catalog/import` route.
- [x] Implement `importTrack` HTTP handler.
- [x] Map `ErrImportRejected` → HTTP 422 in `writeError`.
- [x] Wire media object service into catalog service via `withCatalogMediaReader()` in `Routes()`.
- [x] Add 7 `ImportTrack` catalog service unit tests.
- [x] Add 7 HTTP-layer import tests.
- [x] Update OpenAPI spec with import path and `CatalogImportRequest` schema.
- [x] Update version to 0.41.0.

## Non-goals

- Batch import endpoints.
- Auto-extraction of track metadata from audio file tags.
- Re-import / update-on-conflict semantics.

## Follow-up candidates

- Phase 42: PostgreSQL full-text search over catalog metadata.
- Phase 43: user-facing read-only catalog browse endpoints.
- Phase 44: batch import endpoint accepting an array of mediaObjectIds.
