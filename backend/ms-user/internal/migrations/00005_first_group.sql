-- +goose Up
CREATE TABLE auth.groups (
    id text PRIMARY KEY,
    name text NOT NULL,
    logo_url text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE auth.group_members (
    group_id text NOT NULL REFERENCES auth.groups (id) ON DELETE CASCADE,
    user_id text NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'guest')),
    created_at timestamptz NOT NULL,
    PRIMARY KEY (group_id, user_id)
);

ALTER TABLE auth.instance_settings
    DROP COLUMN name,
    DROP COLUMN logo_url;

-- +goose Down
ALTER TABLE auth.instance_settings
    ADD COLUMN name text NOT NULL DEFAULT '',
    ADD COLUMN logo_url text NOT NULL DEFAULT '';

DROP TABLE IF EXISTS auth.group_members;
DROP TABLE IF EXISTS auth.groups;
