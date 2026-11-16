-- +goose Up
CREATE TABLE IF NOT EXISTS group_activity_logs (
    id          uuid        PRIMARY KEY,
    group_id    uuid        NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    actor_id    uuid        REFERENCES users(id) ON DELETE SET NULL,
    category    text        NOT NULL,
    action      text        NOT NULL,
    target_type text        NOT NULL DEFAULT '',
    target_id   text        NOT NULL DEFAULT '',
    target_name text        NOT NULL DEFAULT '',
    summary     text        NOT NULL DEFAULT '',
    changes     jsonb       NOT NULL DEFAULT '{}'::jsonb,
    metadata    jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_group_activity_logs_group_created
    ON group_activity_logs (group_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_group_activity_logs_group_actor_created
    ON group_activity_logs (group_id, actor_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_group_activity_logs_group_action_created
    ON group_activity_logs (group_id, action, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_group_activity_logs_group_category_created
    ON group_activity_logs (group_id, category, created_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS group_activity_logs;
