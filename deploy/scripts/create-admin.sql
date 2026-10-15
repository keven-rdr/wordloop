-- Cria o administrador (ou gera um novo convite para um admin que ainda nao tem passkey) direto no banco.
-- Cole no SQL Editor do Neon, troque os 3 valores marcados e rode. A resposta e o link do convite.
-- O link vale 15 minutos e so uma vez; abra-o para definir a sua senha (ou criar uma passkey).
-- Resposta vazia: o e-mail ja pertence a uma conta que nao e um admin pendente (nada foi alterado).
WITH p AS (
    SELECT lower(btrim('SEU_EMAIL@EXEMPLO.COM')) AS email,
           btrim('SEU NOME')                      AS nome,
           'https://tst.wordloop.com.br'          AS site,
           translate(encode(decode(replace(gen_random_uuid()::text || gen_random_uuid()::text, '-', ''), 'hex'), 'base64'), '+/=', '-_') AS segredo
), admin AS (
    INSERT INTO app_user (display_name, email, email_verified_at, role)
    SELECT nome, email, now(), 'admin' FROM p
    ON CONFLICT (lower(email)) WHERE email IS NOT NULL
    DO UPDATE SET email = EXCLUDED.email
    WHERE app_user.role = 'admin' AND app_user.deleted_at IS NULL
      AND NOT EXISTS (SELECT 1 FROM passkey_credential c WHERE c.user_id = app_user.id)
      AND NOT EXISTS (SELECT 1 FROM auth_identity i WHERE i.user_id = app_user.id)
      AND app_user.password_hash IS NULL
    RETURNING id
), revogados AS (
    DELETE FROM auth_token
    WHERE user_id IN (SELECT id FROM admin) AND purpose = 'invite' AND used_at IS NULL
), convite AS (
    INSERT INTO auth_token (id_hash, user_id, purpose, expires_at)
    SELECT sha256(convert_to(p.segredo, 'UTF8')), admin.id, 'invite', now() + interval '15 minutes'
    FROM p, admin
    RETURNING 1
)
SELECT p.site || '/invite#' || p.segredo AS link_do_convite, now() + interval '15 minutes' AS vale_ate
FROM p, admin;
