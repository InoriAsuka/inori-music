# Phase 97 — Admin: force-reset user password

## Goals

Allow admins to reset any user's password without knowing the current password, via
`POST /api/v1/admin/users/{id}/change-password`.

## Requirements

- New `ForceChangePassword(ctx, userID, newPassword string) error` on `auth.Service`.
- Validates new password ≥ 8 characters; returns `ErrInvalidUser` otherwise.
- Returns `ErrUserNotFound` for unknown users.
- Handler decodes `{newPassword}`, returns `204 No Content` on success.
- `400 invalid_user` for weak/missing password; `404 not_found` for unknown user; `503` when auth not configured.
- `ForceChangePasswordRequest` schema in OpenAPI components.

## Tasks

- [x] Add `ForceChangePassword` to `auth.Service`.
- [x] Add `forceChangePassword` handler and register route + methodNotAllowed.
- [x] Add 3 `auth.Service` unit tests.
- [x] Add 4 HTTP-layer tests.
- [x] Update OpenAPI contract; add `TestStorageAdminOpenAPIContractAdminForceChangePassword`.
- [x] Bump `VERSION` to `0.97.0`.
- [x] Update `requirement.md`.

## Non-Goals

- Password reset via email or token links.
