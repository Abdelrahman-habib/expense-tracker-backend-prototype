package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/render"
)

// CreateWallet godoc
// @Summary Create a new wallet
// @Description Creates a new wallet for the authenticated user
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body types.WalletCreatePayload true "Wallet creation request"
// @Success 201 {object} payloads.Response{data=types.Wallet}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 404  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /wallets [post]
// @ID CreateWallet
func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	var req types.WalletCreatePayload
	if err := render.Bind(r, &req); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	wallet, err := h.service.CreateWallet(r.Context(), req, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Created(wallet))
}
