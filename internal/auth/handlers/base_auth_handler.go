package handlers

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/service"
	h "github.com/Abdelrahman-habib/expense-tracker/internal/core/handlers"
	"go.uber.org/zap"
)

type AuthHandler struct {
	h.BaseHandler
	service service.Service
	logger  *zap.Logger
}

func NewAuthHandler(service service.Service, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		BaseHandler: h.NewBaseHandler(logger),
		service:     service,
		logger:      logger,
	}
}
