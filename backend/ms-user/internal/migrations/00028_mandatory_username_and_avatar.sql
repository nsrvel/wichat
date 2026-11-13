-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar_url text NOT NULL DEFAULT '';

UPDATE users
SET username = 'user_' || left(replace(id::text, '-', ''), 12)
WHERE username IS NULL OR btrim(username) = '';

UPDATE users
SET username = lower(btrim(username))
WHERE username <> lower(btrim(username));

ALTER TABLE users
    ALTER COLUMN username SET NOT NULL;

DROP INDEX IF EXISTS users_username_unique;
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_unique ON users (lower(username));

ALTER TABLE instance_settings
    DROP COLUMN IF EXISTS username_enabled;

-- +goose Down
ALTER TABLE instance_settings
    ADD COLUMN IF NOT EXISTS username_enabled boolean NOT NULL DEFAULT true;

DROP INDEX IF EXISTS users_username_lower_unique;
CREATE UNIQUE INDEX IF NOT EXISTS users_username_unique ON users (username) WHERE username IS NOT NULL AND username <> '';

ALTER TABLE users
    ALTER COLUMN username DROP NOT NULL;

ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_url;

