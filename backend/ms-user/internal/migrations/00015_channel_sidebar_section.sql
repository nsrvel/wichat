-- +goose Up
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS sidebar_section text NOT NULL DEFAULT 'channels'
        CHECK (sidebar_section IN ('favorites', 'channels'));

UPDATE channels
SET sidebar_section = 'favorites'
WHERE name = 'announcements';

-- +goose Down
ALTER TABLE channels DROP COLUMN IF EXISTS sidebar_section;
