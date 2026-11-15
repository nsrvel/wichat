-- +goose Up
UPDATE group_roles
SET permissions = array_remove(permissions, 'create_invite')
WHERE slug = 'everyone' AND is_default = true AND 'create_invite' = ANY(permissions);

-- +goose Down
-- No-op: cannot distinguish rows that originally had create_invite vs explicitly enabled.
