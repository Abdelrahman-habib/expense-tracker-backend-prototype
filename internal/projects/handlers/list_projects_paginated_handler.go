package handlers

import (
	"net/http"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/google/uuid"
)

// ListProjectsPaginated godoc
// @Summary List projects with pagination
// @Description Returns a paginated list of projects
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query integer false "Number of projects to return" minimum(1) maximum(100) default(10)
// @Param next_token query string false "Token for the next page"
// @Success 200 {object} payloads.Response{data=[]types.Project}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /projects [get]
// @ID ListProjectsPaginated
func (h *ProjectHandler) ListProjectsPaginated(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	// Parse and validate pagination parameters
	params, err := types.ParsePaginationParams(r.URL.Query())
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Set cursor values based on parsed parameters
	var cursor time.Time
	var cursorID uuid.UUID
	if params.Cursor != nil {
		cursor = params.Cursor.Timestamp
		cursorID = params.Cursor.ID
	} else {
		cursor = time.Now()
		cursorID = uuid.Nil
	}

	projects, err := h.service.ListProjectsPaginated(r.Context(), userID, cursor, cursorID, params.Limit)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	var nextToken string
	if len(projects) > 0 && len(projects) == int(params.Limit) {
		lastProject := projects[len(projects)-1]
		nextToken = types.EncodeCursor(lastProject.CreatedAt, lastProject.ProjectID)
	}

	h.Respond(w, r, payloads.Paginated(
		projects,
		nextToken,
		params.Limit,
	))
}
