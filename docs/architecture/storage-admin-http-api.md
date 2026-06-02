# Storage Administration HTTP API

## Scope

Phase 3 exposes static storage backend administration as a versioned HTTP JSON API. It allows an administrator-facing client to validate candidate configuration, register backends, list configured backends, select the default backend, and disable non-default backends.

This API does not yet authenticate administrators or probe real filesystems, mounts, credentials, buckets, or distributed clusters. The server binds to `127.0.0.1:8080` by default so the unauthenticated scaffold is not exposed on every network interface accidentally. Deployments can override the listener with `INORI_HTTP_ADDR` after applying appropriate network controls.

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Return process health. |
| `GET` | `/api/v1/admin/storage/backends` | List configured storage backends. |
| `POST` | `/api/v1/admin/storage/backends` | Validate and register a backend. |
| `POST` | `/api/v1/admin/storage/backends/validate` | Validate a candidate without persisting it. |
| `POST` | `/api/v1/admin/storage/backends/{id}/default` | Select an enabled backend as default. |
| `POST` | `/api/v1/admin/storage/backends/{id}/disable` | Disable a non-default backend. |

## Request Rules

- JSON request bodies must use `Content-Type: application/json`.
- Request bodies are limited to 1 MiB.
- Unknown JSON fields are rejected.
- Each backend configuration must contain exactly one family branch matching its `type`.
- Static validation does not imply external connectivity or permission validation.

## Example

```bash
curl \
  --request POST \
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
| `404` | `not_found` | The route or backend does not exist. |
| `405` | `method_not_allowed` | The route exists but does not support the requested method. |
| `409` | `conflict` | The requested lifecycle transition conflicts with current state. |
| `500` | `internal_error` | An unexpected server failure occurred. |

## Next Security Step

Before exposing these routes beyond a trusted loopback or development environment, add administrator authentication, authorization middleware, secret encryption, audit logging, and deployment-level transport security.
