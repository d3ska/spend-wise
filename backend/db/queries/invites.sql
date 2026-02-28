-- name: InsertInvite :one
INSERT INTO workspace_invites (workspace_id, code, role, created_by, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, workspace_id, code, role, created_by, expires_at, used_by, used_at, created_at;

-- name: GetInviteByCode :one
SELECT
    wi.id, wi.workspace_id, wi.code, wi.role, wi.created_by,
    wi.expires_at, wi.used_by, wi.used_at, wi.created_at,
    w.name AS workspace_name,
    u.display_name AS inviter_name
FROM workspace_invites wi
JOIN workspaces w ON w.id = wi.workspace_id
JOIN users u ON u.id = wi.created_by
WHERE wi.code = $1;

-- name: MarkInviteUsed :exec
UPDATE workspace_invites
SET used_by = $2, used_at = NOW()
WHERE id = $1;
