-- +goose Up
CREATE TABLE identities (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id),
    provider text NOT NULL,
    subject text NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX identities_provider_subject
    ON identities (provider, subject);

-- +goose Down
DROP TABLE IF EXISTS identities;
