package routes

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Router struct {
	handlers *handlers.AuthHandler
	service  service.Service
}

func New(db *db.Queries, logger *zap.Logger, config *types.Config) *Router {
	// Initialize repository
	repo := repository.NewAuthRepository(db, logger)

	// Initialize service
	svc := service.NewService(config, repo, logger)

	// Initialize handlers
	handler := handlers.NewAuthHandler(svc, logger)

	return &Router{
		handlers: handler,
		service:  svc,
	}
}

// GetService returns the auth service for middleware use
func (r *Router) GetService() service.Service {
	return r.service
}

func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route("/auth", func(router chi.Router) {
		router.Get("/{provider}", r.handlers.BeginAuthHandler)
		router.Get("/{provider}/callback", r.handlers.CallbackHandler)
		router.Post("/refresh", r.handlers.RefreshTokenHandler)
		router.Post("/logout", r.handlers.LogoutHandler)
	})
}
