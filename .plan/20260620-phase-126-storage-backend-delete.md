# Phase 126 — Storage Backend Delete Endpoint (v1.26.0)

**Date:** 2026-06-20
**Version:** 1.26.0

## Goal

Add `DELETE /api/v1/admin/storage/backends/{id}` with safety guards preventing deletion of the default backend or backends that still have media objects registered.

## Changes

### `services/api/internal/storage/repository.go`
- Add `Delete(ctx, id) error` to `Repository` interface.
- Implement on `MemoryRepository`: acquire write lock, 404 if missing, `delete(map, id)`.

### `services/api/internal/storage/file_repository.go`
- Implement `Delete(ctx, id) error`: acquire write lock, 404 if missing, remove from map, `persistLocked()`.

### `services/api/internal/storage/postgres/backend_repository.go`
- Implement `Delete(ctx, id) error`: `DELETE FROM storage_backends WHERE id=$1`, 404 if `RowsAffected()==0`.

### `services/api/internal/storage/validation.go`
- Add `ErrBackendIsDefault` and `ErrBackendInUse` sentinel errors.

### `services/api/internal/storage/service.go`
- Add `(*Service).DeleteBackend(ctx, id)`: fetch, guard `IsDefault → ErrBackendIsDefault`, then `repository.Delete`.

### `services/api/internal/httpapi/handler.go`
- Add `deleteStorageBackend`: check media object references via `ListMediaObjects(limit=1)`, guard `Total>0 → ErrBackendInUse`, call `storage.DeleteBackend`, respond 204.
- Add `ErrBackendIsDefault → 409 storage_backend_is_default` and `ErrBackendInUse → 409 storage_backend_in_use` to `writeError`.
- Register `DELETE /api/v1/admin/storage/backends/{id}` in `Routes()`.
- Fix `/validate` and `/refresh` catch-alls to use explicit method prefixes (GET+DELETE) to avoid ServeMux conflict with the new DELETE `{id}` route.

### `packages/api-contract/openapi/storage-admin.v1.json`
- Add `/api/v1/admin/storage/backends/{id}` path with DELETE method; 204/401/404/409/503.
- Add `storage_backend_is_default` and `storage_backend_in_use` to ErrorEnvelope code enum.
- Version bumped to `1.26.0`.

### `VERSION` → `1.26.0`
### `requirement.md` → Current Version `1.26.0`, v1.26.0 history entry added.

## Tests
- 709 existing tests pass.
- Idempotency: deleting already-absent backend returns 404.
- Guard: default backend → 409 storage_backend_is_default.
- Guard: backend with media objects → 409 storage_backend_in_use.
