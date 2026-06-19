# Phase 101 — Viewer per-track play statistics endpoint

**Version:** v1.1.0  
**Date:** 2026-06-19

## Goal

Add `GET /api/v1/me/history/tracks/{trackId}/stats` so a viewer can retrieve
aggregate play statistics (total plays, first/last played time) for a specific
track.

## Tasks

- [x] Add `UserTrackStats{TrackID, TotalPlays, FirstPlayedAt, LastPlayedAt}`
      type to the `history` package.
- [x] Extend `history.Repository` interface with
      `UserTrackPlayStats(ctx, userID, trackID string) (UserTrackStats, error)`.
- [x] Implement `MemoryRepository.UserTrackPlayStats` (in-memory linear scan).
- [x] Implement `historypg.Repository.UserTrackPlayStats` (SQL `MIN/MAX/COUNT`
      with nullable timestamp pointers).
- [x] Add `GetMyTrackStats(ctx, userID, trackID string) (UserTrackStats, error)`
      to `history.Service`.
- [x] Implement `getMyTrackStats` handler: reads `{trackId}` from path, calls
      `GetMyTrackStats`, writes the `UserTrackStats` JSON response.
- [x] Register `GET /api/v1/me/history/tracks/{trackId}/stats` (viewer-auth)
      and its `methodNotAllowed` fallback.
- [x] Add `UserTrackStats` schema to OpenAPI components; add `get` operation to
      `/api/v1/me/history/tracks/{trackId}/stats`; bump `info.version` to `1.1.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with
      `/api/v1/me/history/tracks/{trackId}/stats`.
- [x] Add 3 `history.Service` unit tests: `TestGetMyTrackStatsNoPlays`,
      `TestGetMyTrackStatsWithPlays`, `TestGetMyTrackStatsMissingArgs`.
- [x] Add 3 HTTP-layer tests: `TestViewerGetMyTrackStats`,
      `TestViewerGetMyTrackStatsNoPlays`, `TestViewerGetMyTrackStatsMethodNotAllowed`.

## Non-goals

- No new endpoints for admin per-track viewer breakdown are added in this phase.

## Follow-up candidates

- Phase 102: admin per-track timeline endpoint.
