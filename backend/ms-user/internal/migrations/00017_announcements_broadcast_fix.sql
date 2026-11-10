-- +goose Up
-- Idempotent repair when 00016 type check name differed.
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

ALTER TABLE channels
    ADD CONSTRAINT channels_type_check CHECK (type IN ('public', 'private', 'dm', 'broadcast'));

UPDATE channels
SET
    type = 'broadcast',
    name = 'Announcements'
WHERE lower(name) = 'announcements';

-- +goose Down
-- No-op; 00016 down handles rollback.
