# Phase 79 — Batch delete play events by IDs

## Context

Single-event `DELETE` (Phase 77) requires one HTTP round-trip per event. Clients
that want to purge a selection of specific events (e.g., from a UI checkbox list)
need a batched operation. Phase 79 adds `POST /api/v1/admin/history/batch-delete`
and `POST /api/v1/me/history/batch-delete` to cover this gap.

Admins can delete any set of events; viewer batch-delete silently skips events that
do not belong to the authenticated user (same ownership semantics as `DeleteMyEvent`).

## HTTP API

| Method | Path | Auth | Body | Description |
|--------|------|------|------|-------------|
| `POST` | `/api/v1/admin/history/batch-delete` | admin Bearer | `{"ids":[…]}` | Delete any events by ID list |
| `POST` | `/api/v1/me/history/batch-delete` | viewer session | `{"ids":[…]}` | Delete own events by ID list |

Both return `{"deleted": N}` where N = events actually removed.

### Constraints

- `ids` must be a non-empty array of strings.
- Maximum 100 IDs per request; returns `400 invalid_ids` if exceeded.
- Unknown or already-deleted IDs are silently ignored (not an error).
- Viewer endpoint silently skips IDs owned by other users.

## Repository interface — 2 new methods

```go
DeletePlayEventsByIDs(ctx context.Context, ids []string) (int, error)
DeletePlayEventsByIDsForUser(ctx context.Context, userID string, ids []string) (int, error)
```

Both return the count of deleted rows. Empty `ids` returns `0, nil`.

## Implementations

**`history.MemoryRepository`** — Lock + range-delete; unknown IDs silently skipped.

**`historypg.Repository`** — `DELETE … WHERE id = ANY($1)` and `DELETE … WHERE id = ANY($1) AND user_id = $2`; `tag.RowsAffected()` as deleted count.

## Service

`MaxBatchDeleteIDs = 100` constant.

| Method | Scope | Validation |
|--------|-------|------------|
| `BatchDeleteEvents(ctx, ids)` | admin | len ≥ 1, len ≤ 100 |
| `BatchDeleteMyEvents(ctx, userID, ids)` | viewer | same + userID non-empty |

## OpenAPI changes

- `BatchDeleteRequest` schema: `{ids: string[], minItems:1, maxItems:100}`.
- `BatchDeleteResult` schema: `{deleted: integer}`.
- `POST /api/v1/admin/history/batch-delete` and `POST /api/v1/me/history/batch-delete`.
- `invalid_ids` added to error code enum.
- `info.version` bumped to `0.79.0`.

## Tests

**`history/service_test.go`** — 4 unit tests:
`TestBatchDeleteEvents`, `TestBatchDeleteEventsUnknownIDsIgnored`,
`TestBatchDeleteMyEvents`, `TestBatchDeleteEventsEmpty`.

**`httpapi/handler_test.go`** — 5 HTTP-layer tests:
`TestAdminBatchDeleteEvents`, `TestAdminBatchDeleteEventsEmptyBody`,
`TestViewerBatchDeleteMyEvents`, `TestViewerBatchDeleteSkipsOtherUsersEvents`,
`TestBatchDeleteHistoryNotConfigured`.

## Contract tests

- Extend `TestStorageAdminOpenAPIContractCoversRoutes`.
- Add `TestStorageAdminOpenAPIContractBatchDelete`.

## Non-goals

- Returning which IDs were actually deleted.
- Partial-failure mode (some deleted, some errored).

## Follow-up candidates

- Return `{deleted: N, notFound: [...ids]}` for audit.
- Admin-forced transfer of event ownership.

## Tasks

- [x] Add `DeletePlayEventsByIDs`, `DeletePlayEventsByIDsForUser` to Repository interface.
- [x] Implement on `MemoryRepository`.
- [x] Implement on `historypg.Repository`.
- [x] Add `MaxBatchDeleteIDs`, `BatchDeleteEvents`, `BatchDeleteMyEvents` to Service.
- [x] Add routes + `batchDeleteAdminEvents` / `batchDeleteMyEvents` handlers.
- [x] Add 4 service unit tests.
- [x] Add 5 HTTP-layer tests.
- [x] Update OpenAPI; bump `0.79.0`.
- [x] Add `TestStorageAdminOpenAPIContractBatchDelete`.
- [x] Run `go test ./services/api/...` — 518 passed.
- [x] Update `requirement.md` to `0.79.0`, append phase entry.
- [x] Bump `VERSION` to `0.79.0`.
- [ ] Commit, tag `v0.79.0`, push.
