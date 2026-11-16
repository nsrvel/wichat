-- +goose Up
ALTER TABLE group_members
	ADD COLUMN IF NOT EXISTS timeout_until TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE group_members
	DROP COLUMN IF EXISTS timeout_until;
