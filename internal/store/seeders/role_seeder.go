package seeders

import (
	"context"
	"database/sql"
	"errors"

	"go-web-template/internal/database"

	"go.uber.org/zap"
)

var defaultPermissions = []struct {
	name        string
	description string
}{
	{"system:superadmin", "Bypass all permission checks (root access)"},
	{"users:read", "Read user data"},
	{"users:write", "Create and update users"},
	{"users:delete", "Delete users"},
	// These are intentionally generic examples - customize for your domain
	{"data:read", "Read data"},
	{"data:write", "Create and update data"},
	{"data:delete", "Delete data"},
	{"admin:access", "Access admin panel"},
}

var defaultRoles = []struct {
	name        string
	isDefault   bool
	description string
	permissions []string
}{
	{
		name:        "super-admin",
		isDefault:   false,
		description: "Super administrator with unrestricted access",
		permissions: []string{
			"system:superadmin",
		},
	},
	{
		name:        "admin",
		isDefault:   false,
		description: "Administrator with full access",
		permissions: []string{
			"users:read", "users:write", "users:delete",
			"data:read", "data:write", "data:delete",
			"admin:access",
		},
	},
	{
		name:        "user",
		isDefault:   true,
		description: "Regular user",
		permissions: []string{
			"data:read", "data:write",
		},
	},
}

func SeedRolesAndPermissions(ctx context.Context, q *database.Queries, logger *zap.Logger) error {
	logger.Info("seeding roles and permissions")

	permissionIDs := make(map[string]int64)
	for _, p := range defaultPermissions {
		existing, err := q.GetPermissionByName(ctx, p.name)
		if err == nil {
			permissionIDs[p.name] = existing.ID
			logger.Debug("permission already exists", zap.String("name", p.name))
			continue
		}

		if errors.Is(err, sql.ErrNoRows) {
			created, err := q.CreatePermission(ctx, database.CreatePermissionParams{
				Name:        p.name,
				Description: sql.NullString{String: p.description, Valid: true},
			})
			if err != nil {
				return err
			}
			permissionIDs[p.name] = created.ID
			logger.Info("created permission", zap.String("name", p.name))
		} else {
			return err
		}
	}

	for _, r := range defaultRoles {
		existing, err := q.GetRoleByName(ctx, r.name)
		var roleID int64

		if err == nil {
			roleID = existing.ID
			logger.Debug("role already exists", zap.String("name", r.name))
		} else if errors.Is(err, sql.ErrNoRows) {
			desc := sql.NullString{String: r.description, Valid: true}
			created, err := q.CreateRole(ctx, database.CreateRoleParams{
				Name:        r.name,
				IsDefault:   r.isDefault,
				Description: desc,
			})
			if err != nil {
				return err
			}
			roleID = created.ID
			logger.Info("created role", zap.String("name", r.name))
		} else {
			return err
		}

		// Assign permissions to role
		for _, permName := range r.permissions {
			permID := permissionIDs[permName]
			err := q.AssignPermissionToRole(ctx, database.AssignPermissionToRoleParams{
				RoleID:       roleID,
				PermissionID: permID,
			})
			if err != nil {
				return err
			}
		}
	}

	logger.Info("roles and permissions seeded successfully")
	return nil
}
