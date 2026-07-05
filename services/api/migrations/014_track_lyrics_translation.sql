-- +migrate Up
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS lyrics_translation_media_object_id TEXT;
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS lyrics_source TEXT;

-- +migrate Down
ALTER TABLE tracks DROP COLUMN IF EXISTS lyrics_translation_media_object_id;
ALTER TABLE tracks DROP COLUMN IF EXISTS lyrics_source;
