-- 002_media_objects.sql
-- Media object metadata reference table.

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

CREATE INDEX IF NOT EXISTS media_objects_backend_id_idx
    ON media_objects (backend_id);

CREATE INDEX IF NOT EXISTS media_objects_content_hash_idx
    ON media_objects (content_hash);

CREATE INDEX IF NOT EXISTS media_objects_lifecycle_state_idx
    ON media_objects (lifecycle_state);

CREATE INDEX IF NOT EXISTS media_objects_asset_kind_idx
    ON media_objects (asset_kind);
