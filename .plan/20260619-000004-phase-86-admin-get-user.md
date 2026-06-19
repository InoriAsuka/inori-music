# Phase 86 — Admin single user fetch GET /admin/users/{id}

## Context

The admin API can list all users and disable/delete them, but has no way to fetch a single user
by ID. This is the standard CRUD complement — `GET /api/v1/admin/users/{id}` returns `UserView`.
`auth.Service.GetUser` was already added in Phase 85.

## Affected endpoints

| Path | Method | Notes |
|------|--------|-------|
| `GET /api/v1/admin/users/{id}` | GET | Returns `UserView` for any user by ID |

## Handler changes

Add `getAdminUser(w, r)`: reads path value `id`, calls `authService.GetUser(ctx, id)`, writes `UserView`.

Register (the `methodNotAllowed` catch-all for `/api/v1/admin/users/{id}` already exists from
the `DELETE` handler; the `GET` operation is just added alongside it):
```
GET /api/v1/admin/users/{id}  → requireAdminAuth(getAdminUser)
```

## OpenAPI changes

Add `get` operation to the existing `/api/v1/admin/users/{id}` path item.
Bump `info.version` to `0.86.0`.

## Tests

**`httpapi/handler_test.go`** — 3 HTTP-layer tests:
`TestAdminGetUser`, `TestAdminGetUserNotFound`, `TestAdminGetUserNotConfigured`.

**`httpapi/openapi_contract_test.go`** — extend `TestStorageAdminOpenAPIContractCoversRoutes`
with `get` on `/api/v1/admin/users/{id}`; add `TestStorageAdminOpenAPIContractAdminGetUser`.

## Non-goals

- Viewer fetch own user by ID (already covered by GET /me from Phase 85)
- Patch/update user (Phase 89)

## Follow-up candidates

- Admin enable user (Phase 88)
- Admin patch user (Phase 89)

## Tasks

- [ ] Add `getAdminUser` handler.
- [ ] Register `GET /api/v1/admin/users/{id}`.
- [ ] Add 3 HTTP-layer tests.
- [ ] Update OpenAPI: add `get` to `/api/v1/admin/users/{id}`; bump `0.86.0`.
- [ ] Extend contract tests.
- [ ] Run `go test ./services/api/...` — all pass.
- [ ] Update `requirement.md` to `0.86.0`, append phase entry.
- [ ] Bump `VERSION` to `0.86.0`.
- [ ] Commit, tag `v0.86.0`, push.
