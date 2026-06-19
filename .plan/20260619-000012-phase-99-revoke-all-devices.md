# Phase 99 — Viewer: revoke ALL sessions (including current)

## Goals

Allow an authenticated viewer to revoke ALL their sessions — including the one they are
currently using — via `POST /api/v1/me/sessions/revoke-all-devices`. This is a "sign out
everywhere" action, distinct from `revoke-all` which preserves the current session.

## Requirements

- New `revokeAllMySessions` handler delegates to `auth.Service.RevokeAllSessionsForUser`.
- All sessions for the user are revoked regardless of whether they match the current token.
- Returns `{"revoked": N}` where N includes the calling session.
- `503` when auth service is not configured.

## Tasks

- [x] Add `revokeAllMySessions` handler.
- [x] Register `POST /api/v1/me/sessions/revoke-all-devices` and its `methodNotAllowed` fallback.
- [x] Add `post` operation to `/api/v1/me/sessions/revoke-all-devices` in OpenAPI contract.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes`.
- [x] Add 3 HTTP-layer tests.
- [x] Bump OpenAPI `info.version` to `0.99.0`.
- [x] Bump `VERSION` to `0.99.0`.
- [x] Update `requirement.md`.

## Non-Goals

- Force-logout of a specific session by token (future admin capability).
