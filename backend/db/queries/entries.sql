-- name: InsertEntry :one
INSERT INTO entries (transaction_id, category_id, participant_id, amount, currency, note)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, transaction_id, category_id, participant_id, amount, currency, note;

-- name: ListEntriesByTransaction :many
SELECT id, transaction_id, category_id, participant_id, amount, currency, note
FROM entries
WHERE transaction_id = $1
ORDER BY id;

-- name: UpdateEntryCategory :exec
UPDATE entries SET category_id = $2 WHERE id = $1;

-- name: ListEntriesByTransactionIDs :many
SELECT id, transaction_id, category_id, participant_id, amount, currency, note
FROM entries
WHERE transaction_id = ANY($1::bigint[])
ORDER BY transaction_id, id;

-- name: BulkUpdateFirstEntryCategory :execresult
UPDATE entries SET category_id = $1
WHERE id IN (
    SELECT DISTINCT ON (transaction_id) id
    FROM entries
    WHERE transaction_id = ANY($2::bigint[])
    ORDER BY transaction_id, id
);

-- name: DeleteEntriesByTransaction :exec
DELETE FROM entries WHERE transaction_id = $1;
