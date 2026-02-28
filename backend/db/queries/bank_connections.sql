-- name: InsertBankConnection :one
INSERT INTO bank_connections (user_id, institution_id, institution_name, session_id, status, auth_expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, institution_id, institution_name, session_id, status, auth_expires_at, created_at, updated_at;

-- name: GetBankConnectionByID :one
SELECT id, user_id, institution_id, institution_name, session_id, status, auth_expires_at, created_at, updated_at
FROM bank_connections
WHERE id = $1;

-- name: ListBankConnectionsByUser :many
SELECT id, user_id, institution_id, institution_name, session_id, status, auth_expires_at, created_at, updated_at
FROM bank_connections
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListActiveBankConnections :many
SELECT id, user_id, institution_id, institution_name, session_id, status, auth_expires_at, created_at, updated_at
FROM bank_connections
WHERE status = 'active' AND auth_expires_at > NOW();

-- name: UpdateBankConnectionStatus :exec
UPDATE bank_connections SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateBankConnectionSession :exec
UPDATE bank_connections
SET session_id = $2, status = $3, auth_expires_at = $4, updated_at = NOW()
WHERE id = $1;

-- name: DeleteBankConnection :execresult
DELETE FROM bank_connections WHERE id = $1 AND user_id = $2;
