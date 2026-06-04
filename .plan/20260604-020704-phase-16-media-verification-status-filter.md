# Phase 16: Media Verification Status Filter (v0.16.0)

## Requirement Snapshot

- Add server-side media object listing by latest verification state so administrators can quickly find verified, failed, and never-verified assets.
- Preserve the existing metadata-only model: filtering must not read media bytes and must rely only on persisted `lastVerification` state.
- Keep the HTTP API strict by requiring exactly one list filter at a time.

## Task Checklist

- [x] Extend `MediaObjectRepository` with a verification-status listing method.
- [x] Implement status filtering for the in-memory media object repository.
- [x] Implement status filtering for the file-backed media object repository.
- [x] Add service-level validation for allowed filter states: `verified`, `failed`, and `unknown`.
- [x] Extend `GET /api/v1/admin/media/objects` with the `verificationStatus` query parameter while preserving single-filter semantics.
- [x] Update the OpenAPI contract to document the new query parameter and version.
- [x] Update README, requirements, and architecture docs with the v0.16.0 behavior.
- [x] Add storage and HTTP handler tests for successful and invalid verification-status filtering.
- [x] Run formatting, static checks, test suites, and diff checks.

## Non-Goals

- No full-text search, SQL persistence, or asynchronous verification queue in this phase.
- No object-byte reads are performed by the list endpoint.
- No deletion or lifecycle mutation is introduced.

## Follow-Up Candidates

- Add a paginated/sorted list API before large-library imports.
- Introduce aggregate verification counters for dashboard cards.
- Add a background re-verification scheduler once persistent job storage exists.
