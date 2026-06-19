# Phase 106 — Admin user history summary endpoint

**Version:** v1.6.0  
**Date:** 2026-06-20

## Goal

Add `GET /api/v1/admin/history/users/{userId}/history-summary` as a convenience
endpoint that returns both aggregate stats and top-tracks for a specific user in
a single request. Avoids two round-trips for dashboard displays.

## Tasks

- [x] Add `UserHistorySummary{Stats UserHistoryStats, TopTracks []TrackPlayCount}`
      type to the `history` package.
- [x] Add `GetAdminUserSummary(ctx, userID string, f UserStatsFilter, topN int)
      (UserHistorySummary, error)` to `history.Service`: calls `GetAdminUserStats`
      and `GetAdminUserTopTracks`, returns the combined struct.
- [x] Add `getAdminUserHistorySummary` handler: `GET
      /api/v1/admin/history/users/{userId}/history-summary`; requires admin auth;
      reads `{userId}` from path; accepts optional `since`, `until` query params
      (RFC3339); accepts optional `?topN` (int; default 10; clamped 1–100);
      returns `UserHistorySummary`; `503` when history service not configured.
- [x] Register `GET /api/v1/admin/history/users/{userId}/history-summary`
      (admin-auth) before the existing `{userId}` catch-all fallback; add
      `methodNotAllowed` fallback for the new path.
- [x] Add `UserHistorySummary` schema to OpenAPI components; add `get` operation
      to `/api/v1/admin/history/users/{userId}/history-summary`; bump
      `info.version` to `1.6.0`.
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on
      `/api/v1/admin/history/users/{userId}/history-summary`.
- [x] Add 2 `history.Service` unit tests: `TestGetAdminUserSummary`,
      `TestGetAdminUserSummaryEmpty`.
- [x] Add 3 HTTP-layer tests: `TestAdminGetUserHistorySummary`,
      `TestAdminGetUserHistorySummaryWithTopN`,
      `TestAdminGetUserHistorySummaryNotConfigured`.

## Non-goals

- No repository interface changes required; delegates entirely to existing
  `UserHistoryStats` and `UserTopTracks` repository methods.
- No changes to existing `/users/{userId}/stats` or `/users/{userId}/top-tracks`
  endpoints.

## Follow-up candidates

- Phase 107: admin track history-summary (stats + top-listeners in one call).
