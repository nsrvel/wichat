-- +goose Up
-- ms-chat owns message_mentions; applied via shared ms-user goose.

ALTER TABLE message_mentions
    ADD COLUMN IF NOT EXISTS read_at timestamptz;

CREATE INDEX message_mentions_unread_user_group_created_idx
    ON message_mentions (mentioned_user_id, group_id, created_at DESC)
    WHERE read_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS message_mentions_unread_user_group_created_idx;

ALTER TABLE message_mentions
    DROP COLUMN IF EXISTS read_at;