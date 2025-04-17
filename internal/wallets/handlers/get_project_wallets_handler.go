package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetProjectWallets godoc
// @Summary Get project wallets
// @Description Retrieves all wallets associated with a specific project
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path string true "Project ID" format(uuid)
// @Success 200 {object} payloads.Response{data=[]types.Wallet}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 404  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /projects/{project_id}/wallets [get]
// @ID GetProjectWallets
func (h *WalletHandler) GetProjectWallets(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "project_id"))
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	wallets, err := h.service.GetProjectWallets(r.Context(), projectID, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.OK(wallets))
}
