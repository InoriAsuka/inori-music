# Phase 129 — CatalogTrack isFavorite Annotation (v1.29.0)

**Date:** 2026-06-20
**Version:** 1.29.0

## Goal

Inject `isFavorite: bool` into all viewer-facing Track responses, using a single batch
`AreFavorites` call per request (no N+1). Admin responses always carry `false`.

## Changes

### `httpapi/handler.go`

#### `trackView` struct
```go
type trackView struct {
    catalog.Track
    IsFavorite bool `json:"isFavorite"`
}
```
Embeds `catalog.Track` so all existing fields serialize as before, plus `isFavorite`.

#### `annotateTracksWithFavorites(ctx, userID, tracks) []trackView`
- Calls `favorites.Service.AreFavorites` once for the whole page.
- Best-effort: if service is nil or call fails, `IsFavorite` defaults to `false`.

#### `isViewerPath(r *http.Request) bool`
Checks prefix `/api/v1/catalog/` or `/api/v1/me/`.

#### Handler changes
| Handler | Change |
|---|---|
| `listTracks` | viewer path → `[]trackView`; admin path → `[]catalog.Track` |
| `getTrack` | viewer path → `trackView{...}`; admin path → `catalog.Track` |
| `getPlaylistTracks` | viewer path → annotated page; admin path unchanged |
| `listFavoriteTracks` | resolves IDs to full `Track` objects via `catalogService.GetTrack`; returns `[]trackView` with `IsFavorite=true` for all; falls back to `trackIds` array if catalog unavailable |

### `packages/api-contract/openapi/storage-admin.v1.json`
- `CatalogTrack.isFavorite` boolean field added (`default: false`).
- `FavoritesPage.tracks` array of `CatalogTrack` added.
- Version bumped to `1.29.0`.

## Tests
- 709 tests pass.
- No N+1 — `AreFavorites` is a single `ANY($2)` query per page.
- Best-effort: annotation failure does not break the response.
