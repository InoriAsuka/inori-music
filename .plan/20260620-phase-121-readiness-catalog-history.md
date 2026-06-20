# Phase 121 — Extend Readiness Checks for Catalog and History Services

**Version**: v1.21.0
**Date**: 2026-06-20

## Goal

The `/readyz` endpoint currently reports three checks: `storage_service`, `media_registry`, and
`admin_auth`. Now that catalog and history services are always wired (Phase 120), add
`catalog_service` and `history_service` readiness checks so operators can confirm both domains are
active at startup.

## Requirements

- Extend `readinessReport()` in `handler.go` to append two new `ReadinessCheck` items:
  - `catalog_service`: ok when `handler.catalogService != nil`.
  - `history_service`: ok when `handler.historyService != nil`.
- `ReadinessReport.Ready` becomes `false` if any check — including the two new ones — fails.
- Update the OpenAPI contract: add `catalog_service` and `history_service` as documented possible
  check names in the `ReadinessCheck` schema description (no schema shape change needed).
- Update `openapi_contract_test.go`: extend `TestStorageAdminOpenAPIContractCoversRoutes` or add
  a dedicated `TestReadinessChecks` test asserting the new check names appear in a handler
  configured with all services.
- Add 4 HTTP-layer tests:
  - `TestReadinessAllConfigured`: all five checks pass, `ready: true`.
  - `TestReadinessMissingCatalog`: catalog nil → `ready: false`, `catalog_service` failed.
  - `TestReadinessMissingHistory`: history nil → `ready: false`, `history_service` failed.
  - `TestReadinessMissingAuth`: admin token absent → `ready: false`, `admin_auth` failed.
- Bump VERSION and OpenAPI `info.version` to `1.21.0`.
- Update `requirement.md` Current Version to `1.21.0`.
- Commit with tag `v1.21.0`.

## Non-Goals

- No liveness probe changes (`/healthz` stays unconditional 200).
- No Prometheus metric changes (metrics reflect readiness gauges; will re-derive from new checks
  automatically because `metrics` handler calls `readinessReport()`).

## Follow-Up Candidates

- CORS / preflight middleware (Phase 122).
