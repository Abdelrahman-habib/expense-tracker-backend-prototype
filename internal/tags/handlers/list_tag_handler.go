package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
)

// ListTags godoc
// @Summary List Tags
// @Description Returns a list of Tags
// @Tags Tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} payloads.Response{data=[]types.Tag}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags [get]list
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	tags, err := h.service.ListTags(r.Context(), userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.OK(tags))
}
