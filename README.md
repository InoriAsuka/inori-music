# inori-music

音乐集中平台，目标是构建支持 Web、Android、iOS、PC 等平台的全平台音乐播放系统，同时兼容 B/S 与 C/S 架构。

## Version

Current architecture baseline version: `0.5.0`.

## 0.x Architecture Direction

The 0.x line focuses on a pragmatic architecture that can grow from local self-hosting to production deployments:

- Cross-platform player direction: Flutter-first.
- Server direction: Go modular monolith first.
- Server metadata database: PostgreSQL-first.
- Client local persistence: SQLite.
- Initial search direction: PostgreSQL full-text search with future external search providers when needed.
- Media storage direction: server-managed multi-backend storage.

## Phase 1: Storage Architecture

The first phase establishes the media storage architecture before runtime implementation. The server backend will manage storage backend configuration and support multiple backend families:

- `local`: LocalSystem filesystem storage for development and single-node self-hosting.
- `nfs`: NFS-mounted storage for LAN and NAS-style deployments.
- `smb`: SMB/CIFS-mounted storage for Windows shares and mixed-platform networks.
- `s3`: S3-compatible object storage for cloud and object-storage deployments.
- `distributed`: distributed storage such as Ceph, Garage, SeaweedFS, or similar systems through S3-compatible APIs, mounted filesystems, or future dedicated adapters.

Large audio files are stored outside the relational database. The database stores metadata, backend IDs, object keys, content hashes, lifecycle state, and references.

## Phase 2: Storage Domain Scaffold

The second phase starts the Go API service scaffold and implements the storage administration domain as executable, tested server-side code. The initial domain package validates backend configuration, infers backend capabilities, manages default backend selection, and provides an in-memory repository for early development.

## Phase 3: Storage Admin HTTP API

The third phase exposes the storage administration domain through a runnable, versioned HTTP JSON API. The server provides health, validation, registration, listing, default-selection, and disable operations while keeping real storage probes, authentication, persistence, and OpenAPI contracts as subsequent tasks.

## Phase 4: Admin Authentication

The fourth phase protects storage administration routes with Bearer Token authentication. `/healthz` remains public, while `/api/v1/admin/*` routes fail closed unless `INORI_ADMIN_TOKEN` is configured. Missing or invalid credentials receive stable JSON error envelopes.

## Phase 5: Filesystem Health Probes

The fifth phase adds safe real health probes for LocalSystem, NFS, SMB, and mounted-filesystem distributed backends. Each probe creates, reads, range-reads, and removes only a short-lived server-owned file inside the configured root, then records the latest backend health state.

## Run the API Scaffold

```bash
INORI_ADMIN_TOKEN=change-me-development-token go run ./services/api/cmd/server
```

The HTTP server binds to `127.0.0.1:8080` by default. Admin routes require `Authorization: Bearer <INORI_ADMIN_TOKEN>`. Override the listener with `INORI_HTTP_ADDR` only after applying appropriate network controls. See [`docs/architecture/storage-admin-http-api.md`](docs/architecture/storage-admin-http-api.md) for the current endpoint contract and security limitations.

## Repository Planning Artifacts

- [`requirement.md`](requirement.md): versioned requirement baseline and requirement history.
- [`.plan/`](.plan/): tracked implementation plans split by phase and task checklist.
- [`docs/architecture/`](docs/architecture/): architecture design notes.
- [`docs/adr/`](docs/adr/): architecture decision records.

## Current Documents

- [`docs/architecture/storage-backends.md`](docs/architecture/storage-backends.md)
- [`docs/architecture/storage-admin-http-api.md`](docs/architecture/storage-admin-http-api.md)
- [`docs/adr/ADR-0001-server-managed-multi-backend-media-storage.md`](docs/adr/ADR-0001-server-managed-multi-backend-media-storage.md)
- [`docs/adr/ADR-0002-postgresql-first-database-and-search.md`](docs/adr/ADR-0002-postgresql-first-database-and-search.md)
