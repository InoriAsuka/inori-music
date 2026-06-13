# Phase 45 — Catalog PATCH Metadata Update API

**Date:** 2026-06-14
**Version:** v0.45.0

## Goal

Add partial-update (PATCH) endpoints for catalog artists, albums, and tracks.
Pointer-typed request fields distinguish "not provided" from "explicitly empty"
so clients can clear optional fields without touching fields they did not send.

## Requirements

- `PATCH /api/v1/admin/catalog/artists/{id}` — update name and/or sortName.
- `PATCH /api/v1/admin/catalog/albums/{id}` — update title, sortTitle, artistId, releaseYear.
- `PATCH /api/v1/admin/catalog/tracks/{id}` — update title, sortTitle, artistId, albumId, trackNumber, discNumber, durationMs.
- Only fields present in the JSON body (non-null) are applied; nil → unchanged.
- Validation mirrors create: name/title/artistId may not be set to empty string; numeric fields must be ≥ 0.
- Artist ownership of album is enforced when updating albumId on a track.
- UpdatedAt advances on every successful PATCH.
- Returns 200 with the updated entity on success; 400 on validation error; 404 on not found; 503 when catalog not configured.

## Non-goals

- No PATCH on viewer routes (read-only).
- No partial-update of mediaObjectId on tracks (immutable after creation).

## Tasks

- [x] Add `UpdateArtistRequest`, `UpdateAlbumRequest`, `UpdateTrackRequest` pointer types to `catalog/types.go`
- [x] Add `WithClock` setter on `catalog.Service` for test injection
- [x] Implement `UpdateArtist`, `UpdateAlbum`, `UpdateTrack` on `catalog.Service`
- [x] Add 11 unit tests in `catalog/service_test.go` (4 artist, 4 album, 3 track)
- [x] Add `patchArtist`, `patchAlbum`, `patchTrack` HTTP handlers in `httpapi/handler.go`
- [x] Register `PATCH` routes in `Handler.Routes()`
- [x] Add 7 HTTP-layer tests in `httpapi/handler_test.go`
- [x] Add `CatalogUpdateArtistRequest`, `CatalogUpdateAlbumRequest`, `CatalogUpdateTrackRequest` schemas to OpenAPI
- [x] Add `patch` operations to artist, album, track `/{id}` paths in OpenAPI
- [x] Bump `info.version` to `0.45.0` in OpenAPI contract
- [x] Bump `VERSION` to `0.45.0`
- [x] Update `requirement.md`

## Follow-up candidates

- Viewer-authenticated PATCH if future per-user catalog editing is added.
- Track `mediaObjectId` re-link via dedicated endpoint (e.g. POST /tracks/{id}/relink).
