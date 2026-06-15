# Phase 54 — Catalog recently updated API

## Goal

Add an admin dashboard endpoint that returns a unified newest-first timeline of
recently updated artists, albums, and tracks, complementing the Phase 53
recently-added timeline and letting admin clients surface changed metadata without
issuing separate list calls per entity type.

## Requirements

- `GET /api/v1/admin/catalog/recently-updated` → `UpdatedCatalogResult` (`{items: [{kind, artist|album|track, updatedAt}]}`)
- The endpoint requires admin auth and returns 503 when no catalog service is configured.
- Query parameter `kind` optionally filters to `artist`, `album`, or `track`; invalid values return 400.
- Query parameter `limit` defaults to 20, must be a positive integer, and clamps above 100.
- Results sort newest-first by each entity's `UpdatedAt` timestamp.
- Empty catalog returns a non-nil empty `items` array (not null).
- No new `Repository` interface methods required — results derive from existing `ListArtists`, `ListAlbums`, and `ListTracks`.

## Non-goals

- No playlist timeline entries in this phase.
- No viewer-facing `/api/v1/catalog/recently-updated` endpoint.
- No PostgreSQL-specific ordering or limit pushdown.
- No pagination beyond the existing `limit` cap.

## Tasks

- [x] Add `UpdatedCatalogItem` and `UpdatedCatalogResult` to `catalog/types.go`.
- [x] Add `GetRecentlyUpdated` to `catalog/service.go`.
- [x] Reuse shared kind validation and limit normalization for recent timeline endpoints.
- [x] Add `getRecentlyUpdated` handler to `httpapi/handler.go`.
- [x] Register route `GET /api/v1/admin/catalog/recently-updated`.
- [x] Register 405 fallback for the path.
- [x] Add `catalog.Service` unit tests for empty, ordering, kind filter, invalid kind, and limits.
- [x] Add HTTP-layer tests for empty, populated, kind, limit, invalid inputs, 503, and 405.
- [x] Update `TestStorageAdminOpenAPIContractCoversRoutes` with the new path.
- [x] Update `TestStorageAdminOpenAPIContractSchemasAndErrors` schema name list.
- [x] Add OpenAPI schemas and path operation.
- [x] Bump OpenAPI `info.version` → `0.54.0`.
- [x] Bump `VERSION` → `0.54.0`.
- [x] Update `requirement.md` current version and append v0.54.0 history entry.

## Follow-up candidates

- Add playlist entries to recent timelines once playlist dashboard semantics are defined.
- Add viewer-facing recent timeline endpoints for library browse clients.
- Add repository-level ordered/limited queries when catalog size warrants pushdown.
