# Phase 67 — SQL recent timelines pushdown

## Goal

Replace multi-query recent timeline assembly (4× ListXxx + sort + slice in
the service layer) with single `UNION ALL + ORDER BY + LIMIT` queries pushed
into the database.

## New Repository methods

| Method | SQL pattern |
|---|---|
| `RecentlyAdded(ctx, kind, limit)` | `UNION ALL` of artist/album/track/playlist branches on `created_at DESC LIMIT $1` |
| `RecentlyUpdated(ctx, kind, limit)` | Same structure on `updated_at DESC LIMIT $1` |

Both methods accept `kind=""` (all) or `kind=artist/album/track/playlist`.
Limit is validated/clamped by the service before calling.

## Service changes

`GetRecentlyAdded` and `GetRecentlyUpdated` replaced their 4-branch list-and-merge
logic with a single delegate call to the new repo methods. The `sort` import is
removed from `service.go`.

## Tasks

- [x] Add `RecentlyAdded` and `RecentlyUpdated` to `catalog/types.go` Repository interface.
- [x] Implement on `catalog.MemoryRepository` (in-memory sort+slice, preserving old behaviour).
- [x] Implement on `catalogpg.Repository` in `repository_page.go` (UNION ALL + ORDER BY + LIMIT).
- [x] Add stubs to test `memRepo` in `service_test.go`.
- [x] Add `TestRepositoryRecentlyAdded` PostgreSQL integration test.
- [x] Remove `sort` import from `service.go`; add `time` import to `repository_page.go`.
- [x] Bump OpenAPI `info.version` → `0.67.0`, `VERSION`, `requirement.md`.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.67.0`, push.

## Follow-up candidates

- Playback history / listening activity domain.
- User-editable playlists (non-admin session users).
