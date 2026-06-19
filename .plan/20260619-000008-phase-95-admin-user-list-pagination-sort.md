# Phase 95 — Admin user list: pagination + sorting

## Goals

Add `?limit`, `?offset`, `?sortBy` (username/role/createdAt/updatedAt), and `?sortOrder`
(asc/desc) query parameters to `GET /api/v1/admin/users`. The response now includes a
`pagination` envelope alongside the `users` array.

## Requirements

- `sortBy` defaults to `username`; valid values: `username`, `role`, `createdAt`, `updatedAt`.
- `sortOrder` defaults to `asc`; valid values: `asc`, `desc`. Returns `400 invalid_sort_order` otherwise.
- `limit=0` (or absent) returns all matching users from `offset` onward.
- `limit > 0` returns at most `limit` items; `hasMore` is `true` when more exist.
- `offset` defaults to 0.
- Sort + paginate is performed in the handler over the full `[]UserView` slice; no new repository methods.
- Response: `{"users":[...], "pagination":{"limit":N,"offset":N,"total":N,"hasMore":bool}}`.

## Tasks

- [x] Rewrite `listUsers` handler to sort+paginate with query param parsing.
- [x] Add 5 HTTP-layer tests (`TestAdminListUsersPagination`, `TestAdminListUsersSortByUsername`, `TestAdminListUsersSortDesc`, `TestAdminListUsersInvalidSortOrder`, `TestAdminListUsersInvalidLimit`).
- [x] Update OpenAPI spec: add query params to `GET /api/v1/admin/users`; update 200 to include `pagination`; bump `info.version` to `0.95.0`.
- [x] Bump `VERSION` to `0.95.0`.
- [x] Update `requirement.md`.

## Non-Goals

- Filtering by username/role/enabled (Phase 96).
- SQL-level pagination for PostgreSQL backend (future).
