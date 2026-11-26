-- +goose Up
ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_sidebar_section_check;

ALTER TABLE channels
    ADD CONSTRAINT channels_sidebar_section_check
        CHECK (sidebar_section IN ('favorites', 'channels', 'chats'));

CREATE TABLE dm_participants (
    channel_id uuid NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (channel_id, user_id)
);

CREATE INDEX dm_participants_user_id_idx ON dm_participants (user_id);

-- +goose Down
DROP TABLE IF EXISTS dm_participants;

ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_sidebar_section_check;

ALTER TABLE channels
    ADD CONSTRAINT channels_sidebar_section_check
        CHECK (sidebar_section IN ('favorites', 'channels'));