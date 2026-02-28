-- name: InsertBankAccount :one
INSERT INTO bank_accounts (bank_connection_id, external_id, iban, currency, name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, bank_connection_id, external_id, iban, currency, name, last_synced_at, created_at, updated_at, custom_name;

-- name: ListBankAccountsByConnection :many
SELECT id, bank_connection_id, external_id, iban, currency, name, last_synced_at, created_at, updated_at, custom_name
FROM bank_accounts
WHERE bank_connection_id = $1
ORDER BY created_at;

-- name: ListBankAccountsByWorkspace :many
SELECT ba.id, ba.bank_connection_id, ba.external_id, ba.iban, ba.currency, ba.name, ba.last_synced_at, ba.created_at, ba.updated_at, ba.custom_name,
       bc.institution_name
FROM bank_accounts ba
JOIN workspace_bank_accounts wba ON wba.bank_account_id = ba.id
JOIN bank_connections bc ON bc.id = ba.bank_connection_id
WHERE wba.workspace_id = $1
ORDER BY bc.institution_name, ba.name;

-- name: UpdateBankAccountLastSyncedAt :exec
UPDATE bank_accounts SET last_synced_at = $2, updated_at = NOW() WHERE id = $1;

-- name: LinkBankAccountToWorkspace :exec
INSERT INTO workspace_bank_accounts (workspace_id, bank_account_id)
VALUES ($1, $2)
ON CONFLICT (workspace_id, bank_account_id) DO NOTHING;

-- name: UnlinkBankAccountFromWorkspace :execresult
DELETE FROM workspace_bank_accounts WHERE workspace_id = $1 AND bank_account_id = $2;

-- name: ListWorkspacesByBankAccount :many
SELECT workspace_id FROM workspace_bank_accounts WHERE bank_account_id = $1;

-- name: GetBankAccountByID :one
SELECT id, bank_connection_id, external_id, iban, currency, name, last_synced_at, created_at, updated_at, custom_name
FROM bank_accounts
WHERE id = $1;

-- name: GetBankAccountByIBANAndCurrencyForUser :one
SELECT ba.id, ba.bank_connection_id, ba.external_id, ba.iban, ba.currency, ba.name, ba.last_synced_at, ba.created_at, ba.updated_at, ba.custom_name
FROM bank_accounts ba
JOIN bank_connections bc ON bc.id = ba.bank_connection_id
WHERE ba.iban = $1 AND ba.currency = $2 AND bc.user_id = $3
LIMIT 1;

-- name: UpdateBankAccountConnection :exec
UPDATE bank_accounts
SET external_id = $2, bank_connection_id = $3, updated_at = NOW()
WHERE id = $1;

-- name: ListBankAccountsByUser :many
SELECT ba.id, ba.bank_connection_id, ba.external_id, ba.iban, ba.currency, ba.name, ba.last_synced_at, ba.created_at, ba.updated_at, ba.custom_name,
       bc.institution_name
FROM bank_accounts ba
JOIN bank_connections bc ON bc.id = ba.bank_connection_id
WHERE bc.user_id = $1
ORDER BY bc.institution_name, ba.name;

-- name: UpdateBankAccountCustomName :one
UPDATE bank_accounts SET custom_name = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, bank_connection_id, external_id, iban, currency, name, last_synced_at, created_at, updated_at, custom_name;
