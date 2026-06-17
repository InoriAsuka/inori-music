# Phase 69 — Admin aggregate history stats

## Goal

Expose system-wide playback aggregate statistics to administrators:
most-played tracks, most-active users, and overall event counters.

## Repository interface extensions

```go
// Added to history.Repository
HistoryStats(ctx context.Context) (HistoryStats, error)
TopTracks(ctx context.Context, limit int) ([]TrackPlayCount, error)
TopUsers(ctx context.Context, limit int) ([]UserPlayCount, error)
```

New types added to `history/types.go` (already started before phase completion):

```go
type HistoryStats struct { TotalEvents, UniqueUsers, UniqueTracks int }
type TrackPlayCount struct { TrackID string; PlayCount int }
type UserPlayCount  struct { UserID  string; PlayCount int }
```

## Service methods

- `GetHistoryStats(ctx)` → `(HistoryStats, error)`
- `GetTopTracks(ctx, limit)` → `([]TrackPlayCount, error)` — limit 0 → 10, max 100
- `GetTopUsers(ctx, limit)` → `([]UserPlayCount, error)` — limit 0 → 10, max 100

## HTTP API (admin-auth only)

| Method | Route | Description |
|---|---|---|
| `GET` | `/api/v1/admin/history/stats`      | System-wide aggregate counts |
| `GET` | `/api/v1/admin/history/top-tracks` | Most-played tracks (optional `?limit`) |
| `GET` | `/api/v1/admin/history/top-users`  | Most-active users (optional `?limit`) |

## Tasks

- [x] Extend `history/types.go`: `HistoryStats`, `TrackPlayCount`, `UserPlayCount` + interface methods.
- [x] Implement `HistoryStats`, `TopTracks`, `TopUsers` on `history.MemoryRepository`.
- [x] Implement `HistoryStats`, `TopTracks`, `TopUsers` on `historypg.Repository` (SQL aggregate queries).
- [x] Add `GetHistoryStats`, `GetTopTracks`, `GetTopUsers` to `history.Service`.
- [x] Add 3 routes + 3 405 fallbacks in `handler.go`.
- [x] Add 3 handler functions + `parseHistoryAdminLimit` helper.
- [x] Add 3 history service unit tests (stats, top-tracks, top-users).
- [x] Add 5 HTTP-layer tests (stats, top-tracks with limit, top-users, not-configured 503, 405).
- [x] Add `HistoryStats`, `TrackPlayCount`, `UserPlayCount`, `TopTracksResult`, `TopUsersResult` schemas to OpenAPI.
- [x] Add 3 admin history paths to OpenAPI; bump `info.version` to `0.69.0`.
- [x] Add `TestStorageAdminOpenAPIContractAdminHistoryPaths`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` for 3 new paths.
- [x] Extend `TestStorageAdminOpenAPIContractSchemasAndErrors` for new schema names.
- [x] Run full `go test ./...` — green (434 tests).
- [x] Bump `requirement.md` version to `0.69.0`, append phase entry.
- [x] Commit, tag `v0.69.0`, push.

## Non-goals

- Per-user history admin (viewing another user's history).
- Time-windowed stats (last 7 days, last 30 days).
- Track/user enrichment (names, titles) in stats responses.

## Follow-up candidates

- Time-windowed play counts.
- Per-track/per-user history detail views for admins.
- User-editable playlists.
