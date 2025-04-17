package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/types"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
)

// SearchProject godoc
// @Summary Search project
// @Description Searches for project based on a query string
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query" minLength(1) maxLength(100)
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(50) default(10)
// @Success 200 {object} payloads.Response{data=[]types.Project}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /project/search [get]
// @ID SearchProjects
func (h *ProjectHandler) SearchProjects(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	params, err := types.ParseAndValidateSearchParams(query)
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	projects, err := h.service.SearchProjects(r.Context(), userID, params.Query, params.Limit)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Search(
		projects,
		params.Query,
		params.Limit,
		len(projects),
	))
}
