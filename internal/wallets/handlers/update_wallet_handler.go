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

// UpdateWallet godoc
// @Summary Update a wallet
// @Description Updates an existing wallet
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Wallet ID" format(uuid)
// @Param request body types.WalletUpdatePayload true "Wallet update request"
// @Success 200 {object} payloads.Response{data=types.Wallet}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 404  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /wallets/{id} [put]
// @ID UpdateWallet
func (h *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
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

	// Get existing wallet first
	existingWallet, err := h.service.GetWallet(r.Context(), walletID, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	// Create update payload from existing wallet
	updatePayload := existingWallet.ToUpdatePayload()

	// Use render.Bind to decode and validate
	if err := render.Bind(r, &updatePayload); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	wallet, err := h.service.UpdateWallet(r.Context(), updatePayload, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Updated(wallet))
}
