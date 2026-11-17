-- +goose Up
-- Voice & Video Phase 3 schema (owned by ms-core; goose runs from ms-user).

ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_type_check;
ALTER TABLE channels ADD CONSTRAINT channels_type_check
    CHECK (type IN ('broadcast', 'text_only', 'voice_only', 'text_and_voice', 'forum', 'dm'));

CREATE TABLE IF NOT EXISTS call_sessions (
    id                        uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    public_id                 text        NOT NULL UNIQUE CHECK (public_id ~ '^[0-9a-f]{24}$'),
    group_id                  uuid        REFERENCES groups (id) ON DELETE CASCADE,
    channel_id                uuid        NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    room_name                 text        NOT NULL,
    started_at                timestamptz NOT NULL DEFAULT now(),
    ended_at                  timestamptz,
    peak_participants         integer     NOT NULL DEFAULT 1,
    total_participant_seconds bigint      NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS call_sessions_channel_started_idx
    ON call_sessions (channel_id, started_at DESC);

CREATE INDEX IF NOT EXISTS call_sessions_started_idx
    ON call_sessions (started_at);

CREATE INDEX IF NOT EXISTS call_sessions_active_channel_idx
    ON call_sessions (channel_id)
    WHERE ended_at IS NULL;

CREATE TABLE IF NOT EXISTS call_participants (
    id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id        uuid        NOT NULL REFERENCES call_sessions (id) ON DELETE CASCADE,
    user_id           uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    joined_at         timestamptz NOT NULL DEFAULT now(),
    left_at           timestamptz,
    duration_seconds  integer     NOT NULL DEFAULT 0,
    server_muted      boolean     NOT NULL DEFAULT false,
    server_deafened   boolean     NOT NULL DEFAULT false,
    disconnect_reason text        NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS call_participants_session_user_idx
    ON call_participants (session_id, user_id);

CREATE INDEX IF NOT EXISTS call_participants_user_joined_idx
    ON call_participants (user_id, joined_at DESC);

CREATE TABLE IF NOT EXISTS call_recordings (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    public_id        text        NOT NULL UNIQUE CHECK (public_id ~ '^[0-9a-f]{24}$'),
    session_id       uuid        NOT NULL REFERENCES call_sessions (id) ON DELETE CASCADE,
    channel_id       uuid        NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    egress_id        text        NOT NULL UNIQUE,
    status           text        NOT NULL CHECK (status IN ('starting', 'active', 'stopping', 'completed', 'failed')),
    file_key         text        NOT NULL DEFAULT '',
    file_size_bytes  bigint      NOT NULL DEFAULT 0,
    duration_seconds integer     NOT NULL DEFAULT 0,
    started_by       uuid        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    stopped_by       uuid        REFERENCES users (id) ON DELETE SET NULL,
    started_at       timestamptz NOT NULL DEFAULT now(),
    ended_at         timestamptz
);

CREATE INDEX IF NOT EXISTS call_recordings_channel_started_idx
    ON call_recordings (channel_id, started_at DESC);

CREATE INDEX IF NOT EXISTS call_recordings_session_idx
    ON call_recordings (session_id);

-- +goose Down
DROP TABLE IF EXISTS call_recordings;
DROP TABLE IF EXISTS call_participants;
DROP TABLE IF EXISTS call_sessions;
ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_type_check;
ALTER TABLE channels ADD CONSTRAINT channels_type_check
    CHECK (type IN ('broadcast', 'text_only', 'voice_only', 'forum', 'dm'));
