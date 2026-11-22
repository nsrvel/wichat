-- +goose Up

CREATE INDEX IF NOT EXISTS messages_channel_author_created_idx
    ON messages (channel_id, author_id, created_at DESC);

CREATE INDEX IF NOT EXISTS messages_reply_to_id_idx
    ON messages (reply_to_id)
    WHERE reply_to_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS user_sessions_refresh_hash_active_idx
    ON user_sessions (refresh_token_hash)
    WHERE revoked_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS user_sessions_refresh_hash_active_idx;
DROP INDEX IF EXISTS messages_reply_to_id_idx;
DROP INDEX IF EXISTS messages_channel_author_created_idx;