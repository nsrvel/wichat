-- +goose Up
CREATE TABLE auth.identities (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES auth.users (id),
    provider text NOT NULL,
    subject text NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX identities_provider_subject
    ON auth.identities (provider, subject);

-- +goose Down
DROP TABLE IF EXISTS auth.identities;
