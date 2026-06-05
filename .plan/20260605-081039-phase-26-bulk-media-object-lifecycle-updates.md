# Phase 26: Bulk Media Object Lifecycle Updates (v0.26.0)

## Requirement Snapshot

Add a metadata-only bulk lifecycle update workflow so administrators can move selected media objects between staged, active, archived, and deleted lifecycle states without deleting media bytes or changing storage backend references.

## Task Checklist

- [x] Add media object selection and lifecycle update report models.
- [x] Implement bulk lifecycle updates selected by exactly one metadata filter.
- [x] Preserve terminal `deleted` semantics and report per-object failures instead of deleting bytes.
- [x] Expose an authenticated HTTP route for bulk lifecycle updates.
- [x] Update OpenAPI schemas, route coverage, and request/response documentation.
- [x] Update README, requirements, architecture docs, and version baseline to `0.26.0`.
- [x] Add domain and handler tests for bulk lifecycle reports and invalid filters.
- [x] Run formatting, tests, vet, race tests, JSON validation, and diff checks.

## Non-Goals

- No physical media deletion.
- No storage backend mutation.
- No database schema changes.

## Follow-Up Candidates

- Add audit events for every lifecycle update.
- Add asynchronous job execution for very large selections once PostgreSQL is introduced.
