# Phase 77 — Per-event fetch and delete (admin + viewer)

## Context

The history domain had rich bulk and aggregate endpoints but no way to reference
a single event by ID. Admins could not verify or remove a specific audit entry
without clearing an entire user's or track's history. Viewers could not delete a
single mistaken record without clearing all history. Phase 77 closes both gaps by
adding `GET` and `DELETE` on a new `{eventId}` path segment under both the admin
and viewer history prefixes.

## HTTP API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/admin/history/{eventId}` | admin Bearer | Fetch any play event by ID |
| `DELETE` | `/api/v1/admin/history/{eventId}` | admin Bearer | Delete any play event by ID |
| `GET` | `/api/v1/me/history/{eventId}` | viewer session | Fetch own play event by ID |
| `DELETE` | `/api/v1/me/history/{eventId}` | viewer session | Delete own play event by ID |

Viewer endpoints perform an ownership check: if the event exists but belongs to
a different user the response is `403 event_forbidden`.

## New errors — `history/types.go`

```go
var ErrEventNotFound = errors.New("play event not found")
var ErrEventForbidden = errors.New("play event belongs to another user")
```

## Repository interface — 2 new methods

```go
GetPlayEventByID(ctx context.Context, id string) (PlayEvent, error)
DeletePlayEventByID(ctx context.Context, id string) error
```

Both return `ErrEventNotFound` when the ID does not exist.

## Implementations

**`history.MemoryRepository`** — map lookup under RLock/Lock; `ErrEventNotFound`
when absent. Delete atomically checks then removes.

**`historypg.Repository`** — `SELECT … WHERE id = $1` (scan → `ErrNoRows` →
`ErrEventNotFound`); `DELETE … WHERE id = $1` (check `RowsAffected() == 0`).

## Service — 4 new methods

| Method | Scope | Notes |
|--------|-------|-------|
| `GetEventByID(ctx, id)` | admin | no ownership check |
| `DeleteEventByID(ctx, id)` | admin | no ownership check |
| `GetMyEvent(ctx, userID, id)` | viewer | returns `ErrEventForbidden` if `event.UserID != userID` |
| `DeleteMyEvent(ctx, userID, id)` | viewer | fetch + ownership check before delete |

## Handler

4 new handlers: `getAdminEvent`, `deleteAdminEvent`, `getMyEvent`, `deleteMyEvent`.

`writeError` extended:
- `ErrEventNotFound` → `404 not_found`
- `ErrEventForbidden` → `403 event_forbidden`

Route registration order: fixed paths (`/stats`, `/top-tracks`, `/users/{...}`,
`/tracks/{...}`) registered before `/{eventId}` wildcards to avoid Go `ServeMux`
pattern-overlap panics.

## Tests

**`history/service_test.go`** — 5 unit tests:
`TestGetEventByID`, `TestGetEventByIDNotFound`, `TestDeleteEventByID`,
`TestGetMyEvent`, `TestDeleteMyEvent`.

**`httpapi/handler_test.go`** — 7 HTTP-layer tests:
`TestAdminGetEvent`, `TestAdminGetEventNotFound`, `TestAdminDeleteEvent`,
`TestViewerGetEvent`, `TestViewerGetEventNotOwned`, `TestViewerDeleteEvent`,
`TestPerEventHistoryNotConfigured`.

## OpenAPI changes

- Add `GET`/`DELETE` to `/api/v1/admin/history/{eventId}` and
  `/api/v1/me/history/{eventId}`; 200 responses use `$ref PlayEvent`.
- Add `event_forbidden` to `ErrorEnvelope.error.code` enum.
- Bump `info.version` to `0.77.0`.

## Contract tests

- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get`/`delete` on
  both new paths.
- Add `TestStorageAdminOpenAPIContractPerEventPaths` asserting `PlayEvent` 200
  schema, path params, and `event_forbidden` error code.

## Non-goals

- Editing/patching a play event's `playedAt` timestamp.
- Viewer listing another user's events.

## Follow-up candidates

- `PATCH /api/v1/me/history/{eventId}` to correct `playedAt`.
- Batch delete by IDs.

## Tasks

- [x] Add `ErrEventNotFound`, `ErrEventForbidden` to `history/types.go`.
- [x] Add `GetPlayEventByID`, `DeletePlayEventByID` to Repository interface.
- [x] Implement on `MemoryRepository`.
- [x] Implement on `historypg.Repository`.
- [x] Add `GetEventByID`, `DeleteEventByID`, `GetMyEvent`, `DeleteMyEvent` to Service.
- [x] Map errors in `writeError`.
- [x] Add 4 routes + handlers; fix registration order for ServeMux.
- [x] Add 5 service unit tests.
- [x] Add 7 HTTP-layer tests.
- [x] Update OpenAPI; bump `0.77.0`.
- [x] Add contract test `TestStorageAdminOpenAPIContractPerEventPaths`.
- [x] Run `go test ./services/api/...` — 498 passed.
- [x] Update `requirement.md` to `0.77.0`, append phase entry.
- [x] Bump `VERSION` to `0.77.0`.
- [ ] Commit, tag `v0.77.0`, push.
