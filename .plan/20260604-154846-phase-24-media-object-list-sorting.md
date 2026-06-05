# Phase 24: Media Object List Sorting (v0.24.0)

## Requirement Snapshot

Add deterministic sort controls to media object list APIs so admin clients can request predictable table order before pagination without changing repository persistence semantics.

## Task Checklist

- [x] Add `sortBy` and `sortOrder` fields to the media object list domain filter.
- [x] Normalize and validate supported sort fields and directions.
- [x] Sort filtered media object results before limit/offset pagination.
- [x] Wire HTTP query parameters into the list handler.
- [x] Update OpenAPI contract query parameter documentation.
- [x] Update README, requirements, architecture docs, and version baseline to `0.24.0`.
- [x] Add unit, handler, and OpenAPI contract tests for sort controls.
- [x] Run formatting, tests, vet, race tests, JSON validation, and diff checks.

## Non-Goals

- No database or repository schema changes.
- No changes to filter cardinality rules; exactly one metadata filter remains required.
- No changes to verification, lifecycle, or backend probe behavior.

## Follow-Up Candidates

- Add cursor-based pagination when PostgreSQL persistence is introduced.
- Add compound filters and server-side SQL index planning in the PostgreSQL phase.
