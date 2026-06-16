# Storage Persistence Strategy

## Current Phase

The server supports three persistence tiers selected at startup:

1. **PostgreSQL** (production) — enabled by setting `INORI_DATABASE_URL`. A shared `pgxpool.Pool` is opened on startup, schema migrations run automatically, and both storage backends and media objects are persisted to PostgreSQL tables.
2. **JSON file** (single-node development/self-hosting) — enabled by `INORI_STORAGE_REPOSITORY_FILE` and `INORI_MEDIA_OBJECT_REPOSITORY_FILE`. State is persisted with atomic rename.
3. **In-memory** (default fallback) — state is lost on restart. Suitable for local development and test runs.

PostgreSQL takes precedence when `INORI_DATABASE_URL` is set. File and in-memory tiers are used only when PostgreSQL is not configured.

## Environment Variables

| Variable | Purpose |
|---|---|
| `INORI_DATABASE_URL` | PostgreSQL connection string (e.g. `postgres://user:pass@host:5432/db`). When set, both storage backend and media object repositories use PostgreSQL. |
| `INORI_STORAGE_REPOSITORY_FILE` | JSON file path for storage backends. Used when `INORI_DATABASE_URL` is not set. |
| `INORI_MEDIA_OBJECT_REPOSITORY_FILE` | JSON file path for media objects. Used when `INORI_DATABASE_URL` is not set. |

## Schema

Schema migrations run at startup via `postgres.Migrate`. Migrations use `IF NOT EXISTS` guards and are safe to re-run.

- `storage_backends` — one row per registered storage backend, config and capabilities stored as JSONB.
- `media_objects` — one row per registered media object reference, last verification result stored as JSONB.

## Integration Tests

PostgreSQL repository tests live under `services/api/internal/storage/postgres/` and use `//go:build integration`. They require Docker and spin up a real PostgreSQL container via `testcontainers-go`. Run with:

```
go test -tags integration ./services/api/internal/storage/postgres/...
```

## Future Direction

- Add a versioned migration table to replace `IF NOT EXISTS` idempotency guards.
- Add PostgreSQL-backed probe history and audit log tables.
- Add role-based authorization and session management backed by PostgreSQL.
- Add encrypted backend configuration persistence for credentials stored in the database.
