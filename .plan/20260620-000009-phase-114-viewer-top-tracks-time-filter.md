# Phase 114 — Viewer Top-Tracks Time Filter + OpenAPI Cleanup

**Version**: v1.14.0  
**Date**: 2026-06-20

## Goal

Add explicit `since` / `until` query-parameter declarations to the remaining viewer history endpoints whose handlers already consume them via `parseHistoryAdminFilter` but whose OpenAPI operations do not yet declare those parameters. Specifically: `GET /api/v1/me/history/top-tracks`, `GET /api/v1/me/history/stats`, and `GET /api/v1/me/history/timeline`. Confirm the filter is actually forwarded in each handler; add forwarding if missing.

## Requirements

- Verify `getMyTopTracks` handler parses `?since`/`?until` and passes them as `UserStatsFilter.Since`/`Until` to `GetMyTopTracks`; fix forwarding if missing.
- Verify `getMyHistoryStats` handler does the same for `GetMyStats`.
- Verify `getMyHistoryTimeline` handler does the same for `GetMyTimeline`.
- Update `GET /api/v1/me/history/top-tracks`, `GET /api/v1/me/history/stats`, and `GET /api/v1/me/history/timeline` in OpenAPI to declare `since` and `until` query parameters; bump `info.version` to `1.14.0`.
- Add 1 `history.Service` unit test: `TestGetMyTopTracksTimeWindow`.
- Add 2 HTTP-layer tests: `TestViewerGetMyTopTracksSince`, `TestViewerGetMyTopTracksUntil`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

## Non-Goals

- No new routes or types.
- No major-version bump (continuing minor increments per project policy).

## Follow-Up Candidates

- Full OpenAPI contract audit pass to catch any remaining undeclared parameters.
- Viewer history export endpoint (future).
