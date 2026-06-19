# Phase 88 — Admin enable user POST /admin/users/{id}/enable

## Context

`POST /api/v1/admin/users/{id}/disable` exists for disabling users, but there is no corresponding
`enable` action. A disabled user cannot be re-enabled without directly accessing the database.
This phase adds the symmetric `POST /api/v1/admin/users/{id}/enable` endpoint.

## Affected endpoints

| Path | Method | Notes |
|------|--------|-------|
| `POST /api/v1/admin/users/{id}/enable` | POST | Re-enables a previously disabled user |

## Service changes

Add `EnableUser(ctx, id string) (UserView, error)` to `auth.Service` — mirrors `DisableUser`:
fetches the user by ID, sets `Enabled = true`, updates `UpdatedAt`, saves, returns `UserView`.

## Handler changes

Add `enableUser(w, r)`: calls `authService.EnableUser(ctx, r.PathValue("id"))`, writes `UserView`.

Register:
```
POST /api/v1/admin/users/{id}/enable  → requireAdminAuth(enableUser)
/api/v1/admin/users/{id}/enable       → requireAdminAuth(methodNotAllowed)
```

## OpenAPI changes

Add `POST /api/v1/admin/users/{id}/enable` path item with a `$ref` to `UserId` path parameter.
Bump `info.version` to `0.88.0`.

## Tests

**`auth/service_test.go`** — 2 unit tests:
`TestEnableUser`, `TestEnableUserNotFound`.

**`httpapi/handler_test.go`** — 3 HTTP-layer tests:
`TestEnableUser`, `TestEnableUserNotFound`, `TestEnableUserNotConfigured`.

**`httpapi/openapi_contract_test.go`** — extend `TestStorageAdminOpenAPIContractCoversRoutes`
with `post` on `/api/v1/admin/users/{id}/enable`; add `TestStorageAdminOpenAPIContractEnableUser`.

## Non-goals

- Viewer self-enable (not applicable; disabled users cannot authenticate).
- Admin patch role (Phase 89).

## Follow-up candidates

- Admin patch user role (Phase 89).

## Tasks

- [ ] Add `EnableUser` to `auth.Service`.
- [ ] Add `enableUser` handler.
- [ ] Register `POST /api/v1/admin/users/{id}/enable` and catch-all.
- [ ] Add 2 service unit tests.
- [ ] Add 3 HTTP-layer tests.
- [ ] Update OpenAPI: add `/api/v1/admin/users/{id}/enable`; bump `0.88.0`.
- [ ] Extend contract tests.
- [ ] Run `go test ./services/api/...` — all pass.
- [ ] Update `requirement.md` to `0.88.0`, append phase entry.
- [ ] Bump `VERSION` to `0.88.0`.
- [ ] Commit, tag `v0.88.0`, push.
