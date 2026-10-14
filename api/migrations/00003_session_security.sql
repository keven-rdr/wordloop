-- +goose Up
-- Expand (so adiciona; a versao anterior da API continua funcionando durante o deploy).
-- Papel: so muda por SQL feito pelo dono (docs/features/2026-10-14-sessao-segura-admin.md); nenhuma rota o altera.
ALTER TABLE app_user ADD COLUMN role text NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));

-- public_id identifica a sessao nas listagens (nao e o segredo do cookie); authenticated_at e mfa vem do ID token.
ALTER TABLE auth_session
    ADD COLUMN public_id        uuid        NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN authenticated_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN mfa              boolean     NOT NULL DEFAULT false;
CREATE UNIQUE INDEX auth_session_public_id_idx ON auth_session (public_id);
