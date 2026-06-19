# Phase 98 — Delete user: cascade session revocation

## Goals

Ensure that deleting a user also revokes all their active sessions, so orphaned
session tokens can no longer be used to authenticate after the account is gone.

## Requirements

- `DeleteUser` in `auth.Service` calls `RevokeAllSessionsByUser` before `users.DeleteUser`.
- If session revocation fails, deletion is aborted and the error is returned.
- No new routes, repository interface methods, or OpenAPI paths.

## Tasks

- [x] Modify `DeleteUser` in `auth.Service` to revoke sessions before deleting.
- [x] Add 2 `auth.Service` unit tests: `TestDeleteUserRevokesSessionsFirst`, `TestDeleteUserNotFound`.
- [x] Add 1 HTTP-layer test: `TestAdminDeleteUserRevokesSessionsFirst`.
- [x] Bump OpenAPI `info.version` to `0.98.0`.
- [x] Bump `VERSION` to `0.98.0`.
- [x] Update `requirement.md`.

## Non-Goals

- Cascading deletion of user history or catalog associations.
