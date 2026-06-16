# Phase 40 - Catalog Admin HTTP API

## Requirement

Expose the catalog domain over authenticated admin HTTP endpoints for artists, albums, and tracks.

## Tasks

- [x] Add `WithCatalogService` handler option and `catalogService` field.
- [x] Register 12 catalog routes under `/api/v1/admin/catalog/`.
- [x] Implement artist list/create/get/delete handlers.
- [x] Implement album list (with optional `artistId` filter)/create/get/delete handlers.
- [x] Implement track list (with optional `artistId`/`albumId` filter)/create/get/delete handlers.
- [x] Map catalog domain errors to HTTP status codes in `writeError`.
- [x] Add `MemoryRepository` to the catalog package for in-process testing.
- [x] Add 11 HTTP-layer catalog tests.
- [x] Add `UserId`, `CatalogId` parameters and all catalog paths to the OpenAPI contract.
- [x] Update versioned requirements and release version metadata.

## Non-goals

- Search or full-text query endpoints.
- Track update/PATCH endpoints.
- Playlist, favorites, or playback APIs.

## Follow-up candidates

- Phase 41: catalog import workflow — convert verified media objects into track metadata via import API.
- Phase 42: PostgreSQL full-text search over catalog metadata.
- Phase 43: user-facing read-only catalog browse endpoints.
