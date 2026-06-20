# Phase 127 — Track Genre Field and Filter (v1.27.0)

**Date:** 2026-06-20
**Version:** 1.27.0

## Goal

Add optional `genre` field to `CatalogTrack` with support for:
- Setting genre on create (POST), patch (PATCH), and import.
- Filtering track lists by genre (`?genre=` query param, case-insensitive).
- Sorting by genre (`sortBy=genre`).
- PostgreSQL migration (non-destructive `ADD COLUMN IF NOT EXISTS`).

## Changes

### `catalog/types.go`
- `Track.Genre string` (JSON `"genre,omitempty"`).
- `ListQuery.Genre string` (filter pass-through for track queries).
- `UpdateTrackRequest.Genre *string`.
- `ImportTrackRequest.Genre string`.
- `TrackSortByGenre = "genre"` constant.

### `catalog/service.go`
- `CreateTrack` signature: added `genre string` param after `mediaObjectID`.
- `UpdateTrack`: apply `req.Genre` when non-nil.
- `ImportTrack`: set `track.Genre = strings.TrimSpace(req.Genre)`.

### `catalog/memory_repository.go`
- `ListTracksPage/ListTracksByAlbumPage/ListTracksByArtistPage`: `strings.EqualFold` genre filter.
- `trackLess`: added `TrackSortByGenre` case.

### `catalog/postgres/repository.go`
- `SaveTrack`: INSERT/UPDATE includes `genre` column; `NULLIF($10,'')`.
- `GetTrack/ListTracks/ListTracksByAlbum/ListTracksByArtist`: `COALESCE(genre,'')` in SELECT.
- `scanTrack`: added `&t.Genre`.

### `catalog/postgres/repository_page.go`
- `trackOrderBy`: `TrackSortByGenre → lower(COALESCE(genre,''))`.
- `ListTracksPage/ListTracksByAlbumPage/ListTracksByArtistPage`: conditional `WHERE lower(COALESCE(genre,''))=lower($N)` when `q.Genre != ""`.
- `queryTracksPage`: `COALESCE(genre,'')` in SELECT + `&t.Genre` in Scan.

### `storage/postgres/migrate.go`
- Migration `009_track_genre`: `ALTER TABLE tracks ADD COLUMN IF NOT EXISTS genre TEXT` + partial index.

### `httpapi/handler.go`
- `createTrackRequest`, `patchTrackRequest`, `importTrackRequest`: added `Genre string/Genre *string`.
- `listTracks`: parse `?genre` query param, set `q.Genre`.
- `createTrack`, `patchTrack`, `importTrack`, `batchImportTracks`: pass genre through to service.

### `catalog/service_test.go`
- Updated all `CreateTrack` call sites to include `""` for new genre param.

### `packages/api-contract/openapi/storage-admin.v1.json`
- `CatalogTrack.genre` field added.
- `CatalogUpdateTrackRequest.genre` field added.
- `?genre` query param added to 6 track list endpoints.
- Version bumped to `1.27.0`.

### `VERSION` → `1.27.0`
### `requirement.md` → Current Version `1.27.0`, v1.27.0 history entry added.

## Tests
- 709+ existing tests pass.
- Genre filter is case-insensitive (`Rock` == `rock` == `ROCK`).
- Genre is nullable in DB; missing genre returns `""` in JSON (omitempty).
