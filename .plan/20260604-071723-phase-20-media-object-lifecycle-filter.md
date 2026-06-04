# Phase 20: Media Object Lifecycle Filter (v0.20.0)

## Requirement Snapshot

- Allow administrators to list media object metadata by lifecycle state after lifecycle administration is available.
- Preserve the single-filter list rule while adding `lifecycleState` as a first-class metadata filter.
- Keep the operation metadata-only and compatible with existing pagination.

## Task Checklist

- [x] Extend the media object list filter with `lifecycleState`.
- [x] Add repository support for listing media objects by lifecycle state in stable order.
- [x] Add service-level lifecycle filter validation for `staged`, `active`, `archived`, and `deleted`.
- [x] Extend `GET /api/v1/admin/media/objects` with the `lifecycleState` query parameter and existing pagination.
- [x] Update OpenAPI, requirements, README, and architecture docs for v0.20.0.
- [x] Add domain and HTTP tests for lifecycle filtering and invalid lifecycle filters.
- [x] Run formatting, static checks, JSON contract parsing, unit tests, race tests, and diff checks.

## Non-Goals

- No bulk lifecycle changes in this phase.
- No physical media deletion or cleanup policy.
- No multi-filter query composition until PostgreSQL-backed indexes are designed.

## Follow-Up Candidates

- Add bulk lifecycle state update endpoints for import/admin workflows.
- Add audit events for lifecycle transitions.
- Add SQL-backed lifecycle indexes when persistence moves to PostgreSQL.
