# Phase 19: Media Object Lifecycle Administration (v0.19.0)

## Requirement Snapshot

- Allow administrators to update media object lifecycle metadata after registration.
- Lifecycle updates must be metadata-only and must not delete, move, or rewrite media bytes.
- Preserve integrity history such as `lastVerification` when lifecycle state changes.

## Task Checklist

- [x] Add service-level media object lifecycle state mutation with validation.
- [x] Keep `deleted` as a terminal metadata state that cannot move back to non-deleted states.
- [x] Preserve existing media object metadata and latest verification state while updating `lifecycleState` and `updatedAt`.
- [x] Add authenticated `POST /api/v1/admin/media/objects/{id}/lifecycle`.
- [x] Update OpenAPI with the lifecycle request schema and route.
- [x] Update README, requirements, and architecture documentation for v0.19.0.
- [x] Add domain and HTTP tests for valid lifecycle updates, invalid states, deleted-state conflicts, and authentication.
- [x] Run formatting, static checks, JSON contract parsing, unit tests, race tests, and diff checks.

## Non-Goals

- No physical deletion or movement of media bytes.
- No bulk lifecycle update endpoint in this phase.
- No library ownership or user-facing moderation workflow yet.

## Follow-Up Candidates

- Add bulk lifecycle updates after import job tracking exists.
- Add audit events for lifecycle transitions.
- Add background cleanup policies for deleted metadata after retention settings are designed.
