-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    display_name text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    user_agent text NOT NULL DEFAULT '',
    client_ip text NOT NULL DEFAULT '',
    last_used_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE refresh_tokens (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    used_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX refresh_tokens_one_active
    ON refresh_tokens (session_id)
    WHERE used_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
