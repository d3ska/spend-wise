-- name: GetUserByID :one
SELECT id, email, display_name, avatar_url, provider, provider_id, preferred_language, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, display_name, avatar_url, provider, provider_id, preferred_language, created_at, updated_at
FROM users
WHERE email = $1;

-- name: InsertUser :one
INSERT INTO users (email, display_name, avatar_url, provider, provider_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, display_name, avatar_url, provider, provider_id, preferred_language, created_at, updated_at;

-- name: UpdateUserProvider :one
UPDATE users
SET display_name = $2,
    avatar_url   = $3,
    provider     = $4,
    provider_id  = $5,
    updated_at   = NOW()
WHERE id = $1
RETURNING id, email, display_name, avatar_url, provider, provider_id, preferred_language, created_at, updated_at;

-- name: UpdatePreferredLanguage :one
UPDATE users
SET preferred_language = $2,
    updated_at         = NOW()
WHERE id = $1
RETURNING id, email, display_name, avatar_url, provider, provider_id, preferred_language, created_at, updated_at;
