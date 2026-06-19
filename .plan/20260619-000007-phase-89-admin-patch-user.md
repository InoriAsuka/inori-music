# Phase 89 — Admin user patch PATCH /admin/users/{id}

## Context

The admin API can create, get, list, disable, enable, and delete users, but has no way to update
a user's `role` in-place. This phase adds `PATCH /api/v1/admin/users/{id}` supporting
optional `role` and `username` field updates. Fields absent from the request body are left unchanged.

## Affected endpoints

| Path | Method | Notes |
|------|--------|-------|
| `PATCH /api/v1/admin/users/{id}` | PATCH | Partial update: optional `role`, optional `username` |

## Service changes

Add `PatchUser(ctx, id string, role *Role, username *string) (UserView, error)` to `auth.Service`:
- Fetches the user by ID.
- If `role` is non-nil, validates it is `admin` or `viewer`; applies it.
- If `username` is non-nil, validates with `usernameRe` and checks for conflicts (ErrUserConflict).
- Updates `UpdatedAt`, saves, returns `UserView`.

## Handler changes

Add `patchAdminUser(w, r)`:
- Decodes body: `{role?: string, username?: string}` (all optional).
- Converts non-empty `role` string to `*auth.Role`, non-empty `username` to `*string`.
- Returns `400 invalid_user` if neither field is set (empty PATCH).
- Calls `authService.PatchUser`.

Register:
```
PATCH /api/v1/admin/users/{id}  → requireAdminAuth(patchAdminUser)
```
The `/api/v1/admin/users/{id}` catch-all already covers the `methodNotAllowed`.

## OpenAPI changes

Add `patch` operation to the existing `/api/v1/admin/users/{id}` path item.
Add `PatchUserRequest` schema (`{role?: string(enum), username?: string}`).
Bump `info.version` to `0.89.0`.

## Tests

**`auth/service_test.go`** — 3 unit tests:
`TestPatchUserRole`, `TestPatchUserUsername`, `TestPatchUserConflict`.

**`httpapi/handler_test.go`** — 4 HTTP-layer tests:
`TestAdminPatchUserRole`, `TestAdminPatchUserUsernameConflict`, `TestAdminPatchUserEmpty`,
`TestAdminPatchUserNotConfigured`.

**`httpapi/openapi_contract_test.go`** — extend `TestStorageAdminOpenAPIContractCoversRoutes`
with `patch` on `/api/v1/admin/users/{id}`; add `TestStorageAdminOpenAPIContractAdminPatchUser`.

## Non-goals

- Admin password reset (separate operation from PATCH).
- Viewer self-patch of username.

## Follow-up candidates

- Admin password reset (bypass current-password check).
- Viewer patch own username.

## Tasks

- [ ] Add `PatchUser` to `auth.Service`.
- [ ] Add `patchAdminUser` handler.
- [ ] Register `PATCH /api/v1/admin/users/{id}`.
- [ ] Add 3 service unit tests.
- [ ] Add 4 HTTP-layer tests.
- [ ] Update OpenAPI: add `patch` to `/api/v1/admin/users/{id}`; add `PatchUserRequest`; bump `0.89.0`.
- [ ] Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractAdminPatchUser`.
- [ ] Run `go test ./services/api/...` — all pass.
- [ ] Update `requirement.md` to `0.89.0`, append phase entry.
- [ ] Bump `VERSION` to `0.89.0`.
- [ ] Commit, tag `v0.89.0`, push.
