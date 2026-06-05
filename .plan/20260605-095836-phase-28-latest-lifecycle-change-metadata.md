# Phase 28: Latest Lifecycle Change Metadata (v0.28.0)

## Requirement Snapshot

Persist latest committed media object lifecycle transition metadata so administrators can inspect the most recent lifecycle change and prepare for future durable audit event storage.

## Task Checklist

- [x] Add latest lifecycle change metadata to media object responses.
- [x] Record previous lifecycle state, new lifecycle state, changed time, and single/bulk source for committed single-object lifecycle updates.
- [x] Record the same metadata for committed bulk lifecycle updates.
- [x] Keep dry-run bulk lifecycle previews non-persistent.
- [x] Update OpenAPI schemas and contract tests.
- [x] Update README, requirements, architecture docs, and version baseline to `0.28.0`.
- [x] Add domain tests for single and bulk lifecycle change metadata.
- [x] Run formatting, tests, vet, race tests, JSON validation, and diff checks.

## Non-Goals

- No append-only audit event table or repository.
- No actor identity model beyond existing bootstrap auth.
- No media-byte deletion.

## Follow-Up Candidates

- Add append-only audit events with actor and request metadata.
- Add file-backed and PostgreSQL-backed audit repositories.
