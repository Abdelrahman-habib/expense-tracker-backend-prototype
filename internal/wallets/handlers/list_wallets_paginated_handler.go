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

// ListWalletsPaginated godoc
// @Summary List wallets with pagination
// @Description Returns a paginated list of wallets
// @Tags Wallets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query integer false "Number of wallets to return" minimum(1) maximum(100) default(10)
// @Param next_token query string false "Token for the next page"
// @Success 200 {object} payloads.Response{data=[]types.Wallet}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401  {object} errors.ErrorResponse
// @Failure 429  {object} errors.ErrorResponse
// @Failure 500  {object} errors.ErrorResponse
// @Router /wallets [get]
// @ID ListWalletsPaginated
func (h *WalletHandler) ListWalletsPaginated(w http.ResponseWriter, r *http.Request) {
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

	// Set default cursor values if not provided
	var cursor time.Time
	var cursorID uuid.UUID
	if params.Cursor != nil {
		cursor = params.Cursor.Timestamp
		cursorID = params.Cursor.ID
	} else {
		cursor = time.Now().UTC()
		cursorID = uuid.Nil
	}

	wallets, err := h.service.ListWalletsPaginated(r.Context(), userID, cursor, cursorID, params.Limit)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	var nextToken string
	if len(wallets) > 0 && len(wallets) == int(params.Limit) {
		lastWallet := wallets[len(wallets)-1]
		nextToken = types.EncodeCursor(lastWallet.CreatedAt, lastWallet.WalletID)
	}

	h.Respond(w, r, payloads.Paginated(
		wallets,
		nextToken,
		params.Limit,
	))
}
