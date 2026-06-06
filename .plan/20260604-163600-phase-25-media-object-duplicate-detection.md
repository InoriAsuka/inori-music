# Phase 25: Media Object Duplicate Detection (v0.25.0)

## Requirement Snapshot

Add a metadata-only duplicate detection report for media objects so administrators can identify content hashes referenced by multiple media objects before implementing storage cleanup, deduplication, or PostgreSQL-backed reporting.

## Task Checklist

- [x] Add duplicate report and duplicate group response models to the media object domain.
- [x] Group media objects by `contentHash` and report only groups meeting a configurable `minCopies` threshold.
- [x] Support optional `backendId` scoping without reading media bytes.
- [x] Expose an authenticated HTTP route for duplicate reporting.
- [x] Update OpenAPI schemas, route coverage, and query parameter tests.
- [x] Update README, requirements, architecture docs, and version baseline to `0.25.0`.
- [x] Add domain and handler tests for duplicate reports and invalid `minCopies`.
- [x] Run formatting, tests, vet, race tests, JSON validation, and diff checks.

## Non-Goals

- No automatic file deletion, content rewrite, or deduplication mutation.
- No database schema changes.
- No media-byte reads; this phase uses metadata only.

## Follow-Up Candidates

- Add admin-reviewed deduplication jobs after audit logging exists.
- Add PostgreSQL indexes on `content_hash` and `backend_id` during the persistence phase.
