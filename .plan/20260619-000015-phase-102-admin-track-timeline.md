# Phase 102 — Admin track play timeline endpoint

**Version:** v1.2.0  
**Date:** 2026-06-19

## Goal

Add `GET /api/v1/admin/history/tracks/{trackId}/timeline` so admins can retrieve
play-event counts for a specific track grouped by time bucket (day/week/month).

## Tasks

- [x] Implement `getAdminTrackTimeline` handler: reads `{trackId}` from path,
      parses `since`, `until`, `granularity` query params, calls
      `history.Service.GetHistoryTimeline` with `TrackID` set from path;
      returns `{buckets}`. Validates that `since` and `until` are both present.
- [x] Register `GET /api/v1/admin/history/tracks/{trackId}/timeline` (admin-auth)
      and its `methodNotAllowed` fallback.
- [x] Add `get` operation to `/api/v1/admin/history/tracks/{trackId}/timeline`
      in OpenAPI contract; bump `info.version` to `1.2.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on
      `/api/v1/admin/history/tracks/{trackId}/timeline`.
- [x] Add 3 HTTP-layer tests: `TestAdminGetTrackTimeline`,
      `TestAdminGetTrackTimelineMissingBounds`,
      `TestAdminGetTrackTimelineMethodNotAllowed`.

## Non-goals

- No new service or repository methods required; reuses `GetHistoryTimeline`
  with `TrackID` set in `TimelineFilter`.

## Follow-up candidates

- Phase 103: admin per-user timeline endpoint
  (`GET /api/v1/admin/history/users/{userId}/timeline`).
