# Phase 107 — Admin track history summary endpoint

**Version:** v1.7.0  
**Date:** 2026-06-20

## Goal

Add `GET /api/v1/admin/history/tracks/{trackId}/history-summary` as a convenience
endpoint that returns both aggregate stats and top-listeners for a specific track
in a single request. Mirrors Phase 106's user history-summary for tracks.

## Tasks

- [x] Add `TrackHistorySummary{Stats TrackHistoryStatsResult, TopListeners []UserPlayCount}`
      type to the `history` package.
- [x] Add `GetTrackSummary(ctx, trackID string, f TrackStatsFilter, topN int)
      (TrackHistorySummary, error)` to `history.Service`: calls `GetTrackStats`
      and `GetTrackTopListeners`, returns the combined struct.
- [x] Add `getAdminTrackHistorySummary` handler:
      `GET /api/v1/admin/history/tracks/{trackId}/history-summary`; requires
      admin auth; reads `{trackId}` from path; accepts optional `since`, `until`
      query params (RFC3339); accepts optional `?topN` (int; default 10;
      clamped 1–100); returns `TrackHistorySummary`; `503` when history service
      not configured.
- [x] Register `GET /api/v1/admin/history/tracks/{trackId}/history-summary`
      (admin-auth) before the existing `{trackId}` catch-all fallback; add
      `methodNotAllowed` fallback for the new path.
- [x] Add `TrackHistorySummary` schema to OpenAPI components; add `get` operation
      to `/api/v1/admin/history/tracks/{trackId}/history-summary`; bump
      `info.version` to `1.7.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on
      `/api/v1/admin/history/tracks/{trackId}/history-summary`.
- [x] Add 2 `history.Service` unit tests: `TestGetTrackSummary`,
      `TestGetTrackSummaryEmpty`.
- [x] Add 3 HTTP-layer tests: `TestAdminGetTrackHistorySummary`,
      `TestAdminGetTrackHistorySummaryWithTopN`,
      `TestAdminGetTrackHistorySummaryNotConfigured`.

## Non-goals

- No repository interface changes required.
- No changes to existing track stats or top-listeners endpoints.

## Follow-up candidates

- Phase 108: viewer history summary (stats + top-tracks in one call).
