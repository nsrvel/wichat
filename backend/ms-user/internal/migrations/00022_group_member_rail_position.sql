-- +goose Up
ALTER TABLE group_members ADD COLUMN rail_position int;

-- +goose Down
ALTER TABLE group_members DROP COLUMN IF EXISTS rail_position;
