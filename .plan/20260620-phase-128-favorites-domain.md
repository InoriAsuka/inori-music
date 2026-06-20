# Phase 128 — User Favorites Domain (v1.28.0)

**Date:** 2026-06-20
**Version:** 1.28.0

## Goal

Add a personal favorites system allowing viewers to favorite/unfavorite tracks and list their favorites with pagination. All operations are idempotent.

## New Package: `internal/favorites`

### `types.go`
- `FavoriteEntry{UserID, TrackID, CreatedAt}` — the persisted record.
- `FavoritesPage{TrackIDs []string, Total int}` — paginated list response.
- `Repository` interface: `AddFavorite`, `RemoveFavorite`, `ListFavorites`, `IsFavorite`, `AreFavorites`.

### `service.go`
- `Service.AddFavorite` — idempotent, trims whitespace.
- `Service.RemoveFavorite` — idempotent.
- `Service.ListFavorites` — clamps limit to [1, 200], defaults to 50.
- `Service.IsFavorite` — point lookup.
- `Service.AreFavorites` — batch membership for Phase 129.

### `memory_repository.go`
- In-memory implementation sorted by `CreatedAt DESC`.

### `postgres/repository.go`
- `AddFavorite`: `INSERT … ON CONFLICT DO NOTHING`.
- `RemoveFavorite`: `DELETE WHERE user_id=$1 AND track_id=$2`.
- `ListFavorites`: `COUNT(*) OVER()` window function, `ORDER BY created_at DESC`.
- `AreFavorites`: `WHERE user_id=$1 AND track_id = ANY($2)`.

## Migration `010_user_track_favorites`

```sql
CREATE TABLE IF NOT EXISTS user_track_favorites (
    user_id    TEXT  NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    track_id   TEXT  NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, track_id)
);
CREATE INDEX IF NOT EXISTS user_track_favorites_user_id_created_at_idx
    ON user_track_favorites (user_id, created_at DESC);
```

## HTTP API Changes

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/me/favorites/tracks/{trackId}` | Add favorite (200 idempotent) |
| `DELETE` | `/api/v1/me/favorites/tracks/{trackId}` | Remove favorite (204 idempotent) |
| `GET` | `/api/v1/me/favorites/tracks` | List favorites with `limit`/`offset` |

List response: `{"trackIds":[...], "pagination":{limit,offset,total,hasMore}}`

## `main.go`
- `favoritesRepository(pool)` → PG or in-memory.
- `favorites.NewService(favoritesRepo)` wired as `WithFavoritesService`.

## OpenAPI
- `FavoritesPage` schema added.
- 3 favorites paths added.
- `favorites_not_configured` error code added.
- Version `1.28.0`.

## Tests
- 709 tests pass across 12 packages (favorites + favorites/postgres included).
