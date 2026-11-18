-- name: ListUsersPage :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND (sqlc.narg('after_created_at')::timestamp IS NULL
    OR (created_at, id) < (sqlc.narg('after_created_at')::timestamp, sqlc.narg('after_id')::bigint))
ORDER BY created_at DESC, id DESC
    LIMIT sqlc.arg('limit');

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, password, display_name, role_id)
VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetDefaultRole :one
SELECT * FROM roles
WHERE is_default = true
    LIMIT 1;

-- name: ConfirmUserEmail :exec
UPDATE users SET email_confirmed = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserPassword :exec
UPDATE users SET password = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;