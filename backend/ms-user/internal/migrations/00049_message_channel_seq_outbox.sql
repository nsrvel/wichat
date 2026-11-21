-- +goose Up
-- Owned by ms-chat domain; goose runs from ms-user.

ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS message_seq bigint NOT NULL DEFAULT 0;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS channel_seq bigint;

WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (PARTITION BY channel_id ORDER BY created_at, id) AS seq
    FROM messages
)
UPDATE messages AS m
SET channel_seq = ranked.seq
FROM ranked
WHERE m.id = ranked.id;

UPDATE channels AS c
SET message_seq = COALESCE((
    SELECT MAX(m.channel_seq) FROM messages AS m WHERE m.channel_id = c.id
), 0);

ALTER TABLE messages
    ALTER COLUMN channel_seq SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS messages_channel_seq_key
    ON messages (channel_id, channel_seq);

CREATE INDEX IF NOT EXISTS messages_channel_seq_desc_idx
    ON messages (channel_id, channel_seq DESC);

CREATE TABLE event_outbox (
    id uuid PRIMARY KEY,
    event_id uuid NOT NULL UNIQUE,
    event_type text NOT NULL,
    partition_key text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    publish_attempts int NOT NULL DEFAULT 0,
    last_error text,
    CONSTRAINT event_outbox_partition_key_nonempty CHECK (length(trim(partition_key)) > 0)
);

CREATE INDEX event_outbox_unpublished_idx
    ON event_outbox (created_at)
    WHERE published_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS event_outbox_unpublished_idx;
DROP TABLE IF EXISTS event_outbox;

DROP INDEX IF EXISTS messages_channel_seq_desc_idx;
DROP INDEX IF EXISTS messages_channel_seq_key;

ALTER TABLE messages DROP COLUMN IF EXISTS channel_seq;

ALTER TABLE channels DROP COLUMN IF EXISTS message_seq;