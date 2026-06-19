# Phase 100 — Viewer track play history endpoint

**Version:** v1.0.0  
**Date:** 2026-06-19

## Goal

Add `GET /api/v1/me/history/tracks/{trackId}` so a viewer can retrieve their
own paginated play history for a specific track.

## Tasks

- [x] Register `GET /api/v1/me/history/tracks/{trackId}` (viewer-auth) and its
      `methodNotAllowed` fallback in `Routes()`.
- [x] Implement `getMyTrackHistory` handler: reads `{trackId}` from path,
      parses `limit`, `offset`, `since`, `until`, `order` query params,
      calls `history.Service.ListPlays` with `UserID` from auth context and
      `TrackID` from path; returns `{events, pagination}`.
- [x] Add `get` operation to `/api/v1/me/history/tracks/{trackId}` in the
      OpenAPI contract; bump `info.version` to `1.0.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on
      `/api/v1/me/history/tracks/{trackId}`.
- [x] Add HTTP-layer tests:
      `TestViewerGetMyTrackHistory`,
      `TestViewerGetMyTrackHistoryFiltersToOwnUser`,
      `TestViewerGetMyTrackHistoryMethodNotAllowed`.

## Non-goals

- No new repository or service interface methods required; reuses `ListPlays`
  with `TrackID` filter (already supported by `PlayEventFilter`).

## Follow-up candidates

- Phase 101: viewer stats for a specific track
  (`GET /api/v1/me/history/tracks/{trackId}/stats`).
