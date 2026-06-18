# Phase 78 — PATCH per-event playedAt correction

## Context

Phase 77 added `GET` and `DELETE` on per-event paths but no mutation. Viewers
sometimes record a play with the wrong timestamp (e.g., the client supplied a
stale cached time). Phase 78 closes the gap by adding `PATCH` on both admin and
viewer per-event paths to allow correcting `playedAt`.

Only `playedAt` is mutable. `userId`, `trackId`, and `createdAt` are write-once.

## HTTP API

| Method | Path | Auth | Body | Description |
|--------|------|------|------|-------------|
| `PATCH` | `/api/v1/admin/history/{eventId}` | admin Bearer | `{"playedAt":"<RFC3339>"}` | Update any event's playedAt |
| `PATCH` | `/api/v1/me/history/{eventId}` | viewer session | `{"playedAt":"<RFC3339>"}` | Update own event's playedAt |

Both return the full updated `PlayEvent` on `200`.

### Validation

- Missing or empty `playedAt` → `400 invalid_played_at`
- Non-RFC3339 `playedAt` → `400 invalid_played_at`
- Event not found → `404 not_found`
- Event belongs to another user (viewer only) → `403 event_forbidden`

## Repository interface — 1 new method

```go
UpdatePlayEventByID(ctx context.Context, id string, playedAt time.Time) (PlayEvent, error)
```

Returns `ErrEventNotFound` when the ID does not exist.

## Implementations

**`history.MemoryRepository`** — Lock, map lookup (miss → `ErrEventNotFound`),
update `e.PlayedAt = playedAt.UTC()`, store back, return updated event.

**`historypg.Repository`** — `UPDATE play_events SET played_at = $2 WHERE id = $1 RETURNING …`;
on `ErrNoRows` → `ErrEventNotFound`.

## Service — 2 new methods

```go
UpdateEventByID(ctx, id string, playedAt time.Time) (PlayEvent, error)  // admin
UpdateMyEvent(ctx, userID, id string, playedAt time.Time) (PlayEvent, error)  // viewer
```

`UpdateMyEvent`: `GetPlayEventByID` → ownership check → `UpdatePlayEventByID`.

## OpenAPI changes

- Add `UpdatePlayEventRequest` schema `{playedAt: string/date-time, required}`.
- Add `patch` operation to `/api/v1/admin/history/{eventId}` and `/api/v1/me/history/{eventId}`;
  `requestBody` references `UpdatePlayEventRequest`; 200 response references `PlayEvent`.
- Add `invalid_played_at` to `ErrorEnvelope.error.code` enum.
- Bump `info.version` to `0.78.0`.

## Tests

**`history/service_test.go`** — 3 unit tests:
`TestUpdateEventByID`, `TestUpdateEventByIDNotFound`, `TestUpdateMyEvent`.

**`httpapi/handler_test.go`** — 7 HTTP-layer tests:
`TestAdminPatchEvent`, `TestAdminPatchEventNotFound`, `TestAdminPatchEventInvalidPlayedAt`,
`TestViewerPatchEvent`, `TestViewerPatchEventInvalidPlayedAt`,
`TestViewerPatchEventMissingPlayedAt`, `TestPatchEventHistoryNotConfigured`.

## Contract tests

- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `patch` on both paths.
- Extend `TestStorageAdminOpenAPIContractPerEventPaths` to assert `PATCH` operations,
  `UpdatePlayEventRequest` requestBody ref, and `invalid_played_at` error code.

## Non-goals

- Updating `trackId` or `userId`.
- Partial patch for fields other than `playedAt`.

## Follow-up candidates

- Batch update `playedAt` for multiple events.
- Admin-forced transfer of event ownership (`userId`).

## Tasks

- [x] Add `UpdatePlayEventByID` to Repository interface.
- [x] Implement on `MemoryRepository` (added `time` import).
- [x] Implement on `historypg.Repository` (added `time` import).
- [x] Add `UpdateEventByID` and `UpdateMyEvent` to Service.
- [x] Add `PATCH` routes and `patchAdminEvent`/`patchMyEvent` handlers.
- [x] Add 3 service unit tests.
- [x] Add 7 HTTP-layer tests.
- [x] Update OpenAPI: `UpdatePlayEventRequest` schema, `patch` on both paths, `invalid_played_at` enum; bump `0.78.0`.
- [x] Extend contract tests.
- [x] Run `go test ./services/api/...` — 508 passed.
- [x] Update `requirement.md` to `0.78.0`, append phase entry.
- [x] Bump `VERSION` to `0.78.0`.
- [ ] Commit, tag `v0.78.0`, push.
