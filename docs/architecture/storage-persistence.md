# Storage Persistence Strategy

## Current Phase

Early 0.x uses in-memory repositories and optional JSON file-backed repositories. File repositories are intended for development and single-node self-hosting and persist state with temporary files, sync, and atomic rename.

## Environment Variables

- `INORI_STORAGE_REPOSITORY_FILE`: enable JSON file persistence for storage backends.
- `INORI_MEDIA_OBJECT_REPOSITORY_FILE`: enable JSON file persistence for media objects.

## Future Direction

Production metadata persistence should move to PostgreSQL. Domain services should avoid unnecessary coupling to a specific SQL implementation so migrations, indexes, transactions, and audit logging can be added later.
