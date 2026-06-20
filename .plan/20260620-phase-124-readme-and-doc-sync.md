# Phase 124 — README and Documentation Sync

**Version**: v1.24.0
**Date**: 2026-06-20

## Goal

The README still references architecture baseline version `0.34.0` and lists only the first 34
phases. Bring all documentation into alignment with the current `v1.24.0` baseline, covering
Phases 35–124.

## Requirements

### README.md
- Update `## Version` to `Current architecture baseline version: 1.24.0`.
- Replace the `## 0.x Architecture Direction` section title with `## Architecture Direction`.
- Add a new `## Completed Phases` entry for each phase group since Phase 34:
  - Phase 35: PostgreSQL persistence layer.
  - Phase 36–37: Auth domain and session-based login.
  - Phase 38: Admin user management API.
  - Phase 39–67: Catalog domain (entities, playlists, stats, search, pagination, sorting).
  - Phase 68–119: Playback history domain (events, stats, timelines, summaries, filters).
  - Phase 120: Catalog and history service wiring into main.go.
  - Phase 121: Readiness check coverage for catalog and history services.
  - Phase 122: CORS middleware for browser client access.
  - Phase 123: Request-ID propagation middleware.
  - Phase 124: README and documentation sync (this phase).
- Update the `## Run the API Scaffold` section: add `INORI_CORS_ORIGINS` to the example command.
- Update `## Project Documents` to reference `docs/architecture/frontend-client-constraints.md`.
- Update `## Future Outlook` to reflect the current roadmap: Flutter client, web player, admin
  console.

### requirement.md
- Confirm `## Current Version` reads `1.24.0`.
- Add history entries for Phases 120–124 in the same format as existing entries.

### docs/architecture/frontend-client-constraints.md
- Already cleaned of non-English text in Phase 120 prep; no further changes needed.

### VERSION file
- Set to `1.24.0`.

### OpenAPI
- Bump `info.version` to `1.24.0`.

### Tests
- No new tests required; this phase is documentation-only.
- Run `go test ./...` to confirm no regressions from doc-adjacent file changes.

## Non-Goals

- Translating `.plan/` files retroactively (historical artefacts, kept as-is).
- Adding new API functionality.
- Updating Dockerfile or CI workflows (no functional change).

## Follow-Up Candidates

- Begin `packages/web/` and `packages/admin/` scaffolding.
- Flutter `packages/app/` initialisation.
