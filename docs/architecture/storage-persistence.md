# Storage Administration Persistence

## Scope

Phase 9 adds an optional durable repository for server-managed storage backend configuration before the PostgreSQL persistence layer is introduced.

The API server keeps `MemoryRepository` as the default for tests and ephemeral development. Operators can set `INORI_STORAGE_REPOSITORY_FILE` to enable the file-backed repository:

```bash
INORI_STORAGE_REPOSITORY_FILE=/var/lib/inori-music/storage-backends.json
```

## File Repository Semantics

The file repository stores backend configuration, server-owned health state, capacity reports, timestamps, and inferred capabilities in a JSON document. It does not store raw storage credentials; S3-compatible backends continue to reference environment variable names through `accessKeySecretRef` and `secretKeySecretRef`.

Writes use a conservative local-filesystem sequence:

1. Create parent directories with owner-only permissions.
2. Encode the complete repository document to a temporary file in the same directory.
3. Sync and close the temporary file.
4. Atomically rename the temporary file over the configured repository file.

This makes single-process writes crash-tolerant on normal local filesystems. It is not a distributed lock manager and should not be shared by multiple API server processes.

## Recommended Use

Use the file repository for:

- Local development.
- Single-node self-hosted deployments.
- Bootstrap environments before PostgreSQL is available.
- Migration staging and fixtures.

Use PostgreSQL, once implemented, for multi-user administration, horizontal API deployments, and production metadata durability.

## Future PostgreSQL Direction

The long-term server metadata direction remains PostgreSQL-first. The eventual PostgreSQL repository should preserve the same domain-level `storage.Repository` interface while adding migrations, transactional default-backend selection, audit fields, and optimistic concurrency.
