package routes

import (
	"github.com/Abdelrahman-habib/expense-tracker/config"
	authService "github.com/Abdelrahman-habib/expense-tracker/internal/auth/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/users/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/users/repository"
	userService "github.com/Abdelrahman-habib/expense-tracker/internal/users/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Router struct {
	Handlers *handlers.UserHandler
}

func New(db db.Service, logger *zap.Logger, auth authService.Service, clerkConfig *config.ClerkConfig) *Router {
	// Initialize repository
	repo := repository.NewUsersRepository(db.Queries(), logger, nil)

	// Initialize service
	us := userService.NewUsersService(repo, logger)

	// Initialize handler
	handler := handlers.NewUserHandler(us, logger, clerkConfig, auth)

	return &Router{
		Handlers: handler,
	}
}

func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route("/users", func(router chi.Router) {
		router.Use(r.Handlers.WithUser)
		router.Get("/{id}", r.Handlers.GetUser)
		router.Get("/contacts", r.Handlers.GetUserContacts)
	})
}
