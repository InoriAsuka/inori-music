# Phase 17: Media Object List Pagination (v0.17.0)

## Requirement Snapshot

- Bound media-object list responses before large library imports produce thousands of metadata records.
- Preserve the existing single-filter rule from v0.16.0 while adding explicit `limit` and `offset` controls.
- Return pagination metadata so admin clients can build deterministic review workflows.

## Task Checklist

- [x] Add media object list filter and page response domain types.
- [x] Add default and maximum list limits for server-side bounded responses.
- [x] Implement service-level pagination for backend, content-hash, and verification-status list filters.
- [x] Extend the HTTP media-object list route with `limit` and `offset` query parameters.
- [x] Return `pagination.limit`, `pagination.offset`, `pagination.total`, and `pagination.hasMore` in list responses.
- [x] Update OpenAPI, requirements, README, and architecture docs for v0.17.0.
- [x] Add domain and HTTP tests for paginated responses and invalid pagination inputs.
- [x] Run formatting, static checks, JSON contract parsing, unit tests, race tests, and diff checks.

## Non-Goals

- No cursor pagination or database-backed keyset pagination yet.
- No full-library unfiltered listing endpoint.
- No search or sorting customization beyond the existing stable backend/object-key order.

## Follow-Up Candidates

- Add cursor or keyset pagination once PostgreSQL persistence lands.
- Add aggregate verification counters for failed and unknown objects.
- Add admin UI pagination controls generated from the OpenAPI contract.
