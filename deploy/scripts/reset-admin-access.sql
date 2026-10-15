-- Esqueci a senha do admin: gera um novo convite (15 minutos, uso unico) para o admin que ja existe, apaga todas as
-- sessoes dele e destrava a conta. Cole no SQL Editor do Neon, troque o e-mail e rode; a resposta e o link.
-- O link abre a tela de definir a senha. Resposta vazia: nao ha admin com esse e-mail (nada foi alterado).
WITH p AS (
    SELECT lower(btrim('SEU_EMAIL@EXEMPLO.COM')) AS email,
           'https://tst.wordloop.com.br'          AS site,
           translate(encode(decode(replace(gen_random_uuid()::text || gen_random_uuid()::text, '-', ''), 'hex'), 'base64'), '+/=', '-_') AS segredo
), alvo AS (
    SELECT u.id FROM app_user u, p
    WHERE lower(u.email) = p.email AND u.role = 'admin' AND u.deleted_at IS NULL
), revogados AS (
    DELETE FROM auth_token
    WHERE user_id IN (SELECT id FROM alvo) AND purpose = 'invite' AND used_at IS NULL
), sessoes AS (
    DELETE FROM auth_session WHERE user_id IN (SELECT id FROM alvo)
), destravado AS (
    UPDATE app_user SET failed_logins = 0, last_failed_at = NULL, locked_until = NULL
    WHERE id IN (SELECT id FROM alvo)
), convite AS (
    INSERT INTO auth_token (id_hash, user_id, purpose, expires_at)
    SELECT sha256(convert_to(p.segredo, 'UTF8')), alvo.id, 'invite', now() + interval '15 minutes'
    FROM p, alvo
    RETURNING 1
)
SELECT p.site || '/invite#' || p.segredo AS link_do_convite, now() + interval '15 minutes' AS vale_ate
FROM p, alvo;
