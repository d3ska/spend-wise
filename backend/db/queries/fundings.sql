-- name: UpsertFunding :one
INSERT INTO fundings (workspace_id, user_id, year_month, amount, currency, category_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (workspace_id, user_id, year_month, COALESCE(category_id, 0))
DO UPDATE SET amount = EXCLUDED.amount, currency = EXCLUDED.currency, updated_at = NOW()
RETURNING id, workspace_id, user_id, year_month, amount, currency, category_id, created_at, updated_at;

-- name: ListFundingsByWorkspace :many
SELECT id, workspace_id, user_id, year_month, amount, currency, category_id, created_at, updated_at
FROM fundings
WHERE workspace_id = $1 AND year_month >= $2 AND year_month <= $3
ORDER BY year_month, user_id;

-- name: GetFundingsByWorkspaceAndUsers :many
SELECT id, workspace_id, user_id, year_month, amount, currency, category_id, created_at, updated_at
FROM fundings
WHERE workspace_id = $1 AND year_month >= $2 AND year_month <= $3
ORDER BY user_id, year_month;

-- name: DeleteFunding :execresult
DELETE FROM fundings
WHERE id = $1 AND workspace_id = $2;

-- name: GetOverallBudgetByWorkspace :many
SELECT id, workspace_id, user_id, year_month, amount, currency, category_id, created_at, updated_at
FROM fundings
WHERE workspace_id = $1 AND year_month >= $2 AND year_month <= $3 AND category_id IS NULL
ORDER BY year_month;

-- name: GetCategoryBudgetsByWorkspace :many
SELECT id, workspace_id, user_id, year_month, amount, currency, category_id, created_at, updated_at
FROM fundings
WHERE workspace_id = $1 AND year_month >= $2 AND year_month <= $3 AND category_id IS NOT NULL
ORDER BY category_id, year_month;
