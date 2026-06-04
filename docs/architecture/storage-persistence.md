# Storage Administration Persistence

## Scope

Phase 9 adds an optional durable repository for server-managed storage backend configuration before the PostgreSQL persistence layer is introduced. Phase 12 applies the same bootstrap persistence pattern to media object metadata.

The API server keeps `MemoryRepository` and `MemoryMediaObjectRepository` as the defaults for tests and ephemeral development. Operators can set `INORI_STORAGE_REPOSITORY_FILE` to enable the file-backed storage backend repository:

```bash
INORI_STORAGE_REPOSITORY_FILE=/var/lib/inori-music/storage-backends.json
```

Set `INORI_MEDIA_OBJECT_REPOSITORY_FILE` to enable durable media object metadata references:

```bash
INORI_MEDIA_OBJECT_REPOSITORY_FILE=/var/lib/inori-music/media-objects.json
```

## File Repository Semantics

The storage backend file repository stores backend configuration, server-owned health state, capacity reports, timestamps, and inferred capabilities in a JSON document. It does not store raw storage credentials; S3-compatible backends continue to reference environment variable names through `accessKeySecretRef` and `secretKeySecretRef`.

The media object file repository stores metadata-only binary asset references, including backend ID, object key, content hash, byte size, MIME type, asset kind, lifecycle state, and server-owned timestamps. It does not store media bytes.

Writes use a conservative local-filesystem sequence:

1. Create parent directories with owner-only permissions.
2. Encode the complete repository document to a temporary file in the same directory.
3. Sync and close the temporary file.
4. Atomically rename the temporary file over the configured repository file.

This makes single-process writes crash-tolerant on normal local filesystems. It is not a distributed lock manager and should not be shared by multiple API server processes.

## Recommended Use

Use file repositories for:

- Local development.
- Single-node self-hosted deployments.
- Bootstrap environments before PostgreSQL is available.
- Migration staging and fixtures.

Use PostgreSQL, once implemented, for multi-user administration, horizontal API deployments, and production metadata durability.

## Future PostgreSQL Direction

The long-term server metadata direction remains PostgreSQL-first. The eventual PostgreSQL repository should preserve the same domain-level `storage.Repository` interface while adding migrations, transactional default-backend selection, audit fields, and optimistic concurrency.
