# Storage Administration HTTP API

## Scope

The HTTP API exposes authenticated admin endpoints for managing storage backends, running probes, refreshing health, reading capacity, registering media objects, running read-only integrity checks, updating lifecycle metadata, listing metadata pages, and reading metadata statistics.

## Authentication

`/healthz` remains public. `/api/v1/admin/*` requires `Authorization: Bearer <INORI_ADMIN_TOKEN>`. If no admin token is configured, admin routes fail closed with `503 admin_auth_not_configured`.

## Main Endpoints

- `GET /healthz`: process health.
- `GET/POST /api/v1/admin/storage/backends`: list or register storage backends.
- `POST /api/v1/admin/storage/backends/validate`: validate a candidate backend without persisting it.
- `POST /api/v1/admin/storage/backends/refresh`: refresh backend health and supported capacity state.
- `POST /api/v1/admin/storage/backends/{id}/probe`: run a safe backend probe.
- `GET /api/v1/admin/media/objects`: list media objects by exactly one metadata filter with sorting and pagination.
- `POST /api/v1/admin/media/objects`: register media object metadata.
- `GET /api/v1/admin/media/objects/stats`: read metadata-only statistics.
- `POST /api/v1/admin/media/objects/{id}/lifecycle`: update lifecycle metadata.
- `POST /api/v1/admin/media/objects/{id}/verify`: verify one media object in read-only mode.
- `POST /api/v1/admin/media/objects/verify`: batch-verify by backend or content hash.

## OpenAPI

The API contract lives at `packages/api-contract/openapi/storage-admin.v1.json`. Contract tests verify route, parameter, security, schema, and error-code coverage.
