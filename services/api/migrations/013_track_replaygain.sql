-- +migrate Up
ALTER TABLE tracks ADD COLUMN IF NOT EXISTS replay_gain_db REAL;

-- +migrate Down
ALTER TABLE tracks DROP COLUMN IF EXISTS replay_gain_db;
