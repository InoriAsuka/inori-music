# Storage Administration HTTP API

## Scope

Phase 3 exposes static storage backend administration as a versioned HTTP JSON API. It allows an administrator-facing client to validate candidate configuration, register backends, list configured backends, select the default backend, and disable non-default backends.

This API authenticates storage administration routes with an administrator bearer token. It does not yet provide role-based authorization, user sessions, audit logging, persistent storage, or real filesystem, mount, credential, bucket, and distributed-cluster probes. The server binds to `127.0.0.1:8080` by default; deployments can override the listener with `INORI_HTTP_ADDR` after applying appropriate network controls.

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Return process health. |
| `GET` | `/api/v1/admin/storage/backends` | List configured storage backends. |
| `POST` | `/api/v1/admin/storage/backends` | Validate and register a backend. |
| `POST` | `/api/v1/admin/storage/backends/validate` | Validate a candidate without persisting it. |
| `POST` | `/api/v1/admin/storage/backends/{id}/default` | Select an enabled backend as default. |
| `POST` | `/api/v1/admin/storage/backends/{id}/disable` | Disable a non-default backend. |

## Authentication

Administrative endpoints require:

```http
Authorization: Bearer <INORI_ADMIN_TOKEN>
```

`GET /healthz` is intentionally unauthenticated for process and deployment health checks.

Runtime configuration:

| Environment variable | Purpose |
|---|---|
| `INORI_ADMIN_TOKEN` | Required admin bearer token. It must be at least 32 characters. |
| `INORI_INSECURE_DEV_AUTH=1` | Allows admin routes without a token for local development only. |
| `INORI_HTTP_ADDR` | Optional listener override. Defaults to `127.0.0.1:8080`. |

## Request Rules

- Administrative requests must include a valid bearer token unless explicit insecure development mode is enabled.
- JSON request bodies must use `Content-Type: application/json`.
- Request bodies are limited to 1 MiB.
- Unknown JSON fields are rejected.
- Each backend configuration must contain exactly one family branch matching its `type`.
- Static validation does not imply external connectivity or permission validation.

## Example

```bash
curl \
  --request POST \
  --header "Authorization: Bearer ${INORI_ADMIN_TOKEN}" \
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
| `401` | `unauthorized` | A valid admin bearer token is required. |
| `404` | `not_found` | The route or backend does not exist. |
| `405` | `method_not_allowed` | The route exists but does not support the requested method. |
| `409` | `conflict` | The requested lifecycle transition conflicts with current state. |
| `500` | `internal_error` | An unexpected server failure occurred. |
| `503` | `auth_not_configured` | Admin routes were constructed without a token and insecure mode is disabled. |

## Next Security Step

Before exposing these routes beyond a trusted loopback or development environment, add role-based administrator authorization, secret encryption, audit logging, OpenAPI security schemes, and deployment-level transport security.
