-- +goose Up
CREATE TABLE user_mfa_totp (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    secret_encrypted bytea NOT NULL,
    last_used_step bigint NOT NULL DEFAULT 0,
    enabled_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE user_webauthn_credentials (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name text NOT NULL,
    credential_id bytea NOT NULL UNIQUE,
    public_key bytea NOT NULL,
    attestation_type text NOT NULL DEFAULT 'none',
    aaguid bytea NOT NULL DEFAULT ''::bytea,
    sign_count bigint NOT NULL DEFAULT 0,
    clone_warning boolean NOT NULL DEFAULT false,
    transports text[] NOT NULL DEFAULT '{}',
    backup_eligible boolean NOT NULL DEFAULT false,
    backup_state boolean NOT NULL DEFAULT false,
    last_used_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE INDEX user_webauthn_credentials_user_id_idx
    ON user_webauthn_credentials (user_id);

CREATE TABLE user_mfa_recovery_codes (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    code_hash bytea NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE INDEX user_mfa_recovery_codes_unused_idx
    ON user_mfa_recovery_codes (user_id)
    WHERE used_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS user_mfa_recovery_codes;
DROP TABLE IF EXISTS user_webauthn_credentials;
DROP TABLE IF EXISTS user_mfa_totp;

