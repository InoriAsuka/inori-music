-- +migrate Up
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS lyrics_media_object_id TEXT;

-- +migrate Down
ALTER TABLE tracks DROP COLUMN IF EXISTS lyrics_media_object_id;
