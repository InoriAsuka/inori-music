# Phase 72 — Admin per-user and per-track history detail views

## Goal

Add two admin-only endpoints that expose paginated play event lists scoped to
a specific user or a specific track, allowing admins to inspect individual
listening histories without needing to be that user.

## HTTP API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/admin/history/users/{userId}` | Paginated play events for any user; optional `?trackId` filter |
| `GET` | `/api/v1/admin/history/tracks/{trackId}` | Paginated play events for any track; optional `?userId` filter |

Both endpoints support `?limit` (1–500, default 50) and `?offset` pagination.
Response shape is identical to `GET /api/v1/me/history`: `{events, pagination}`.

## Repository change

New `ListPlayEventsByTrack(ctx, AdminPlayEventFilter)` method on `Repository`:
- `AdminPlayEventFilter` carries `TrackID` (required), optional `UserID`, `Limit`, `Offset`.
- `GetUserHistory` in `Service` reuses the existing `ListPlayEvents` (already accepts any `userID`).
- PostgreSQL uses a window-function `COUNT(*) OVER()` for stable pagination totals.

## Tasks

- [x] Add `AdminPlayEventFilter` to `history/types.go`.
- [x] Add `ListPlayEventsByTrack` to `Repository` interface.
- [x] Implement `ListPlayEventsByTrack` on `history.MemoryRepository`.
- [x] Implement `ListPlayEventsByTrack` on `historypg.Repository` (SQL with optional user filter).
- [x] Add `GetUserHistory` and `GetTrackHistory` to `history.Service`.
- [x] Register 2 routes + 2 `methodNotAllowed` fallbacks in `handler.go`.
- [x] Add `getAdminUserHistory`, `getAdminTrackHistory`, `parseHistoryAdminPagination` helpers.
- [x] Add 2 service unit tests (`TestGetUserHistory`, `TestGetTrackHistory`).
- [x] Add 4 HTTP-layer tests (user history, track history, 405, 503 not-configured).
- [x] Add 2 new paths to OpenAPI with `PlayEventList` response ref; bump `info.version` to `0.72.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` for 2 new paths.
- [x] Add `TestStorageAdminOpenAPIContractAdminHistoryDetailPaths`.
- [x] Run full `go test ./...` — green (453 tests).
- [x] Bump `requirement.md` version to `0.72.0`, append phase entry.
- [x] Commit, tag `v0.72.0`, push.

## Non-goals

- Track/user enrichment (names, titles) in responses.
- User-editable playlists.

## Follow-up candidates

- User-editable playlists.
- Track/user name enrichment in history responses.
- Admin bulk-delete history (by user, by track, or by time window).
