-- +goose Up
UPDATE group_roles SET name = 'everyone' WHERE slug = 'everyone' AND name = '@everyone';

-- +goose Down
UPDATE group_roles SET name = '@everyone' WHERE slug = 'everyone' AND name = 'everyone';
