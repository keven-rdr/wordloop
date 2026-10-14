-- name: FindAccountByEmail :one
SELECT u.id, u.role,
       (EXISTS (SELECT 1 FROM auth_identity i WHERE i.user_id = u.id)
        OR EXISTS (SELECT 1 FROM passkey_credential p WHERE p.user_id = u.id))::boolean AS has_credentials
FROM app_user u
WHERE lower(u.email) = lower(sqlc.arg(email)::text) AND u.deleted_at IS NULL;

-- name: CreateAdminUser :one
INSERT INTO app_user (display_name, email, email_verified_at, role)
VALUES ($1, $2, $3, 'admin')
RETURNING id;

-- name: ReplaceInvite :exec
WITH revoked AS (
    DELETE FROM auth_token WHERE user_id = $1 AND purpose = 'invite' AND used_at IS NULL
)
INSERT INTO auth_token (id_hash, user_id, purpose, created_at, expires_at)
VALUES ($2, $1, 'invite', $3, $4);
