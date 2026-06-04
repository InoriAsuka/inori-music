# Phase 18: Media Object Metadata Statistics (v0.18.0)

## Requirement Snapshot

- Provide metadata-only aggregate statistics for administrator dashboards and operational review.
- Summaries must not read media bytes and must use only registered media object metadata and persisted `lastVerification` state.
- Support both all-backend summaries and optional per-backend summaries.

## Task Checklist

- [x] Add a `MediaObjectStats` domain response with total object count, total size, backend, asset-kind, lifecycle, and verification-status buckets.
- [x] Add repository support for listing all media object metadata in stable order.
- [x] Implement service-level statistics aggregation without touching storage backends.
- [x] Add authenticated `GET /api/v1/admin/media/objects/stats` with optional `backendId` filtering.
- [x] Update OpenAPI with the stats path and schema.
- [x] Update README, requirements, and architecture documentation for v0.18.0.
- [x] Add domain and HTTP tests for global/per-backend statistics and authentication.
- [x] Run formatting, static checks, JSON contract parsing, unit tests, race tests, and diff checks.

## Non-Goals

- No time-series metrics, Prometheus exporter, or dashboard UI in this phase.
- No media-byte verification or background re-verification is triggered by the stats endpoint.
- No PostgreSQL aggregate query implementation until database persistence lands.

## Follow-Up Candidates

- Add dashboard-friendly failed/unknown verification counters to the future web admin UI.
- Add aggregate import progress and library ownership dimensions after library models exist.
- Move aggregation to indexed SQL queries when PostgreSQL persistence is introduced.
