package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/render"
)

// CreateTag godoc
// @Summary Create a new tag
// @Description Creates a new tag for the authenticated user
// @Tags Tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body types.TagCreatePayload true "Tag creation request"
// @Success 201 {object} payloads.Response{data=types.Tag}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags [post]
// @ID CreateTag
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	var req types.TagCreatePayload
	if err := render.Bind(r, &req); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	tag, err := h.service.CreateTag(r.Context(), userID, req)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}
	h.Respond(w, r, payloads.Created(tag))
}
