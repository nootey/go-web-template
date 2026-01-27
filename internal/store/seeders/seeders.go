package seeders

import (
	"context"
	"database/sql"
	"go-web-template/internal/database"

	"go-web-template/internal/config"

	"go.uber.org/zap"
)

func SeedDatabase(ctx context.Context, db *sql.DB, logger *zap.Logger, cfg *config.Config, seedType string) error {
	queries := database.New(db)

	switch seedType {
	case "core":
		return seedCore(ctx, queries, logger, cfg)
	case "full":
		return seedFull(ctx, queries, logger, cfg)
	default:
		return nil
	}
}

func seedCore(ctx context.Context, q *database.Queries, logger *zap.Logger, cfg *config.Config) error {
	logger.Info("seeding core data")

	if err := SeedRolesAndPermissions(ctx, q, logger); err != nil {
		return err
	}

	if cfg.Seed.RootUser != "" && cfg.Seed.RootPassword != "" {
		if err := SeedRootUser(ctx, q, logger, cfg); err != nil {
			return err
		}
	} else {
		logger.Info("skipping root user seed - credentials not provided")
	}

	return nil
}

func seedFull(ctx context.Context, q *database.Queries, logger *zap.Logger, cfg *config.Config) error {
	logger.Info("seeding full data (dev mode)")

	if err := seedCore(ctx, q, logger, cfg); err != nil {
		return err
	}

	// Add more as needed...

	return nil
}
