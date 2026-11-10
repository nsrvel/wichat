-- +goose Up
UPDATE channels
SET name = 'General'
WHERE lower(name) = 'general';

-- +goose Down
UPDATE channels
SET name = 'general'
WHERE name = 'General';
