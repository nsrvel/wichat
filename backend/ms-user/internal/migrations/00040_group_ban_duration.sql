-- +goose Up
ALTER TABLE group_bans
	ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE group_bans
	DROP COLUMN IF EXISTS expires_at;
