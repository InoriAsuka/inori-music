# Phase 96 — Admin user list: filter by username/role/enabled

## Goals

Add `?username`, `?role`, and `?enabled` filter query parameters to `GET /api/v1/admin/users`.
Filters are applied in the handler before sort and pagination.

## Requirements

- `?username=<value>`: exact match on username; case-sensitive.
- `?role=admin|viewer`: filter by role; returns `400 invalid_role` for other values.
- `?enabled=true|false`: filter by enabled status; returns `400 invalid_enabled` for other values.
- Multiple filters may be combined; they are applied in order (username → role → enabled).
- No new repository interface methods required.

## Tasks

- [x] Extend `listUsers` handler to apply `username`, `role`, and `enabled` filters before sort+paginate.
- [x] Add 5 HTTP-layer tests (`TestAdminListUsersFilterByRole`, `TestAdminListUsersFilterByEnabled`, `TestAdminListUsersFilterByUsername`, `TestAdminListUsersFilterInvalidRole`, `TestAdminListUsersFilterInvalidEnabled`).
- [x] Extend `GET /api/v1/admin/users` in OpenAPI contract with `username`, `role`, `enabled` query params; bump `info.version` to `0.96.0`.
- [x] Bump `VERSION` to `0.96.0`.
- [x] Update `requirement.md`.

## Non-Goals

- Full-text or substring search on username.
- Admin force-password-reset (Phase 97).
