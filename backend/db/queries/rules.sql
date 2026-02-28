-- name: GetRuleByID :one
SELECT id, scope, owner_id, workspace_id, match_pattern, target_category_id, priority, enabled, created_at, updated_at, amount_min, amount_max, bank_account_id, counterparty_iban
FROM categorization_rules
WHERE id = $1;

-- name: InsertRule :one
INSERT INTO categorization_rules (scope, owner_id, workspace_id, match_pattern, target_category_id, priority, enabled, amount_min, amount_max, bank_account_id, counterparty_iban)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, scope, owner_id, workspace_id, match_pattern, target_category_id, priority, enabled, created_at, updated_at, amount_min, amount_max, bank_account_id, counterparty_iban;

-- name: UpdateRule :one
UPDATE categorization_rules
SET match_pattern = $2, target_category_id = $3, priority = $4, amount_min = $5, amount_max = $6, bank_account_id = $7, counterparty_iban = $8, updated_at = NOW()
WHERE id = $1
RETURNING id, scope, owner_id, workspace_id, match_pattern, target_category_id, priority, enabled, created_at, updated_at, amount_min, amount_max, bank_account_id, counterparty_iban;

-- name: DeleteRule :exec
DELETE FROM categorization_rules WHERE id = $1;

-- name: ListRulesByWorkspace :many
SELECT id, scope, owner_id, workspace_id, match_pattern, target_category_id, priority, enabled, created_at, updated_at, amount_min, amount_max, bank_account_id, counterparty_iban
FROM categorization_rules
WHERE workspace_id = $1
ORDER BY priority DESC;

-- name: ToggleRuleEnabled :one
UPDATE categorization_rules
SET enabled = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, scope, owner_id, workspace_id, match_pattern, target_category_id, priority, enabled, created_at, updated_at, amount_min, amount_max, bank_account_id, counterparty_iban;

-- name: ListEnabledRulesForResolution :many
SELECT id, scope, owner_id, workspace_id, match_pattern, target_category_id, priority, enabled, created_at, updated_at, amount_min, amount_max, bank_account_id, counterparty_iban
FROM categorization_rules
WHERE (workspace_id = $1 OR owner_id = $2 OR scope = 'system')
  AND enabled = true
ORDER BY
  CASE scope WHEN 'workspace' THEN 3 WHEN 'user' THEN 2 WHEN 'system' THEN 1 ELSE 0 END DESC,
  priority DESC;
