-- +goose Up
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS image_url text NOT NULL DEFAULT '';

ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_type_check;

ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (type IN ('public', 'private', 'dm', 'broadcast'));

UPDATE channels
SET
    type = 'broadcast',
    name = 'Announcements',
    image_url = ''
WHERE lower(name) = 'announcements';

-- +goose Down
UPDATE channels
SET type = 'public'
WHERE type = 'broadcast';

ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_type_check;

ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (type IN ('public', 'private', 'dm'));

ALTER TABLE channels DROP COLUMN IF EXISTS image_url;
