package server

import (
	"fmt"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/config"
	authRoutes "github.com/Abdelrahman-habib/expense-tracker/internal/auth/routes"
	contactRoutes "github.com/Abdelrahman-habib/expense-tracker/internal/contacts/routes"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	projectRoutes "github.com/Abdelrahman-habib/expense-tracker/internal/projects/routes"
	"github.com/Abdelrahman-habib/expense-tracker/internal/server/middleware"
	tagRoutes "github.com/Abdelrahman-habib/expense-tracker/internal/tags/routes"
	userRoutes "github.com/Abdelrahman-habib/expense-tracker/internal/users/routes"
	walletRoutes "github.com/Abdelrahman-habib/expense-tracker/internal/wallets/routes"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type APIServer struct {
	config        *config.Config
	db            db.Service
	logger        *zap.Logger
	middleware    *middleware.Middleware
	authRoutes    *authRoutes.Router
	tagRoutes     *tagRoutes.Router
	userRoutes    *userRoutes.Router
	projectRoutes *projectRoutes.Router
	walletRoutes  *walletRoutes.Router
	contactRoutes *contactRoutes.Router
}

type ServerDependencies struct {
	Config *config.Config
	DB     db.Service
	Logger *zap.Logger
}

func NewAPIServer(deps ServerDependencies) *APIServer {
	// Create server instance
	server := &APIServer{
		config:        deps.Config,
		db:            deps.DB,
		logger:        deps.Logger,
		authRoutes:    authRoutes.New(deps.DB.Queries(), deps.Logger, &deps.Config.Auth),
		userRoutes:    userRoutes.New(deps.DB, deps.Logger, nil, &deps.Config.Clerk),
		tagRoutes:     tagRoutes.New(deps.DB, deps.Logger),
		projectRoutes: projectRoutes.New(deps.DB, deps.Logger),
		walletRoutes:  walletRoutes.New(deps.DB, deps.Logger),
		contactRoutes: contactRoutes.New(deps.DB, deps.Logger),
	}

	// Initialize middleware after auth service is created
	server.middleware = middleware.NewMiddleware(deps.Logger, server.authRoutes.GetService(), deps.DB, deps.Config.Server, nil)

	return server
}

// NewHTTPServer creates and returns a configured http.Server
func (s *APIServer) NewHTTPServer() *http.Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  s.config.Server.IdleTimeout,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	s.logger.Info("server initialized",
		zap.Int("port", s.config.Server.Port),
		zap.Duration("idle_timeout", s.config.Server.IdleTimeout),
		zap.Duration("read_timeout", s.config.Server.ReadTimeout),
		zap.Duration("write_timeout", s.config.Server.WriteTimeout),
	)

	return server
}

func (s *APIServer) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(s.middleware.Timeout(s.config.Server.RequestTimeout))
	r.Use(s.middleware.Recovery)
	r.Use(s.middleware.Logger)
	r.Use(s.middleware.CORS())
	r.Use(s.middleware.RateLimiter)

	// Public routes
	r.Group(func(r chi.Router) {
		s.logger.Debug("registering public routes")
		// Register auth routes
		s.authRoutes.RegisterRoutes(r)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		s.logger.Debug("registering protected routes")
		r.Use(s.middleware.Authenticate)
		r.Route("/api/v1", func(r chi.Router) {
			// User routes
			s.userRoutes.RegisterRoutes(r)
			// Register tag routes
			s.tagRoutes.RegisterRoutes(r)
			// Register project routes
			s.projectRoutes.RegisterRoutes(r)
			// Register wallet Routes
			s.walletRoutes.RegisterRoutes(r)
			// Register contact Routes
			s.contactRoutes.RegisterRoutes(r)
		})
	})

	s.logger.Info("routes registered successfully")
	return r
}
