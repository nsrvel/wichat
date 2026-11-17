-- +goose Up
-- Owned by ms-chat domain; goose runs from ms-user until a separate migration runner exists.

CREATE INDEX IF NOT EXISTS messages_channel_created_id_idx
    ON messages (channel_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_message_reactions_msg_created_cover
    ON message_reactions (message_id, created_at ASC, emoji, user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_message_reactions_msg_created_cover;
DROP INDEX IF EXISTS messages_channel_created_id_idx;
