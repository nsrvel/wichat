-- +goose Up
-- Legacy installs created tables under schema auth; move to public (no-op on fresh DBs).

-- +goose StatementBegin
DO $wichat$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.schemata WHERE schema_name = 'auth'
    ) THEN
        RETURN;
    END IF;

    ALTER TABLE auth.users SET SCHEMA public;
    ALTER TABLE auth.user_credentials SET SCHEMA public;
    ALTER TABLE auth.user_connections SET SCHEMA public;
    ALTER TABLE auth.user_sessions SET SCHEMA public;
    ALTER TABLE auth.groups SET SCHEMA public;
    ALTER TABLE auth.group_settings SET SCHEMA public;
    ALTER TABLE auth.group_members SET SCHEMA public;
    ALTER TABLE auth.instance_settings SET SCHEMA public;

    DROP SCHEMA auth;
END $wichat$;
-- +goose StatementEnd

-- +goose Down
CREATE SCHEMA IF NOT EXISTS auth;

ALTER TABLE public.users SET SCHEMA auth;
ALTER TABLE public.user_credentials SET SCHEMA auth;
ALTER TABLE public.user_connections SET SCHEMA auth;
ALTER TABLE public.user_sessions SET SCHEMA auth;
ALTER TABLE public.groups SET SCHEMA auth;
ALTER TABLE public.group_settings SET SCHEMA auth;
ALTER TABLE public.group_members SET SCHEMA auth;
ALTER TABLE public.instance_settings SET SCHEMA auth;
