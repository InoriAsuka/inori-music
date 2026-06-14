# Phase 47 — Playlist Track Reorder API

**Date:** 2026-06-14
**Version:** v0.47.0

## Goal

Add `PUT /api/v1/admin/catalog/playlists/{id}/tracks` to atomically replace the
entire ordered track list of a playlist in a single request — enabling reorder,
bulk-add, bulk-remove, and clear in one atomic call.

## Requirements

- `PUT /api/v1/admin/catalog/playlists/{id}/tracks` — atomically replaces the
  playlist's ordered track list with the supplied `trackIds` array.
- Every supplied track ID must exist; unknown IDs return 404.
- An empty `trackIds` array is valid and clears the playlist.
- Duplicate entries in the incoming list are preserved (client-controlled).
- Returns 200 with the updated `Playlist` on success.
- 404 when the playlist ID is unknown.
- 400 when `trackIds` is missing from the body.
- 503 when no catalog service is configured.

## Non-goals

- No viewer-facing PUT (read-only viewer routes).
- No partial reorder (cursor-swap, move-by-index). Full replacement is the primitive.
- No repository interface change — `SavePlaylist` already handles atomic track-list replacement.

## Tasks

- [x] Add `SetPlaylistTracks(ctx, playlistID, trackIDs)` to `catalog.Service`
- [x] Add `setPlaylistTracksRequest` struct and `setPlaylistTracks` handler to `httpapi/handler.go`
- [x] Register `PUT /api/v1/admin/catalog/playlists/{id}/tracks` route
- [x] Add 5 service unit tests in `catalog/service_test.go` (reorder, clear, duplicate, unknown track, unknown playlist)
- [x] Add 7 HTTP-layer tests in `httpapi/handler_test.go` (happy path, clear, unknown track, unknown playlist, missing trackIds, no-catalog 503, pre-existing method-not-allowed still passes)
- [x] Add `SetPlaylistTracksRequest` schema and `put` operation to OpenAPI contract
- [x] Fix pre-existing Phase 46 OpenAPI regressions: path-level `parameters` missing on playlist `{id}` paths; `BearerAuth` vs `bearerAuth` case inconsistency
- [x] Bump `info.version` → `0.47.0`
- [x] Bump `VERSION` → `0.47.0`
- [x] Update `requirement.md` current version and append v0.47.0 history entry

## Follow-up candidates

- Viewer-authenticated PATCH if future per-user catalog editing is added.
- Track `mediaObjectId` re-link via dedicated endpoint (e.g. `POST /tracks/{id}/relink`).
- Playlist-level ordering guarantees in PostgreSQL (position column already present).
