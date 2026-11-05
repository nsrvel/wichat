-- +goose Up
CREATE TABLE channels (
    id uuid PRIMARY KEY,
    public_id text NOT NULL,
    group_id uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    type text NOT NULL CHECK (type IN ('public', 'private', 'dm')),
    name text NOT NULL,
    created_by uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT channels_public_id_format CHECK (public_id ~ '^[0-9a-f]{24}$')
);

CREATE UNIQUE INDEX channels_public_id_key ON channels (public_id);
CREATE UNIQUE INDEX channels_group_general_name_key ON channels (group_id, name)
    WHERE name = 'general';

INSERT INTO channels (id, public_id, group_id, type, name, created_by, created_at, updated_at)
SELECT
    gen_random_uuid(),
    substring(
        replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''),
        1,
        24
    ),
    g.id,
    'public',
    'general',
    COALESCE(
        (
            SELECT m.user_id
            FROM group_members AS m
            WHERE m.group_id = g.id AND m.role = 'owner'
            ORDER BY m.created_at ASC
            LIMIT 1
        ),
        (
            SELECT m.user_id
            FROM group_members AS m
            WHERE m.group_id = g.id
            ORDER BY m.created_at ASC
            LIMIT 1
        )
    ),
    g.created_at,
    g.updated_at
FROM groups AS g
WHERE NOT EXISTS (
    SELECT 1 FROM channels AS c WHERE c.group_id = g.id AND c.name = 'general'
);

-- +goose Down
DROP TABLE IF EXISTS channels;
