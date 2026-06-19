# Phase 103 — Admin user play timeline endpoint

**Version:** v1.3.0  
**Date:** 2026-06-19

## Goal

Add `GET /api/v1/admin/history/users/{userId}/timeline` so admins can retrieve
play-event counts for a specific user grouped by time bucket (day/week/month).

## Tasks

- [x] Implement `getAdminUserTimeline` handler: reads `{userId}` from path,
      parses `since`, `until`, `granularity` query params, calls
      `history.Service.GetHistoryTimeline` with `UserID` set from path;
      returns `{buckets}`. Validates that `since` and `until` are both present.
- [x] Register `GET /api/v1/admin/history/users/{userId}/timeline` (admin-auth)
      and its `methodNotAllowed` fallback.
- [x] Add `get` operation to `/api/v1/admin/history/users/{userId}/timeline`
      in OpenAPI contract; bump `info.version` to `1.3.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on
      `/api/v1/admin/history/users/{userId}/timeline`.
- [x] Add 3 HTTP-layer tests: `TestAdminGetUserTimeline`,
      `TestAdminGetUserTimelineMissingBounds`,
      `TestAdminGetUserTimelineMethodNotAllowed`.

## Non-goals

- No new service or repository methods required; reuses `GetHistoryTimeline`
  with `UserID` set in `TimelineFilter`.

## Follow-up candidates

- Phase 104: `?trackId` filter on `GET /api/v1/me/history/timeline`.
