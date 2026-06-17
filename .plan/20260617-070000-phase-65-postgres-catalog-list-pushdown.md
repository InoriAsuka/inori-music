# Phase 65 — PostgreSQL catalog list pushdown

## Goal

All catalog list operations (artists, albums, tracks, playlists) previously
fetched every row then sorted and paginated in the handler. This phase adds
`ListXxxPage` methods that push ORDER BY + LIMIT/OFFSET + `COUNT(*) OVER()`
into SQL, so the database does the work and only the requested page is
transferred to the application.

## Design

Add parallel `ListXxxPage` methods to the `Repository` interface alongside the
existing `ListXxx` methods (which remain for internal service use). The new
methods accept a `ListQuery` struct and return `ListPage[T]` carrying `Items`
and the full unfiltered `Total`.

### Types added to `catalog/types.go`

```go
type ListQuery struct {
    SortBy    string
    SortOrder string // "asc" | "desc"
    Limit     int
    Offset    int
}

type ListPage[T any] struct {
    Items []T
    Total int // COUNT(*) OVER () — unaffected by LIMIT
}
```

### New Repository methods (× 7)

`ListArtistsPage`, `ListAlbumsPage`, `ListAlbumsByArtistPage`,
`ListTracksPage`, `ListTracksByAlbumPage`, `ListTracksByArtistPage`,
`ListPlaylistsPage`.

## Implementation files

| File | Change |
|---|---|
| `catalog/types.go` | Add `ListQuery`, `ListPage[T]`, 7 new interface methods |
| `catalog/service.go` | Add 7 thin pass-through methods |
| `catalog/memory_repository.go` | Implement 7 Page methods with Go sort+slice (adds `sort` import) |
| `catalog/postgres/repository_page.go` | New file — SQL with `COUNT(*) OVER ()`, `ORDER BY` mapping, `LIMIT`/`OFFSET` |
| `catalog/service_test.go` | Add stub Page methods on test `memRepo`; add 4 ListPage service tests |
| `catalog/postgres/repository_integration_test.go` | Add `TestRepositoryListArtistsPage` and `TestRepositoryListAlbumsPageByArtist` |
| `httpapi/handler.go` | Update 7 list handlers to call `ListXxxPage` and remove in-handler sort/paginate |

## Non-goals

- Existing `ListXxx` methods unchanged (still used by stats, recent timelines, search helpers).
- No changes to HTTP API shape — sort/paginate behavior is identical to Phase 61-62.

## Tasks

- [x] Add `ListQuery`, `ListPage[T]` to `catalog/types.go`; update interface.
- [x] Add 7 service pass-through methods.
- [x] Implement 7 `ListXxxPage` methods on `MemoryRepository`.
- [x] Create `catalog/postgres/repository_page.go` with SQL pushdown.
- [x] Update `listArtists`, `listAlbums`, `listTracks`, `listPlaylists`, `listAlbumsByArtist`, `listTracksByArtist`, `listTracksByAlbum` handlers.
- [x] Add Page stubs to test `memRepo`; add 4 `ListXxxPage` service unit tests.
- [x] Add 2 Postgres integration tests (build tag: integration).
- [x] Bump `VERSION` → `0.65.0` + requirement.md + OpenAPI.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.65.0`, push.

## Follow-up candidates

- SQL pushdown for recent timelines (`GetRecentlyAdded`/`GetRecentlyUpdated`).
- SQL aggregate stats instead of full-list-then-count.
- Playback history / listening activity domain.
