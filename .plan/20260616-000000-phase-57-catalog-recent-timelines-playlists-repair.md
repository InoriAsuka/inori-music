# Phase 57 — Catalog recent timelines playlist repair

## Goal

The v0.56.0 requirement and OpenAPI contract advertised playlist entries in the
recently-added and recently-updated unified catalog timelines, but the Go domain
code was not updated. Service tests already referenced `catalog.RecentItemPlaylist`
and playlist payload fields, and the handlers relied on the absent constant, so
the baseline was contract-inconsistent. This phase makes the implementation match
the published contract and releases as `0.57.0`.

## Requirements

- `RecentItemKind` enum must include `"playlist"` as a valid constant.
- `RecentCatalogItem` and `UpdatedCatalogItem` must carry an optional `Playlist
  *Playlist` payload field.
- `GetRecentlyAdded` and `GetRecentlyUpdated` on `catalog.Service` must iterate
  `ListPlaylists` when `kind` is empty or `"playlist"`.
- `validateRecentItemKind` must accept `"playlist"`; the error text must name all
  four valid kinds.
- All existing `catalog.Service` unit tests and HTTP-layer tests that reference
  playlist timeline symbols must compile and pass.
- New HTTP-layer tests assert `kind=playlist` returns playlist payloads on both
  admin and viewer recent endpoints.
- OpenAPI `RecentItemKind` enum already includes `"playlist"`; verified by new
  contract test asserting the enum and payload `$ref`.
- Missing error codes `relink_rejected`, `validation_error`, and `invalid_limit`
  added to OpenAPI `ErrorEnvelope.error.code` enum.
- `PATCH` on artist and album `{id}` paths added to the route coverage test.
- `capacity.go` corrected: duplicate `FilesystemCapacityProvider` body removed
  now that build-tagged `capacity_unix.go` / `capacity_unsupported.go` exist.

## Non-goals

- No new `Repository` interface methods.
- No PostgreSQL pushdown for recent timelines.
- No pagination beyond the existing `limit` cap.
- No viewer-facing catalog stats endpoints.
- No playback URL or streaming changes.

## Tasks

- [x] Add `RecentItemPlaylist RecentItemKind = "playlist"` to `catalog/types.go`.
- [x] Add `Playlist *Playlist` field to `RecentCatalogItem` and `UpdatedCatalogItem`; update type comments.
- [x] Extend `GetRecentlyAdded` to iterate playlists via `repo.ListPlaylists`.
- [x] Extend `GetRecentlyUpdated` to iterate playlists via `repo.ListPlaylists`.
- [x] Update `validateRecentItemKind` to accept `playlist`; update error text.
- [x] Fix duplicate `FilesystemCapacityProvider` declaration in `capacity.go`.
- [x] Add HTTP-layer tests: admin recently-added/updated `kind=playlist` + unified playlist presence.
- [x] Add viewer HTTP-layer tests: `kind=playlist` on both viewer recent endpoints.
- [x] Add `TestStorageAdminOpenAPIContractRecentTimelineSchemas` asserting `RecentItemKind` enum and playlist payload refs.
- [x] Fix contract route test to include `patch` on artist and album `{id}` paths.
- [x] Add `relink_rejected`, `validation_error`, `invalid_limit` to OpenAPI error enum and contract test.
- [x] Bump OpenAPI `info.version` → `0.57.0`.
- [x] Bump `VERSION` → `0.57.0`.
- [x] Update `requirement.md` current version and append v0.57.0 history entry.
- [x] Run full `go test ./...` and confirm green.
- [x] Commit, tag `v0.57.0`, push.

## Follow-up candidates

- Phase 58: track playback URL / streaming access bootstrap for viewer clients.
- Repository-level ordered/limited timeline queries for large catalogs.
- Viewer-facing catalog stats (`GET /api/v1/catalog/stats`).
