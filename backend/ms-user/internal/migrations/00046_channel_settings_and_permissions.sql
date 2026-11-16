-- +goose Up
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS slowmode_seconds integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS thread_slowmode_seconds integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS content_visibility text NOT NULL DEFAULT 'default'
        CHECK (content_visibility IN ('default', 'spoiler', 'age_restricted')),
    ADD COLUMN IF NOT EXISTS forum_default_layout text NOT NULL DEFAULT 'list'
        CHECK (forum_default_layout IN ('list', 'gallery')),
    ADD COLUMN IF NOT EXISTS forum_sort_order text NOT NULL DEFAULT 'recent_activity'
        CHECK (forum_sort_order IN ('recent_activity', 'creation_time')),
    ADD COLUMN IF NOT EXISTS forum_tag_matching text NOT NULL DEFAULT 'match_some'
        CHECK (forum_tag_matching IN ('match_some', 'match_all')),
    ADD COLUMN IF NOT EXISTS bitrate_kbps integer NOT NULL DEFAULT 64,
    ADD COLUMN IF NOT EXISTS video_quality text NOT NULL DEFAULT 'auto'
        CHECK (video_quality IN ('auto', '720p')),
    ADD COLUMN IF NOT EXISTS user_limit integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS rtc_region text NOT NULL DEFAULT 'automatic';

CREATE TABLE channel_permission_overrides (
    id uuid PRIMARY KEY,
    channel_id uuid NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
    target_type text NOT NULL CHECK (target_type IN ('role', 'member')),
    target_id uuid NOT NULL,
    allow_permissions text[] NOT NULL DEFAULT '{}',
    deny_permissions text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT channel_perm_overrides_unique UNIQUE (channel_id, target_type, target_id)
);

CREATE INDEX channel_perm_overrides_chan_idx
    ON channel_permission_overrides (channel_id, target_type);

-- +goose Down
DROP TABLE IF EXISTS channel_permission_overrides;

ALTER TABLE channels
    DROP COLUMN IF EXISTS rtc_region,
    DROP COLUMN IF EXISTS user_limit,
    DROP COLUMN IF EXISTS video_quality,
    DROP COLUMN IF EXISTS bitrate_kbps,
    DROP COLUMN IF EXISTS forum_tag_matching,
    DROP COLUMN IF EXISTS forum_sort_order,
    DROP COLUMN IF EXISTS forum_default_layout,
    DROP COLUMN IF EXISTS content_visibility,
    DROP COLUMN IF EXISTS thread_slowmode_seconds,
    DROP COLUMN IF EXISTS slowmode_seconds;
