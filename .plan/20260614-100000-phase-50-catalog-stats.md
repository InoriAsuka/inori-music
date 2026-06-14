# Phase 50 — Catalog entity count statistics

## Goal

Add `GET /api/v1/admin/catalog/stats` endpoint that returns metadata-only
aggregate counts for artists, albums, tracks, and playlists in a single request,
enabling admin dashboards to display a quick summary without N separate list calls.

## Requirements

- Response shape: `{"artists": N, "albums": N, "tracks": N, "playlists": N}`
- All counts are derived from in-memory counts via the existing `List*` repo calls;
  no new `Repository` interface methods required.
- Returns 503 when no catalog service is configured.
- Returns 405 on non-GET requests.

## Non-goals

- No per-artist/album breakdown (follow-up candidate).
- No PostgreSQL COUNT query optimisation — sequential list calls are adequate.

## Tasks

- [x] Add `CatalogStats` struct and `GetCatalogStats(ctx)` to `catalog.Service`
- [x] Add `getCatalogStats` handler and register `GET /api/v1/admin/catalog/stats` route
- [x] Register 405 fallback for `/api/v1/admin/catalog/stats`
- [x] Add 3 `catalog.Service` unit tests (empty, populated, error propagation)
- [x] Add 4 HTTP-layer tests (happy path, no-catalog 503, method-not-allowed 405, populated counts)
- [x] Add `CatalogStats` schema and `get` operation to OpenAPI contract
- [x] Bump `info.version` → `0.50.0`
- [x] Bump `VERSION` → `0.50.0`
- [x] Update `requirement.md` current version and append v0.50.0 history entry

## Follow-up candidates

- Per-artist album/track breakdown stats.
- PostgreSQL-backed COUNT query for performance with large catalogs.
