-- +goose Up
ALTER TABLE auth.users
    ADD COLUMN is_instance_owner boolean NOT NULL DEFAULT false;

CREATE TABLE auth.instance_settings (
    id smallint PRIMARY KEY CHECK (id = 1),
    name text NOT NULL DEFAULT '',
    logo_url text NOT NULL DEFAULT '',
    username_enabled boolean NOT NULL DEFAULT true,
    registration_mode text NOT NULL DEFAULT 'invite_only'
        CHECK (registration_mode IN ('invite_only', 'approval_required', 'open')),
    guest_enabled boolean NOT NULL DEFAULT false,
    setup_completed boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

INSERT INTO auth.instance_settings (
    id, name, logo_url, username_enabled, registration_mode, guest_enabled, setup_completed, created_at, updated_at
) VALUES (
    1, '', '', true, 'invite_only', false, false, now() AT TIME ZONE 'utc', now() AT TIME ZONE 'utc'
);

-- +goose Down
DROP TABLE IF EXISTS auth.instance_settings;

ALTER TABLE auth.users
    DROP COLUMN IF EXISTS is_instance_owner;
