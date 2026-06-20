# Phase 123 — Request-ID Propagation Middleware

**Version**: v1.23.0
**Date**: 2026-06-20

## Goal

Add a lightweight request-ID middleware that reads `X-Request-ID` from the incoming request (or
generates a new one) and echoes it on every response. This lets browser clients and operators
correlate log lines with specific API calls without requiring a tracing backend.

## Requirements

### Middleware behaviour
- Read `X-Request-ID` from the request header.
- If present and non-empty: use it as-is (trusts the upstream proxy / client).
- If absent or empty: generate a new ID — 16 hex bytes from `crypto/rand` formatted as a
  32-character lowercase hex string.
- Set `X-Request-ID` on the response header before writing any body.
- Inject the request ID into the request context so handlers can reference it in log lines.

### Implementation
- `internal/httpapi/requestid.go` — `requestIDMiddleware() func(http.Handler) http.Handler`.
- `requestIDFromContext(ctx context.Context) string` helper for handler use.
- Chain order in `Routes()`: `requestIDMiddleware` wraps the mux first, then `corsMiddleware`
  wraps the result (so CORS headers are added after the request ID is set).

### Tests
- Add `internal/httpapi/requestid_test.go` with 4 tests:
  - `TestRequestIDPassthroughExisting`: existing header echoed back unchanged.
  - `TestRequestIDGeneratedWhenAbsent`: generated ID is 32 lowercase hex chars.
  - `TestRequestIDPresentOnAllRoutes`: request to `/healthz` includes the header.
  - `TestRequestIDInjectedIntoContext`: context value accessible from a handler stub.
- Bump VERSION and OpenAPI `info.version` to `1.23.0`.
- Update `requirement.md` Current Version to `1.23.0`.
- Commit with tag `v1.23.0`.

## Non-Goals

- Structured logging / log injection (out of scope for 1.x).
- Trace context propagation (W3C `traceparent` — future work).
- Request ID length / format configuration via env var.

## Follow-Up Candidates

- README and documentation sync (Phase 124).
