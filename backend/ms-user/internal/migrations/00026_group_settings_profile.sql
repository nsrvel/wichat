-- +goose Up
ALTER TABLE group_settings
    ADD COLUMN description text NOT NULL DEFAULT '',
    ADD COLUMN is_private boolean NOT NULL DEFAULT false,
    ADD COLUMN banner_color text NOT NULL DEFAULT '',
    ADD COLUMN banner_url text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE group_settings
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS is_private,
    DROP COLUMN IF EXISTS banner_color,
    DROP COLUMN IF EXISTS banner_url;
