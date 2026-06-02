# Plan: Phase 6 S3-Compatible Object Probes

## Requirement Version

v0.6.0

## Goals

- Add real S3-compatible object probes for cloud and object-storage deployments.
- Keep probe behavior safe by operating only on a short-lived application-owned object key.
- Resolve credentials through configured secret reference names without logging or returning secret values.
- Test S3 probe behavior with a local fake S3-compatible HTTP server.

## Phase 1: Requirement Update

- [x] Append `v0.6.0` to `requirement.md`.
- [x] Create this phase plan under `.plan/`.
- [x] Bump `VERSION` and README baseline to `0.6.0`.

## Phase 2: S3 Probe Domain

- [x] Add standard-library S3-compatible probe client with AWS Signature Version 4 signing. Network access blocked adding AWS SDK dependencies, so no third-party dependency was introduced.
- [x] Add a composite prober that tries filesystem probes first and then S3-compatible probes.
- [x] Add S3-compatible probe support for `s3` backends.
- [x] Add S3-compatible probe support for distributed backends with `adapter: s3-compatible`.
- [x] Resolve access and secret keys from environment variable names declared in backend config.
- [x] Put, full-read, range-read, and delete a short-lived probe object.
- [x] Ensure best-effort cleanup after object creation.

## Phase 3: Verification

- [x] Add fake S3-compatible HTTP server tests for put, full-read, range-read, and delete behavior.
- [x] Add tests for missing credentials and distributed S3-compatible probe routing.
- [x] Run `gofmt`.
- [x] Run `git diff --check`.
- [x] Run `go vet ./services/api/...`.
- [x] Run `go test ./services/api/...`.
- [x] Run `go test -race ./services/api/...`.

## Future Implementation Tasks

- [ ] Add scheduled probe refresh and capacity reporting.
- [ ] Add PostgreSQL persistence for latest health state and probe history.
- [ ] Add probe timeouts and retry policy configuration.
- [ ] Add object-storage provider compatibility notes for MinIO, R2, B2, Ceph RGW, Garage, and SeaweedFS.
- [ ] Add dedicated distributed adapter probes beyond S3-compatible and mounted-filesystem adapters.

## Completion Notes

This phase validates conservative S3-compatible object semantics through a standard-library HTTP client with AWS Signature Version 4 signing against configurable endpoints. It does not claim support for provider-specific lifecycle, versioning, object lock, or event-notification features.
