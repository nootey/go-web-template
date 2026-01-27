package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"go-web-template/internal/config"
	"go-web-template/internal/database"
	"go-web-template/internal/handlers"
	"go-web-template/internal/services"
	"go-web-template/internal/store"
	"go-web-template/pkg/logging"
)

func main() {
	// Load config
	if err := config.Load(); err != nil {
		panic("failed to load config: " + err.Error())
	}
	cfg := config.Get()

	// Initialize logger
	logger := logging.InitLogger(cfg.App.Environment == "production", cfg.App.LogLevel)
	defer func() {
		_ = logger.Sync()
	}()
	logger = logger.Named("api")

	logger.Info("starting application",
		zap.String("environment", cfg.App.Environment),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database
	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("database connected")

	// Create SQLC queries instance
	queries := database.New(db)
	
	// Initialize services
	userService := services.NewUserService(queries, logger)
	// Add more services as needed

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, logger)
	// Add more handlers as needed

	// Setup router
	r := setupRouter(cfg, userHandler, logger)

	// Start server
	startServer(cfg, r, logger)
}

func setupRouter(cfg *config.Config, userHandler *handlers.UserHandler, logger *zap.Logger) *chi.Mux {
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

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Mount("/users", userHandler.Routes())
		// Mount more handlers as needed
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
