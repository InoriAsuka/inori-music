# Phase 120 — Wire Catalog and History Services into main.go

**Version**: v1.20.0
**Date**: 2026-06-20

## Goal

Instantiate `catalog.Service` and `history.Service` in `main.go` and inject them into the HTTP
handler via `WithCatalogService` and `WithHistoryService`. All catalog and history routes
currently return 503 in production because the services are never wired up. Fix that without
touching any domain logic.

## Requirements

- Import `catalogpg "inori-music/services/api/internal/catalog/postgres"` and
  `historypg "inori-music/services/api/internal/history/postgres"` in `main.go`.
- Add `catalogRepository()` and `historyRepository()` constructor helpers (mirroring the existing
  `storageRepository` / `mediaObjectRepository` pattern):
  - When `pool != nil`: return the PostgreSQL-backed repository.
  - Otherwise: return `catalog.NewMemoryRepository()` / `history.NewMemoryRepository()`.
- Instantiate `catalog.NewService(catalogRepo)` and `history.NewService(historyRepo)`.
- Append `httpapi.WithCatalogService(catalogService)` and
  `httpapi.WithHistoryService(historyService)` to `handlerOpts`.
- No new environment variables; no new schema migrations (008 already covers `play_events`,
  005–007 already cover catalog tables — all in the shared `pgstore.Migrate` runner).
- Add 1 `main_test.go` smoke test confirming that `NewHandler` configured with all four services
  reports `Ready: true` from `readinessReport()`, and that without catalog/history services the
  report still returns the same `admin_auth`, `storage_service`, `media_registry` checks as
  before (catalog/history services do not participate in the readiness contract).
- Bump VERSION and OpenAPI `info.version` to `1.20.0`.
- Update `requirement.md` Current Version to `1.20.0`.
- Commit with tag `v1.20.0`.

## Non-Goals

- No new HTTP endpoints.
- No CORS, rate-limiting, or gateway middleware (covered in later phases).
- No in-memory fallback environment variables for catalog/history repositories (always follow the
  PostgreSQL-or-memory pattern of the other domains).

## Follow-Up Candidates

- Readiness check coverage for catalog and history services (Phase 121).
- README sync (Phase 124).
