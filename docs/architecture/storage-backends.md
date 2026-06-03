# Media Storage Backend Architecture

## Purpose

Inori Music must support multiple media storage backends without coupling playback, import, catalog, or administrative logic to one storage product. The server backend owns storage configuration and exposes a stable abstraction to the rest of the system.

## Design Principles

- Backend-agnostic media references: business records store storage backend IDs and object keys, not absolute local paths or vendor-specific URLs.
- Server-managed configuration: administrators manage storage backends through server-side APIs and administrative UI flows.
- Safe defaults: local filesystem storage is the simplest 0.x default, while production deployments can use S3-compatible or distributed storage.
- Capability-aware behavior: the server detects what each backend can do instead of assuming all backends support every object-storage feature.
- Streaming readiness: every backend must support server-mediated range reads or a safe equivalent for playback.
- Migration readiness: media objects track source backend, content hash, size, lifecycle state, and migration status.

## Backend Families

### LocalSystem

`local` storage writes media assets to a server-local path. It is the default backend for development, single-node deployments, and simple self-hosted installations.

Typical configuration:

```yaml
storage:
  backends:
    - id: local-main
      type: local
      displayName: Local media library
      enabled: true
      default: true
      local:
        rootPath: ./data/media
```

### NFS

`nfs` storage represents an NFS-mounted filesystem path. The server treats it as a mounted filesystem backend but records it separately so administrators can understand operational ownership, mount dependencies, and health checks.

Typical configuration:

```yaml
storage:
  backends:
    - id: nfs-library
      type: nfs
      displayName: Studio NFS library
      enabled: true
      nfs:
        mountPath: /mnt/inori-media
        expectedRemote: 10.0.0.20:/exports/inori-media
```

### SMB

`smb` storage represents an SMB/CIFS-mounted filesystem path. It supports NAS appliances, Windows shares, and mixed-platform local networks. Credentials should be managed outside plain repository files through secrets or host-level mount configuration.

Typical configuration:

```yaml
storage:
  backends:
    - id: smb-nas
      type: smb
      displayName: Family NAS music share
      enabled: true
      smb:
        mountPath: /mnt/smb/inori-media
        expectedShare: //nas/music
```

### S3-Compatible Object Storage

`s3` storage supports cloud object storage and S3-compatible products. The application should rely only on a conservative S3 capability baseline unless a backend advertises additional features.

Typical configuration:

```yaml
storage:
  backends:
    - id: s3-prod
      type: s3
      displayName: Production object storage
      enabled: true
      s3:
        endpoint: https://s3.example.com
        region: auto
        bucket: inori-media
        pathStyle: true
        accessKeySecretRef: INORI_S3_ACCESS_KEY
        secretKeySecretRef: INORI_S3_SECRET_KEY
```

### Distributed Storage

`distributed` storage covers systems such as Ceph, Garage, SeaweedFS, or equivalent distributed backends. Distributed systems can be connected through S3-compatible APIs, filesystem mounts, or future dedicated adapters.

Typical configuration:

```yaml
storage:
  backends:
    - id: ceph-rgw
      type: distributed
      displayName: Ceph RGW cluster
      enabled: true
      distributed:
        adapter: s3-compatible
        endpoint: https://rgw.example.com
        bucket: inori-media
```

## Capability Model

Each backend exposes capabilities so the service can choose safe behavior.

| Capability | LocalSystem | NFS | SMB | S3-compatible | Distributed |
|---|---:|---:|---:|---:|---:|
| Server range reads | Yes | Yes | Yes | Yes | Depends on adapter |
| Presigned URLs | No | No | No | Usually | Depends on adapter |
| Multipart upload | No | No | No | Usually | Depends on adapter |
| Native lifecycle policy | No | No | No | Sometimes | Depends on adapter |
| Cross-node access | No | Yes | Yes | Yes | Yes |
| Requires mount validation | No | Yes | Yes | No | Depends on adapter |
| Requires credential validation | No | Sometimes | Sometimes | Yes | Depends on adapter |

## Server Administrative Model

