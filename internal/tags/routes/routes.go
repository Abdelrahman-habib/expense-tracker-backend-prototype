package routes

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Router encapsulates the tag routes setup
type Router struct {
	handler *handlers.TagHandler
}

// New creates a new tag router with proper dependency injection
func New(dbService db.Service, logger *zap.Logger) *Router {
	// Get queries from db service
	queries := dbService.Queries()

	// Initialize repository
	repo := repository.NewTagRepository(queries)

	// Initialize service with repository
	tagService := service.NewTagService(repo, logger)

	// Initialize handler with service
	handler := handlers.NewTagHandler(tagService, logger)

	return &Router{
		handler: handler,
	}
}

// RegisterRoutes registers all tag routes
func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route("/tags", func(router chi.Router) {
		router.Get("/", r.handler.ListTags)
		router.Post("/", r.handler.CreateTag)
		router.Delete("/", r.handler.DeleteUserTags)

		router.Route("/{id}", func(router chi.Router) {
			router.Get("/", r.handler.GetTag)
			router.Put("/", r.handler.UpdateTag)
			router.Delete("/", r.handler.DeleteTag)
		})
	})
}
