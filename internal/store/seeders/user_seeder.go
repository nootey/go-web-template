package seeders

import (
	"context"
	"database/sql"
	"errors"
	"go-web-template/internal/config"
	"go-web-template/internal/database"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func SeedRootUser(ctx context.Context, q *database.Queries, logger *zap.Logger, cfg *config.Config) error {

	logger.Info("seeding root user")

	// Check if root user already exists
	_, err := q.GetUserByEmail(ctx, cfg.Seed.RootUser)
	if err == nil {
		logger.Info("root user already exists, skipping")
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	adminRole, err := q.GetRoleByName(ctx, "super-admin")
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(cfg.Seed.RootPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	user, err := q.CreateUser(ctx, database.CreateUserParams{
		Email:       cfg.Seed.RootUser,
		Password:    string(hashedPassword),
		DisplayName: "Root Administrator",
		RoleID:      adminRole.ID,
	})
	if err != nil {
		return err
	}

	logger.Info("root user created",
		zap.String("email", user.Email),
		zap.Int64("id", user.ID),
		zap.String("role", "super-admin"),
	)

	return nil
}
