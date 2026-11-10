-- +goose Up
CREATE UNIQUE INDEX channels_group_announcements_name_key ON channels (group_id, name)
    WHERE name = 'announcements';

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
    'announcements',
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
    SELECT 1 FROM channels AS c WHERE c.group_id = g.id AND c.name = 'announcements'
);

-- +goose Down
DROP INDEX IF EXISTS channels_group_announcements_name_key;

DELETE FROM channels WHERE name = 'announcements';
