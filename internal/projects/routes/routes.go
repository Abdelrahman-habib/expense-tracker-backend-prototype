package routes

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Router encapsulates the project routes setup
type Router struct {
	handler *handlers.ProjectHandler
}

// New creates a new project router with proper dependency injection
func New(dbService db.Service, logger *zap.Logger) *Router {
	// Get queries from db service
	queries := dbService.Queries()

	// Initialize repository
	repo := repository.NewProjectRepository(queries)

	// Initialize service with repository
	projectService := service.NewProjectService(repo, logger)

	// Initialize handler with service
	handler := handlers.NewProjectHandler(projectService, logger)

	return &Router{
		handler: handler,
	}
}

// RegisterRoutes registers all project routes
func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route("/projects", func(router chi.Router) {
		router.Get("/", r.handler.ListProjects)
		router.Get("/search", r.handler.SearchProjects)
		router.Get("/paginated", r.handler.ListProjectsPaginated)
		router.Post("/", r.handler.CreateProject)
		router.Route("/{id}", func(router chi.Router) {
			router.Get("/", r.handler.GetProject)
			router.Put("/", r.handler.UpdateProject)
			router.Delete("/", r.handler.DeleteProject)
			// router.Get("/wallets", r.handler.GetProjectWallets) // handled by wallets feature
		})
	})
}
