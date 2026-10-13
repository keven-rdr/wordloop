-- +goose Up
-- Expand: identidade do IdP no perfil e sessao do BFF. O sub de um IdP externo pode nao ser UUID, entao o id
-- passa a ser gerado pelo app e a identidade e o par (idp_issuer, idp_subject).
ALTER TABLE app_user ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE app_user ADD COLUMN idp_issuer text, ADD COLUMN idp_subject text;
CREATE UNIQUE INDEX app_user_idp_identity_idx ON app_user (idp_issuer, idp_subject);

-- Sessao no servidor: o navegador so tem o cookie httpOnly; aqui fica o HASH do id (nunca o id) e o refresh token cifrado.
CREATE TABLE auth_session (
    id_hash           bytea PRIMARY KEY,
    user_id           uuid NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    refresh_token_enc text NOT NULL,
    created_at        timestamptz NOT NULL DEFAULT now(),
    last_used_at      timestamptz NOT NULL DEFAULT now(),
    expires_at        timestamptz NOT NULL,
    user_agent        text
);
CREATE INDEX auth_session_user_idx ON auth_session (user_id);
