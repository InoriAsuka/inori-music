# ADR-0001: Server-Managed Multi-Backend Media Storage

## Status

Accepted

## Date

2026-06-02

## Context

Inori Music needs to support local development, personal self-hosting, NAS environments, cloud deployments, and distributed storage clusters. A single hard-coded object storage dependency would make the system harder to adopt across these environments.

The system must support LocalSystem, NFS, SMB, S3-compatible object storage, and distributed storage backends while keeping playback, catalog, import, and synchronization logic independent from backend-specific details.

## Decision

The server backend will own media storage configuration and expose storage backends as administrative resources. Media assets will be stored through a storage abstraction with backend-specific adapters.

The initial 0.x backend families are:

- `local` for local filesystem storage.
- `nfs` for NFS-mounted filesystem storage.
- `smb` for SMB/CIFS-mounted filesystem storage.
- `s3` for S3-compatible object storage.
- `distributed` for Ceph, Garage, SeaweedFS, or equivalent systems through S3-compatible, mounted filesystem, or future dedicated adapters.

Large audio files must not be stored directly in the relational database.

## Consequences

### Positive

- The project can support local, NAS, cloud, and distributed deployments.
- The default 0.x deployment can remain simple with local filesystem storage.
- Production deployments can use S3-compatible or distributed storage without changing domain logic.
- Storage migrations can be planned around backend IDs and media object metadata.

### Negative

- The server must maintain a capability model because not every backend supports presigned URLs, multipart uploads, lifecycle policies, or native object metadata.
- Administrative validation workflows are required before a backend can become default.
- Mounted filesystem backends require host-level mount reliability and operational documentation.

## Alternatives Considered

### S3-Compatible Only

Rejected for 0.x because it increases adoption friction for local, NAS, and simple self-hosted deployments.

### Local Filesystem Only

Rejected as the long-term architecture because it does not naturally support multi-node production deployments.

### Database BLOB Storage

Rejected for large media because it would bloat the metadata database, make range playback awkward, and complicate backups and CDN integration.
