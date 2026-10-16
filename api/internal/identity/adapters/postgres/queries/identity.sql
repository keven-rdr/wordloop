-- name: GetUserIDByIdentity :one
SELECT user_id FROM auth_identity WHERE provider = $1 AND subject = $2;

-- name: GetUserIDByLegacyIdentity :one
SELECT id FROM app_user WHERE idp_issuer = $1 AND idp_subject = $2;

-- name: LinkIdentity :exec
INSERT INTO auth_identity (user_id, provider, subject) VALUES ($1, $2, $3)
ON CONFLICT (provider, subject) DO NOTHING;

-- name: GetUserIDByEmail :one
SELECT id FROM app_user WHERE lower(email) = lower(sqlc.arg(email)::text) AND deleted_at IS NULL;

-- name: CreateUserWithIdentity :one
WITH u AS (
    INSERT INTO app_user (idp_issuer, idp_subject, display_name, email, email_verified_at, last_seen_at)
    VALUES ($1, $2, $3, $4, $5, now())
    RETURNING id
), l AS (
    INSERT INTO auth_identity (user_id, provider, subject) SELECT id, $1, $2 FROM u
)
SELECT id FROM u;

-- name: TouchUser :one
UPDATE app_user
SET last_seen_at = now(),
    display_name = COALESCE(NULLIF(sqlc.arg(display_name)::text, ''), display_name),
    picture_url = COALESCE(NULLIF(sqlc.arg(picture_url)::text, ''), picture_url)
WHERE id = sqlc.arg(id)
RETURNING id, display_name, role, mfa_required, picture_url;

-- name: GetUser :one
SELECT id, display_name, role, mfa_required, picture_url FROM app_user WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateSession :exec
INSERT INTO auth_session (id_hash, user_id, refresh_token_enc, created_at, last_used_at, expires_at, authenticated_at, mfa, user_agent)
VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8);

-- name: GetSession :one
SELECT public_id, user_id, created_at, last_used_at, expires_at, authenticated_at, mfa, user_agent
FROM auth_session WHERE id_hash = $1;

-- name: TouchSession :exec
UPDATE auth_session SET last_used_at = $2 WHERE id_hash = $1;

-- name: DeleteSession :exec
DELETE FROM auth_session WHERE id_hash = $1;

-- name: ListSessionsByUser :many
SELECT public_id, user_id, created_at, last_used_at, expires_at, authenticated_at, mfa, user_agent
FROM auth_session WHERE user_id = $1 ORDER BY last_used_at DESC;

-- name: DeleteSessionByPublicID :execrows
DELETE FROM auth_session WHERE user_id = $1 AND public_id = $2;

-- name: DeleteSessionsByUser :exec
DELETE FROM auth_session WHERE user_id = $1;
