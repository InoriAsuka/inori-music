# Phase 49 — Playlist tracks expand endpoint

## Goal

Add `GET /api/v1/admin/catalog/playlists/{id}/tracks` and
`GET /api/v1/catalog/playlists/{id}/tracks` so clients can retrieve the
full ordered list of `Track` objects for a playlist in a single HTTP
round-trip, instead of N separate per-track fetches after reading
`Playlist.trackIds`.

## Requirements

- Returns `{"tracks": [...Track...]}` with Track objects in playlist-defined order.
- An empty playlist returns `{"tracks": []}`.
- Duplicate entries (same track ID at multiple positions) are expanded once per
  occurrence — the list reflects the raw `trackIds` array in order.
- Returns 404 `not_found` when the playlist ID is unknown.
- Admin route: 503 when catalog service is not configured.
- Viewer route: 401 when no valid bearer token; viewer token accepted.

## Non-goals

- No pagination — playlists are assumed client-manageable in size for now.
- No PostgreSQL-side join optimisation — sequential `GetTrack` calls per
  `MemoryRepository` are adequate for the service-test tier.

## Tasks

- [x] Add `GetPlaylistTracks(ctx, playlistID)` to `catalog.Service`
- [x] Add `getPlaylistTracks` handler to `httpapi/handler.go`
- [x] Register `GET /api/v1/admin/catalog/playlists/{id}/tracks` (admin-auth)
- [x] Register `GET /api/v1/catalog/playlists/{id}/tracks` (viewer-auth)
- [x] Register 405 fallback for `/api/v1/catalog/playlists/{id}/tracks`
- [x] Add 4 service unit tests (ordered, empty, not-found, duplicate expansion)
- [x] Add 6 HTTP-layer tests (admin, empty, 404, viewer, no-catalog 503, 405)
- [x] Add `PlaylistTracksResult` schema and `get` ops to OpenAPI contract
- [x] Extend `TestStorageAdminOpenAPIContractCoversRoutes` for all playlist paths
- [x] Bump `info.version` → `0.49.0`
- [x] Bump `VERSION` → `0.49.0`
- [x] Update `requirement.md` current version and append v0.48.0 + v0.49.0 history entries

## Follow-up candidates

- Pagination support on playlist track lists for very large playlists.
- PostgreSQL-backed `GetPlaylistTracks` using a single JOIN query for efficiency.
- Viewer-authenticated playlist creation / edit if per-user library management is added.
