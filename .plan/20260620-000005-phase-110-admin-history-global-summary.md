# Phase 110 — Admin Global History Summary

**Version**: v1.10.0  
**Date**: 2026-06-20

## Goal

Add a single combined admin endpoint that returns the system-wide play-event aggregate stats, top tracks, and top users in one response — the global counterpart to the per-user and per-track summary endpoints added in Phases 106–107.

## Requirements

- Add `GlobalHistorySummary{Stats HistoryStats, TopTracks []TrackPlayCount, TopUsers []UserPlayCount}` type to the `history` package.
- Add `GetGlobalSummary(ctx, f StatsFilter, topN int) (GlobalHistorySummary, error)` to `history.Service`: calls `GetHistoryStats`, `GetTopTracks`, and `GetTopUsers` in sequence; `topN ≤ 0` defaults to 10, clamped to 100.
- Add `getAdminHistorySummary` handler: `GET /api/v1/admin/history/summary`; requires admin auth; accepts optional `since`, `until` (RFC3339) and optional `?topN` (int; default 10; clamped 1–100) query params; calls `history.Service.GetGlobalSummary`; returns `GlobalHistorySummary`; `503` when history service not configured.
- Register `GET /api/v1/admin/history/summary` (admin-auth) before the existing `/api/v1/admin/history/{eventId}` wildcard; add `methodNotAllowed` fallback.
- Add `GlobalHistorySummary` schema to OpenAPI components; add `get` operation to `/api/v1/admin/history/summary`; bump `info.version` to `1.10.0`.
- Extend `TestStorageAdminOpenAPIContractCoversRoutes` with `get` on `/api/v1/admin/history/summary`.
- Add 2 `history.Service` unit tests: `TestGetGlobalSummary`, `TestGetGlobalSummaryWithTopN`.
- Add 3 HTTP-layer tests: `TestAdminGetHistorySummary`, `TestAdminGetHistorySummaryWithTopN`, `TestAdminGetHistorySummaryNotConfigured`.
- The phase output is version-tracked and covered by the relevant tests or documentation checks.

## Non-Goals

- No viewer-facing global summary.
- No change to the `HistoryStats`, `TrackPlayCount`, or `UserPlayCount` types.

## Follow-Up Candidates

- Viewer per-track summary (Phase 111).
