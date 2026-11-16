-- +goose Up
CREATE TABLE channel_categories (
    id uuid PRIMARY KEY,
    public_id text NOT NULL,
    group_id uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    name text NOT NULL,
    is_private boolean NOT NULL DEFAULT false,
    position integer NOT NULL DEFAULT 0,
    created_by uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT channel_categories_public_id_format CHECK (public_id ~ '^[0-9a-f]{24}$')
);

CREATE UNIQUE INDEX channel_categories_public_id_key ON channel_categories (public_id);
CREATE INDEX channel_categories_group_pos_idx ON channel_categories (group_id, position, created_at);

ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS category_id uuid REFERENCES channel_categories (id) ON DELETE SET NULL;

DROP INDEX IF EXISTS channels_group_general_name_key;

-- +goose StatementBegin
DO $$
DECLARE
    con record;
BEGIN
    FOR con IN
        SELECT c.conname
        FROM pg_constraint c
            JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'channels'
          AND c.contype = 'c'
          AND pg_get_constraintdef(c.oid) LIKE '%type%'
    LOOP
        EXECUTE format('ALTER TABLE channels DROP CONSTRAINT %I', con.conname);
    END LOOP;
END $$;
-- +goose StatementEnd

UPDATE channels
SET type = 'text_only'
WHERE type = 'text_and_voice';

ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (
        type IN ('broadcast', 'text_only', 'voice_only', 'forum', 'dm')
    );

INSERT INTO channel_categories (id, public_id, group_id, name, is_private, position, created_by, created_at, updated_at)
SELECT
    gen_random_uuid(),
    substring(replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''), 1, 24),
    g.id,
    'Text Channels',
    false,
    0,
    COALESCE(
        (SELECT m.user_id FROM group_members AS m WHERE m.group_id = g.id AND m.role = 'owner' ORDER BY m.created_at ASC LIMIT 1),
        (SELECT m.user_id FROM group_members AS m WHERE m.group_id = g.id ORDER BY m.created_at ASC LIMIT 1)
    ),
    g.created_at,
    g.updated_at
FROM groups AS g
WHERE NOT EXISTS (
    SELECT 1 FROM channel_categories AS cat WHERE cat.group_id = g.id AND cat.name = 'Text Channels'
);

INSERT INTO channel_categories (id, public_id, group_id, name, is_private, position, created_by, created_at, updated_at)
SELECT
    gen_random_uuid(),
    substring(replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''), 1, 24),
    g.id,
    'Voice Channels',
    false,
    1,
    COALESCE(
        (SELECT m.user_id FROM group_members AS m WHERE m.group_id = g.id AND m.role = 'owner' ORDER BY m.created_at ASC LIMIT 1),
        (SELECT m.user_id FROM group_members AS m WHERE m.group_id = g.id ORDER BY m.created_at ASC LIMIT 1)
    ),
    g.created_at,
    g.updated_at
FROM groups AS g
WHERE NOT EXISTS (
    SELECT 1 FROM channel_categories AS cat WHERE cat.group_id = g.id AND cat.name = 'Voice Channels'
);

UPDATE channels AS c
SET category_id = cat.id
FROM channel_categories AS cat
WHERE c.group_id = cat.group_id
  AND c.sidebar_section = 'channels'
  AND c.category_id IS NULL
  AND ((c.type = 'voice_only' AND cat.name = 'Voice Channels') OR (c.type <> 'voice_only' AND cat.name = 'Text Channels'));

INSERT INTO channels (id, public_id, group_id, category_id, type, is_private, name, sidebar_section, created_by, created_at, updated_at)
SELECT
    gen_random_uuid(),
    substring(replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''), 1, 24),
    g.id,
    cat.id,
    'voice_only',
    false,
    'General',
    'channels',
    cat.created_by,
    g.created_at,
    g.updated_at
FROM groups AS g
JOIN channel_categories AS cat ON cat.group_id = g.id AND cat.name = 'Voice Channels'
WHERE NOT EXISTS (
    SELECT 1 FROM channels AS c WHERE c.group_id = g.id AND lower(c.name) = 'general' AND c.type = 'voice_only'
);

-- +goose Down
ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_type_check;
ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (
        type IN ('broadcast', 'text_only', 'voice_only', 'text_and_voice', 'forum', 'dm')
    );
ALTER TABLE channels DROP COLUMN IF EXISTS category_id;
DROP TABLE IF EXISTS channel_categories;
