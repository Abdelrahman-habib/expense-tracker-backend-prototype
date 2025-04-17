package routes

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Router encapsulates the wallet routes setup
type Router struct {
	handler *handlers.WalletHandler
}

// New creates a new wallet router with proper dependency injection
func New(dbService db.Service, logger *zap.Logger) *Router {
	// Get queries from db service
	queries := dbService.Queries()

	// Initialize repository
	repo := repository.NewWalletRepository(queries)

	// Initialize service with repository
	walletService := service.NewWalletService(repo, logger)

	// Initialize handler with service
	handler := handlers.NewWalletHandler(walletService, logger)

	return &Router{
		handler: handler,
	}
}

// RegisterRoutes registers all wallet routes
func (r *Router) RegisterRoutes(router chi.Router) {
	router.Route("/wallets", func(router chi.Router) {
		router.Get("/search", r.handler.SearchWallets)
		router.Get("/paginated", r.handler.ListWalletsPaginated)
		router.Post("/", r.handler.CreateWallet)
		router.Route("/{id}", func(router chi.Router) {
			router.Get("/", r.handler.GetWallet)
			router.Put("/", r.handler.UpdateWallet)
			router.Delete("/", r.handler.DeleteWallet)
		})
	})
	router.Get("/projects/{id}/wallets", r.handler.GetProjectWallets)
}
