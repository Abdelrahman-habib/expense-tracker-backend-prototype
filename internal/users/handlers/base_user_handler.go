package handlers

import (
	"github.com/Abdelrahman-habib/expense-tracker/config"
	authService "github.com/Abdelrahman-habib/expense-tracker/internal/auth/service"
	h "github.com/Abdelrahman-habib/expense-tracker/internal/core/handlers"
	userService "github.com/Abdelrahman-habib/expense-tracker/internal/users/service"
	"go.uber.org/zap"
)

type UserHandler struct {
	h.BaseHandler
	service userService.UsersService
	logger  *zap.Logger
	clerk   *config.ClerkConfig
	auth    authService.Service
}

func NewUserHandler(service userService.UsersService, logger *zap.Logger, clerk *config.ClerkConfig, auth authService.Service) *UserHandler {
	return &UserHandler{
		BaseHandler: h.NewBaseHandler(logger),
		service:     service,
		logger:      logger,
		clerk:       clerk,
		auth:        auth,
	}
}
