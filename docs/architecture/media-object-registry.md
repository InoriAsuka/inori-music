# Media Object Registry Architecture

## Goal

The media object registry stores metadata references to binary assets, not the large media bytes themselves. Objects are described by backend ID, object key, content hash, size, MIME type, asset kind, lifecycle state, and latest verification state.

## Registration Rules

A media object must reference an enabled storage backend. Object keys must be relative, clean, slash-delimited keys and must not contain absolute paths, parent traversal, or backslashes. Content hashes use the `algorithm:value` shape.

## Listing and Filters

List endpoints use `limit` and `offset` pagination and require exactly one filter per request. Supported filters are `backendId`, `contentHash`, `verificationStatus`, `lifecycleState`, and `assetKind`. Filtering reads metadata only. Optional sort controls run before pagination and support `backend_object_key`, `created_at`, `updated_at`, `size_bytes`, `object_key`, and `id` with `asc` or `desc` order.

## Verification and Lifecycle

Integrity verification is read-only and currently focuses on filesystem-backed size and `sha256` checks. Lifecycle updates change only `lifecycleState` and `updatedAt`, preserving storage references and latest verification results. Single-object and bulk lifecycle updates are metadata-only; bulk updates require exactly one selection filter and can run in dry-run mode to preview matched objects without persisting changes. `deleted` is terminal metadata and does not delete bytes from storage.

## Statistics and Duplicate Detection

The statistics endpoint calculates object counts, total referenced bytes, backend distribution, asset kind, lifecycle state, and latest verification state from metadata only. It does not open media files or trigger probes. The duplicate report groups objects that share the same content hash, optionally scoped to one backend, so administrators can plan deduplication or cleanup without reading media bytes.
