# Phase 80 — since/until time-window on list endpoints

## Context

`GET /api/v1/admin/history` (Phase 75) and the three admin aggregate endpoints (phases 69–71)
all support `?since`/`?until`. However the three individual-event list endpoints
(`GET /api/v1/me/history`, `GET /api/v1/admin/history/users/{userId}`,
`GET /api/v1/admin/history/tracks/{trackId}`) were missing time-window filters,
making it impossible to paginate events within a time range on those endpoints.
Phase 80 closes this gap, bringing all four paginated list endpoints to parity.

## Affected endpoints

| Path | Added params |
|------|-------------|
| `GET /api/v1/me/history` | `?since`, `?until` |
| `GET /api/v1/admin/history/users/{userId}` | `?since`, `?until` |
| `GET /api/v1/admin/history/tracks/{trackId}` | `?since`, `?until` |

Existing behaviour is unchanged when neither param is supplied.

## Type changes — `history/types.go`

```go
// Added to PlayEventFilter and AdminPlayEventFilter:
Since   time.Time // optional lower bound on played_at (inclusive)
Until   time.Time // optional upper bound on played_at (exclusive)
```

## Repository changes

**`history.MemoryRepository`** — `ListPlayEvents` and `ListPlayEventsByTrack`
each gain two `Since`/`Until` guards mirroring the existing pattern in
`ListAllPlayEvents`.

**`historypg.Repository`** — Both `ListPlayEvents` and `ListPlayEventsByTrack`
are refactored from a two-branch (`TrackID`/no-`TrackID`) approach to a single
unified dynamic `WHERE` clause builder, accepting all four optional filter fields.

## Handler changes

`listPlayEvents`, `getAdminUserHistory`, and `getAdminTrackHistory` each gain:

```go
tf, ok := parseHistoryAdminFilter(w, r)  // reuse existing helper
if !ok { return }
// … pass tf.Since and tf.Until to the filter struct
```

## OpenAPI changes

- Add `since` (string/date-time, optional, inclusive) and `until` (string/date-time,
  optional, exclusive) query params to the three GET operations.
- Bump `info.version` to `0.80.0`.

## Tests

**`history/service_test.go`** — 3 unit tests:
`TestListPlaysSinceFilter`, `TestListPlaysUntilFilter`, `TestGetUserHistorySinceFilter`.

**`httpapi/handler_test.go`** — 4 HTTP-layer tests:
`TestListPlayEventsSinceFilter`, `TestListPlayEventsUntilFilter`,
`TestAdminUserHistorySinceUntilFilter`, `TestAdminTrackHistorySinceFilter`.

## Contract tests

`TestStorageAdminOpenAPIContractListSinceUntilParams` — asserts `since` and `until`
on all three paths.

## Non-goals

- Adding `since`/`until` to `batch-delete` endpoints.
- Client-side caching hints.

## Follow-up candidates

- Cursor-based pagination using `(played_at, id)` as a stable cursor.
- `trackId` filter on `GET /api/v1/admin/history` (already there via GlobalPlayEventFilter).

## Tasks

- [x] Add `Since`, `Until` to `PlayEventFilter` and `AdminPlayEventFilter`.
- [x] Update `MemoryRepository.ListPlayEvents` and `ListPlayEventsByTrack` guards.
- [x] Refactor `historypg.Repository.ListPlayEvents` and `ListPlayEventsByTrack` to dynamic WHERE.
- [x] Thread `Since`/`Until` into 3 handlers via `parseHistoryAdminFilter`.
- [x] Add 3 service unit tests.
- [x] Add 4 HTTP-layer tests.
- [x] Update OpenAPI: `since`/`until` on 3 paths; bump `0.80.0`.
- [x] Add `TestStorageAdminOpenAPIContractListSinceUntilParams`.
- [x] Run `go test ./services/api/...` — 526 passed.
- [x] Update `requirement.md` to `0.80.0`, append phase entry.
- [x] Bump `VERSION` to `0.80.0`.
- [ ] Commit, tag `v0.80.0`, push.
