package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
)

// SearchContacts godoc
// @Summary Search Contacts
// @Description Searches for Contacts based on a query string
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query" minLength(1) maxLength(100)
// @Param limit query integer false "Maximum number of results" minimum(1) maximum(50) default(10)
// @Success 200 {object} payloads.Response{data=[]types.Contact}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /contacts/search [get]
// @ID SearchContacts
func (h *ContactHandler) SearchContacts(w http.ResponseWriter, r *http.Request) {
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

	var contacts []types.Contact
	if params.SearchByPhone {
		contacts, err = h.service.SearchContactsByPhone(r.Context(), userID, params.Query, params.Limit)
	} else {
		contacts, err = h.service.SearchContacts(r.Context(), userID, params.Query, params.Limit)
	}

	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Search(
		contacts,
		params.Query,
		params.Limit,
		len(contacts),
	))
}
