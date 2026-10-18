-- +goose Up
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.users (
    id text PRIMARY KEY,
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    display_name text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE auth.sessions (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES auth.users (id),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    user_agent text NOT NULL DEFAULT '',
    client_ip text NOT NULL DEFAULT '',
    last_used_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE auth.refresh_tokens (
    id text PRIMARY KEY,
    session_id text NOT NULL REFERENCES auth.sessions (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    used_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX refresh_tokens_one_active
    ON auth.refresh_tokens (session_id)
    WHERE used_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS auth.refresh_tokens;
DROP TABLE IF EXISTS auth.sessions;
DROP TABLE IF EXISTS auth.users;
DROP SCHEMA IF EXISTS auth;
