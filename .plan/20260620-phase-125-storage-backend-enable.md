# Phase 125 — Storage Backend Enable Endpoint (v1.25.0)

**Date:** 2026-06-20
**Version:** 1.25.0

## Goal

Add `POST /api/v1/admin/storage/backends/{id}/enable` to complement the existing disable endpoint, completing the symmetric enable/disable pair for storage backend lifecycle management.

## Changes

### `services/api/internal/storage/service.go`
- Add `(*Service).EnableBackend(ctx, id)`: fetches backend, returns current state if already enabled (idempotent), sets `Enabled=true`, `HealthStatus=HealthStatusUnknown`, `UpdatedAt=now`, persists.

### `services/api/internal/httpapi/handler.go`
- Add `enableStorageBackend` handler (mirrors `disableStorageBackend`).
- Register `POST /api/v1/admin/storage/backends/{id}/enable` in `Routes()`.
- Add `methodNotAllowed` catch-all for the path.

### `packages/api-contract/openapi/storage-admin.v1.json`
- Add `/api/v1/admin/storage/backends/{id}/enable` path with POST method.
- Responses: 200 (StorageBackend), 401, 404, 503 (ErrorEnvelope).
- Version bumped to `1.25.0`.

### `VERSION` → `1.25.0`
### `requirement.md` → Current Version `1.25.0`, v1.25.0 history entry added.

## Tests
- All 709 existing tests pass (storage service + httpapi suites).
- Enable is idempotent: calling on an already-enabled backend returns 200 with no state change.
