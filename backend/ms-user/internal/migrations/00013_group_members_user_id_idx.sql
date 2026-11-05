-- +goose Up
CREATE INDEX group_members_user_id_idx ON group_members (user_id);

CREATE INDEX channels_group_id_idx ON channels (group_id);

-- +goose Down
DROP INDEX IF EXISTS channels_group_id_idx;
DROP INDEX IF EXISTS group_members_user_id_idx;
