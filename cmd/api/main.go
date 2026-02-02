package main

import (
	"context"
	"database/sql"
	"errors"
	"go-web-template/internal/domains/auth"
	"go-web-template/internal/domains/user"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mWare "go-web-template/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"go-web-template/internal/config"
	"go-web-template/internal/database"
	"go-web-template/internal/store"
	"go-web-template/pkg/logging"
)

type Handlers struct {
	Auth *auth.AuthHandler
	User *user.UserHandler
}

func main() {

	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	cfg := config.Get()

	logger := logging.InitLogger(cfg.App.Environment == "production", cfg.App.LogLevel)
	defer func() {
		_ = logger.Sync()
	}()
	logger = logger.Named("api")

	logger.Info("starting application",
		zap.String("environment", cfg.App.Environment),
		zap.String("port", cfg.Server.Port),
	)

	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatal("failed to close database connection", zap.Error(err))
		}
	}(db)
	logger.Info("database connected")

	// Create SQLC queries instance
	queries := database.New(db)

	// Initialize auth middleware
	authMiddleware := mWare.NewAuthMiddleware(
		cfg,
		logger,
		cfg.Auth.AccessTTL,
		cfg.Auth.RefreshTTLShort,
		cfg.Auth.RefreshTTLLong,
	)

	// Initialize services
	userService := user.NewUserService(queries)
	authService := auth.NewAuthService(queries)
	// Add more services as needed

	// Initialize handlers
	userHandler := user.NewUserHandler(userService)
	authHandler := auth.NewAuthHandler(authService, authMiddleware)
	// Add more handlers as needed

	h := Handlers{
		Auth: authHandler,
		User: userHandler,
	}

	r := setupRouter(cfg, &h, authMiddleware, logger)

	startServer(cfg, r, logger)
}

func setupRouter(cfg *config.Config, h *Handlers, authMiddleware *mWare.AuthMiddleware, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			logger.Fatal("failed to write health response", zap.Error(err))
		}
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Mount("/auth", h.Auth.Routes())

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.WebClientAuthentication)

			r.Mount("/users", h.User.Routes())
			// Mount more protected handlers as needed
		})
	})

	logger.Info("router configured")
	return r
}

func startServer(cfg *config.Config, handler http.Handler, logger *zap.Logger) {
	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("server starting",
			zap.String("address", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped gracefully")
}
