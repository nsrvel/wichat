-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS status_note_text text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS status_note_emoji text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS status_note_expires_at timestamptz;

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS status_note_expires_at,
    DROP COLUMN IF EXISTS status_note_emoji,
    DROP COLUMN IF EXISTS status_note_text;
