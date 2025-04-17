package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/config"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/server"
	"github.com/Abdelrahman-habib/expense-tracker/internal/server/lifecycle"
	"go.uber.org/zap"
)

// App represents the application and its dependencies
type App struct {
	config     *config.Config
	logger     *zap.Logger
	db         db.Service
	httpServer *http.Server
}

// New creates a new application instance
func New(cfg *config.Config) (*App, error) {
	// Initialize logger
	logger := zap.Must(zap.NewProduction())
	if cfg.Logger.Environment == "development" {
		logger = zap.Must(zap.NewDevelopment())
	}

	// Initialize database
	dbService := db.NewService(cfg.Database)

	// Create API server
	apiServer := server.NewAPIServer(server.ServerDependencies{
		Config: cfg,
		DB:     dbService,
		Logger: logger,
	})

	// Create HTTP server
	httpServer := apiServer.NewHTTPServer()

	return &App{
		config:     cfg,
		logger:     logger,
		db:         dbService,
		httpServer: httpServer,
	}, nil
}

// Start starts the application
func (a *App) Start() error {
	// Start server with graceful shutdown
	done := lifecycle.GracefulShutdown(a.httpServer, a.logger)

	a.logger.Info("starting server", zap.String("addr", a.httpServer.Addr))
	if err := a.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	a.logger.Info("server shutdown complete")
	return nil
}

// Stop stops the application
func (a *App) Stop(ctx context.Context) error {
	// Stop HTTP server
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	// Close database connections
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("error closing database: %w", err)
	}

	// Sync logger
	if err := a.logger.Sync(); err != nil {
		return fmt.Errorf("error syncing logger: %w", err)
	}

	return nil
}
