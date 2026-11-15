-- +goose Up
CREATE TABLE IF NOT EXISTS user_activity_logs (
    id          uuid        PRIMARY KEY,
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id    uuid        REFERENCES users(id) ON DELETE SET NULL,
    category    text        NOT NULL,
    action      text        NOT NULL,
    target_type text        NOT NULL DEFAULT '',
    target_id   text        NOT NULL DEFAULT '',
    summary     text        NOT NULL DEFAULT '',
    changes     jsonb       NOT NULL DEFAULT '{}'::jsonb,
    metadata    jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_activity_logs_user_created
    ON user_activity_logs (user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_user_activity_logs_user_category_created
    ON user_activity_logs (user_id, category, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_user_activity_logs_action_created
    ON user_activity_logs (action, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_activity_logs;
