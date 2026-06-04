# Phase 21: Media Object Asset Kind Filter (v0.21.0)

## Requirement Snapshot

- Allow administrators to list media object metadata by asset kind after lifecycle and statistics support are available.
- Preserve the single-filter list rule while adding `assetKind` as a first-class metadata filter.
- Keep filtering metadata-only and compatible with existing pagination.

## Task Checklist

- [x] Extend the media object list filter with `assetKind`.
- [x] Add repository support for listing media objects by asset kind in stable order.
- [x] Add service-level asset-kind filter validation for supported media asset kinds.
- [x] Extend `GET /api/v1/admin/media/objects` with the `assetKind` query parameter and existing pagination.
- [x] Update OpenAPI, requirements, README, and architecture docs for v0.21.0.
- [x] Add domain and HTTP tests for asset-kind filtering and invalid asset-kind filters.
- [x] Run formatting, static checks, JSON contract parsing, unit tests, race tests, and diff checks.

## Non-Goals

- No multi-filter query composition in this phase.
- No user library, album, or track domain models yet.
- No PostgreSQL indexing until database persistence is introduced.

## Follow-Up Candidates

- Add metadata import jobs that register media objects by asset kind.
- Add SQL-backed indexes for asset-kind queries when persistence moves to PostgreSQL.
- Add admin UI filters generated from the OpenAPI contract.
