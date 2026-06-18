# Phase 76 — Sort order parameter for history list endpoints

## Context

All play-event list endpoints previously returned results in a hard-coded
`played_at DESC` order (newest first). Clients that want to replay history
chronologically, build timelines, or export in chronological order had no
way to request ascending order. Phase 76 adds an optional `?order=asc|desc`
parameter (default `desc`) to all four list endpoints.

## Affected endpoints

| Method | Path |
|--------|------|
| `GET` | `/api/v1/me/history` |
| `GET` | `/api/v1/admin/history/users/{userId}` |
| `GET` | `/api/v1/admin/history/tracks/{trackId}` |
| `GET` | `/api/v1/admin/history` |

## New query parameter

| Name | Type | Required | Values | Default |
|------|------|----------|--------|---------|
| `order` | string enum | no | `asc`, `desc` | `desc` |

Returns `400 invalid_order` for any value other than `"asc"` or `"desc"`.

## Type changes — `history/types.go`

```go
// Added to PlayEventFilter, AdminPlayEventFilter, and GlobalPlayEventFilter:
Asc bool   // false → played_at DESC (default); true → played_at ASC
```

## Repository changes

**`history.MemoryRepository`** — `ListPlayEvents`, `ListPlayEventsByTrack`,
`ListAllPlayEvents`: sort comparators updated to branch on `f.Asc`.
Tie-breaking by `id` also flips direction for deterministic ordering.

**`historypg.Repository`** — new `eventOrder(asc bool) string` helper
returns either `"played_at ASC, id ASC"` or `"played_at DESC, id DESC"`.
All three list methods call `eventOrder(f.Asc)` in their `ORDER BY` clause.

## Handler changes

New `parseHistoryOrder(w, r) (asc bool, ok bool)` helper in `handler.go`:
- `?order=""` or omitted → `desc` (false, true)
- `?order=asc` → (true, true)
- `?order=desc` → (false, true)
- anything else → `400 invalid_order` (false, false)

All four handlers call `parseHistoryOrder`; the resulting `Asc` value is
threaded into the filter struct passed to the service.

## Tests

**`history/service_test.go`** — 2 new unit tests:
- `TestListPlaysAscOrder` — verifies desc returns newest-first, asc oldest-first.
- `TestGetAllHistoryAscOrder` — same for the global list.

**`httpapi/handler_test.go`** — 4 new HTTP-layer tests:
- `TestListPlayEventsAscOrder` — `GET /api/v1/me/history?order=asc` returns oldest event first.
- `TestListPlayEventsInvalidOrder` — `?order=random` → 400 `invalid_order`.
- `TestAdminGetAllHistoryAscOrder` — `GET /api/v1/admin/history?order=asc` returns oldest first.
- `TestAdminGetAllHistoryInvalidOrder` — `?order=newest` → 400 `invalid_order`.

## OpenAPI changes

- Add `order` query param (string enum `["asc","desc"]`, optional, default `"desc"`)
  to all four affected `GET` operations.
- Add `invalid_order` to `ErrorEnvelope.error.code` enum.
- Bump `info.version` to `0.76.0`.

## Contract tests

- Add `TestStorageAdminOpenAPIContractHistoryOrderParam` asserting the `order`
  param is present on all four paths and `invalid_order` is in the error enum.

## Non-goals

- Sort by fields other than `played_at` (e.g., `createdAt`, `trackId`).
- Per-page sort changes (the order applies to the entire logical result set).

## Follow-up candidates

- Sort by additional fields (`createdAt`).
- Cursor-based pagination using `played_at + id` as a stable cursor.

## Tasks

- [x] Add `Asc bool` to `PlayEventFilter`, `AdminPlayEventFilter`, `GlobalPlayEventFilter`.
- [x] Update `MemoryRepository` sort comparators.
- [x] Add `eventOrder` helper; update `historypg.Repository` ORDER BY clauses.
- [x] Add `parseHistoryOrder` helper to `handler.go`.
- [x] Thread `Asc` through all four list handlers.
- [x] Add 2 service unit tests.
- [x] Add 4 HTTP-layer tests.
- [x] Update OpenAPI: `order` param on 4 paths, `invalid_order` in error enum; bump `0.76.0`.
- [x] Add `TestStorageAdminOpenAPIContractHistoryOrderParam`.
- [x] Run `go test ./services/api/...` — 485 passed.
- [x] Update `requirement.md` to `0.76.0`, append phase entry.
- [x] Bump `VERSION` to `0.76.0`.
- [ ] Commit, tag `v0.76.0`, push.
