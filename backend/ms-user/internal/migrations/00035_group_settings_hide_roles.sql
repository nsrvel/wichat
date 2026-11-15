-- +goose Up
ALTER TABLE group_settings
    ADD COLUMN hide_roles boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE group_settings
    DROP COLUMN IF EXISTS hide_roles;
