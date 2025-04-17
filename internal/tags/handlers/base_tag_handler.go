package handlers

import (
	h "github.com/Abdelrahman-habib/expense-tracker/internal/core/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/service"
	"go.uber.org/zap"
)

type TagHandler struct {
	h.BaseHandler
	service service.TagService
}

func NewTagHandler(service service.TagService, logger *zap.Logger) *TagHandler {
	return &TagHandler{
		BaseHandler: h.NewBaseHandler(logger),
		service:     service,
	}
}
