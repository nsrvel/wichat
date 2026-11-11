-- +goose Up
ALTER TABLE instance_settings
    DROP COLUMN IF EXISTS guest_enabled;

-- +goose Down
ALTER TABLE instance_settings
    ADD COLUMN guest_enabled boolean NOT NULL DEFAULT false;

UPDATE instance_settings
SET guest_enabled = false,
    updated_at = now() AT TIME ZONE 'utc'
WHERE id = 1;
