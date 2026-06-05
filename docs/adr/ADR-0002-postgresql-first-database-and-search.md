# ADR-0002: PostgreSQL-First Database and Search Strategy

## Status

Accepted.

## Context

The system needs reliable server-side metadata storage, transactions, indexes, and baseline full-text search. Clients still need lightweight local persistence for caches and offline queues.

## Decision

The 0.x server uses PostgreSQL as the primary metadata database and starts with PostgreSQL full-text search. Client local persistence uses SQLite. External search engines remain future extensions rather than 0.x requirements.

## Consequences

This reduces early operational complexity while preserving a path to external search services when scale, language quality, or fuzzy matching needs grow.
