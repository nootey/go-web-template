package integration

import (
	"context"
	"database/sql"
	"go-web-template/internal/domains/user"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"go-web-template/internal/config"
	"go-web-template/internal/database"
	"go-web-template/internal/store/seeders"
)

type TestContainer struct {
	Container *postgres.PostgresContainer
	DB        *sql.DB
	Queries   *database.Queries
	Services  *Services
}

type Services struct {
	UserService *user.UserService
	// Add more services here
}

type ServiceIntegrationSuite struct {
	suite.Suite
	TC  *TestContainer
	Ctx context.Context
}

func (s *ServiceIntegrationSuite) SetupSuite() {
	s.Ctx = context.Background()

	// Start PostgreSQL container
	container, err := postgres.Run(s.Ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err, "failed to start container")

	// Connect to container
	connStr, err := container.ConnectionString(s.Ctx, "sslmode=disable")
	s.Require().NoError(err, "failed to get connection string")

	db, err := sql.Open("postgres", connStr)
	s.Require().NoError(err, "failed to connect to database")

	// Run migrations
	err = goose.SetDialect("postgres")
	s.Require().NoError(err, "failed to set goose dialect")

	migrationsPath := filepath.Join("..", "..", "..", "migrations")
	err = goose.Up(db, migrationsPath)
	s.Require().NoError(err, "migrations failed")

	// Create queries
	queries := database.New(db)

	// Seed basic data (roles, permissions)
	logger := zap.NewNop()
	cfg := &config.Config{} // empty config for tests
	err = seeders.SeedDatabase(s.Ctx, db, logger, cfg, "core")
	s.Require().NoError(err, "seeding failed")

	// Initialize services
	userService := user.NewUserService(queries)

	s.TC = &TestContainer{
		Container: container,
		DB:        db,
		Queries:   queries,
		Services: &Services{
			UserService: userService,
		},
	}
}

func (s *ServiceIntegrationSuite) SetupTest() {
	// Truncate users between tests (keep roles)
	_, err := s.TC.DB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	s.Require().NoError(err, "failed to truncate users table")
}

func (s *ServiceIntegrationSuite) TearDownSuite() {
	if s.TC.DB != nil {
		_ = s.TC.DB.Close()
	}

	if s.TC.Container != nil {
		if err := s.TC.Container.Terminate(s.Ctx); err != nil {
			s.T().Logf("container cleanup warning: %s", err)
		}
	}
}
