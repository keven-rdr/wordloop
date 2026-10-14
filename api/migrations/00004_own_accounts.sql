-- +goose Up
ALTER TABLE app_user
    ADD COLUMN email             text,
    ADD COLUMN email_verified_at timestamptz;
CREATE UNIQUE INDEX app_user_email_idx ON app_user (lower(email)) WHERE email IS NOT NULL;

CREATE TABLE auth_identity (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    provider   text        NOT NULL,
    subject    text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, subject)
);
CREATE INDEX auth_identity_user_idx ON auth_identity (user_id);

INSERT INTO auth_identity (user_id, provider, subject)
SELECT id, idp_issuer, idp_subject FROM app_user WHERE idp_issuer IS NOT NULL AND idp_subject IS NOT NULL;

CREATE TABLE auth_token (
    id_hash    bytea       PRIMARY KEY,
    user_id    uuid        NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    purpose    text        NOT NULL CHECK (purpose IN ('invite', 'recovery', 'verify_email')),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    used_at    timestamptz
);
CREATE INDEX auth_token_user_idx ON auth_token (user_id, purpose);
