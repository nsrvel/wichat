-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS locale text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS time_format text NOT NULL DEFAULT 'auto';

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS time_format,
    DROP COLUMN IF EXISTS locale;
