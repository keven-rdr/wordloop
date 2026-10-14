-- +goose Up
CREATE TABLE passkey_credential (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid        NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    credential_id   bytea       NOT NULL UNIQUE,
    public_key      bytea       NOT NULL,
    sign_count      bigint      NOT NULL DEFAULT 0,
    transports      text[]      NOT NULL DEFAULT '{}',
    user_verified   boolean     NOT NULL DEFAULT false,
    backup_eligible boolean     NOT NULL DEFAULT false,
    backup_state    boolean     NOT NULL DEFAULT false,
    aaguid          bytea,
    name            text        NOT NULL,
    created_at      timestamptz NOT NULL,
    last_used_at    timestamptz
);
CREATE INDEX passkey_credential_user_idx ON passkey_credential (user_id);

CREATE TABLE auth_challenge (
    id_hash     bytea       PRIMARY KEY,
    purpose     text        NOT NULL CHECK (purpose IN ('register', 'login')),
    user_id     uuid        REFERENCES app_user (id) ON DELETE CASCADE,
    invite_hash bytea,
    state       bytea       NOT NULL,
    created_at  timestamptz NOT NULL,
    expires_at  timestamptz NOT NULL
);
CREATE INDEX auth_challenge_expires_idx ON auth_challenge (expires_at);
