-- +goose Up
-- messages_channel_seq_key (unique) supports ordered scans; drop duplicate direction indexes.

DROP INDEX IF EXISTS messages_channel_seq_desc_idx;
DROP INDEX IF EXISTS messages_channel_seq_asc_idx;

-- +goose Down
CREATE INDEX IF NOT EXISTS messages_channel_seq_desc_idx
    ON messages (channel_id, channel_seq DESC);

CREATE INDEX IF NOT EXISTS messages_channel_seq_asc_idx
    ON messages (channel_id, channel_seq ASC);