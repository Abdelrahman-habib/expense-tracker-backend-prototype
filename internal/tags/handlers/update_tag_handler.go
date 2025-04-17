package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

// UpdateTag godoc
// @Summary Update a tag
// @Description Updates an existing tag
// @Tags Tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tag ID" format(uuid)
// @Param request body types.TagUpdatePayload true "Tag update request"
// @Success 200 {object} payloads.Response{data=types.Tag}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags/{id} [put]
// @ID UpdateTag
func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	tagID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Get existing tag first
	existingTag, err := h.service.GetTag(r.Context(), userID, tagID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	// Create update payload from existing tag
	updatePayload := existingTag.ToUpdatePayload()

	// Use render.Bind to decode and validate
	if err := render.Bind(r, &updatePayload); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	tag, err := h.service.UpdateTag(r.Context(), userID, updatePayload)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Updated(tag))
}
