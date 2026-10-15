-- +goose Up
ALTER TABLE app_user
    ADD COLUMN password_hash  text,
    ADD COLUMN failed_logins  integer     NOT NULL DEFAULT 0,
    ADD COLUMN last_failed_at timestamptz,
    ADD COLUMN locked_until   timestamptz;
