# Phase 108 — Viewer history summary endpoint

**Version:** v1.8.0  
**Date:** 2026-06-20

## Goal

Add `GET /api/v1/me/history/summary` as a convenience endpoint for viewers that
returns both aggregate stats and top-tracks in one request. Provides the same
composite view as the admin user history-summary, but scoped to the authenticated
viewer and exposed on the viewer-auth surface.

## Tasks

- [ ] Add `getMyHistorySummary` handler: `GET /api/v1/me/history/summary`;
      requires viewer auth; accepts optional `since`, `until` query params
      (RFC3339); accepts optional `?topN` (int; default 10; clamped 1–100);
      calls `history.Service.GetMyStats` and `history.Service.GetMyTopTracks`
      with `UserID` from auth context; returns
      `{"stats": UserHistoryStats, "topTracks": []TrackPlayCount}`;
      `503` when history service not configured.
- [ ] Register `GET /api/v1/me/history/summary` (viewer-auth) before the
      `/api/v1/me/history/{eventId}` wildcard; add `methodNotAllowed` fallback.
- [ ] Add `get` operation to `/api/v1/me/history/summary` in OpenAPI contract
      (response inline: `UserHistoryStats` ref + `TrackPlayCount` array);
      bump `info.version` to `1.8.0`.
- [ ] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on
      `/api/v1/me/history/summary`.
- [ ] Add 3 HTTP-layer tests: `TestViewerGetMyHistorySummary`,
      `TestViewerGetMyHistorySummaryWithTopN`,
      `TestViewerGetMyHistorySummaryNotConfigured`.

## Non-goals

- No new service or repository methods required; reuses `GetMyStats` and
  `GetMyTopTracks` from `history.Service`.
- No changes to `/me/history/stats` or `/me/history/top-tracks`.

## Follow-up candidates

- Phase 109: `?since`/`?until` time-bound filters on
  `GET /me/history/tracks/{trackId}/stats`.
