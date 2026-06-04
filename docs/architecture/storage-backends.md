# Storage Backend Architecture

## Goal

The server manages media storage configuration and hides differences between local filesystems, NFS, SMB, S3-compatible object storage, and distributed storage systems.

## Backend Families

- `local`: development, single-node self-hosting, or already-mounted NAS paths.
- `nfs`: NFS-mounted storage for LAN and homelab deployments.
- `smb`: SMB/CIFS-mounted storage for Windows shares, NAS devices, and mixed-platform networks.
- `s3`: cloud or self-hosted S3-compatible object storage.
- `distributed`: systems such as Ceph, Garage, or SeaweedFS exposed through S3-compatible APIs, mounted filesystems, or future dedicated adapters.

## Capability Model

Backends expose capabilities such as server range reads, presigned URLs, multipart upload, native lifecycle policy support, cross-node access, mount validation, and credential validation. Static validation infers capabilities from configuration, while probes update runtime health state.

## Safety Principles

Probes may only create server-owned temporary files or objects and must clean them up. The application does not mount or unmount NFS/SMB shares. Credentials are resolved through environment variable references and must not be written into repository files.
