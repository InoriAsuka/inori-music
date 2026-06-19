# Phase 111 — Viewer Track Summary

**Version**: v1.11.0  
**Date**: 2026-06-20

## Goal

Expose a combined viewer-scoped stats snapshot for a single track: how many times the authenticated user has played it plus the viewer's personal top tracks on a single endpoint. This complements the admin `GET /admin/history/tracks/{trackId}/history-summary` (Phase 107) with a viewer counterpart.

## Requirements

- Add `MyTrackSummary{Stats UserTrackStats, RecentTracks []TrackPlayCount}` type to the `history` package; `RecentTracks` is the viewer's overall top-N tracks (not scoped to `trackId`) for cross-track context.
- Add `GetMyTrackSummary(ctx, userID, trackID string, f UserStatsFilter, topN int) (MyTrackSummary, error)` to `history.Service`: calls `GetMyTrackStats` and `GetMyTopTracks`; `topN ≤ 0` defaults to 10, clamped to 100.
- Add `getMyTrackSummary` handler: `GET /api/v1/me/history/tracks/{trackId}/summary`; requires viewer auth; reads `{trackId}` from path; accepts optional `since`, `until` (RFC3339) and optional `?topN` (int; default 10; clamped 1–100) query params; calls `history.Service.GetMyTrackSummary`; returns `MyTrackSummary`; `503` when history service not configured.
- Register `GET /api/v1/me/history/tracks/{trackId}/summary` (viewer-auth) before the existing `{trackId}` wildcard fallback; add `methodNotAllowed` fallback.
- Add `MyTrackSummary` schema to OpenAPI components; add `get` operation to `/api/v1/me/history/tracks/{trackId}/summary`; bump `info.version` to `1.11.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/me/history/tracks/{trackId}/summary`.
- Add 2 `history.Service` unit tests: `TestGetMyTrackSummary`, `TestGetMyTrackSummaryEmpty`.
- Add 3 HTTP-layer tests: `TestViewerGetMyTrackSummary`, `TestViewerGetMyTrackSummaryWithTopN`, `TestViewerGetMyTrackSummaryNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

## Non-Goals

- No admin counterpart beyond what Phase 107 already provides.
- No mutation of track history data.

## Follow-Up Candidates

- Admin global history summary (Phase 110).
- Viewer history export / download (future).
