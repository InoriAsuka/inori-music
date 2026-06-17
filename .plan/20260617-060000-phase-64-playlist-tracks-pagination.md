# Phase 64 — Playlist tracks pagination

## Goal

`GET /api/v1/catalog/playlists/{id}/tracks` and its admin counterpart returned an
unbounded ordered track array. This phase adds `limit`/`offset` pagination so
clients can page through large playlists without fetching the entire list.

## Key constraint

Playlist track order is **user-curated** and must be preserved exactly.
`sortBy`/`sortOrder` parameters are intentionally NOT added — the user-defined
sequence is the canonical sort. Pagination simply windows the existing ordered list.

## Tasks

- [x] Update `getPlaylistTracks` handler to parse `limit`/`offset`, call `paginateCatalog`, and return `{"tracks": page, "pagination": meta}`.
- [x] Add 2 HTTP-layer tests: admin pagination (order-preserved, limit, offset, invalid params) and viewer pagination (limit + hasMore).
- [x] Add `limit`/`offset` query params and `pagination` property to both playlist tracks paths in the OpenAPI contract.
- [x] Add `TestStorageAdminOpenAPIContractPlaylistTracksPagination` asserting `limit`/`offset` present and `sortBy`/`sortOrder` absent.
- [x] Bump OpenAPI `info.version` → `0.64.0`.
- [x] Bump `VERSION` → `0.64.0`.
- [x] Update `requirement.md` current version and append v0.64.0 history entry.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.64.0`, push.

## Follow-up candidates

- PostgreSQL ORDER BY + LIMIT/OFFSET pushdown for large catalogs.
- Playback history / listening activity domain.
- User-editable playlists (create/edit by non-admin session users).