The server must manage storage backends as first-class administrative resources.

Recommended resource fields:

```text
StorageBackend
├── id
├── type
├── display_name
├── enabled
├── is_default
├── priority
├── health_status
├── last_health_check_at
├── capabilities
├── encrypted_config
├── created_at
└── updated_at
```

Recommended operations:

- `CreateStorageBackend`
- `UpdateStorageBackend`
- `DisableStorageBackend`
- `ValidateStorageBackend`
- `SetDefaultStorageBackend`
- `ListStorageBackends`
- `GetStorageBackendHealth`
- `PlanStorageMigration`

## Media Object Model

Media records should reference storage objects rather than embedding files in the metadata database.

Recommended fields:

```text
MediaObject
├── id
├── backend_id
├── object_key
├── content_hash
├── size_bytes
├── mime_type
├── asset_kind
├── lifecycle_state
├── created_at
└── updated_at
```

## Validation Workflow

When an administrator creates or updates a storage backend, the server should:

1. Validate required fields for the backend type.
2. Validate secrets are available without logging secret values.
3. Check path existence or endpoint reachability.
4. Verify read/write permissions using a short-lived probe object or probe file.
5. Verify range-read behavior where applicable.
6. Record backend capabilities and health state.
7. Reject making a backend default until validation succeeds.

## Security Requirements

- Never store plaintext credentials in repository files.
- Store secrets through environment variables, secret managers, or encrypted server configuration.
- Avoid exposing local filesystem paths to clients.
- Prefer server-mediated streaming for filesystem backends.
- Use presigned URLs only when the backend supports them and policy allows direct object access.
- Log object IDs and backend IDs, not secret values or full credential-bearing URLs.

## 0.x Implementation Order

1. LocalSystem adapter.
2. Mounted filesystem support for NFS and SMB using validated mount paths.
3. S3-compatible adapter and safe short-lived object probes.
4. Health check and capability tracking for LocalSystem, NFS, SMB, S3-compatible, and distributed mounted-filesystem or S3-compatible adapters.
5. Distributed storage adapters through S3-compatible or mounted filesystem strategies.
6. Dedicated distributed-storage adapters only when a real deployment requires them.

## Filesystem Probe Safety

LocalSystem, NFS, SMB, and `distributed` backends using the `mounted-filesystem` adapter can be verified through a real filesystem probe. The probe intentionally performs only these operations inside the configured root:

1. Create one application-owned `.inori-music-probe-*` temporary file.
2. Write and sync a fixed probe payload.
3. Perform full-read and range-read verification.
4. Close and remove the same probe file.

The probe never scans, modifies, or deletes unrelated media files. NFS and SMB mounting remain host-level operational responsibilities. S3-compatible object probes are implemented separately with short-lived `.inori-music-probe/*` object keys. Dedicated distributed probes beyond mounted-filesystem and S3-compatible adapters require later adapters.

## S3-Compatible Probe Safety

S3-compatible backends can be verified through a conservative object probe. The probe resolves credentials from `accessKeySecretRef` and `secretKeySecretRef` environment variable names, then performs only these operations inside a server-owned `.inori-music-probe/` prefix:

1. Put one short-lived probe object.
2. Perform full-read verification.
3. Perform range-read verification.
4. Delete the same probe object with best-effort cleanup if a later step fails.

The S3-compatible probe validates basic object API behavior only. It does not validate provider-specific lifecycle policies, versioning, object lock, event notifications, or bucket-level administration.

## Refresh Scheduling and Capacity

Administrators can refresh all enabled backends on demand. A refresh run isolates backend failures, records supported health results, skips disabled backends, and continues processing the remaining backends. The server can also run the same refresh periodically when `INORI_STORAGE_REFRESH_INTERVAL` contains a positive Go duration such as `15m`.

Filesystem-backed storage reports total, available, and used bytes from mounted-path filesystem statistics. S3-compatible services do not provide one portable bucket-capacity API, so S3-compatible capacity intentionally remains unsupported until provider-specific quota integrations are designed.
