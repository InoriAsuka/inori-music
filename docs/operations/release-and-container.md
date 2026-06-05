# Release and Container Automation

## Scope

The repository uses GitHub Actions to validate the Go API, publish tagged release binaries, and publish the API container image to GitHub Container Registry.

## Workflows

- `Build` runs on pull requests and pushes to `main` or `work`. It checks formatting, runs `go vet`, executes unit and race tests, validates the OpenAPI JSON, and performs a Docker build smoke test.
- `Release` runs on semantic version tags such as `v0.30.0`. It builds Linux, macOS, and Windows API server binaries and publishes them to the GitHub release with SHA-256 checksum files.
- `Docker` runs on pushes to `main`, semantic version tags, and manual dispatch. It builds multi-architecture `linux/amd64` and `linux/arm64` images and pushes them to `ghcr.io/<owner>/<repo>/api`.

## Container Defaults

The Docker image listens on `0.0.0.0:8080`, stores bootstrap JSON repositories under `/data`, exposes port `8080`, and requires `INORI_ADMIN_TOKEN` at runtime for admin endpoints. Mount `/data` if the deployment uses the JSON repositories before PostgreSQL persistence is introduced. Release binaries and container images inject non-sensitive version metadata, which is available from the public `/versionz` endpoint.

## Release Process

1. Merge the versioned change to `main`.
2. Push the matching semantic version tag, for example `v0.30.0`.
3. Confirm that `Build`, `Release`, and `Docker` workflows complete successfully.
4. Use the generated GitHub release assets and GHCR image tags for deployment validation.
