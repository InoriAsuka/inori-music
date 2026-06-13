# Phase 43 — User-Facing Read-Only Catalog Browse API

**Date:** 2026-06-13  
**Version:** 0.43.0

## Requirement

Expose the existing catalog read and search capabilities to session-authenticated viewer and admin clients via a new `/api/v1/catalog/` namespace. All admin mutation endpoints remain under `/api/v1/admin/catalog/`. No unauthenticated public access.

## Tasks

- [x] Bump `VERSION` to `0.43.0`
- [x] Add `### v0.43.0` entry to `requirement.md`
- [x] Sync OpenAPI `info.version` to `0.43.0`
- [x] Create `.plan/20260613-130000-phase-43-catalog-viewer-api.md`
- [x] Add `requireViewerAuth` middleware to `handler.go`
- [x] Register 7 viewer catalog GET routes in `Routes()`
- [x] Register 7 viewer catalog method-not-allowed fallbacks in `Routes()`
- [x] Add missing 405 fallback for `/api/v1/admin/catalog/search` in `Routes()`
- [x] Add `/api/v1/catalog/` authenticated not-found catch-all
- [x] Add 7 viewer catalog paths to `storage-admin.v1.json`
- [x] Add 7 viewer paths to `expected` map in `openapi_contract_test.go`
- [x] Add `newViewerTestHandler` helper to `handler_test.go`
- [x] Add 11 HTTP-layer tests for viewer catalog browse

## Non-goals

- Unauthenticated (zero-auth) public catalog access
- Catalog create/update/delete under `/api/v1/catalog/`
- Pagination, fuzzy search (pg_trgm), batch import, playback URLs
- Database schema changes

## Follow-up candidates

- Phase 44: Batch import endpoint (`POST /api/v1/admin/catalog/batch-import`)
- Phase 45: pg_trgm fuzzy search extension
- Phase 46: Pagination on catalog list endpoints
- Future: Streaming/playback URL generation for tracks
