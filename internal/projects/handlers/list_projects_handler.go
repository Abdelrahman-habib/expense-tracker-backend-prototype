package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
)

// ListProjects godoc
// @Summary List projects
// @Description Returns all project for a user (since projects are limited to 10 we usually won't want to paginate the date)
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} payloads.Response{data=[]types.Project}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /projects [get]
// @ID ListProjects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	projects, err := h.service.ListProjects(r.Context(), userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.OK(projects))
}
