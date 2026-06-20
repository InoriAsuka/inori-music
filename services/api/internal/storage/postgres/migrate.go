package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Migrate applies the embedded SQL schema migrations in order.
// It is idempotent: every statement uses IF NOT EXISTS guards.
func Migrate(ctx context.Context, conn *pgx.Conn) error {
	for _, m := range migrations {
		if _, err := conn.Exec(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
	}
	return nil
}

type migration struct {
	name string
	sql  string
}

// migrations lists all schema migrations in application order.
// Each statement must be idempotent (IF NOT EXISTS).
var migrations = []migration{
	{
		name: "001_storage_backends",
		sql: `
CREATE TABLE IF NOT EXISTS storage_backends (
    id                   TEXT        NOT NULL PRIMARY KEY,
    type                 TEXT        NOT NULL,
    display_name         TEXT        NOT NULL,
    enabled              BOOLEAN     NOT NULL DEFAULT TRUE,
    is_default           BOOLEAN     NOT NULL DEFAULT FALSE,
    priority             INTEGER     NOT NULL DEFAULT 0,
    health_status        TEXT        NOT NULL DEFAULT 'unknown',
    last_health_check_at TIMESTAMPTZ,
    last_capacity        JSONB,
    capabilities         JSONB       NOT NULL DEFAULT '{}',
    config               JSONB       NOT NULL DEFAULT '{}',
    created_at           TIMESTAMPTZ NOT NULL,
    updated_at           TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS storage_backends_priority_id_idx
    ON storage_backends (priority, id);`,
	},
	{
		name: "002_media_objects",
		sql: `
CREATE TABLE IF NOT EXISTS media_objects (
    id                 TEXT        NOT NULL PRIMARY KEY,
    backend_id         TEXT        NOT NULL,
    object_key         TEXT        NOT NULL,
    content_hash       TEXT        NOT NULL,
    size_bytes         BIGINT      NOT NULL DEFAULT 0,
    mime_type          TEXT        NOT NULL,
    asset_kind         TEXT        NOT NULL,
    lifecycle_state    TEXT        NOT NULL,
    last_verification  JSONB,
    created_at         TIMESTAMPTZ NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS media_objects_backend_id_idx ON media_objects (backend_id);
CREATE INDEX IF NOT EXISTS media_objects_content_hash_idx ON media_objects (content_hash);
CREATE INDEX IF NOT EXISTS media_objects_lifecycle_state_idx ON media_objects (lifecycle_state);
CREATE INDEX IF NOT EXISTS media_objects_asset_kind_idx ON media_objects (asset_kind);`,
	},
	{
		name: "003_users",
		sql: `
CREATE TABLE IF NOT EXISTS users (
    id            TEXT        NOT NULL PRIMARY KEY,
    username      TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'viewer',
    enabled       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS users_username_idx ON users (lower(username));`,
	},
	{
		name: "004_sessions",
		sql: `
CREATE TABLE IF NOT EXISTS sessions (
    token_hash TEXT        NOT NULL PRIMARY KEY,
    user_id    TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS sessions_user_id_idx  ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at);`,
	},
	{
		name: "005_catalog",
		sql: `
CREATE TABLE IF NOT EXISTS artists (
    id         TEXT        NOT NULL PRIMARY KEY,
    name       TEXT        NOT NULL,
    sort_name  TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS artists_sort_name_idx ON artists (lower(sort_name), lower(name), id);

CREATE TABLE IF NOT EXISTS albums (
    id           TEXT        NOT NULL PRIMARY KEY,
    title        TEXT        NOT NULL,
    sort_title   TEXT        NOT NULL DEFAULT '',
    artist_id    TEXT        NOT NULL REFERENCES artists(id) ON DELETE RESTRICT,
    release_year INTEGER     NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS albums_artist_id_idx ON albums (artist_id);
CREATE INDEX IF NOT EXISTS albums_sort_title_idx ON albums (lower(sort_title), lower(title), id);

CREATE TABLE IF NOT EXISTS tracks (
    id              TEXT        NOT NULL PRIMARY KEY,
    title           TEXT        NOT NULL,
    sort_title      TEXT        NOT NULL DEFAULT '',
    artist_id       TEXT        NOT NULL REFERENCES artists(id) ON DELETE RESTRICT,
    album_id        TEXT        REFERENCES albums(id) ON DELETE SET NULL,
    media_object_id TEXT        NOT NULL REFERENCES media_objects(id) ON DELETE RESTRICT,
    track_number    INTEGER     NOT NULL DEFAULT 0,
    disc_number     INTEGER     NOT NULL DEFAULT 0,
    duration_ms     INTEGER     NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS tracks_artist_id_idx ON tracks (artist_id);
CREATE INDEX IF NOT EXISTS tracks_album_id_idx ON tracks (album_id);
CREATE INDEX IF NOT EXISTS tracks_media_object_id_idx ON tracks (media_object_id);
CREATE INDEX IF NOT EXISTS tracks_sort_title_idx ON tracks (lower(sort_title), lower(title), id);`,
	},
	{
		name: "006_catalog_fts",
		sql: `
-- Add tsvector columns for full-text search on catalog entities.
ALTER TABLE artists
    ADD COLUMN IF NOT EXISTS search_vector tsvector
        GENERATED ALWAYS AS (
            setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(sort_name, '')), 'B')
        ) STORED;

ALTER TABLE albums
    ADD COLUMN IF NOT EXISTS search_vector tsvector
        GENERATED ALWAYS AS (
            setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(sort_title, '')), 'B')
        ) STORED;

ALTER TABLE tracks
    ADD COLUMN IF NOT EXISTS search_vector tsvector
        GENERATED ALWAYS AS (
            setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(sort_title, '')), 'B')
        ) STORED;

CREATE INDEX IF NOT EXISTS artists_search_vector_idx ON artists USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS albums_search_vector_idx  ON albums  USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS tracks_search_vector_idx  ON tracks  USING GIN (search_vector);`,
	},
	{
		name: "007_playlists",
		sql: `
CREATE TABLE IF NOT EXISTS playlists (
    id          TEXT        NOT NULL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS playlists_name_idx ON playlists (lower(name));

-- Ordered mapping of playlists to tracks.
-- position is a zero-based integer that controls playback order.
-- ON DELETE CASCADE removes entries when a playlist or track is deleted.
CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id TEXT    NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id    TEXT    NOT NULL REFERENCES tracks(id)    ON DELETE CASCADE,
    position    INTEGER NOT NULL,
    PRIMARY KEY (playlist_id, position)
);
CREATE INDEX IF NOT EXISTS playlist_tracks_playlist_id_idx ON playlist_tracks (playlist_id);`,
	},
	{
		name: "008_play_events",
		sql: `
CREATE TABLE IF NOT EXISTS play_events (
    id         TEXT        NOT NULL PRIMARY KEY,
    user_id    TEXT        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    track_id   TEXT        NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    played_at  TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS play_events_user_id_played_at_idx ON play_events (user_id, played_at DESC);
CREATE INDEX IF NOT EXISTS play_events_track_id_idx           ON play_events (track_id);`,
	},
	{
		name: "009_track_genre",
		sql: `
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS genre TEXT;
CREATE INDEX IF NOT EXISTS tracks_genre_idx ON tracks (lower(genre)) WHERE genre IS NOT NULL;`,
	},
}
