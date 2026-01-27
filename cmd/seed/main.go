package main

import (
	"context"
	"database/sql"
	"go-web-template/internal/store/seeders"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"go-web-template/internal/config"
	"go-web-template/internal/store"
	"go-web-template/pkg/logging"

	"go.uber.org/zap"
)

func main() {

	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	cfg := config.Get()

	logger := logging.InitLogger(cfg.App.Environment == "production", cfg.App.LogLevel)
	defer func(log *zap.Logger) {
		_ = log.Sync() // Ignore sync errors
	}(logger)
	logger = logger.Named("seeder")

	// Get seed type from args
	args := os.Args[1:]
	seedType := "help"
	if len(args) > 0 {
		seedType = args[0]
	}

	validTypes := map[string]bool{
		"core": true,
		"full": true,
	}

	if !validTypes[seedType] {
		logger.Fatal("invalid seed type", zap.String("type", seedType))
	}

	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	logger.Info("starting database seeding", zap.String("type", seedType))

	if err := seeders.SeedDatabase(ctx, db, logger, cfg, seedType); err != nil {
		logger.Fatal("seeding failed", zap.Error(err))
	}

	logger.Info("seeding completed successfully")
}
