# inori-music

音乐集中平台，目标是构建支持 Web、Android、iOS、PC 等平台的全平台音乐播放系统，同时兼容 B/S 与 C/S 架构。

## Version

Current architecture baseline version: `0.1.0`.

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

## Repository Planning Artifacts

- [`requirement.md`](requirement.md): versioned requirement baseline and requirement history.
- [`.plan/`](.plan/): tracked implementation plans split by phase and task checklist.
- [`docs/architecture/`](docs/architecture/): architecture design notes.
- [`docs/adr/`](docs/adr/): architecture decision records.

## Current Documents

- [`docs/architecture/storage-backends.md`](docs/architecture/storage-backends.md)
- [`docs/adr/ADR-0001-server-managed-multi-backend-media-storage.md`](docs/adr/ADR-0001-server-managed-multi-backend-media-storage.md)
- [`docs/adr/ADR-0002-postgresql-first-database-and-search.md`](docs/adr/ADR-0002-postgresql-first-database-and-search.md)
