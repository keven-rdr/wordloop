-- name: FindRegistrationAccount :one
SELECT id, role, (email_verified_at IS NOT NULL)::boolean AS verified
FROM app_user
WHERE lower(email) = lower(sqlc.arg(email)::text) AND deleted_at IS NULL;

-- name: CreatePendingUser :one
INSERT INTO app_user (display_name, email, role)
VALUES ($1, $2, 'user')
RETURNING id;

-- name: CountRecentTokens :one
SELECT count(*)::int
FROM auth_token
WHERE user_id = $1 AND purpose IN ('verify_email', 'recovery') AND created_at > $2;

-- name: ReplaceToken :exec
WITH revoked AS (
    UPDATE auth_token SET expires_at = $4
    WHERE user_id = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > $4
)
INSERT INTO auth_token (id_hash, user_id, purpose, created_at, expires_at)
VALUES ($3, $1, $2, $4, $5);

-- name: PeekToken :one
SELECT t.user_id, t.purpose, u.email, u.role
FROM auth_token t JOIN app_user u ON u.id = t.user_id
WHERE t.id_hash = $1 AND t.purpose IN ('invite', 'verify_email', 'recovery')
  AND t.used_at IS NULL AND t.expires_at > $2 AND u.deleted_at IS NULL
  AND (t.purpose <> 'verify_email' OR u.email_verified_at IS NULL);

-- name: ConsumeToken :one
UPDATE auth_token SET used_at = $2
WHERE id_hash = $1 AND purpose IN ('invite', 'verify_email', 'recovery') AND used_at IS NULL AND expires_at > $2
  AND (purpose <> 'verify_email' OR EXISTS (
      SELECT 1 FROM app_user u WHERE u.id = auth_token.user_id AND u.email_verified_at IS NULL))
RETURNING user_id, purpose;

-- name: MarkEmailVerified :exec
UPDATE app_user SET email_verified_at = COALESCE(email_verified_at, $2) WHERE id = $1;
