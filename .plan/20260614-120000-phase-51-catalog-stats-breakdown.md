# Phase 51 — Catalog stats breakdown API

## Goal

Extend the v0.50.0 aggregate catalog stats endpoint with two per-entity breakdown
endpoints for admin dashboards: per-artist album/track counts and per-album track
counts, enabling display of richly annotated entity tables without N separate list calls.

## Requirements

- `GET /api/v1/admin/catalog/stats/artists` → `CatalogArtistStatsBreakdown` (`{artists: [{artistId, name, albumCount, trackCount}]}`)
- `GET /api/v1/admin/catalog/stats/albums` → `CatalogAlbumStatsBreakdown` (`{albums: [{albumId, title, artistId, trackCount}]}`)
- Both endpoints require admin auth and return 503 when no catalog service is configured.
- Both return 405 on non-GET verbs.
- Empty catalog returns a non-nil empty array (not null).
- No new `Repository` interface methods required — counts derived from existing `ListAlbumsByArtist`, `ListTracksByArtist`, `ListTracksByAlbum`.

## Non-goals

- No PostgreSQL COUNT(*) query optimisation — sequential list calls are adequate at current scale.
- No viewer-facing (`/api/v1/catalog/stats/...`) endpoints.
- No playlist membership breakdown.
- No pagination on breakdown responses.

## Tasks

- [x] Add `ArtistStatItem`, `ArtistStatsBreakdown`, `AlbumStatItem`, `AlbumStatsBreakdown` to `catalog/types.go`
- [x] Add `GetArtistStatsBreakdown` and `GetAlbumStatsBreakdown` to `catalog/service.go`
- [x] Add `getArtistStatsBreakdown` and `getAlbumStatsBreakdown` handlers to `httpapi/handler.go`
- [x] Register routes `GET /api/v1/admin/catalog/stats/artists` and `GET /api/v1/admin/catalog/stats/albums`
- [x] Register 405 fallbacks for both paths
- [x] Add 4 `catalog.Service` unit tests (empty/populated for each breakdown type)
- [x] Add 8 HTTP-layer tests (empty, populated, 503, 405 per endpoint)
- [x] Update `TestStorageAdminOpenAPIContractCoversRoutes` with both new paths
- [x] Update `TestStorageAdminOpenAPIContractSchemasAndErrors` schema name list
- [x] Add 4 schemas to OpenAPI contract
- [x] Add 2 path operations to OpenAPI contract
- [x] Bump `info.version` → `0.51.0`
- [x] Bump `VERSION` → `0.51.0`
- [x] Update `requirement.md` current version and append v0.51.0 history entry

## Follow-up candidates

- PostgreSQL `SELECT artist_id, COUNT(*) FROM albums GROUP BY artist_id` for O(1) breakdown at scale.
- Viewer-accessible catalog stats endpoint (`/api/v1/catalog/stats`).
- Playlist membership count per artist/album in breakdown response.
