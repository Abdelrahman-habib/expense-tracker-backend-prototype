package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
)

// SearchWallets godoc
// @Summary Search wallets
// @Description Searches for wallets based on a query string
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query" minLength(1) maxLength(100)
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(50) default(10)
// @Success 200 {object} payloads.Response{data=[]types.Wallet}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /wallets/search [get]
// @ID SearchWallets
func (h *WalletHandler) SearchWallets(w http.ResponseWriter, r *http.Request) {
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

	wallets, err := h.service.SearchWallets(r.Context(), userID, params.Query, params.Limit)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Search(
		wallets,
		params.Query,
		params.Limit,
		len(wallets),
	))
}
