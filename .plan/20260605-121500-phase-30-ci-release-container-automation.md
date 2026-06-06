# Phase 30: CI Release and Container Automation

## Version

v0.30.0

## Requirement

Add repository automation that validates the Go API, publishes release binaries for semantic version tags, and builds/publishes Docker images for the API service.

## Tasks

- [x] Add a Dockerfile for the API server with container-friendly defaults and JSON repository storage under `/data`.
- [x] Add a build workflow for formatting, vetting, unit tests, race tests, OpenAPI JSON validation, and Docker build smoke checks.
- [x] Add a release workflow that builds tagged cross-platform API binaries and publishes GitHub release assets with checksums.
- [x] Add a Docker workflow that publishes multi-architecture API images to GitHub Container Registry.
- [x] Document release and container automation in English and bump the baseline to v0.30.0.

## Notes

The workflows are intentionally scoped to the current Go API service. Future client packages can add separate jobs once Web, Android, iOS, or desktop client scaffolds are introduced.
