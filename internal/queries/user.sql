-- name: ListUsersPaginated :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
    LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, password, display_name, role_id)
VALUES ($1, $2, $3, $4)
    RETURNING *;