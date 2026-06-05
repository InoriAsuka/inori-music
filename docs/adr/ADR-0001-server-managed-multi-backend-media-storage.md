# ADR-0001: Server-Managed Multi-Backend Media Storage

## Status

Accepted.

## Context

The music system must support local self-hosting, NAS paths, object storage, and distributed storage. Clients should not decide storage placement directly, and large media bytes should not be stored in the relational database.

## Decision

The server is the source of truth for storage backend configuration, validation, health state, capacity metadata, and default backend selection. Media objects store reference metadata only; bytes live in configured storage backends.

## Consequences

This simplifies clients and enables migration, probing, audit, and permission controls. It also requires a stable server-side backend abstraction and conservative probe semantics.
