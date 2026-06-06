# Phase 33: Runtime Metrics Endpoint

## Version

v0.33.0

## Requirement

Expose non-sensitive runtime metrics so deployment monitoring can scrape readiness and build metadata without administrator credentials.

## Tasks

- [x] Add a public `/metrics` endpoint using Prometheus text exposition format.
- [x] Export readiness gauge metrics aligned with `/readyz` checks.
- [x] Export API build metadata as an info gauge aligned with `/versionz`.
- [x] Extend OpenAPI and contract tests for the public metrics route.
- [x] Update README, requirements, operations docs, and phase tracking for v0.33.0.

## Notes

The metrics endpoint intentionally avoids media object, backend secret, and credential details. It is designed for bootstrap deployment monitoring and can be expanded later with latency histograms.
