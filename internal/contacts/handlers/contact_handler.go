package handlers

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/handlers"
	"go.uber.org/zap"
)

type ContactHandler struct {
	handlers.BaseHandler
	service service.ContactService
}

func NewContactHandler(service service.ContactService, logger *zap.Logger) *ContactHandler {
	return &ContactHandler{
		BaseHandler: handlers.NewBaseHandler(logger),
		service:     service,
	}
}
