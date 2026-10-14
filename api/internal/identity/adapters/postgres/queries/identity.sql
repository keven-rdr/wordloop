-- name: UpsertUser :one
INSERT INTO app_user (idp_issuer, idp_subject, display_name, last_seen_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (idp_issuer, idp_subject) DO UPDATE
SET last_seen_at = now(),
    display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), app_user.display_name)
RETURNING id, display_name, role;

-- name: GetUser :one
SELECT id, display_name, role FROM app_user WHERE id = $1 AND deleted_at IS NULL;

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
