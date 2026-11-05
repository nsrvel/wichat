-- +goose Up
CREATE TABLE groups (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    logo_url text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE group_members (
    group_id uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'guest')),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (group_id, user_id)
);

ALTER TABLE instance_settings
    DROP COLUMN name,
    DROP COLUMN logo_url;

-- +goose Down
ALTER TABLE instance_settings
    ADD COLUMN name text NOT NULL DEFAULT '',
    ADD COLUMN logo_url text NOT NULL DEFAULT '';

DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
