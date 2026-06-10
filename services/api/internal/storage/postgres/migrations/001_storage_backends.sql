-- 001_storage_backends.sql
-- Storage backend configuration table.

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
    ON storage_backends (priority, id);
