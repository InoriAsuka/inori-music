# Phase 70 — Time-windowed admin history stats

## Goal

Add an optional `since` query parameter (RFC3339) to the three admin
aggregate history endpoints so callers can limit results to events that
occurred at or after a given timestamp.

## HTTP API changes

| Endpoint | New param |
|---|---|
| `GET /api/v1/admin/history/stats` | `?since=<RFC3339>` |
| `GET /api/v1/admin/history/top-tracks` | `?since=<RFC3339>` |
| `GET /api/v1/admin/history/top-users` | `?since=<RFC3339>` |

`since` is optional; omitting it returns all-time data (Phase 69 behaviour).
An unparseable value returns `400 invalid_since`.

## Repository interface changes

`StatsFilter` struct added to `history/types.go`:

```go
type StatsFilter struct {
    Since time.Time // zero → all time
}
```

All three aggregate `Repository` methods updated to accept `StatsFilter`:

```go
HistoryStats(ctx context.Context, f StatsFilter) (HistoryStats, error)
TopTracks(ctx context.Context, f StatsFilter, limit int) ([]TrackPlayCount, error)
TopUsers(ctx context.Context, f StatsFilter, limit int) ([]UserPlayCount, error)
```

## Tasks

- [x] Add `StatsFilter` to `history/types.go`; update `Repository` interface.
- [x] Implement `since` filter on `history.MemoryRepository` (`played_at >= since`).
- [x] Implement `since` filter on `historypg.Repository` (`WHERE played_at >= $N`).
- [x] Update `GetHistoryStats`, `GetTopTracks`, `GetTopUsers` in `history.Service`.
- [x] Add `parseHistoryAdminFilter` helper in `handler.go`; thread filter through 3 handlers.
- [x] Add 3 service unit tests for windowed filtering (stats, top-tracks, top-users).
- [x] Add 2 HTTP-layer tests (`TestAdminHistorySinceFilter`, `TestAdminHistorySinceInvalid`).
- [x] Add `since` query param to 3 admin history GET paths in OpenAPI; add `invalid_since` error code; bump `info.version` to `0.70.0`.
- [x] Add `TestStorageAdminOpenAPIContractAdminHistorySinceParam`.
- [x] Extend `TestStorageAdminOpenAPIContractSchemasAndErrors` with `invalid_since`.
- [x] Run full `go test ./...` — green (440 tests).
- [x] Bump `requirement.md` version to `0.70.0`, append phase entry.
- [x] Commit, tag `v0.70.0`, push.

## Non-goals

- `until` / end-of-window param (only lower-bound supported).
- Per-user history admin (viewing another user's history).
- Track/user enrichment (names, titles) in stats responses.

## Follow-up candidates

- `until` upper-bound param for the stats window.
- Per-track/per-user history detail views for admins.
- User-editable playlists.
