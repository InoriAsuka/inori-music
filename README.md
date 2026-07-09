# inori-music

A self-hosted centralized music platform with cross-platform clients (Web, Android, iOS, desktop) supporting both browser/server and client/server architectures.

## Current Version

**4.8.0** — see [requirement.md](requirement.md) for the authoritative version history and phase log.

## Architecture

**services/api** — Go 1.23 modular monolith. PostgreSQL (pgx/v5) for metadata, Meilisearch for full-text search, pluggable storage backends (local filesystem, S3-compatible). Exposes 146 OpenAPI operations across catalog, auth, favorites, history, playlists, storage management, admin.

**services/web** — Next.js 15 / React 19 / Tailwind 4. Viewer-facing web player (port 3000). Consumes all 109 user-facing API endpoints.

**services/admin** — Next.js 15 / React 19 / Tailwind 4. Admin console (port 3001) for user/catalog/storage management.

**services/mobile** — Flutter 3.x / Dart ≥3.12.2. Cross-platform mobile client (Android/iOS) using Riverpod state management and just_audio playback engine.

**Search:** Meilisearch via `MEILI_HOST`/`MEILI_SEARCH_KEY` env vars (fallback to no-op if not configured).

**Storage:** Local filesystem, NFS, SMB, S3-compatible object stores. Media bytes stored external to Postgres; only metadata/keys/hashes in DB.

## Quick Start

### API Server

```bash
cd services/api
go run ./cmd/server/main.go
```

Required env vars:
- `INORI_DATABASE_URL` — Postgres connection string
- `INORI_ADMIN_TOKEN` — Bootstrap admin token (optional but recommended)
- `INORI_INITIAL_ADMIN_USER` / `INORI_INITIAL_ADMIN_PASSWORD` — Seed admin account
- `MEILI_HOST` / `MEILI_SEARCH_KEY` — Meilisearch connection (optional)
- `INORI_STORAGE_REPOSITORY_FILE` / `INORI_MEDIA_OBJECT_REPOSITORY_FILE` — Storage config JSON paths
- `INORI_CORS_ORIGINS` — Comma-separated CORS origins (defaults to permissive mode)

Defaults to `:8080` (override via `INORI_HTTP_ADDR`).

### Web Player

```bash
cd services/web
npm install
npm run generate:api  # Generate TypeScript client from OpenAPI spec
npm run dev           # Starts on port 3000
```

### Admin Console

```bash
cd services/admin
npm install
npm run generate:api
npm run dev           # Starts on port 3001
```

### Mobile App

```bash
cd services/mobile
flutter pub get
flutter run
```

### Docker Compose (Production)

```bash
docker compose -f docker-compose.prod.yml up
```

Stands up: API server, Postgres, Meilisearch, and optionally pre-built web/admin containers. Check [docker-compose.prod.yml](docker-compose.prod.yml) for service definitions and required env var overrides.

## Version History

All phases and release notes are maintained in [requirement.md](requirement.md). This is the single authoritative source for version history — the README does not duplicate the phase log.

## Documentation Policy

Repository Markdown documentation is maintained in English. Historical phase plans, requirements, ADRs, and architecture documents are kept in English so implementation records remain consistent across releases.

## Future Outlook

**v5.x — Productization / External Readiness** (planned through v5.4.0): Security baseline (HMAC-signed streaming URLs, login rate limiting, Meilisearch key enforcement, nginx security headers), Web/Mobile feature parity (lyrics panel, search highlighting, audio engine parity with ReplayGain/gapless, user playlists, playback speed, sleep timer), and cross-device resume sync (`/me/player-state` + `/me/search-history` endpoints). v5.4.0 closes this phase.

**v6.x — Media Ingestion & Catalog Management** (directional, not yet phase-planned): Making "putting music into the system" a first-class product capability. Current gaps: only lyrics have an upload endpoint; `storage.Service.PutObject` exists but no audio upload endpoint or library scanner; `transcoded_audio` asset kind enumerated but no producer.

**v7.x — Intelligent Experience & Data** (directional): Recommendation, smart playlists, listening analytics.

**v8.x — Sharing & Multi-User** (directional): Shared playlists, social features, multi-tenant support.

v6/v7/v8 are direction-level only and will be broken into concrete phases when v5 work completes.
