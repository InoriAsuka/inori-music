# Phase 55 — Catalog recent timelines include playlists

## Goal

Extend the admin recent catalog timelines so playlists participate alongside
artists, albums, and tracks. Admin dashboard clients should be able to surface
newly created and recently changed playlists without issuing a separate playlist
list call or merging timelines client-side.

## Requirements

- Extend `GET /api/v1/admin/catalog/recently-added` so `RecentCatalogResult` items may include `kind: "playlist"`, a `playlist` payload, and `addedAt` copied from `Playlist.CreatedAt`.
- Extend `GET /api/v1/admin/catalog/recently-updated` so `UpdatedCatalogResult` items may include `kind: "playlist"`, a `playlist` payload, and `updatedAt` copied from `Playlist.UpdatedAt`.
- Query parameter `kind` on both endpoints must accept `artist`, `album`, `track`, or `playlist`; invalid values still return 400.
- Existing default and maximum `limit` behavior remains unchanged: default 20, must be positive, clamp above 100.
- Results remain newest-first across all included entity types, with playlists sorted by their timestamp field just like other entities.
- Empty catalogs still return a non-nil empty `items` array.
- No new `Repository` interface methods required — playlist entries derive from existing `ListPlaylists`.
- Preserve admin-only scope; no viewer-facing recent timeline endpoints in this phase.

## Non-goals

- No changes to playlist CRUD or playlist track expansion semantics.
- No pagination beyond the existing `limit` cap.
- No PostgreSQL-specific ordered/limited query pushdown.
- No media object or storage timeline entries.
- No viewer-facing `/api/v1/catalog/recently-added` or `/api/v1/catalog/recently-updated` endpoints.

## Tasks

- [x] Extend `RecentItemKind` validation to include `playlist`.
- [x] Add optional `Playlist *Playlist` payload fields to `RecentCatalogItem` and `UpdatedCatalogItem` in `catalog/types.go`.
- [x] Update `GetRecentlyAdded` in `catalog/service.go` to read `ListPlaylists`, append playlist timeline items, support `kind=playlist`, and preserve newest-first limit behavior.
- [x] Update `GetRecentlyUpdated` in `catalog/service.go` to read `ListPlaylists`, append playlist timeline items, support `kind=playlist`, and preserve newest-first limit behavior.
- [x] Add `catalog.Service` unit tests for recently-added playlist inclusion, recently-added playlist-only filter, recently-updated playlist inclusion, recently-updated playlist-only filter, invalid kind preservation, and limit behavior with playlists.
- [x] Add HTTP-layer tests for playlist payloads and `kind=playlist` on both recent endpoints.
- [x] Update OpenAPI `RecentItemKind` enum to include `playlist`.
- [x] Update OpenAPI `RecentCatalogItem` and `UpdatedCatalogItem` schemas with optional `playlist` payloads.
- [x] Bump OpenAPI `info.version` → `0.55.0`.
- [x] Bump `VERSION` → `0.55.0`.
- [x] Update `requirement.md` current version and append v0.55.0 history entry.
- [x] Run the relevant Go tests and OpenAPI contract tests.
- [x] Refresh the codegraph index if changed symbols are not reflected after implementation.

## Follow-up candidates

- Add viewer-facing recent timeline endpoints once product requirements for browse clients are finalized.
- Add repository-level ordered/limited queries when catalog size warrants pushdown.
- Add media-object-backed listening or playback activity timelines in a separate domain-specific phase.
