# Phase 73 — Admin bulk-delete play event history

## Goal

Add admin endpoints to bulk-delete play events scoped by user, track, or time window,
enabling admins to prune history data without requiring the target user to self-delete.

## HTTP API

| Method   | Path                                       | Description                                              |
|----------|--------------------------------------------|----------------------------------------------------------|
| `DELETE` | `/api/v1/admin/history/users/{userId}`     | Delete all play events for a specific user               |
| `DELETE` | `/api/v1/admin/history/tracks/{trackId}`   | Delete all play events for a specific track (all users)  |
| `DELETE` | `/api/v1/admin/history`                    | Delete play events matching a time-window filter (`?since` / `?until`; at least one required) |

Each endpoint returns `204 No Content` on success.
`DELETE /api/v1/admin/history` without at least one time bound returns `400 missing_time_filter`.

## Repository changes

Three new methods added to `history.Repository`:

```go
DeletePlayEventsByUserAdmin(ctx context.Context, userID string) error
DeletePlayEventsByTrack(ctx context.Context, trackID string) error
DeletePlayEventsInWindow(ctx context.Context, f StatsFilter) error  // StatsFilter reused; at least one bound required
```

- `MemoryRepository`: in-memory iteration + delete under lock.
- `historypg.Repository`: single `DELETE FROM play_events WHERE …` with `statsWhere(f)` reused for window deletion.

## Service changes

Three new methods on `history.Service`:
```go
AdminDeleteUserHistory(ctx, userID string) error
AdminDeleteTrackHistory(ctx, trackID string) error
AdminDeleteHistoryWindow(ctx, f StatsFilter) error  // validates at least one bound
```

## HTTP handler changes

- `deleteAdminUserHistory`: `DELETE /api/v1/admin/history/users/{userId}` → 204
- `deleteAdminTrackHistory`: `DELETE /api/v1/admin/history/tracks/{trackId}` → 204
- `deleteAdminHistoryWindow`: `DELETE /api/v1/admin/history` → 204; reuses `parseHistoryAdminFilter`; requires at least one bound
- Register routes + `methodNotAllowed` fallbacks under `requireAdminAuth`
- Existing GET handlers for `users/{userId}` and `tracks/{trackId}` are unaffected; the mux distinguishes method.

## Tests

Service unit tests (3):
- `TestAdminDeleteUserHistory` — deletes only target user's events, leaves others intact.
- `TestAdminDeleteTrackHistory` — deletes only target track's events, leaves others intact.
- `TestAdminDeleteHistoryWindow` — validates no-bound rejection; validates since-only, until-only, and both-bounds deletion.

HTTP-layer tests (5):
- `TestAdminDeleteUserHistory` — 204, events gone for that user.
- `TestAdminDeleteTrackHistory` — 204, events gone for that track.
- `TestAdminDeleteHistoryWindow` — 204 with `?since=`, events outside window preserved.
- `TestAdminDeleteHistoryWindowMissingFilter` — 400 `missing_time_filter` when no bounds given.
- `TestAdminBulkDeleteHistoryNotConfigured` — 503 `history_not_configured`.

## OpenAPI changes

- Add `delete` operations on `/api/v1/admin/history/users/{userId}` and `/api/v1/admin/history/tracks/{trackId}` paths.
- Add new path `/api/v1/admin/history` with `delete` operation; `since`/`until` query params (optional individually, but at least one required at runtime).
- Add `missing_time_filter` to error code enum.
- Bump `info.version` to `0.73.0`.

## Contract test changes

- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `delete` on both `users/{userId}` and `tracks/{trackId}`.
- Add `TestStorageAdminOpenAPIContractAdminHistoryBulkDelete` asserting the new `/api/v1/admin/history` DELETE path, `since`/`until` params, and `missing_time_filter` error code.

## Non-goals

- Track/user name enrichment in history list responses.
- User-editable playlists.
- Soft-delete or audit log for deleted events.

## Follow-up candidates

- Track/user name enrichment in history responses.
- User-editable playlists.
- History export (CSV/JSON download).

## Tasks

- [x] Add `DeletePlayEventsByUserAdmin`, `DeletePlayEventsByTrack`, `DeletePlayEventsInWindow` to `history.Repository` interface.
- [x] Implement all three on `history.MemoryRepository`.
- [x] Implement all three on `historypg.Repository` (`DELETE … WHERE` with `statsWhere` reuse for window).
- [x] Add `AdminDeleteUserHistory`, `AdminDeleteTrackHistory`, `AdminDeleteHistoryWindow` to `history.Service`.
- [x] Add `deleteAdminUserHistory`, `deleteAdminTrackHistory`, `deleteAdminHistoryWindow` handlers.
- [x] Register 3 routes + 3 `methodNotAllowed` fallbacks in `handler.go`.
- [x] Add 3 service unit tests.
- [x] Add 5 HTTP-layer tests.
- [x] Update OpenAPI contract: 3 new `delete` operations, `missing_time_filter` error code; bump `info.version` to `0.73.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` for new delete methods.
- [x] Add `TestStorageAdminOpenAPIContractAdminHistoryBulkDelete`.
- [x] Run full `go test ./...` — green (462 tests).
- [x] Bump `requirement.md` version to `0.73.0`, append phase entry.
- [ ] Commit, tag `v0.73.0`, push.
