package routes

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Router encapsulates the contact routes setup
type Router struct {
	handler *handlers.ContactHandler
}

// New creates a new contact router with proper dependency injection
func New(dbService db.Service, logger *zap.Logger) *Router {
	// Get queries from db service
	queries := dbService.Queries()

	// Initialize repository
	repo := repository.New(queries)

	// Initialize service with repository
	contactservice := service.NewContactService(repo, logger)

	// Initialize handler with service
	handler := handlers.NewContactHandler(contactservice, logger)

	return &Router{
		handler: handler,
	}
}

// RegisterRoutes registers all contact routes
func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route("/contacts", func(router chi.Router) {
		router.Get("/", r.handler.ListContactsPaginated)
		router.Get("/paginated", r.handler.ListContactsPaginated)
		router.Get("/search", r.handler.SearchContacts)
		router.Post("/", r.handler.CreateContact)
		router.Route("/{id}", func(router chi.Router) {
			router.Get("/", r.handler.GetContact)
			router.Put("/", r.handler.UpdateContact)
			router.Delete("/", r.handler.DeleteContact)
		})
	})
}
