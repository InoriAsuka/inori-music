# Phase 122 — CORS Middleware for Browser Client Access

**Version**: v1.22.0
**Date**: 2026-06-20

## Goal

Add a CORS middleware layer to the HTTP handler so browser-based clients (`inori-web`,
`inori-admin`) can make cross-origin requests to the API. The middleware must be configurable via
environment variables and must never expose sensitive headers.

## Requirements

### Middleware behaviour
- Intercept every request before the mux; add `Access-Control-Allow-Origin`,
  `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`,
  `Access-Control-Allow-Credentials`, and `Access-Control-Max-Age` headers on responses.
- Handle `OPTIONS` preflight requests: respond with `204 No Content` and the same CORS headers;
  do not forward to any route handler.
- Allowed methods (fixed): `GET, POST, PUT, PATCH, DELETE, OPTIONS`.
- Allowed headers (fixed): `Authorization, Content-Type, X-Request-ID`.
- `Access-Control-Allow-Credentials: true` always (the API uses Bearer tokens, not cookies, but
  browsers need this for header access on credentialed requests).
- `Access-Control-Max-Age: 86400` (24 h preflight cache).

### Origin policy
- New env var `INORI_CORS_ORIGINS` (comma-separated list of allowed origins, e.g.
  `http://localhost:5173,https://inori.example.com`).
- If `INORI_CORS_ORIGINS` is empty or unset: reflect the request `Origin` back (permissive mode
  for local development); log a warning at startup.
- If set: only reflect the origin when it matches a value in the list exactly; otherwise omit
  the `Access-Control-Allow-Origin` header (browser will block).
- Wildcard `*` is not permitted when `Allow-Credentials: true` (browser restriction); the
  reflect-or-match strategy above is the correct substitute.

### Implementation
- Implement as a pure `func corsMiddleware(origins []string) func(http.Handler) http.Handler`
  in `internal/httpapi/cors.go`.
- Wrap the mux returned by `Routes()` with this middleware before returning from `Routes()`.
- Pass `origins` from `main.go` (parsed from `INORI_CORS_ORIGINS`).
- Expose `WithCORSOrigins(origins []string) HandlerOption` so tests can configure the middleware
  without env vars.

### Tests
- Add `services/api/internal/httpapi/cors_test.go` with 6 unit tests:
  - `TestCORSPreflightReturns204`.
  - `TestCORSPreflightHeadersPresent`.
  - `TestCORSAllowedOriginReflected`.
  - `TestCORSDisallowedOriginOmitted`.
  - `TestCORSPermissiveModeReflectsAnyOrigin` (empty origins list).
  - `TestCORSNonPreflightPassesThrough`.
- Bump VERSION and OpenAPI `info.version` to `1.22.0`.
- Update `requirement.md` Current Version to `1.22.0`.
- Commit with tag `v1.22.0`.

## Non-Goals

- Rate limiting (not in scope for 1.x).
- Per-route CORS policy (single global policy is sufficient).
- `Vary: Origin` header management (clients are SPA; caching proxies are out of scope).

## Follow-Up Candidates

- `X-Request-ID` propagation middleware (Phase 123).
