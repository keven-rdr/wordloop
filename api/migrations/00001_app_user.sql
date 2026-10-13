-- +goose Up
-- Perfil minimo do app. Identidade (e-mail, senha, vinculos) fica no Keycloak; id = claim "sub" (criado no 1o login).
CREATE TABLE app_user (
    id           uuid PRIMARY KEY,
    display_name text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz,
    deleted_at   timestamptz
);
