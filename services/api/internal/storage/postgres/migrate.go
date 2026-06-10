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
}
