-- +goose Up
UPDATE channels
SET name = 'General'
WHERE lower(name) = 'general' AND name <> 'General';

UPDATE channels
SET
    type = 'broadcast',
    name = 'Announcements'
WHERE lower(name) = 'announcements';

-- +goose Down
-- No-op.
