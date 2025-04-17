package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetWallet godoc
// @Summary Get a wallet
// @Description Retrieves a wallet by ID
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Wallet ID" format(uuid)
// @Success 200 {object} payloads.Response{data=types.Wallet}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 404  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /wallets/{id} [get]
// @ID GetWallet
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	wallet, err := h.service.GetWallet(r.Context(), walletID, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.OK(wallet))
}
