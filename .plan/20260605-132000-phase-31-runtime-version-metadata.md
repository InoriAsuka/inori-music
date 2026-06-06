# Phase 31: Runtime Version Metadata

## Version

v0.31.0

## Requirement

Expose non-sensitive runtime build metadata so operators can verify deployed binaries and containers against release tags, commits, and build timestamps.

## Tasks

- [x] Add a public `/versionz` endpoint that returns API service name, version, commit, and build time.
- [x] Wire build metadata from the server command into the HTTP handler while keeping safe development defaults.
- [x] Update release binary and Docker image workflows to inject version metadata during builds.
- [x] Extend the OpenAPI contract and tests for the public version endpoint and schema.
- [x] Update the versioned requirements, README, and phase plan for v0.31.0.

## Notes

The endpoint intentionally exposes only non-secret build metadata. It remains public alongside `/healthz` so deployment automation can validate running artifacts without administrator credentials.
