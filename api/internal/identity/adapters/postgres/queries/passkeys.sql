-- name: ListPasskeysByUser :many
SELECT id, user_id, credential_id, public_key, sign_count, transports, user_verified, backup_eligible, backup_state,
       aaguid, name, created_at, last_used_at
FROM passkey_credential WHERE user_id = $1 ORDER BY created_at;

-- name: GetPasskeyWithOwner :one
SELECT p.id, p.user_id, p.credential_id, p.public_key, p.sign_count, p.transports, p.user_verified, p.backup_eligible,
       p.backup_state, p.aaguid, p.name, p.created_at, p.last_used_at, u.email, u.display_name
FROM passkey_credential p
JOIN app_user u ON u.id = p.user_id AND u.deleted_at IS NULL
WHERE p.credential_id = $1;

-- name: InsertPasskey :exec
INSERT INTO passkey_credential (credential_id, user_id, public_key, sign_count, transports, user_verified,
                                backup_eligible, backup_state, aaguid, name, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: CountPasskeysByUser :one
SELECT count(*) FROM passkey_credential WHERE user_id = $1;

-- name: CountIdentitiesByUser :one
SELECT count(*) FROM auth_identity WHERE user_id = $1;

-- name: RecordPasskeyLogin :exec
UPDATE passkey_credential
SET sign_count = $2, user_verified = $3, backup_state = $4, last_used_at = $5
WHERE credential_id = $1;

-- name: DeletePasskey :execrows
DELETE FROM passkey_credential WHERE user_id = $1 AND id = $2;

-- name: LockUser :one
SELECT id FROM app_user WHERE id = $1 FOR UPDATE;

-- name: GetOwner :one
SELECT id, email, display_name FROM app_user WHERE id = $1 AND deleted_at IS NULL;

-- name: PeekInvite :one
SELECT user_id FROM auth_token
WHERE id_hash = $1 AND purpose = 'invite' AND used_at IS NULL AND expires_at > $2;

-- name: ConsumeInvite :one
UPDATE auth_token SET used_at = $2
WHERE id_hash = $1 AND purpose = 'invite' AND used_at IS NULL AND expires_at > $2
RETURNING user_id;

-- name: SaveChallenge :exec
INSERT INTO auth_challenge (id_hash, purpose, user_id, invite_hash, state, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: TakeChallenge :one
DELETE FROM auth_challenge
WHERE id_hash = $1 AND purpose = $2 AND expires_at > $3
RETURNING user_id, invite_hash, state;

-- name: PurgeChallenges :exec
DELETE FROM auth_challenge WHERE expires_at <= $1;
