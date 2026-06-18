# Phase 75 — Admin global play-event list

## Context

Phases 68–74 built a rich history domain: record/list/clear per-user events, admin per-user
and per-track detail views, admin bulk-delete, admin aggregate stats, and viewer stats. However
the `DELETE /api/v1/admin/history` path has no `GET` counterpart — admins cannot browse the raw
event table without knowing a specific userId or trackId upfront.

Phase 75 closes this gap by adding `GET /api/v1/admin/history` as a cross-cutting global list
endpoint that accepts optional filters and returns a `PlayEventList` in the same paginated shape
used by all other history list endpoints.

## HTTP API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/admin/history` | admin Bearer token | Global paginated list of all play events |

### Query Parameters

| Name | Type | Required | Default | Notes |
|------|------|----------|---------|-------|
| `userId` | string | no | — | Filter to events for a specific user |
| `trackId` | string | no | — | Filter to events for a specific track |
| `since` | RFC3339 | no | — | Include events at or after this time (inclusive) |
| `until` | RFC3339 | no | — | Include events before this time (exclusive) |
| `limit` | integer 1–500 | no | 50 | Max events to return |
| `offset` | integer ≥ 0 | no | 0 | Events to skip |

### Response (200)

```json
{
  "events": [ PlayEvent, … ],
  "pagination": { "limit": 50, "offset": 0, "total": 123, "hasMore": true }
}
```

## New type — `history/types.go`

```go
// GlobalPlayEventFilter scopes admin queries that list all events across every user and track.
// All fields are optional filters; zero values mean "no restriction".
type GlobalPlayEventFilter struct {
    UserID  string    // optional
    TrackID string    // optional
    Since   time.Time // optional, inclusive
    Until   time.Time // optional, exclusive
    Limit   int       // 0 → default (50); clamped to 500
    Offset  int
}
```

## Repository interface — 1 new method

```go
ListAllPlayEvents(ctx context.Context, f GlobalPlayEventFilter) ([]PlayEvent, int, error)
```

## Implementations

**`history.MemoryRepository`** — iterates all events applying each non-zero filter field as a
guard, then sorts by `played_at DESC, id ASC` and applies `Offset`/`Limit` slicing.

**`historypg.Repository`** — builds a dynamic `WHERE` clause from non-zero filter fields
(identical pattern to `statsWhere`), with `$1=limit`, `$2=offset`, and filter args at `$3+`.
Uses `COUNT(*) OVER()` window function for pagination total in a single query.

## Service — 1 new method

```go
func (s *Service) GetAllHistory(ctx, f GlobalPlayEventFilter) ([]PlayEvent, int, error)
```

Clamps/defaults limit. No required fields (unfiltered query is valid).

## Handler

`getAdminAllHistory` in `handler.go`:
- `requireHistoryService` guard
- `parseHistoryAdminFilter` for `Since`/`Until`
- `parseHistoryAdminPagination` for `Limit`/`Offset`
- Raw query params `userId`/`trackId` from `r.URL.Query()`
- Calls `historyService.GetAllHistory`; responds `{events, pagination}`

Route:
```
GET /api/v1/admin/history  →  requireAdminAuth(getAdminAllHistory)
```
(Existing `DELETE` and wildcard fallback routes are unchanged.)

## Tests

**`history/service_test.go`** — 3 new unit tests:
- `TestGetAllHistory` — no filter → all events returned.
- `TestGetAllHistoryUserFilter` — userId filter scopes to one user.
- `TestGetAllHistoryTimeWindow` — since/until window isolates correct events.

**`httpapi/handler_test.go`** — 4 new HTTP-layer tests:
- `TestAdminGetAllHistory` — 200, correct total after recording events.
- `TestAdminGetAllHistoryTrackFilter` — `?trackId=` reduces total correctly.
- `TestAdminGetAllHistoryNotConfigured` — 503 `history_not_configured`.
- `TestAdminGetAllHistoryMethodNotAllowed` — POST → 405.

## OpenAPI changes

- Add `get` operation to `/api/v1/admin/history` with 6 query params and
  `PlayEventList` response schema ref.
- Bump `info.version` to `0.75.0`.

## Contract tests

- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history`.
- Add `TestStorageAdminOpenAPIContractAdminHistoryGlobalList` asserting all 6 params and
  `PlayEventList` response schema ref.

## Non-goals

- Enrichment of event payloads with track title or user name.
- CSV/JSON export.
- Server-side sort field selection.

## Follow-up candidates

- Sort control (`?sort=playedAt|createdAt`).
- CSV/JSON history export.
- Track/user name enrichment for richer admin views.

## Tasks

- [x] Add `GlobalPlayEventFilter` to `history/types.go`.
- [x] Add `ListAllPlayEvents` to `Repository` interface.
- [x] Implement on `history.MemoryRepository`.
- [x] Implement on `historypg.Repository`.
- [x] Add `GetAllHistory` to `history.Service`.
- [x] Add `getAdminAllHistory` handler + `GET /api/v1/admin/history` route.
- [x] Add 3 service unit tests.
- [x] Add 4 HTTP-layer tests.
- [x] Update OpenAPI contract: `get` on `/api/v1/admin/history`; bump `0.75.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add contract test.
- [x] Run full `go test ./services/api/...` — green (478 passed).
- [x] Bump `requirement.md` to `0.75.0`, append phase entry.
- [x] Bump `VERSION` to `0.75.0`.
- [ ] Commit, tag `v0.75.0`, push.
