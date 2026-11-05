-- +goose Up
-- Normalizes identity: credentials, connections, user_sessions, group_settings.
-- Existing refresh rows in Postgres are dropped; clients must sign in again after upgrade.

CREATE TABLE user_credentials (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    password_hash text NOT NULL,
    updated_at timestamptz NOT NULL
);

INSERT INTO user_credentials (user_id, password_hash, updated_at)
SELECT id, password_hash, updated_at
FROM users
WHERE password_hash IS NOT NULL AND password_hash <> '';

CREATE TABLE user_connections (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('login', 'integration')),
    provider text NOT NULL,
    subject text NOT NULL,
    scopes text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

INSERT INTO user_connections (id, user_id, kind, provider, subject, scopes, created_at, updated_at)
SELECT id, user_id, 'login', provider, subject, '{}', created_at, created_at
FROM identities;

CREATE UNIQUE INDEX user_connections_login_provider_subject
    ON user_connections (provider, subject)
    WHERE kind = 'login';

CREATE TABLE group_settings (
    group_id uuid PRIMARY KEY REFERENCES groups (id) ON DELETE CASCADE,
    name text NOT NULL,
    logo_url text NOT NULL DEFAULT '',
    retention_days int NOT NULL DEFAULT 0,
    trash_retention_days int NOT NULL DEFAULT 30,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

INSERT INTO group_settings (group_id, name, logo_url, retention_days, trash_retention_days, created_at, updated_at)
SELECT id, name, logo_url, 0, 30, created_at, updated_at
FROM groups;

ALTER TABLE groups
    DROP COLUMN name,
    DROP COLUMN logo_url;

CREATE TABLE user_sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    user_agent text NOT NULL DEFAULT '',
    client_ip text NOT NULL DEFAULT '',
    device_label text NOT NULL DEFAULT '',
    platform text NOT NULL DEFAULT '',
    location_label text NOT NULL DEFAULT '',
    refresh_id text NOT NULL DEFAULT '',
    last_used_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);

INSERT INTO user_sessions (
    id, user_id, expires_at, revoked_at, user_agent, client_ip,
    device_label, platform, location_label, refresh_id, last_used_at, created_at
)
SELECT
    id, user_id, expires_at, revoked_at, user_agent, client_ip,
    '', '', '', '', last_used_at, created_at
FROM sessions;

ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS identities;

-- +goose Down
CREATE TABLE identities (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id),
    provider text NOT NULL,
    subject text NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX identities_provider_subject ON identities (provider, subject);

INSERT INTO identities (id, user_id, provider, subject, created_at)
SELECT id, user_id, provider, subject, created_at
FROM user_connections
WHERE kind = 'login';

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

INSERT INTO sessions (id, user_id, expires_at, revoked_at, user_agent, client_ip, last_used_at, created_at)
SELECT id, user_id, expires_at, revoked_at, user_agent, client_ip, last_used_at, created_at
FROM user_sessions;

CREATE TABLE refresh_tokens (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    used_at timestamptz,
    created_at timestamptz NOT NULL
);

ALTER TABLE users ADD COLUMN password_hash text;

UPDATE users u
SET password_hash = c.password_hash
FROM user_credentials c
WHERE c.user_id = u.id;

ALTER TABLE groups
    ADD COLUMN name text NOT NULL DEFAULT '',
    ADD COLUMN logo_url text NOT NULL DEFAULT '';

UPDATE groups g
SET name = s.name, logo_url = s.logo_url
FROM group_settings s
WHERE s.group_id = g.id;

DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS group_settings;
DROP TABLE IF EXISTS user_connections;
DROP TABLE IF EXISTS user_credentials;
