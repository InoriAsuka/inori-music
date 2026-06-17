# Phase 59 — Viewer-facing catalog stats

## Goal

Expose the existing admin catalog stats endpoints to session-authenticated viewer
clients so they can build dashboard-style summaries without requiring admin
access.

## Requirements

- Add viewer-accessible `GET /api/v1/catalog/stats` returning `CatalogStats`.
- Add viewer-accessible `GET /api/v1/catalog/stats/artists` returning `CatalogArtistStatsBreakdown`.
- Add viewer-accessible `GET /api/v1/catalog/stats/albums` returning `CatalogAlbumStatsBreakdown`.
- Add viewer-accessible `GET /api/v1/catalog/stats/playlists` returning `CatalogPlaylistStatsBreakdown`.
- All four endpoints require session authentication (`requireViewerAuth`); static bootstrap token not accepted.
- Reuse existing handler functions `getCatalogStats`, `getArtistStatsBreakdown`, `getAlbumStatsBreakdown`, `getPlaylistStatsBreakdown` without modification.
- Add 405 fallbacks for all four new viewer paths.
- No new domain types or service methods required.

## Non-goals

- No admin-side changes.
- No schema changes (all four schemas already exist).
- No new catalog service logic.

## Tasks

- [x] Register four `GET /api/v1/catalog/stats*` routes under `requireViewerAuth` in `handler.go`.
- [x] Register four 405 fallback routes for the new viewer paths.
- [x] Add 14 HTTP-layer tests covering empty stats, populated counts, admin session acceptance, no-catalog-service 503, and method-not-allowed for all four endpoints.
- [x] Add four viewer paths to OpenAPI contract under the "Catalog" tag.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` to assert the four new paths.
- [x] Bump OpenAPI `info.version` → `0.59.0`.
- [x] Bump `VERSION` → `0.59.0`.
- [x] Update `requirement.md` current version and append v0.59.0 history entry.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.59.0`, push.

## Follow-up candidates

- Phase 60: S3 presigned URL generation for backends with `PresignedURLs` capability.
- Repository-level SQL pushdown for stats and timeline queries.
- Playback history / listening activity domain.
