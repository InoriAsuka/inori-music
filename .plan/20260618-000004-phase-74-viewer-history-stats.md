# Phase 74 — Viewer personal history stats

## Context

The history domain has rich admin-facing aggregate endpoints (phases 69–71) but viewers can
only list / record / clear their own events. Phase 74 closes that gap by giving each viewer
user-scoped equivalents: a personal stat summary and a personal top-tracks leaderboard.

This mirrors the admin aggregate API shape but restricts data to the authenticated user.
The clean approach uses a dedicated `UserStatsFilter` (carries `UserID` + optional time
bounds) rather than adding `UserID` to the existing `StatsFilter` — keeping the admin/viewer
interface boundary explicit, exactly as `PlayEventFilter` vs `AdminPlayEventFilter` does for
list queries.

## HTTP API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/me/history/stats` | viewer session | Personal stats: total events, unique tracks |
| `GET` | `/api/v1/me/history/top-tracks` | viewer session | Personal top tracks by play count |

Both accept optional `?since` / `?until` (RFC3339) time-window params, same validation as
admin endpoints. `top-tracks` also accepts optional `?limit` (default 10, max 100).

## New types — `history/types.go`

```go
// UserStatsFilter scopes aggregate queries to a single user's play events.
type UserStatsFilter struct {
    UserID string    // required — injected from auth context
    Since  time.Time // optional lower bound (inclusive)
    Until  time.Time // optional upper bound (exclusive)
}

// UserHistoryStats holds per-user playback aggregate counts.
// UniqueUsers is always 1 so it is omitted; only UniqueTracks is meaningful.
type UserHistoryStats struct {
    TotalEvents  int `json:"totalEvents"`
    UniqueTracks int `json:"uniqueTracks"`
}
```

## Repository interface — 2 new methods

```go
UserTopTracks(ctx context.Context, f UserStatsFilter, limit int) ([]TrackPlayCount, error)
UserHistoryStats(ctx context.Context, f UserStatsFilter) (UserHistoryStats, error)
```

`TrackPlayCount` is already defined and reusable.

## Implementations

**`history.MemoryRepository`** — scope the existing in-memory loops with `e.UserID == f.UserID`
plus the existing `Since`/`Until` guards. Same sort + slice pattern as `TopTracks`.

**`historypg.Repository`** — extend `statsWhere` to accept a `userID string` param (or add
a dedicated `userStatsWhere` helper) that adds `AND user_id = $N` when non-empty. The SQL
queries are structurally identical to `TopTracks`/`HistoryStats` with a mandatory user clause.

## Service — 2 new methods

```go
func (s *Service) GetMyTopTracks(ctx, f UserStatsFilter, limit int) ([]TrackPlayCount, error)
func (s *Service) GetMyStats(ctx, f UserStatsFilter) (UserHistoryStats, error)
```

Both clamp/default `limit` and validate `UserID != ""`.

## Handler

Two new handlers in the viewer history block of `handler.go`:

- **`getMyHistoryStats`** — `userFromContext` for `UserID`; reuse
  `parseHistoryAdminFilter` for `Since`/`Until`; call `historyService.GetMyStats`.
- **`getMyTopTracks`** — same plus `parseHistoryAdminLimit` for `?limit`;
  call `historyService.GetMyTopTracks`; respond `{"tracks": [...]}`.

Route registration (added alongside existing `/api/v1/me/history` group):
```
GET /api/v1/me/history/stats        → requireViewerAuth(getMyHistoryStats)
GET /api/v1/me/history/top-tracks   → requireViewerAuth(getMyTopTracks)
/api/v1/me/history/stats            → requireViewerAuth(methodNotAllowed)
/api/v1/me/history/top-tracks       → requireViewerAuth(methodNotAllowed)
```

## Tests

**`history/service_test.go`** — 3 new unit tests:
- `TestGetMyStats` — verifies totalEvents and uniqueTracks are user-scoped.
- `TestGetMyTopTracks` — verifies only the authenticated user's plays rank; other users excluded.
- `TestGetMyTopTracksTimeWindow` — verifies `Since`/`Until` filtering within user scope.

**`httpapi/handler_test.go`** — 4 new HTTP-layer tests:
- `TestGetMyHistoryStats` — 200, correct counts after recording plays.
- `TestGetMyTopTracks` — 200, tracks ranked by play count.
- `TestGetMyHistoryStatsTimeWindow` — `?since=` filters correctly.
- `TestGetMyHistoryStatsNotConfigured` — 503 `history_not_configured`.

## OpenAPI changes

- New schema `UserHistoryStats` with `totalEvents` and `uniqueTracks` integer properties.
- New path `GET /api/v1/me/history/stats` — tagged `History`, `bearerAuth`, optional
  `since`/`until` params; 200 response `$ref UserHistoryStats`.
- New path `GET /api/v1/me/history/top-tracks` — tagged `History`, `bearerAuth`, optional
  `limit`/`since`/`until` params; 200 response `{"tracks": TrackPlayCount[]}` (inline array
  with `$ref TrackPlayCount` items — same shape as admin top-tracks).
- Bump `info.version` to `0.74.0`.

## Contract tests

- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on both new paths.
- Add `TestStorageAdminOpenAPIContractViewerHistoryStatsPaths` asserting schema presence,
  param names, and response shapes.

## Non-goals

- Track/user name enrichment in responses.
- User-editable playlists.
- History export.

## Follow-up candidates

- Track/user name enrichment (inline `trackTitle` in `TrackPlayCount` for personal stats).
- User-editable playlists.
- History export (CSV/JSON download for the authenticated user).

## Tasks

- [ ] Add `UserStatsFilter`, `UserHistoryStats` to `history/types.go`.
- [ ] Add `UserTopTracks`, `UserHistoryStats` methods to `Repository` interface.
- [ ] Implement both on `history.MemoryRepository`.
- [ ] Implement both on `historypg.Repository` (add `userStatsWhere` helper).
- [ ] Add `GetMyTopTracks`, `GetMyStats` to `history.Service`.
- [ ] Add `getMyHistoryStats`, `getMyTopTracks` handlers + routes + 405 fallbacks.
- [ ] Add 3 service unit tests.
- [ ] Add 4 HTTP-layer tests.
- [ ] Update OpenAPI contract: `UserHistoryStats` schema, 2 new paths; bump `0.74.0`.
- [ ] Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add contract test.
- [ ] Run full `go test ./...` — green.
- [ ] Bump `requirement.md` to `0.74.0`, append phase entry.
- [ ] Commit, tag `v0.74.0`, push.
