-- +goose Up
ALTER TABLE group_members
    ADD COLUMN IF NOT EXISTS join_method text NOT NULL DEFAULT 'unknown'
        CHECK (join_method IN ('owner', 'open', 'invite', 'invite_link', 'unknown')),
    ADD COLUMN IF NOT EXISTS invite_code text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS invited_by_user_id uuid REFERENCES users (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE group_members
    DROP COLUMN IF EXISTS invited_by_user_id,
    DROP COLUMN IF EXISTS invite_code,
    DROP COLUMN IF EXISTS join_method;
