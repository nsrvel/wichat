-- +goose Up
CREATE TABLE user_relationships (
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    target_user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('ignore', 'block')),
    created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    PRIMARY KEY (user_id, target_user_id),
    CHECK (user_id <> target_user_id)
);

CREATE INDEX user_relationships_user_kind_idx
    ON user_relationships (user_id, kind);

-- +goose Down
DROP TABLE IF EXISTS user_relationships;
