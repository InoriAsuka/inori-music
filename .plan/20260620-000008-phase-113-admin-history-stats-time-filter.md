# Phase 113 — Admin History Stats Time Filter

**Version**: v1.13.0  
**Date**: 2026-06-20

## Goal

Extend `GET /api/v1/admin/history/stats` (system-wide), `GET /api/v1/admin/history/top-tracks`, and `GET /api/v1/admin/history/top-users` to accept optional `?since` / `?until` — making all three "global" admin aggregate endpoints consistently time-bounded.

## Requirements

- `history.Repository.HistoryStats`, `TopTracks`, and `TopUsers` already accept `StatsFilter` with `Since`/`Until`; confirm the Postgres and memory implementations filter correctly when bounds are set; add clauses if any are missing.
- The `getAdminHistoryStats`, `getAdminTopTracks`, and `getAdminTopUsers` handlers already call `parseHistoryAdminFilter` which parses `since`/`until`; verify those parsed values are forwarded into `StatsFilter.Since`/`Until`. If they already are, no handler change is needed.
- Update `GET /api/v1/admin/history/stats`, `GET /api/v1/admin/history/top-tracks`, and `GET /api/v1/admin/history/top-users` in OpenAPI to explicitly declare `since` and `until` query parameters (if not yet declared); bump `info.version` to `1.13.0`.
- Add 1 `history.Service` unit test: `TestGetHistoryStatsTimeWindow`.
- Add 2 HTTP-layer tests: `TestAdminGetHistoryStatsSince`, `TestAdminGetHistoryStatsUntil`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

## Non-Goals

- No new routes.
- No change to the `GlobalHistorySummary` type (Phase 110).

## Follow-Up Candidates

- Viewer top-tracks time filter (Phase 114).
