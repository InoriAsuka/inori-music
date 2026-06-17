# Phase 71 — until upper-bound param for admin history stats window

## Goal

Complement Phase 70's `?since` lower bound with an optional `?until=<RFC3339>`
upper bound (exclusive) on the three admin aggregate history endpoints.
When both `since` and `until` are provided, validate that `since < until`.

## HTTP API changes

| Endpoint | New param |
|---|---|
| `GET /api/v1/admin/history/stats` | `?until=<RFC3339>` (exclusive upper bound) |
| `GET /api/v1/admin/history/top-tracks` | `?until=<RFC3339>` |
| `GET /api/v1/admin/history/top-users` | `?until=<RFC3339>` |

All three params are optional and compose with `since`. Error codes:
- `400 invalid_until` — unparseable value
- `400 invalid_time_range` — `since >= until` when both present

## Repository interface changes

`StatsFilter` gains `Until time.Time`:

```go
type StatsFilter struct {
    Since time.Time // optional lower bound on played_at (inclusive)
    Until time.Time // optional upper bound on played_at (exclusive)
}
```

PostgreSQL uses a shared `statsWhere(f)` helper that builds the `WHERE` clause
and args dynamically for any combination of since/until, eliminating four-branch
duplication.

## Tasks

- [x] Add `Until time.Time` to `StatsFilter` in `history/types.go`.
- [x] Add `until` guard (`played_at < until`) to `history.MemoryRepository` aggregate methods.
- [x] Replace per-method branch logic in `historypg.Repository` with `statsWhere` helper; add until clause.
- [x] Extend `parseHistoryAdminFilter` in `handler.go` to parse `until` and validate `since < until`.
- [x] Add 3 service unit tests (until-stats, since+until-tracks, until-combined).
- [x] Add 4 HTTP-layer tests (TestAdminHistoryUntilFilter, TestAdminHistoryUntilInvalid, TestAdminHistoryInvalidTimeRange, extended since test).
- [x] Add `until` query param to 3 admin history GET paths in OpenAPI; add `invalid_until` and `invalid_time_range` to error code enum; bump `info.version` to `0.71.0`.
- [x] Add `TestStorageAdminOpenAPIContractAdminHistoryUntilParam`.
- [x] Extend `TestStorageAdminOpenAPIContractSchemasAndErrors` with `invalid_until`, `invalid_time_range`.
- [x] Run full `go test ./...` — green (446 tests).
- [x] Bump `requirement.md` version to `0.71.0`, append phase entry.
- [x] Commit, tag `v0.71.0`, push.

## Non-goals

- Per-user history admin (viewing another user's history).
- Track/user enrichment (names, titles) in stats responses.

## Follow-up candidates

- Per-track/per-user history detail views for admins.
- User-editable playlists.
