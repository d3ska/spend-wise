-- name: InsertWorkspace :one
INSERT INTO workspaces (name, description, owner_id)
VALUES ($1, $2, $3)
RETURNING id, name, description, owner_id, created_at, updated_at;

-- name: GetWorkspaceByID :one
SELECT id, name, description, owner_id, created_at, updated_at
FROM workspaces
WHERE id = $1;

-- name: ListWorkspacesByUser :many
SELECT w.id, w.name, w.description, w.owner_id, w.created_at, w.updated_at,
       (SELECT COUNT(*) FROM workspace_members wm2 WHERE wm2.workspace_id = w.id)::bigint AS member_count
FROM workspaces w
JOIN workspace_members wm ON w.id = wm.workspace_id
WHERE wm.user_id = $1
ORDER BY w.created_at DESC;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET name        = $2,
    description = $3,
    updated_at  = NOW()
WHERE id = $1
RETURNING id, name, description, owner_id, created_at, updated_at;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces WHERE id = $1;

-- name: InsertMember :one
INSERT INTO workspace_members (workspace_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING workspace_id, user_id, role, created_at;

-- name: GetMember :one
SELECT workspace_id, user_id, role, created_at
FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: ListMembers :many
SELECT workspace_id, user_id, role, created_at
FROM workspace_members
WHERE workspace_id = $1
ORDER BY created_at;

-- name: ListMembersWithProfiles :many
SELECT wm.workspace_id, wm.user_id, wm.role, wm.created_at,
       u.display_name, u.email, u.avatar_url
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.created_at;

-- name: UpdateMemberRole :exec
UPDATE workspace_members
SET role = $3
WHERE workspace_id = $1 AND user_id = $2;

-- name: DeleteMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;
