-- +goose Up
ALTER TABLE app_user ADD COLUMN role text NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));

ALTER TABLE auth_session
    ADD COLUMN public_id        uuid        NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN authenticated_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN mfa              boolean     NOT NULL DEFAULT false;
CREATE UNIQUE INDEX auth_session_public_id_idx ON auth_session (public_id);
