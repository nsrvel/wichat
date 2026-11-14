-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS disabled_at timestamptz NULL,
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

CREATE INDEX IF NOT EXISTS idx_users_active_status
    ON users (id)
    WHERE disabled_at IS NULL AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_active_status;
ALTER TABLE users
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS disabled_at;
