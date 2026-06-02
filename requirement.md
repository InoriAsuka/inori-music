# Requirement

## Project

Inori Music is a full-platform centralized music playback system that supports browser/server and client/server usage models across Web, Android, iOS, and desktop platforms.

## Current Scope

The current 0.x line focuses on architecture bootstrapping, requirements traceability, and the first runnable service foundations. The first phase defines the server-managed media storage architecture before feature implementation begins.

## Storage Requirements

### Media Storage Scope

The system must store and manage media-related binary assets through a stable storage abstraction instead of binding business logic to a single storage vendor or protocol.

Media assets include:

- Original audio files.
- Transcoded audio files.
- Album artwork and artist images.
- Lyrics files.
- Waveform and audio-analysis artifacts.
- Import packages and backup artifacts.

### Supported Backend Families

The server backend must manage media storage configurations and support multiple backend families:

- `local`: local filesystem storage for development, single-node self-hosting, and simple NAS-mounted paths.
- `nfs`: NFS-mounted filesystem storage for LAN and homelab deployments.
- `smb`: SMB/CIFS-mounted filesystem storage for Windows shares, NAS appliances, and mixed-platform networks.
- `s3`: S3-compatible object storage for cloud deployments and object-storage products.
- `distributed`: distributed storage backends such as Ceph, Garage, SeaweedFS, or similar systems exposed through either filesystem mounts, S3-compatible APIs, or dedicated adapters.

### Server-Managed Configuration

The server backend is the source of truth for storage backend configuration. It must provide administrative capabilities to:

- Register storage backends.
- Enable or disable storage backends.
- Validate backend connectivity and permissions.
- Select a default backend for new media assets.
- Track backend health and capacity metadata.
- Prevent accidental deletion of referenced media assets.
- Support migration planning between backends.

### Storage Safety Requirements

The system must not store large audio files directly inside the relational database. The database stores metadata, references, checksums, ownership, storage keys, and lifecycle state.

The storage abstraction must support:

- Object identity and immutable content references.
- Range reads for streaming playback.
- Atomic metadata updates at the application level.
- Content hashing and deduplication planning.
- Backend capability detection.
- Pluggable URL generation, including direct server streaming and presigned URLs where supported.

## Database Requirements

The server-side primary database for 0.x should be PostgreSQL-first. The client-side local database should be SQLite for offline queues, cache indexes, and local search. Multi-primary-database compatibility is not a 0.x delivery requirement, but server code should avoid unnecessary coupling between domain logic and SQL implementation details.

## Search Requirements

The 0.x server-side search should begin with PostgreSQL full-text search, normalized fields, aliases, and ranking rules. External search engines are optional future integrations when scale, language quality, or typo-tolerance requirements exceed PostgreSQL capabilities.

## Requirement History

### v0.1.0 - 2026-06-02

- Established the first-phase architecture scope for server-managed media storage.
- Defined storage backend families: LocalSystem, NFS, SMB, S3-compatible, and distributed storage.
- Required the server backend to own storage configuration, validation, health tracking, and default-backend selection.
- Confirmed PostgreSQL-first server metadata storage and SQLite client-side local storage direction.
- Confirmed PostgreSQL-first search for 0.x with future external search extensibility.
