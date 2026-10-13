-- name: UpsertUser :one
INSERT INTO app_user (idp_issuer, idp_subject, display_name, last_seen_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (idp_issuer, idp_subject) DO UPDATE
SET last_seen_at = now(),
    display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), app_user.display_name)
RETURNING id, display_name;

-- name: GetUser :one
SELECT id, display_name FROM app_user WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateSession :exec
INSERT INTO auth_session (id_hash, user_id, refresh_token_enc, expires_at, user_agent)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSession :one
SELECT user_id, expires_at FROM auth_session WHERE id_hash = $1;

-- name: TouchSession :exec
UPDATE auth_session SET last_used_at = now() WHERE id_hash = $1;

-- name: DeleteSession :exec
DELETE FROM auth_session WHERE id_hash = $1;
