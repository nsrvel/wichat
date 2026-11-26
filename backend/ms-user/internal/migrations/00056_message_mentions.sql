-- +goose Up
-- Mention inbox rows (written by ms-chat on message create).

CREATE TABLE message_mentions (
    message_id uuid NOT NULL REFERENCES messages (id) ON DELETE CASCADE,
    group_id uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    channel_id uuid NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    mentioned_user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    author_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (message_id, mentioned_user_id)
);

CREATE INDEX message_mentions_user_group_created_idx
    ON message_mentions (mentioned_user_id, group_id, created_at DESC);

CREATE INDEX message_mentions_group_created_idx
    ON message_mentions (group_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS message_mentions_group_created_idx;
DROP INDEX IF EXISTS message_mentions_user_group_created_idx;
DROP TABLE IF EXISTS message_mentions;