-- +goose Up
ALTER TABLE groups ADD COLUMN public_id text;

UPDATE groups
SET public_id = substring(
    replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''),
    1,
    24
)
WHERE public_id IS NULL;

ALTER TABLE groups
    ALTER COLUMN public_id SET NOT NULL;

ALTER TABLE groups
    ADD CONSTRAINT groups_public_id_format CHECK (public_id ~ '^[0-9a-f]{24}$');

CREATE UNIQUE INDEX groups_public_id_key ON groups (public_id);

-- +goose Down
DROP INDEX IF EXISTS groups_public_id_key;
ALTER TABLE groups DROP CONSTRAINT IF EXISTS groups_public_id_format;
ALTER TABLE groups DROP COLUMN IF EXISTS public_id;
