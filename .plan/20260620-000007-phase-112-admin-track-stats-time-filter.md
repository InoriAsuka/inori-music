# Phase 112 — Admin Track Stats Time Filter

**Version**: v1.12.0  
**Date**: 2026-06-20

## Goal

Extend `GET /api/v1/admin/history/tracks/{trackId}/stats` and its top-listeners sibling to accept optional `?since` / `?until` time bounds — mirroring what Phase 109 did for the viewer's per-track stats endpoint.

## Requirements

- `history.Repository.TrackHistoryStats` and `TrackTopListeners` already accept `TrackStatsFilter` with `Since`/`Until` fields. No interface changes are needed.
- Verify that `historypg.Repository.TrackHistoryStats` already filters on `played_at` when `f.Since` / `f.Until` are non-zero; if not, add the `AND played_at >= $N` / `AND played_at < $N` clauses.
- Verify that `history.MemoryRepository.TrackHistoryStats` already filters on `played_at`; if not, add the skip logic.
- Update `getAdminTrackStats` handler to parse optional `?since` / `?until` query params (RFC3339) and pass them into `TrackStatsFilter`; return `400 invalid_since` / `400 invalid_until` on parse failure.
- Update `getAdminTrackTopListeners` handler similarly.
- Update `GET /api/v1/admin/history/tracks/{trackId}/stats` and `GET /api/v1/admin/history/tracks/{trackId}/top-listeners` in OpenAPI to declare `since` and `until` query parameters; bump `info.version` to `1.12.0`.
- Add 1 `history.Service` unit test: `TestGetAdminTrackStatsTimeWindow`.
- Add 2 HTTP-layer tests: `TestAdminGetTrackStatsSince`, `TestAdminGetTrackStatsUntil`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

## Non-Goals

- No new route — only query-parameter additions.
- The `TrackHistorySummary` endpoint (Phase 107) already propagates `since`/`until` via `TrackStatsFilter`; no change needed there.

## Follow-Up Candidates

- Admin user stats time filter (Phase 113 already targets `UserHistoryStats`).
