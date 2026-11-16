-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    grp RECORD;
    fav_cat_id uuid;
    owner_id uuid;
    now_ts timestamptz := now();
BEGIN
    FOR grp IN
        SELECT DISTINCT c.group_id
        FROM channels AS c
        WHERE c.sidebar_section = 'favorites'
          AND c.category_id IS NULL
    LOOP
        SELECT cc.id INTO fav_cat_id
        FROM channel_categories AS cc
        WHERE cc.group_id = grp.group_id
          AND lower(cc.name) = 'favorites'
        LIMIT 1;

        IF fav_cat_id IS NULL THEN
            UPDATE channel_categories
            SET position = position + 1
            WHERE group_id = grp.group_id;

            SELECT COALESCE(
                (SELECT m.user_id FROM group_members AS m WHERE m.group_id = grp.group_id AND m.role = 'owner' ORDER BY m.created_at ASC LIMIT 1),
                (SELECT m.user_id FROM group_members AS m WHERE m.group_id = grp.group_id ORDER BY m.created_at ASC LIMIT 1)
            ) INTO owner_id;

            IF owner_id IS NOT NULL THEN
                fav_cat_id := gen_random_uuid();
                INSERT INTO channel_categories (id, public_id, group_id, name, is_private, position, created_by, created_at, updated_at)
                VALUES (
                    fav_cat_id,
                    substring(replace(gen_random_uuid()::text, '-', '') || replace(gen_random_uuid()::text, '-', ''), 1, 24),
                    grp.group_id,
                    'Favorites',
                    false,
                    0,
                    owner_id,
                    now_ts,
                    now_ts
                );
            END IF;
        END IF;

        IF fav_cat_id IS NOT NULL THEN
            UPDATE channels
            SET category_id = fav_cat_id,
                sidebar_section = 'channels'
            WHERE group_id = grp.group_id
              AND sidebar_section = 'favorites'
              AND category_id IS NULL;
        END IF;
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- Non-destructive; regular Favorites categories remain normal categories.
SELECT 1;
