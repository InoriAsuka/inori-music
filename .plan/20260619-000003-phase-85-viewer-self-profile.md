# Phase 85 — Viewer self-profile GET /me

## Context

The viewer currently has session-backed access to history and catalog endpoints, but has no way to
inspect their own profile (ID, username, role, enabled flag, timestamps). The authenticated `User`
is already carried in the request context by `requireViewerAuth`; `getMe` simply projects it.

## Affected endpoints

| Path | Method | Notes |
|------|--------|-------|
| `GET /api/v1/me` | GET | Returns `UserView` for the authenticated viewer |

## Service changes

Add `GetUser(ctx context.Context, id string) (UserView, error)` to `auth.Service` — delegates to
`users.GetUser(ctx, id)` and wraps the result with `toView`. Used by both this phase and Phase 86.

## Handler changes

Add `getMe(w, r)`: reads the authenticated user from `userFromContext(r)` and writes it as `UserView`.

Register:
```
GET /api/v1/me       → requireViewerAuth(getMe)
/api/v1/me           → requireViewerAuth(methodNotAllowed)
```

## OpenAPI changes

Add `GET /api/v1/me` operation: no path parameters; 200 refs `UserView` schema; 401 ErrorEnvelope.
Bump `info.version` to `0.85.0`.

## Tests

**`auth/service_test.go`** — 1 unit test: `TestGetUser`.

**`httpapi/handler_test.go`** — 3 HTTP-layer tests:
`TestGetMe`, `TestGetMeUnauthenticated`, `TestGetMeNotConfigured`.

**`httpapi/openapi_contract_test.go`** — extend `TestStorageAdminOpenAPIContractCoversRoutes`
with `GET /api/v1/me`; add `TestStorageAdminOpenAPIContractGetMe`.

## Non-goals

- PATCH /me (profile update)
- Admin variant (Phase 86 adds GET /admin/users/{id})

## Follow-up candidates

- Admin single-user GET (Phase 86)
- Viewer password change (Phase 87)

## Tasks

- [ ] Add `GetUser` to `auth.Service`.
- [ ] Add `getMe` handler.
- [ ] Register `GET /api/v1/me` and its methodNotAllowed catch-all.
- [ ] Add 1 service unit test.
- [ ] Add 3 HTTP-layer tests.
- [ ] Update OpenAPI: `GET /api/v1/me`; bump `0.85.0`.
- [ ] Extend `TestStorageAdminOpenAPIContractCoversRoutes`; add `TestStorageAdminOpenAPIContractGetMe`.
- [ ] Run `go test ./services/api/...` — all pass.
- [ ] Update `requirement.md` to `0.85.0`, append phase entry.
- [ ] Bump `VERSION` to `0.85.0`.
- [ ] Commit, tag `v0.85.0`, push.
