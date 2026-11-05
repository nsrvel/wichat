-- +goose Up
-- Drop the old local-dev seed group (grp_dev_wichat_corp → fixed UUID). Not used in production flows.
DELETE FROM groups
WHERE id = '00000000-0000-4000-8000-00000000d301'::uuid;

-- +goose Down
-- Data restore not supported; re-seed manually if needed for archaeology.
