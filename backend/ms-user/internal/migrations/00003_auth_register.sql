-- +goose Up
ALTER TABLE users
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users
    ADD COLUMN username text,
    ADD COLUMN birth_date date,
    ADD COLUMN is_guest boolean NOT NULL DEFAULT false;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX users_email_unique ON users (email) WHERE email IS NOT NULL AND email <> '';
CREATE UNIQUE INDEX users_username_unique ON users (username) WHERE username IS NOT NULL AND username <> '';

-- +goose Down
DROP INDEX IF EXISTS users_username_unique;
DROP INDEX IF EXISTS users_email_unique;

ALTER TABLE users
    DROP COLUMN IF EXISTS is_guest,
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS username;

ALTER TABLE users
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN password_hash SET NOT NULL;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
