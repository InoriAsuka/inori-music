# Plan: Phase 4 Admin Authentication

## Requirement Version

v0.4.0

## Goals

- Fail closed for storage administration routes unless an administrator token is configured.
- Keep `/healthz` public so process supervisors, containers, and local smoke tests can check service liveness.
- Provide a minimal authentication boundary before adding persistent storage, probe checks, or broader administrative capabilities.
- Keep the authentication implementation standard-library only and easy to replace with stronger identity providers later.

## Phase 1: Requirement Update

- [x] Append `v0.4.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.4.0`.

## Phase 2: HTTP API Authentication

- [x] Add handler options for administrator Bearer Token authentication.
- [x] Require `Authorization: Bearer <token>` for `/api/v1/admin/*` routes.
- [x] Keep `/healthz` unauthenticated.
- [x] Fail closed with `503 admin_auth_not_configured` when no admin token is configured.
- [x] Return `401 unauthorized` for missing, malformed, or invalid credentials when auth is configured.
- [x] Use constant-time token comparison.

## Phase 3: Server Bootstrap

- [x] Read `INORI_ADMIN_TOKEN` in the server entry point.
- [x] Warn when admin routes are disabled because no token is configured.
- [x] Preserve the loopback default bind address.

## Phase 4: Verification

- [x] Add handler tests for public health, configured auth, missing auth, invalid auth, malformed auth, and auth-not-configured behavior.
- [x] Update existing handler tests to use authenticated admin requests.
- [x] Run `gofmt`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add persistent administrator users or service accounts.
- [ ] Add token hashing or external secret-manager integration for persisted credentials.
- [ ] Add authorization roles and audit logging.
- [ ] Add OpenAPI security scheme documentation.
- [ ] Add TLS and reverse-proxy deployment guidance.

## Completion Notes

This phase adds a bootstrap authentication boundary only. It is sufficient to prevent accidental unauthenticated local admin access, but it is not a replacement for production identity, authorization, audit, transport security, or secret rotation.
