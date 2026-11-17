-- +goose Up
-- Owned by ms-chat domain; goose runs from ms-user until a separate migration runner exists.

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS reply_to_id uuid REFERENCES messages (id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS edited_at timestamptz,
    ADD COLUMN IF NOT EXISTS is_pinned boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS pinned_at timestamptz,
    ADD COLUMN IF NOT EXISTS pinned_by uuid REFERENCES users (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS messages_channel_pinned_idx
    ON messages (channel_id, is_pinned, pinned_at DESC)
    WHERE is_pinned = true;

CREATE TABLE IF NOT EXISTS message_reactions (
    id         uuid        PRIMARY KEY,
    message_id uuid        NOT NULL REFERENCES messages (id) ON DELETE CASCADE,
    user_id    uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    emoji      text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT message_reactions_unique UNIQUE (message_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_message_reactions_message_id
    ON message_reactions (message_id);

-- +goose Down
DROP TABLE IF EXISTS message_reactions;
DROP INDEX IF EXISTS messages_channel_pinned_idx;
ALTER TABLE messages
    DROP COLUMN IF EXISTS pinned_by,
    DROP COLUMN IF EXISTS pinned_at,
    DROP COLUMN IF EXISTS is_pinned,
    DROP COLUMN IF EXISTS edited_at,
    DROP COLUMN IF EXISTS reply_to_id;
