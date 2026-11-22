-- +goose Up
-- Idempotent client sends: one row per (channel, author, client_msg_id).

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS client_msg_id uuid;

CREATE UNIQUE INDEX IF NOT EXISTS messages_channel_author_client_msg_key
    ON messages (channel_id, author_id, client_msg_id)
    WHERE client_msg_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS messages_channel_author_client_msg_key;
ALTER TABLE messages DROP COLUMN IF EXISTS client_msg_id;