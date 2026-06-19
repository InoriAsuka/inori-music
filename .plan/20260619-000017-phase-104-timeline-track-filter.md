# Phase 104 — ?trackId filter on viewer history timeline

**Version:** v1.4.0  
**Date:** 2026-06-19

## Goal

Expose the existing `?trackId` query parameter on `GET /api/v1/me/history/timeline`
so viewers can scope their own play-event timeline to a single track.

## Tasks

- [x] The `getMyHistoryTimeline` handler already passes `TrackID: q.Get("trackId")`
      to `GetMyTimeline` — no handler code change needed.
- [x] The OpenAPI contract for `/api/v1/me/history/timeline` already declares
      `trackId` as a query parameter — no schema change needed.
- [x] Add HTTP-layer test `TestViewerGetHistoryTimelineTrackIdFilter` confirming
      that `?trackId` scopes the timeline to the specified track.
- [x] Bump `info.version` to `1.4.0`.

## Non-goals

- No new repository or service interface methods required.
- No changes to the `/api/v1/admin/history/timeline` endpoint (already supports
  `?trackId` as an optional filter in the existing handler).

## Follow-up candidates

- Consider exposing `?trackId` filter on the admin history timeline endpoint in
  a future phase.
