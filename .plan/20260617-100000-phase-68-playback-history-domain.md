# Phase 68 — Playback history domain

## Goal

Add a playback history / listening activity domain so session-authenticated
users can record which tracks they have played and query their personal history.

## New domain: `services/api/internal/history/`

### Types

```go
type PlayEvent struct {
    ID        string    `json:"id"`
    UserID    string    `json:"userId"`
    TrackID   string    `json:"trackId"`
    PlayedAt  time.Time `json:"playedAt"`   // client-reported
    CreatedAt time.Time `json:"createdAt"`  // server record time
}

type PlayEventFilter struct {
    UserID  string
    TrackID string // optional
    Limit   int
    Offset  int
}

type Repository interface {
    SavePlayEvent(ctx, e) error
    ListPlayEvents(ctx, f) ([]PlayEvent, int, error)  // items, total
    DeletePlayEventsByUser(ctx, userID) error
}
```

### Service methods

- `RecordPlay(ctx, userID, trackID, playedAt)` → `(PlayEvent, error)`
- `ListPlays(ctx, PlayEventFilter)` → `([]PlayEvent, int, error)`
- `ClearHistory(ctx, userID)` → `error`

## Database migration

`008_play_events` (appended to `storage/postgres/migrate.go`):
```sql
CREATE TABLE IF NOT EXISTS play_events (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    played_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS play_events_user_id_played_at_idx ON play_events (user_id, played_at DESC);
CREATE INDEX IF NOT EXISTS play_events_track_id_idx ON play_events (track_id);
```

## HTTP API (viewer-auth, user-scoped)

| Method | Route | Description |
|---|---|---|
| `POST`   | `/api/v1/me/history` | Record a play event |
| `GET`    | `/api/v1/me/history` | List calling user's events, newest first |
| `DELETE` | `/api/v1/me/history` | Delete all calling user's events |

## Middleware extension

`requireViewerAuth` now injects the authenticated `auth.User` into the request
context via `context.WithValue`. Handlers retrieve it with `userFromContext(r)`.
Same injection added to `requireAdminAuth` session path.

## Tasks

- [x] Create `history/types.go`, `history/service.go`, `history/memory_repository.go`.
- [x] Create `history/postgres/repository.go`.
- [x] Add migration `008_play_events` to `storage/postgres/migrate.go`.
- [x] Inject `auth.User` into context in `requireViewerAuth` and `requireAdminAuth`.
- [x] Add `historyService` field + `WithHistoryService` option to `Handler`.
- [x] Add `import history` to handler.go; add `historyNewService` test helper.
- [x] Add 3 routes + 405 fallback + `requireHistoryService` guard + 3 handler funcs.
- [x] Add 5 service unit tests (`service_test.go`).
- [x] Add 5 handler tests (record, list w/ pagination, clear, not-configured 503, 405).
- [x] Add `PlayEvent`, `PlayEventList` schemas + `/api/v1/me/history` path to OpenAPI.
- [x] Add `history_not_configured` error code to OpenAPI enum + contract test.
- [x] Add `/api/v1/me/history` to `TestStorageAdminOpenAPIContractCoversRoutes`.
- [x] Bump OpenAPI `info.version` → `0.68.0`, `VERSION`, `requirement.md`.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.68.0`, push.

## Follow-up candidates

- Admin aggregate history (most-played tracks, per-user stats).
- User-editable playlists.
- Full-text search extension to playlists.
