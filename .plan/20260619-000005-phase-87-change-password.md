# Phase 87 — Viewer password change POST /me/change-password

## Context

Authenticated viewers have no way to change their own password. This phase adds
`POST /api/v1/me/change-password` with current-password verification before accepting the new one.
The operation requires no admin privileges — it is self-service for the authenticated viewer.

## Affected endpoints

| Path | Method | Notes |
|------|--------|-------|
| `POST /api/v1/me/change-password` | POST | Body: `{currentPassword, newPassword}` |

## Service changes

Add `ChangePassword(ctx, userID, currentPassword, newPassword string) error` to `auth.Service`:
- Fetches the user by ID (`users.GetUser`).
- Verifies `currentPassword` against `user.PasswordHash` using `CheckPassword`.
- Returns `ErrBadCredentials` if the current password is wrong.
- Hashes `newPassword` (minimum 8 chars, same rule as CreateUser) with `HashPassword`.
- Saves the updated user with `users.SaveUser`, updating `PasswordHash` and `UpdatedAt`.

## Handler changes

Add `changePassword(w, r)`: reads `userFromContext`, decodes `{currentPassword, newPassword}`,
calls `authService.ChangePassword`, returns `204 No Content` on success.

Register:
```
POST /api/v1/me/change-password  → requireViewerAuth(changePassword)
/api/v1/me/change-password       → requireViewerAuth(methodNotAllowed)
```

## OpenAPI changes

Add `POST /api/v1/me/change-password` operation: request body refs `ChangePasswordRequest` schema
(`{currentPassword: string, newPassword: string}`); 204 No Content on success; 400/401 ErrorEnvelope.
Bump `info.version` to `0.87.0`.

## Tests

**`auth/service_test.go`** — 3 unit tests:
`TestChangePassword`, `TestChangePassword_WrongCurrent`, `TestChangePassword_WeakNew`.

**`httpapi/handler_test.go`** — 4 HTTP-layer tests:
`TestChangePassword`, `TestChangePasswordWrongCurrent`, `TestChangePasswordUnauthenticated`,
`TestChangePasswordNotConfigured`.

**`httpapi/openapi_contract_test.go`** — extend `TestStorageAdminOpenAPIContractCoversRoutes`
with `post` on `/api/v1/me/change-password`; add `TestStorageAdminOpenAPIContractChangePassword`.

## Non-goals

- Admin-forced password reset.
- Password complexity rules beyond minimum length.
- Session invalidation after password change.

## Follow-up candidates

- Admin reset user password (bypass current-password check).
- Session invalidation on password change.

## Tasks

- [ ] Add `ChangePassword` to `auth.Service`.
- [ ] Add `changePassword` handler.
- [ ] Register `POST /api/v1/me/change-password` and methodNotAllowed catch-all.
- [ ] Add 3 service unit tests.
- [ ] Add 4 HTTP-layer tests.
- [ ] Update OpenAPI: add `POST /api/v1/me/change-password`; bump `0.87.0`.
- [ ] Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractChangePassword`.
- [ ] Run `go test ./services/api/...` — all pass.
- [ ] Update `requirement.md` to `0.87.0`, append phase entry.
- [ ] Bump `VERSION` to `0.87.0`.
- [ ] Commit, tag `v0.87.0`, push.
