# Phase 63 — Nested catalog browse routes

## Goal

Add dedicated sub-collection routes so clients can browse albums under an artist
and tracks under an artist or album without applying a flat filter on the collection
endpoint. The parent entity is validated (404 on unknown ID), and results support
the same `limit`/`offset`/`sortBy`/`sortOrder` parameters established in Phases 61-62.

## Routes added (admin + viewer for each)

| Pattern | Handler |
|---|---|
| `GET /api/v1/{admin/}catalog/artists/{id}/albums` | `listAlbumsByArtist` |
| `GET /api/v1/{admin/}catalog/artists/{id}/tracks` | `listTracksByArtist` |
| `GET /api/v1/{admin/}catalog/albums/{id}/tracks`  | `listTracksByAlbum`  |

All 6 new routes are protected by the same auth middleware as existing sibling routes
(`requireAdminAuth` for admin, `requireViewerAuth` for viewer). A 405 fallback is
registered for each.

## Non-goals

- No changes to Repository or Service interfaces.
- No SQL LIMIT/OFFSET pushdown.

## Tasks

- [x] Register 6 new GET routes in `handler.go`.
- [x] Register 6 corresponding 405 fallback routes.
- [x] Add `listAlbumsByArtist`, `listTracksByArtist`, `listTracksByAlbum` handlers (parent-exist validation + sort + paginate).
- [x] Add 4 HTTP-layer tests: albums-by-artist (sort/paginate/404/405), tracks-by-artist (404/405), tracks-by-album (sort/paginate/404/405), viewer access to all three.
- [x] Add 6 new paths to OpenAPI contract with pagination+sort params and typed response schemas.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` for 6 new paths.
- [x] Extend `TestStorageAdminOpenAPIContractCatalogListSortParams` for 6 new paths.
- [x] Bump OpenAPI `info.version` → `0.63.0`.
- [x] Bump `VERSION` → `0.63.0`.
- [x] Update `requirement.md` current version and append v0.63.0 history entry.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.63.0`, push.

## Follow-up candidates

- PostgreSQL ORDER BY + LIMIT/OFFSET pushdown for large catalogs.
- Playback history / listening activity domain.
- `GET /catalog/playlists/{id}/tracks` pagination (currently returns all tracks unbounded).
