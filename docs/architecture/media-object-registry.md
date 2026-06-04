# Media Object Registry

## Scope

Phase 10 introduces a domain scaffold for media object metadata. A media object is a server-owned reference to a binary asset stored in one configured storage backend.

The registry stores metadata such as backend ID, object key, content hash, byte size, MIME type, asset kind, lifecycle state, and timestamps. It does not store audio, artwork, lyrics, waveform, or backup bytes inside the API service or relational database.

## Validation Rules

A media object must reference an existing enabled storage backend. Object keys must be relative storage keys and must not be absolute paths, empty paths, current-directory aliases, parent-directory traversal, or backslash-delimited Windows paths.

Content hashes use an `algorithm:value` shape so future import and deduplication workflows can distinguish algorithms such as `sha256` or `blake3`. Sizes must be non-negative, MIME types must be present, and both asset kind and lifecycle state must be from the supported domain constants.

## Supported Asset Kinds

- `original_audio`
- `transcoded_audio`
- `artwork`
- `lyrics`
- `waveform`
- `analysis`
- `import_package`
- `backup`

## Lifecycle States

- `staged`: registered during an import or verification workflow.
- `active`: usable by library, playback, and client synchronization flows.
- `archived`: retained but not preferred for normal playback or display.
- `deleted`: metadata tombstone for future safe-delete and audit workflows.

## HTTP Administration API

Phase 11 exposes authenticated administrator endpoints for metadata-only workflows:

- `POST /api/v1/admin/media/objects` registers a media object reference for an enabled backend.
- `GET /api/v1/admin/media/objects/{id}` fetches one media object reference.
- `GET /api/v1/admin/media/objects?backendId=...` lists references by backend.
- `GET /api/v1/admin/media/objects?contentHash=...` lists references by content hash for future deduplication workflows.

`POST /api/v1/admin/media/objects/{id}/verify` performs read-only integrity verification for one filesystem-backed object by checking file existence, regular-file shape, byte size, and `sha256` content hash. `POST /api/v1/admin/media/objects/verify?backendId=...` and `?contentHash=...` run the same checks for filtered object groups and continue after individual object failures. Each verification updates the media object's `lastVerification` metadata with the latest status, timestamp, size, content hash, and failure message when present.

These endpoints still do not upload, stream, delete, move, or repair media bytes.

## Future Direction

The first implementation uses an in-memory repository for domain tests. Phase 12 adds `INORI_MEDIA_OBJECT_REPOSITORY_FILE` for optional single-node JSON persistence of media object metadata before database migrations exist. PostgreSQL should later own media object metadata, with indexes for backend ID, object key, content hash, asset kind, lifecycle state, and ownership/library relationships.

## Verification Status Listing

Media-object list requests can filter by `verificationStatus=verified|failed|unknown`. The filter reads only persisted `lastVerification` metadata: `verified` and `failed` match the latest recorded result, while `unknown` returns objects that have not been verified yet. The list endpoint still accepts exactly one filter per request to avoid ambiguous admin workflows.
