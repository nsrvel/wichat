-- +goose Up
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS position integer NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS channels_group_pos_idx ON channels (group_id, position, name);

WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY group_id, category_id
            ORDER BY name ASC, created_at ASC
        ) - 1 AS pos
    FROM channels
)
UPDATE channels AS c
SET position = ranked.pos
FROM ranked
WHERE c.id = ranked.id;

-- +goose Down
DROP INDEX IF EXISTS channels_group_pos_idx;
ALTER TABLE channels DROP COLUMN IF EXISTS position;
