package handlers

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/service"
	"go.uber.org/zap"
)

type ProjectHandler struct {
	handlers.BaseHandler
	service service.ProjectService
}

func NewProjectHandler(service service.ProjectService, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		BaseHandler: handlers.NewBaseHandler(logger),
		service:     service,
	}
}
