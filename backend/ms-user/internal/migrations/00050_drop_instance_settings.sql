-- +goose Up
DROP TABLE IF EXISTS instance_settings;

-- +goose Down
CREATE TABLE instance_settings (
    id smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    name text NOT NULL DEFAULT 'WiChat',
    logo_url text NOT NULL DEFAULT '',
    username_enabled boolean NOT NULL DEFAULT true,
    registration_mode text NOT NULL DEFAULT 'invite_only'
        CHECK (registration_mode IN ('invite_only', 'approval_required', 'open')),
    setup_completed boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    updated_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

INSERT INTO instance_settings (
    id, name, logo_url, username_enabled, registration_mode, setup_completed, created_at, updated_at
) VALUES (1, 'WiChat', '', true, 'open', true, now() AT TIME ZONE 'utc', now() AT TIME ZONE 'utc');