# Phase 34: HTTP Request Metrics

## Version

v0.34.0

## Requirement

Expand runtime metrics with low-cardinality HTTP request counters and duration sums so operators can monitor traffic and status outcomes without exposing raw URLs or credentials.

## Tasks

- [x] Wrap the HTTP mux with request instrumentation that captures method, route pattern, status, and elapsed duration.
- [x] Export `inori_api_http_requests_total` counters from `/metrics`.
- [x] Export cumulative `inori_api_http_request_duration_seconds_sum` metrics from `/metrics`.
- [x] Keep labels low-cardinality by using Go route patterns and status codes instead of raw request paths.
- [x] Add tests and update docs, requirements, README, and version tracking for v0.34.0.

## Notes

The metrics remain process-local and in-memory. Future production telemetry can add histograms or external metrics sinks once deployment topology is better defined.
