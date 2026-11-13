-- +goose Up
CREATE TABLE group_roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    public_id text NOT NULL,
    group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    slug text NOT NULL,
    name text NOT NULL,
    color text NOT NULL DEFAULT '#99aab5',
    hoist boolean NOT NULL DEFAULT false,
    mentionable boolean NOT NULL DEFAULT false,
    is_default boolean NOT NULL DEFAULT false,
    is_system boolean NOT NULL DEFAULT false,
    position integer NOT NULL DEFAULT 0,
    permissions text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT group_roles_public_id_format CHECK (public_id ~ '^[0-9a-f]{24}$'),
    CONSTRAINT group_roles_public_id_unique UNIQUE (public_id),
    CONSTRAINT group_roles_group_slug_unique UNIQUE (group_id, slug)
);

CREATE INDEX group_roles_group_position_idx ON group_roles (group_id, position);

CREATE TABLE group_member_roles (
    group_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role_id uuid NOT NULL REFERENCES group_roles(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id, role_id),
    FOREIGN KEY (group_id, user_id) REFERENCES group_members(group_id, user_id) ON DELETE CASCADE
);

CREATE INDEX group_member_roles_group_user_idx ON group_member_roles (group_id, user_id);
CREATE INDEX group_member_roles_role_id_idx ON group_member_roles (role_id);

CREATE TABLE group_bans (
    group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    banned_by uuid REFERENCES users(id) ON DELETE SET NULL,
    reason text NOT NULL DEFAULT '',
    delete_message_seconds integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

-- Backfill @everyone and Member roles for existing groups
INSERT INTO group_roles (
    id, public_id, group_id, slug, name, color, hoist, mentionable,
    is_default, is_system, position, permissions, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    substring(replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''), 1, 24),
    g.id,
    'everyone',
    '@everyone',
    '#99aab5',
    false,
    false,
    true,
    true,
    0,
    ARRAY['view_channels', 'send_messages', 'read_message_history', 'create_invite', 'attach_files', 'add_reactions', 'connect', 'speak', 'use_voice_activity'],
    g.created_at,
    g.created_at
FROM groups g;

INSERT INTO group_roles (
    id, public_id, group_id, slug, name, color, hoist, mentionable,
    is_default, is_system, position, permissions, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    substring(replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''), 1, 24),
    g.id,
    'member',
    'Member',
    '#5865f2',
    false,
    false,
    false,
    true,
    1,
    '{}'::text[],
    g.created_at,
    g.created_at
FROM groups g;

-- Assign all existing members to the default Member role
INSERT INTO group_member_roles (group_id, user_id, role_id, created_at)
SELECT
    gm.group_id,
    gm.user_id,
    gr.id,
    gm.created_at
FROM group_members gm
JOIN group_roles gr ON gr.group_id = gm.group_id AND gr.slug = 'member'
ON CONFLICT (group_id, user_id, role_id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS group_bans;
DROP TABLE IF EXISTS group_member_roles;
DROP TABLE IF EXISTS group_roles;

