-- +goose Up
ALTER TABLE app_user ADD COLUMN mfa_required boolean NOT NULL DEFAULT false;
