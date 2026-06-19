# Phase 105 — Viewer track-scoped play timeline

**Version:** v1.5.0  
**Date:** 2026-06-20

## Goal

Add `GET /api/v1/me/history/tracks/{trackId}/timeline` so viewers can see their
own play-event counts for a specific track grouped by time bucket (day/week/month).
This mirrors the existing admin endpoint
`GET /api/v1/admin/history/tracks/{trackId}/timeline` but is scoped to the
authenticated user's events only.

## Tasks

- [x] Add `getMyTrackTimeline` handler: reads `{trackId}` from path; requires
      viewer auth; accepts `since` (required), `until` (required),
      `granularity` (optional; day/week/month) query params; calls
      `history.Service.GetMyTimeline` with both `UserID` (from auth context)
      and `TrackID` (from path) set; returns `{buckets}`;
      `503` when history service not configured.
- [x] Register `GET /api/v1/me/history/tracks/{trackId}/timeline` (viewer-auth)
      before the existing wildcard fallback `/api/v1/me/history/tracks/{trackId}`.
- [x] Add `get` operation to `/api/v1/me/history/tracks/{trackId}/timeline` in
      OpenAPI contract; bump `info.version` to `1.5.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on
      `/api/v1/me/history/tracks/{trackId}/timeline`.
- [x] Add 3 HTTP-layer tests: `TestViewerGetMyTrackTimeline`,
      `TestViewerGetMyTrackTimelineMissingBounds`,
      `TestViewerGetMyTrackTimelineMethodNotAllowed`.

## Non-goals

- No new service or repository methods required; reuses `GetMyTimeline` with
  both `UserID` and `TrackID` set in `TimelineFilter`.
- No changes to the admin track timeline endpoint.

## Follow-up candidates

- Phase 106: admin user history-summary (stats + top-tracks in one call).
