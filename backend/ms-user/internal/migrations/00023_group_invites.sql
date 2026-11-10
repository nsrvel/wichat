-- +goose Up
CREATE TABLE group_invites (
    id uuid PRIMARY KEY,
    token text NOT NULL UNIQUE,
    group_id uuid NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    channel_id uuid REFERENCES channels (id) ON DELETE SET NULL,
    email text,
    invited_by_user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind text NOT NULL DEFAULT 'email'
        CHECK (kind IN ('email', 'share_link')),
    expires_at timestamptz NOT NULL,
    redeemed_at timestamptz,
    redeemed_by_user_id uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE INDEX group_invites_group_id_idx ON group_invites (group_id);
CREATE INDEX group_invites_token_idx ON group_invites (token);
CREATE INDEX group_invites_email_open_idx ON group_invites (group_id, lower(email))
    WHERE email IS NOT NULL AND redeemed_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS group_invites;
