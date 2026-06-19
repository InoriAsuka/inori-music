# Phase 109 — ?since/?until time filters on viewer track stats

**Version:** v1.9.0  
**Date:** 2026-06-20

## Goal

Extend `GET /api/v1/me/history/tracks/{trackId}/stats` with optional `?since`
and `?until` query parameters so viewers can scope the aggregate play counts to
a time window (e.g. "how many times did I play this track this month?").

## Tasks

- [x] Rename `UserTrackPlayStats(ctx, userID, trackID string) (UserTrackStats, error)`
      to `UserTrackPlayStats(ctx, userID, trackID string, f UserStatsFilter)
      (UserTrackStats, error)` in `history.Repository`; `f.Since` and `f.Until`
      bound the query when non-zero. (The `UserStatsFilter.UserID` field is
      ignored; caller passes `userID` directly.)
- [x] Update `history.MemoryRepository.UserTrackPlayStats` to apply `f.Since`
      and `f.Until` bounds when non-zero.
- [x] Update `historypg.Repository.UserTrackPlayStats` to inject
      `AND played_at >= $3` / `AND played_at < $4` clauses when non-zero.
- [x] Update `history.Service.GetMyTrackStats(ctx, userID, trackID string,
      f UserStatsFilter) (UserTrackStats, error)` signature to forward `f`.
- [x] Update `getMyTrackStats` handler to parse optional `?since` / `?until`
      query params (RFC3339); return `400 invalid_since` / `400 invalid_until`
      on parse failure; pass them via `UserStatsFilter` to `GetMyTrackStats`.
- [x] Update `GET /api/v1/me/history/tracks/{trackId}/stats` in OpenAPI to
      declare `since` and `until` query parameters; bump `info.version` to
      `1.9.0`.
- [x] Update `history.Service` unit tests for `GetMyTrackStats` to cover the
      time-filtered path: `TestGetMyTrackStatsTimeWindow`.
- [x] Add 2 HTTP-layer tests: `TestViewerGetMyTrackStatsSince`,
      `TestViewerGetMyTrackStatsUntil`.

## Non-goals

- No changes to the admin track stats endpoint (`/admin/history/tracks/{trackId}/stats`)
  which already accepts `since`/`until`.
- No changes to `UserTrackStats` response shape.

## Follow-up candidates

- Apply similar time-filter extension to `UserTrackPlayStats` for admin use.
