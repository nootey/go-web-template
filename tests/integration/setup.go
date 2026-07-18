package integration

import (
	"context"
	"database/sql"
	"go-web-template/internal/domains/auth"
	"go-web-template/internal/domains/user"
	"go-web-template/internal/tokens"
	"path/filepath"
	"time"

	"github.com/alicebob/miniredis/v2"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
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
	Redis     *miniredis.Miniredis
	Mailer    *CaptureMailer
	Services  *Services
}

type Services struct {
	UserService *user.UserService
	AuthService *auth.AuthService
	// Add more services here
}

// CaptureMailer records the last link sent, so tests can read the confirm/reset
// token that would otherwise only reach the user by email.
type CaptureMailer struct {
	ConfirmationLink string
	ResetLink        string
}

func (m *CaptureMailer) SendConfirmationEmail(to, name, link string) error {
	m.ConfirmationLink = link
	return nil
}

func (m *CaptureMailer) SendPasswordResetEmail(to, name, link string) error {
	m.ResetLink = link
	return nil
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
	cfg := testConfig()
	err = seeders.SeedDatabase(s.Ctx, db, logger, cfg, "core")
	s.Require().NoError(err, "seeding failed")

	// In-memory redis backs the token store.
	mr := miniredis.RunT(s.T())
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	tokenStore := tokens.NewStore(rdb, cfg.Token)
	captureMailer := &CaptureMailer{}

	// Initialize services
	userService := user.NewUserService(queries)
	authService := auth.NewAuthService(queries, tokenStore, captureMailer, cfg)

	s.TC = &TestContainer{
		Container: container,
		DB:        db,
		Queries:   queries,
		Redis:     mr,
		Mailer:    captureMailer,
		Services: &Services{
			UserService: userService,
			AuthService: authService,
		},
	}
}

// testConfig returns the minimal config the wired services need.
func testConfig() *config.Config {
	cfg := &config.Config{}
	cfg.App.Environment = "local"
	cfg.Server.Host = "localhost"
	cfg.Server.Port = "8080"
	cfg.WebClient.Domain = "localhost"
	cfg.WebClient.Port = "5173"
	cfg.Token.ConfirmTTLHours = 24
	cfg.Token.ResetTTLMinutes = 60
	return cfg
}

func (s *ServiceIntegrationSuite) SetupTest() {
	// Truncate users between tests (keep roles)
	_, err := s.TC.DB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	s.Require().NoError(err, "failed to truncate users table")

	// Clear tokens/sessions and the captured mail between tests.
	s.TC.Redis.FlushAll()
	s.TC.Mailer.ConfirmationLink = ""
	s.TC.Mailer.ResetLink = ""
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
