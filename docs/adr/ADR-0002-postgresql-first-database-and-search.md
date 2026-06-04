# ADR-0002: PostgreSQL-First Database and Search Direction

## Status

Accepted

## Date

2026-06-02

## Context

Inori Music requires structured metadata for tracks, artists, albums, playlists, user libraries, devices, storage objects, import jobs, and audit records. It also needs search capabilities for titles, artists, albums, aliases, tags, and lyrics.

Supporting many server-side database engines in 0.x would increase migration, indexing, query, testing, and operational complexity before the core product has stabilized.

## Decision

The 0.x server-side primary database will be PostgreSQL-first. Client-side local persistence will use SQLite where offline playback queues, cache indexes, and local search are needed.

Server-side search will begin with PostgreSQL full-text search, normalized fields, alias tables, ranking rules, and appropriate indexes. External search engines remain future options when search scale or quality requirements exceed PostgreSQL capabilities.

## Consequences

### Positive

- PostgreSQL supports relational modeling, transactions, JSONB metadata, indexes, and built-in full-text search.
- SQLite is well suited for client-local offline state and cache indexes.
- The project avoids a large 0.x database compatibility matrix.
- Search can start simple while preserving an explicit `SearchService` boundary for future providers.

### Negative

- Deployments that require MySQL or MariaDB are not first-class 0.x targets.
- Advanced multilingual search, typo tolerance, and large-scale lyrics search may require a future external engine.
- Repository and migration code should still be structured carefully to avoid unnecessary domain coupling.

## Future Triggers

Consider an external search provider when one or more of these become product requirements:

- Strong Chinese, Japanese, or Korean search quality beyond PostgreSQL defaults.
- Typo tolerance and high-quality search-as-you-type.
- Large-scale lyrics indexing.
- Search latency consistently exceeds product targets.
- Search load begins to affect the primary database.
