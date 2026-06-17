# Phase 61 — Catalog list pagination

## Goal

All four catalog list endpoints return unbounded arrays. This phase adds
`limit`/`offset` pagination to `listArtists`, `listAlbums`, `listTracks`,
and `listPlaylists` so clients can build predictable paginated tables without
fetching the entire catalog. The pattern mirrors `listMediaObjects` but lives
entirely in the handler layer — no Repository or Service changes are needed.

## Requirements

- Add `limit` and `offset` query parameters to all four catalog list endpoints
  (artist, album, track, playlist) on both admin and viewer routes.
- `limit` defaults to 50, max 500; must be a positive integer or 400 `invalid_limit`.
- `offset` defaults to 0; must be a non-negative integer or 400 `invalid_offset`.
- Each response includes the existing entity array key plus a `pagination` envelope
  with `limit`, `offset`, `total`, and `hasMore`.
- Existing `artistId` and `albumId` filter params on albums and tracks continue
  to work and are applied before pagination.
- No changes to `catalog.Repository`, `catalog.Service`, or any List* signatures.

## Non-goals

- No `sortBy` or `sortOrder` parameters in this phase.
- No SQL-level LIMIT/OFFSET pushdown (in-service pagination is sufficient for now).
- No new nested browse routes.

## Tasks

- [x] Add `CatalogPaginationMeta` type to `catalog/types.go`.
- [x] Add `parseCatalogPage` and `paginateCatalog[T]` helpers to `handler.go`.
- [x] Update `listArtists`, `listAlbums`, `listTracks`, `listPlaylists` to use pagination.
- [x] Add 5 HTTP-layer pagination tests covering limit, offset, hasMore, invalid params, and viewer access.
- [x] Add `CatalogPaginationMeta` schema to OpenAPI contract.
- [x] Add `limit`/`offset` query params and `pagination` response property to the 8 catalog list paths.
- [x] Add `invalid_offset` to OpenAPI error code enum and contract test.
- [x] Add `CatalogPaginationMeta` to schema name list in contract test.
- [x] Bump OpenAPI `info.version` → `0.61.0`.
- [x] Bump `VERSION` → `0.61.0`.
- [x] Update `requirement.md` current version and append v0.61.0 history entry.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.61.0`, push.

## Follow-up candidates

- `sortBy`/`sortOrder` on catalog list endpoints (phase 62).
- Nested browse routes (`/catalog/artists/{id}/albums` etc.).
- PostgreSQL LIMIT/OFFSET pushdown for large catalogs.
- Playback history / listening activity domain.
