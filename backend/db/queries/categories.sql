-- name: InsertCategory :one
INSERT INTO categories (workspace_id, name, icon, slug)
VALUES ($1, $2, $3, $4)
RETURNING id, workspace_id, name, icon, slug, created_at, updated_at;

-- name: GetCategoryByID :one
SELECT id, workspace_id, name, icon, slug, created_at, updated_at
FROM categories
WHERE id = $1;

-- name: ListCategoriesByWorkspace :many
SELECT id, workspace_id, name, icon, slug, created_at, updated_at
FROM categories
WHERE workspace_id = $1
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET name       = $2,
    icon       = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, workspace_id, name, icon, slug, created_at, updated_at;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: GetCategoryByName :one
SELECT id, workspace_id, name, icon, slug, created_at, updated_at
FROM categories
WHERE workspace_id = $1 AND name = $2;

-- name: GetCategoryBySlug :one
SELECT id, workspace_id, name, icon, slug, created_at, updated_at
FROM categories
WHERE workspace_id = $1 AND slug = $2;
