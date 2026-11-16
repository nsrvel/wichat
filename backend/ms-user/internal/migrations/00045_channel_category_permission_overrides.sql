-- +goose Up
CREATE TABLE channel_category_permission_overrides (
    id uuid PRIMARY KEY,
    category_id uuid NOT NULL REFERENCES channel_categories (id) ON DELETE CASCADE,
    target_type text NOT NULL CHECK (target_type IN ('role', 'member')),
    target_id uuid NOT NULL,
    allow_permissions text[] NOT NULL DEFAULT '{}',
    deny_permissions text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT channel_category_perm_overrides_unique UNIQUE (category_id, target_type, target_id)
);

CREATE INDEX channel_category_perm_overrides_cat_idx
    ON channel_category_permission_overrides (category_id, target_type);

-- +goose Down
DROP TABLE IF EXISTS channel_category_permission_overrides;
