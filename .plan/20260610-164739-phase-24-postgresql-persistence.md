# Phase 24: PostgreSQL Persistence

## Requirement Snapshot

Add PostgreSQL-backed repository implementations for storage backends and media objects, with automatic schema migration and a shared connection pool selected by `INORI_DATABASE_URL`. File and in-memory repositories remain available when the environment variable is unset.

## Task Checklist

- [x] Define this phase scope and non-goals.
- [x] Complete the corresponding code, API, or documentation updates.
- [x] Add or update the required tests.
- [x] Record the phase outcome for later review.

## Non-Goals

- Do not introduce role-based authorization or multi-user identity in this phase.
- Do not add encrypted configuration persistence in this phase.
- Do not migrate the OpenAPI contract routes in this phase.

## Follow-Up Candidates

- Add PostgreSQL-backed probe history and audit log tables.
- Add role-based administrator authorization backed by PostgreSQL user records.
- Add encrypted backend configuration persistence for secrets stored in the database.
- Add a dedicated database migration versioning table to replace idempotent IF NOT EXISTS guards.
