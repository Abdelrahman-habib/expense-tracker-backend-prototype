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

// UpdateProject godoc
// @Summary Update a project
// @Description Updates an existing project
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "project ID" format(uuid)
// @Param request body types.ProjectUpdatePayload true "project update request"
// @Success 200 {object} payloads.Response{data=types.Project}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /projects/{id} [put]
// @ID UpdateProject
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Get existing project first
	existingProject, err := h.service.GetProject(r.Context(), userID, projectID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	// Create update payload from existing project
	updatePayload := existingProject.ToUpdatePayload()

	// Use render.Bind to decode and validate
	if err := render.Bind(r, &updatePayload); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	project, err := h.service.UpdateProject(r.Context(), userID, updatePayload)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Updated(project))
}
