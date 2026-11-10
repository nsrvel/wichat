-- +goose Up
-- Owned by ms-chat domain; goose runs from ms-user until a separate migration runner exists.

CREATE TABLE messages (
    id uuid PRIMARY KEY,
    public_id text NOT NULL UNIQUE,
    channel_id uuid NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    author_id uuid NOT NULL REFERENCES users (id),
    body text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT messages_public_id_format CHECK (public_id ~ '^[0-9a-f]{24}$')
);

CREATE INDEX messages_channel_created_at ON messages (channel_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS messages;
