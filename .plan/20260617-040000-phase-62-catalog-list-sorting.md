# Phase 62 — Catalog list sorting

## Goal

Add `sortBy` and `sortOrder` query parameters to all four catalog list endpoints
(artists, albums, tracks, playlists) on both admin and viewer routes, enabling
clients to request deterministic ordering before pagination.

## Requirements

- Add `sortBy` and `sortOrder` query parameters to the same 8 list paths that
  received `limit`/`offset` in Phase 61.
- `sortOrder` must be `asc` (default) or `desc`; any other value returns
  `400 invalid_sort_order`.
- `sortBy` defaults silently to the entity's primary sort field when empty or
  unrecognised (no 400 — unknown fields fall through to the default).

### Valid sortBy values

| Entity | Fields |
|---|---|
| Artist | `name` (default), `sortName`, `createdAt`, `updatedAt` |
| Album  | `title` (default), `sortTitle`, `releaseYear`, `createdAt`, `updatedAt` |
| Track  | `title` (default), `sortTitle`, `trackNumber`, `discNumber`, `durationMs`, `createdAt`, `updatedAt` |
| Playlist | `name` (default), `createdAt`, `updatedAt` |

- Sort is applied before pagination so `limit`/`offset` windows are stable.
- No Repository or Service interface changes.

## Non-goals

- No PostgreSQL ORDER BY pushdown (in-handler sort is sufficient for now).
- No multi-key compound sort.

## Tasks

- [x] Add sort constants and `CatalogSortOrderAsc/Desc` to `catalog/types.go`.
- [x] Extend `parseCatalogPage` to also return `sortBy` and `sortOrder`.
- [x] Add `normalizeSortOrder` helper and per-entity sort functions (`sortCatalogArtists`, `sortCatalogAlbums`, `sortCatalogTracks`, `sortCatalogPlaylists`) to `handler.go`.
- [x] Update `listArtists`, `listAlbums`, `listTracks`, `listPlaylists` to sort before paginating.
- [x] Add 6 HTTP-layer sort tests (per-entity sort direction + invalid sortOrder + viewer session).
- [x] Add `sortBy`/`sortOrder` params, sort field descriptions, and `invalid_sort_order` error code to OpenAPI contract.
- [x] Add `TestStorageAdminOpenAPIContractCatalogListSortParams` contract test.
- [x] Bump OpenAPI `info.version` → `0.62.0`.
- [x] Bump `VERSION` → `0.62.0`.
- [x] Update `requirement.md` current version and append v0.62.0 history entry.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.62.0`, push.

## Follow-up candidates

- Nested browse routes (`GET /catalog/artists/{id}/albums`, `GET /catalog/albums/{id}/tracks`).
- PostgreSQL ORDER BY + LIMIT/OFFSET pushdown for large catalogs.
- Playback history / listening activity domain.
