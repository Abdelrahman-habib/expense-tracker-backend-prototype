package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
)

// DeleteUserTags godoc
// @Summary Delete all user tags
// @Description Deletes all tags for the authenticated user
// @Tags Tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} payloads.Response
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags [delete]
// @ID DeleteUserTags
func (h *TagHandler) DeleteUserTags(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	if err := h.service.DeleteUserTags(r.Context(), userID); err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Deleted())
}
