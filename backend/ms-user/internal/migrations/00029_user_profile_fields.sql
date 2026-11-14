-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS pronouns text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS bio text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS banner_color text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS banner_color,
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS pronouns;

