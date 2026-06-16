# Phase 44 — Catalog Batch Import API

**Date:** 2026-06-13  
**Version:** 0.44.0

## Requirement

Add a batch-import endpoint that accepts a list of media object import requests and
processes each one independently, returning a per-item result that indicates success
or a typed error. A partially-successful batch returns HTTP 207 Multi-Status;
a fully-successful batch returns HTTP 200 OK. The service must not stop on individual
item failures.

## Tasks

- [x] Bump `VERSION` to `0.44.0`
- [x] Add `### v0.44.0` entry to `requirement.md`
- [x] Sync OpenAPI `info.version` to `0.44.0`
- [x] Create `.plan/20260613-140000-phase-44-catalog-batch-import.md`
- [x] Add `BatchImportItem`, `BatchImportResult`, `BatchImportResultItem` types to `catalog/types.go`
- [x] Add `BatchImportTracks(ctx, items []ImportTrackRequest) BatchImportResult` to `catalog.Service`
- [x] Add 5 `BatchImportTracks` unit tests in `catalog/service_test.go`
- [x] Add `POST /api/v1/admin/catalog/batch-import` route to `Routes()` in `handler.go`
- [x] Add `/api/v1/admin/catalog/batch-import` method-not-allowed fallback
- [x] Add `batchImportTracks` HTTP handler to `handler.go`
- [x] Add 6 HTTP-layer tests for batch import in `handler_test.go`
- [x] Update OpenAPI contract with `/api/v1/admin/catalog/batch-import` path and schemas
- [x] Commit, tag, push

## Non-goals

- Transactional rollback on partial failure (items are independent)
- Batch size limits enforced at the database layer
- Pagination or continuation tokens for large batches

## Follow-up candidates

- Phase 45: pg_trgm fuzzy search extension
- Phase 46: Pagination on catalog list endpoints
- Future: Streaming/playback URL generation for tracks
