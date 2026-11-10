-- +goose Up
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS is_private boolean NOT NULL DEFAULT false;

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
SET type = 'text_and_voice', is_private = false
WHERE type = 'public';

UPDATE channels
SET type = 'text_and_voice', is_private = true
WHERE type = 'private';

UPDATE channels
SET is_private = false
WHERE type = 'dm';

ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (
        type IN ('broadcast', 'text_only', 'voice_only', 'text_and_voice', 'forum', 'dm')
    );

ALTER TABLE channels
    ADD CONSTRAINT channels_dm_not_private_check CHECK (type <> 'dm' OR is_private = false);

-- +goose Down
ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_dm_not_private_check;
ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_type_check;

ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (type IN ('public', 'private', 'dm', 'broadcast'));

UPDATE channels SET type = 'public' WHERE type = 'text_and_voice' AND is_private = false;
UPDATE channels SET type = 'private' WHERE type = 'text_and_voice' AND is_private = true;

ALTER TABLE channels DROP COLUMN IF EXISTS is_private;
