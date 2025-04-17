package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/render"
)

// CreateProject godoc
// @Summary Create a new project
// @Description Creates a new project for the authenticated user
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body types.ProjectCreatePayload true "project creation request"
// @Success 201 {object} payloads.Response{data=types.Project}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /projects [post]
// @ID CreateProject
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	var req types.ProjectCreatePayload
	if err := render.Bind(r, &req); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	project, err := h.service.CreateProject(r.Context(), userID, req)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Created(project))
}
