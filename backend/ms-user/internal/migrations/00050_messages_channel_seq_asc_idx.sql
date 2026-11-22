-- +goose Up
CREATE INDEX IF NOT EXISTS messages_channel_seq_asc_idx
    ON messages (channel_id, channel_seq ASC);

-- +goose Down
DROP INDEX IF EXISTS messages_channel_seq_asc_idx;