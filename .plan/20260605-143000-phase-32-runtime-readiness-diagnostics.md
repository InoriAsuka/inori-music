# Phase 32: Runtime Readiness Diagnostics

## Version

v0.32.0

## Requirement

Expose public readiness diagnostics so deployment automation can distinguish process liveness from API readiness before routing administrator traffic.

## Tasks

- [x] Add a public `/readyz` endpoint that reports storage service, media registry, and admin-auth readiness checks.
- [x] Return `200` when all readiness checks pass and `503` when one or more required checks fail.
- [x] Extend OpenAPI and contract tests with readiness report schemas and public security behavior.
- [x] Add handler tests for ready and not-ready responses.
- [x] Add a Docker liveness healthcheck and update release/container documentation.

## Notes

`/healthz` remains a lightweight process liveness probe. `/readyz` is intended for deployment readiness gates and intentionally exposes no secrets.
