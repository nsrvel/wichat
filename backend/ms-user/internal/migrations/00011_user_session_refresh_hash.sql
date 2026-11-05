-- +goose Up
ALTER TABLE user_sessions ADD COLUMN refresh_token_hash bytea;

-- +goose Down
ALTER TABLE user_sessions DROP COLUMN IF EXISTS refresh_token_hash;
