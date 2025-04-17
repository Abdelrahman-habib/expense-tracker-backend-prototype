package handlers

import (
	"fmt"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	logger *zap.Logger
}

func NewBaseHandler(logger *zap.Logger) BaseHandler {
	return BaseHandler{
		logger: logger,
	}
}

// Respond is a helper function to handle all responses
func (h *BaseHandler) Respond(w http.ResponseWriter, r *http.Request, renderer render.Renderer) {
	if err := render.Render(w, r, renderer); err != nil {
		h.logger.Error("failed to render response", zap.Error(err))
		render.Render(w, r, errors.ErrRender(err))
	}
}

// RespondError is a helper function to handle all error responses
func (h *BaseHandler) RespondError(w http.ResponseWriter, r *http.Request, err interface{}) {
	if renderer, ok := err.(render.Renderer); ok {
		if typedErr, ok := renderer.(*errors.ErrorResponse); ok {
			h.logger.Error("handler error",
				zap.String("type", string(typedErr.Type)),
				zap.String("message", typedErr.Message),
				zap.Error(typedErr.Err),
			)
		}
		render.Render(w, r, renderer)
		return
	}

	if stdErr, ok := err.(error); ok {
		h.logger.Error("unexpected error", zap.Error(stdErr))
		render.Render(w, r, errors.ErrInternal(stdErr))
		return
	}

	h.logger.Error("unexpected error type", zap.Any("error", err))
	render.Render(w, r, errors.ErrInternal(fmt.Errorf("unexpected error type: %v", err)))
}

func (h *BaseHandler) HandleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.IsErrorType(err, errors.ErrorTypeNotFound) {
		h.RespondError(w, r, errors.ErrNotFound())
		return
	}
	h.RespondError(w, r, errors.ErrDatabase(err))
}
