# Storage Administration HTTP API

## Scope

Phase 3 exposes static storage backend administration as a versioned HTTP JSON API. It allows an administrator-facing client to validate candidate configuration, register backends, list configured backends, select the default backend, and disable non-default backends.

This API authenticates administrator requests with a bootstrap Bearer Token from `INORI_ADMIN_TOKEN`. The server binds to `127.0.0.1:8080` by default. Deployments can override the listener with `INORI_HTTP_ADDR` after applying appropriate network controls.

If `INORI_ADMIN_TOKEN` is not configured, `/api/v1/admin/*` routes fail closed with `503 admin_auth_not_configured`. `/healthz` remains public for process supervisors, container health checks, and local smoke tests.

This API supports safe real filesystem probes for LocalSystem, NFS, SMB, and `distributed` backends using the `mounted-filesystem` adapter. Filesystem probes create, write, full-read, range-read, and remove only a short-lived server-owned probe file inside the configured root. The server does not mount or unmount remote filesystems.

This API also supports conservative S3-compatible object probes for `s3` backends and `distributed` backends using the `s3-compatible` adapter. S3-compatible probes put, full-read, range-read, and delete only a short-lived server-owned probe object under `.inori-music-probe/`. Static validation still only checks request shape and configuration consistency; the explicit probe endpoint checks supported mounted filesystem or S3-compatible semantics.

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Return process health. |
| `GET` | `/api/v1/admin/storage/backends` | List configured storage backends. |
| `POST` | `/api/v1/admin/storage/backends` | Validate and register a backend. |
| `POST` | `/api/v1/admin/storage/backends/validate` | Validate a candidate without persisting it. |
| `POST` | `/api/v1/admin/storage/backends/{id}/default` | Select an enabled backend as default. |
| `POST` | `/api/v1/admin/storage/backends/{id}/disable` | Disable a non-default backend. |
| `POST` | `/api/v1/admin/storage/backends/{id}/probe` | Run a safe real backend probe where supported. |
| `GET` | `/api/v1/admin/storage/backends/{id}/health` | Read the latest recorded backend health state. |

## Authentication

Admin routes require:

```http
Authorization: Bearer <INORI_ADMIN_TOKEN>
```

Missing, malformed, or invalid credentials return `401 unauthorized` when the token is configured. If the token is not configured, admin routes return `503 admin_auth_not_configured`.

## Request Rules

- JSON request bodies must use `Content-Type: application/json`.
- Request bodies are limited to 1 MiB.
- Unknown JSON fields are rejected.
- Each backend configuration must contain exactly one family branch matching its `type`.
- Static validation does not imply external connectivity or permission validation.
- Filesystem probes operate only on an application-owned `.inori-music-probe-*` temporary file and clean it up after the check.
- S3-compatible probes operate only on an application-owned `.inori-music-probe/*` object key and clean it up after the check.
- NFS and SMB mounts must already exist at the host level; the application does not mount remote shares.
- S3-compatible credentials are resolved from environment variable names in `accessKeySecretRef` and `secretKeySecretRef`; secret values must not be stored in repository files.

## Example

```bash
curl \
  --request POST \
  --header 'Authorization: Bearer <INORI_ADMIN_TOKEN>' \
  --header 'Content-Type: application/json' \
  --data '{
    "id": "local-main",
    "type": "local",
    "displayName": "Local media",
    "enabled": true,
    "isDefault": true,
    "config": {
      "local": {
        "rootPath": "/srv/inori/media"
      }
    }
  }' \
  http://127.0.0.1:8080/api/v1/admin/storage/backends
```

## Error Envelope

Errors use a stable JSON envelope:

```json
{
  "error": {
    "code": "invalid_backend",
    "message": "invalid storage backend: id is required"
  }
}
```

Current error codes:

| HTTP status | Code | Meaning |
|---:|---|---|
| `400` | `invalid_backend` | JSON or backend configuration is invalid. |
| `401` | `unauthorized` | The admin Bearer Token is missing, malformed, or invalid. |
| `404` | `not_found` | The route or backend does not exist. |
| `405` | `method_not_allowed` | The route exists but does not support the requested method. |
| `409` | `conflict` | The requested lifecycle transition conflicts with current state, such as probing a disabled backend. |
| `422` | `probe_unsupported` | The backend does not yet have a real probe adapter, such as a future dedicated distributed adapter; its health state remains unchanged. |
| `422` | `probe_failed` | A supported real probe could not complete successfully. |
| `500` | `internal_error` | An unexpected server failure occurred. |
| `503` | `admin_auth_not_configured` | No bootstrap admin token has been configured. |

## Next Security Step

Before exposing these routes beyond a trusted loopback or development environment, replace the bootstrap token with persistent administrator identities or service accounts, add authorization roles, encrypt stored secrets, record audit logs, define token rotation, and apply deployment-level transport security.
