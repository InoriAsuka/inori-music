# Phase 60 — S3 presigned playback URL generation

## Goal

Extend the `GET /api/v1/catalog/tracks/{id}/playback` viewer endpoint to return
a pre-signed AWS Signature Version 4 GET URL when the linked media object's
storage backend has `PresignedURLs` capability. Clients can use the URL to fetch
audio bytes directly from S3-compatible storage without the API server streaming
bytes or exposing credentials.

## Requirements

- Add `presignS3URL` in `s3_probe.go`: generate an AWS SigV4 presigned GET URL
  using query-parameter signing (`X-Amz-Algorithm`, `X-Amz-Credential`,
  `X-Amz-Date`, `X-Amz-Expires`, `X-Amz-SignedHeaders`, `X-Amz-Signature`).
  Reuse existing `s3ObjectURL`, `s3SigningKey`, `hmacSHA256` helpers.
  Default empty `Region` to `us-east-1`. TTL encoded as `X-Amz-Expires` seconds.
- Add `storage.Service.GetBackend(ctx, id)` — single-backend lookup via
  `repository.Get`; used by the handler instead of scanning the full list.
- Add `storage.DefaultPresignedURLTTL = 15 * time.Minute` constant.
- Add `storage.Service.GeneratePresignedURL(ctx, backendID, objectKey, ttl)`:
  calls `repository.Get` → checks `Capabilities.PresignedURLs` → `s3ProbeConfig`
  → `resolveS3ProbeCredentials` → `presignS3URL`. Returns `ErrProbeUnsupported`
  when the backend lacks capability; `ErrProbeFailed` when credentials missing.
- Extend `trackPlaybackDescriptor` with optional `PresignedURL string \`json:"presignedUrl,omitempty"\``.
- Update `getTrackPlayback` handler: replace full-list scan with `GetBackend`;
  if `backend.Capabilities.PresignedURLs`, call `GeneratePresignedURL` and
  populate `PresignedURL`; presign failures are non-fatal (field left empty).
- Add `presignedUrl` property to `TrackPlaybackDescriptor` OpenAPI schema
  (optional, not in `required`).
- Bump OpenAPI `info.version` → `0.60.0`.

## Non-goals

- No S3 streaming proxy or byte serving through the API server.
- No signed URLs for local/NFS/SMB backends.
- No credential management or secrets manager integration.
- No TTL configurability (15 min is fixed via constant).

## Tasks

- [x] Add `presignS3URL` to `services/api/internal/storage/s3_probe.go`.
- [x] Add `GetBackend`, `DefaultPresignedURLTTL`, `GeneratePresignedURL` to `services/api/internal/storage/service.go`.
- [x] Extend `trackPlaybackDescriptor` and `getTrackPlayback` in `handler.go`.
- [x] Add 4 `presignS3URL` unit tests to `s3_probe_test.go`.
- [x] Add 4 `GetBackend`/`GeneratePresignedURL` service tests to `service_test.go`.
- [x] Add `TestGetTrackPlaybackDescriptorPresignedURL` handler test.
- [x] Update `TestStorageAdminOpenAPIContractTrackPlaybackDescriptor` to assert `presignedUrl` property present and not required.
- [x] Add `presignedUrl` property to `TrackPlaybackDescriptor` in OpenAPI contract.
- [x] Bump OpenAPI `info.version` → `0.60.0`.
- [x] Bump `VERSION` → `0.60.0`.
- [x] Update `requirement.md` current version and append v0.60.0 history entry.
- [x] Run full `go test ./...` — green.
- [x] Commit, tag `v0.60.0`, push.

## Follow-up candidates

- Configurable TTL via request header or query parameter.
- Repository-level SQL pushdown for stats and timeline queries.
- Playback history / listening activity domain.
- Playlist recommendations or catalog search ranking improvements.
