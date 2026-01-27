package main

import (
	"database/sql"
	"fmt"
	"go-web-template/pkg/logging"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"go-web-template/internal/config"
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
	logger = logger.Named("migrate")

	dbURL := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Fatal("failed to set dialect:", zap.Error(err))
	}

	args := os.Args[1:]
	if len(args) == 0 {
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "up":
		if err := goose.Up(db, "migrations"); err != nil {
			logger.Fatal("migration up failed:", zap.Error(err))
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := goose.Down(db, "migrations"); err != nil {
			logger.Fatal("migration down failed:", zap.Error(err))
		}
		fmt.Println("Migration rolled back")

	case "status":
		if err := goose.Status(db, "migrations"); err != nil {
			logger.Fatal("migration status failed:", zap.Error(err))
		}

	case "create":
		if len(commandArgs) == 0 {
			logger.Fatal("migration name required: go run cmd/migrate/main.go create <name>")
		}
		if err := goose.Create(db, "migrations", commandArgs[0], "sql"); err != nil {
			logger.Fatal("migration create failed:", zap.Error(err))
		}
		fmt.Printf("Created migration: %s\n", commandArgs[0])

	case "reset":
		if err := goose.Reset(db, "migrations"); err != nil {
			logger.Fatal("migration reset failed:", zap.Error(err))
		}
		fmt.Println("All migrations rolled back")

	default:
		logger.Error("migration command not recognized", zap.String("command", command))
		os.Exit(1)
	}
}
