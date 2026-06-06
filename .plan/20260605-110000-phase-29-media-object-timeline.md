# Phase 29: Media Object Metadata Timeline

## Version

v0.29.0

## Requirement

Expose a read-only metadata timeline for one media object so admin tools can inspect the currently retained registration, latest verification, and latest lifecycle transition summary without opening or mutating media bytes.

## Tasks

- [x] Add a `MediaObjectTimeline` domain model and service method that derives events from existing media object metadata.
- [x] Expose `GET /api/v1/admin/media/objects/{id}/timeline` behind the existing admin bearer-token authentication.
- [x] Extend the OpenAPI contract and contract tests with the timeline path and schemas.
- [x] Add domain and HTTP handler tests for timeline ordering, details, not-found behavior, and authentication.
- [x] Update versioned requirements, README, and architecture notes for v0.29.0.

## Notes

The timeline intentionally summarizes only metadata that is already retained in the current JSON-backed repositories. Full append-only audit events remain a future PostgreSQL-backed phase.
