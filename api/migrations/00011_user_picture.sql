-- +goose Up
-- Foto de perfil vinda do login social (Google, GitHub); so a URL, sem upload.
ALTER TABLE app_user ADD COLUMN picture_url text;

-- +goose Down
ALTER TABLE app_user DROP COLUMN picture_url;
