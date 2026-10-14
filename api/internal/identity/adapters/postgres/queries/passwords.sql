-- name: FindPasswordAccount :one
SELECT id, password_hash, locked_until
FROM app_user
WHERE lower(email) = lower(sqlc.arg(email)::text) AND deleted_at IS NULL;

-- name: BeginPasswordAttempt :one
UPDATE app_user SET
    failed_logins  = CASE WHEN last_failed_at IS NULL OR last_failed_at < sqlc.arg(at)::timestamptz - interval '15 minutes' THEN 1 ELSE failed_logins + 1 END,
    last_failed_at = sqlc.arg(at)::timestamptz,
    locked_until   = CASE WHEN (CASE WHEN last_failed_at IS NULL OR last_failed_at < sqlc.arg(at)::timestamptz - interval '15 minutes' THEN 1 ELSE failed_logins + 1 END) >= 5
        THEN sqlc.arg(at)::timestamptz + LEAST(interval '15 minutes', interval '30 seconds' * power(2, LEAST((CASE WHEN last_failed_at IS NULL OR last_failed_at < sqlc.arg(at)::timestamptz - interval '15 minutes' THEN 1 ELSE failed_logins + 1 END) - 5, 10)))
    END
WHERE id = sqlc.arg(id) AND (locked_until IS NULL OR locked_until <= sqlc.arg(at)::timestamptz)
RETURNING id;

-- name: ResetPasswordFailures :exec
UPDATE app_user SET failed_logins = 0, last_failed_at = NULL, locked_until = NULL WHERE id = $1;

-- name: PeekInviteOwner :one
SELECT t.user_id, u.email
FROM auth_token t JOIN app_user u ON u.id = t.user_id
WHERE t.id_hash = $1 AND t.purpose = 'invite' AND t.used_at IS NULL AND t.expires_at > $2 AND u.deleted_at IS NULL;

-- name: SetPasswordHash :exec
UPDATE app_user
SET password_hash = $2, failed_logins = 0, last_failed_at = NULL, locked_until = NULL
WHERE id = $1;
