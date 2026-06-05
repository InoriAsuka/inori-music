# Phase 27: Bulk Lifecycle Dry Runs (v0.27.0)

## Requirement Snapshot

Add dry-run support to metadata-only bulk media object lifecycle updates so administrators can preview selected objects and per-object outcomes before persisting lifecycle changes.

## Task Checklist

- [x] Add dry-run options and report counters to the bulk lifecycle domain flow.
- [x] Return `would_update` results without saving media object metadata when dry-run mode is enabled.
- [x] Wire `dryRun` through the authenticated bulk lifecycle HTTP route.
- [x] Update OpenAPI request and response schemas for dry-run previews.
- [x] Update README, requirements, architecture docs, and version baseline to `0.27.0`.
- [x] Add domain and handler tests proving dry runs do not persist lifecycle changes.
- [x] Run formatting, tests, vet, race tests, JSON validation, and diff checks.

## Non-Goals

- No asynchronous job execution.
- No audit event persistence.
- No physical media deletion.

## Follow-Up Candidates

- Add audit logging for dry-run and committed lifecycle requests.
- Add job IDs for large lifecycle updates once PostgreSQL persistence exists.
