-- +goose Up
-- Convert legacy text entity ids to uuid (no-op when 00001–00006 already use uuid).

-- +goose StatementBegin
DO $wiz$
DECLARE
    users_id_type text;
    dev_old text := 'grp_dev_wichat_corp';
    dev_new uuid := '00000000-0000-4000-8000-00000000d301';
BEGIN
    SELECT data_type INTO users_id_type
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'id';

    IF users_id_type IS NULL OR users_id_type = 'uuid' THEN
        RETURN;
    END IF;

    UPDATE group_members SET group_id = dev_new::text WHERE group_id = dev_old;
    UPDATE group_settings SET group_id = dev_new::text WHERE group_id = dev_old;
    UPDATE groups SET id = dev_new::text WHERE id = dev_old;

    ALTER TABLE user_credentials DROP CONSTRAINT IF EXISTS user_credentials_user_id_fkey;
    ALTER TABLE user_connections DROP CONSTRAINT IF EXISTS user_connections_user_id_fkey;
    ALTER TABLE user_sessions DROP CONSTRAINT IF EXISTS user_sessions_user_id_fkey;
    ALTER TABLE group_members DROP CONSTRAINT IF EXISTS group_members_user_id_fkey;
    ALTER TABLE group_members DROP CONSTRAINT IF EXISTS group_members_group_id_fkey;
    ALTER TABLE group_settings DROP CONSTRAINT IF EXISTS group_settings_group_id_fkey;

    ALTER TABLE users ALTER COLUMN id TYPE uuid USING id::uuid;
    ALTER TABLE groups ALTER COLUMN id TYPE uuid USING id::uuid;

    ALTER TABLE user_credentials ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
    ALTER TABLE user_connections ALTER COLUMN id TYPE uuid USING id::uuid;
    ALTER TABLE user_connections ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
    ALTER TABLE user_sessions ALTER COLUMN id TYPE uuid USING id::uuid;
    ALTER TABLE user_sessions ALTER COLUMN user_id TYPE uuid USING user_id::uuid;
    ALTER TABLE group_settings ALTER COLUMN group_id TYPE uuid USING group_id::uuid;
    ALTER TABLE group_members ALTER COLUMN group_id TYPE uuid USING group_id::uuid;
    ALTER TABLE group_members ALTER COLUMN user_id TYPE uuid USING user_id::uuid;

    ALTER TABLE user_credentials
        ADD CONSTRAINT user_credentials_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
    ALTER TABLE user_connections
        ADD CONSTRAINT user_connections_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
    ALTER TABLE user_sessions
        ADD CONSTRAINT user_sessions_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
    ALTER TABLE group_settings
        ADD CONSTRAINT group_settings_group_id_fkey
        FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE;
    ALTER TABLE group_members
        ADD CONSTRAINT group_members_group_id_fkey
        FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE;
    ALTER TABLE group_members
        ADD CONSTRAINT group_members_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
END $wiz$;
-- +goose StatementEnd

-- +goose Down
-- Irreversible without data loss; fresh dev DBs should reset instead.
