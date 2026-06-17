# Phase 66 — SQL aggregate catalog stats

## Goal

Replace full-table-scan stats computation (load all rows, count in Go) with
SQL `COUNT(*)`/`GROUP BY` aggregate queries so stats endpoints scale to large
catalogs without fetching every row.

## New Repository methods

| Method | SQL |
|---|---|
| `CountEntities(ctx)` | Scalar subqueries: `SELECT (SELECT COUNT(*) FROM artists) AS artists, …` |
| `ArtistAlbumTrackCounts(ctx)` | `FROM artists LEFT JOIN albums LEFT JOIN tracks GROUP BY artist` |
| `AlbumTrackCounts(ctx)` | `FROM albums LEFT JOIN tracks GROUP BY album` |
| `PlaylistTrackCounts(ctx)` | `FROM playlists LEFT JOIN playlist_tracks GROUP BY playlist` |

## Service changes

`GetCatalogStats`, `GetArtistStatsBreakdown`, `GetAlbumStatsBreakdown`,
`GetPlaylistStatsBreakdown` replaced their N+1 list-then-count logic with a
single delegate call to the new aggregate repo methods.

## Tasks

- [x] Add 4 aggregate method signatures to `catalog/types.go` Repository interface.
- [x] Implement on `catalog.MemoryRepository` (in-memory counting).
- [x] Implement on `catalogpg.Repository` in `repository_page.go` (SQL COUNT/GROUP BY).
- [x] Add 4 aggregate stubs on test `memRepo` in `service_test.go`.
- [x] Add 2 Postgres integration tests (`TestRepositoryCountEntities`, `TestRepositoryArtistAlbumTrackCounts`).
- [x] Bump OpenAPI `info.version` → `0.66.0`.
- [x] Bump `VERSION` → `0.66.0`.
- [x] Update `requirement.md`.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.66.0`, push.

## Follow-up candidates

- SQL pushdown for recent timelines.
- Playback history / listening activity domain.
