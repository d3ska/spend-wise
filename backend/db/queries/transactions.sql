-- name: InsertTransaction :one
INSERT INTO transactions (workspace_id, created_by, total_amount, currency, description, date, fingerprint, source, transaction_type, notes, bank_account_id, iban, counterparty_iban)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, workspace_id, created_by, total_amount, currency, description, date, fingerprint, source, transaction_type, notes, bank_account_id, iban, counterparty_iban, created_at, updated_at;

-- name: GetTransactionByID :one
SELECT id, workspace_id, created_by, total_amount, currency, description, date, fingerprint, source, transaction_type, notes, bank_account_id, iban, counterparty_iban, created_at, updated_at
FROM transactions
WHERE id = $1;

-- name: ListTransactionsByWorkspace :many
SELECT t.id, t.workspace_id, t.created_by, t.total_amount, t.currency, t.description, t.date, t.fingerprint, t.source, t.transaction_type, t.notes, t.bank_account_id, t.iban, t.counterparty_iban, t.created_at, t.updated_at,
       COALESCE(bc.institution_name, '') AS bank_name
FROM transactions t
LEFT JOIN bank_accounts ba ON ba.id = t.bank_account_id
LEFT JOIN bank_connections bc ON bc.id = ba.bank_connection_id
WHERE t.workspace_id = $1 AND t.date >= $2 AND t.date < $3
  AND (sqlc.narg('transaction_type')::transaction_type IS NULL OR t.transaction_type = sqlc.narg('transaction_type'))
  AND (sqlc.narg('currency')::text IS NULL OR t.currency = sqlc.narg('currency'))
  AND (sqlc.narg('category_id')::bigint IS NULL OR EXISTS (
    SELECT 1 FROM entries e WHERE e.transaction_id = t.id AND e.category_id = sqlc.narg('category_id')
  ))
  AND (sqlc.narg('bank_account_id')::bigint IS NULL OR t.bank_account_id = sqlc.narg('bank_account_id'))
  AND (sqlc.narg('amount_min')::numeric IS NULL OR ABS(t.total_amount) >= sqlc.narg('amount_min'))
  AND (sqlc.narg('amount_max')::numeric IS NULL OR ABS(t.total_amount) <= sqlc.narg('amount_max'))
ORDER BY t.date DESC, t.id DESC
LIMIT $4 OFFSET $5;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1;

-- name: UpdateTransaction :one
UPDATE transactions
SET description = $2, date = $3, transaction_type = $4, notes = $5, updated_at = NOW()
WHERE id = $1
RETURNING id, workspace_id, created_by, total_amount, currency, description, date, fingerprint, source, transaction_type, notes, bank_account_id, iban, counterparty_iban, created_at, updated_at;

-- name: ListAllTransactionEntries :many
SELECT t.id, t.description, t.created_by, e.id AS entry_id, e.category_id, t.transaction_type, t.total_amount, t.bank_account_id, t.counterparty_iban
FROM transactions t
JOIN entries e ON e.transaction_id = t.id
WHERE t.workspace_id = $1
ORDER BY t.id;

-- name: ExistsByFingerprint :one
SELECT EXISTS (
    SELECT 1 FROM transactions
    WHERE workspace_id = $1 AND fingerprint = $2
) AS exists;

-- name: BulkDeleteTransactions :execresult
DELETE FROM transactions
WHERE workspace_id = $1 AND id = ANY($2::bigint[]);

-- name: GetTransactionsByIDs :many
SELECT id, workspace_id, created_by, total_amount, currency, description, date, fingerprint, source, transaction_type, notes, bank_account_id, iban, counterparty_iban, created_at, updated_at
FROM transactions
WHERE workspace_id = $1 AND id = ANY($2::bigint[]);

-- name: RelinkOrphanedTransactions :execresult
UPDATE transactions
SET bank_account_id = $1, updated_at = NOW()
WHERE iban = $2 AND currency = $3 AND bank_account_id IS NULL AND workspace_id = $4;
