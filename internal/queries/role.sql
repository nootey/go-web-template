-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: CreateRole :one
INSERT INTO roles (name, is_default, description)
VALUES ($1, $2, $3)
    RETURNING *;

-- name: GetPermissionByName :one
SELECT * FROM permissions WHERE name = $1;

-- name: CreatePermission :one
INSERT INTO permissions (name, description)
VALUES ($1, $2)
    RETURNING *;

-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
    ON CONFLICT DO NOTHING;