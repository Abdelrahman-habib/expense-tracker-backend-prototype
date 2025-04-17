package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
)

// TODO: cache the last request for user and check if the token is repeated

// GetUserContacts godoc
// @Summary      Get user's Google contacts
// @Description  Retrieves the authenticated user's Google contacts with optional pagination
// @Tags         Users
// @x-badges 	[{"name":"External", "position":"before", "color":"purple"}]
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pageToken  query     string  false  "Page token for pagination"
// @Success      200  {object}  payloads.Response{data=types.PaginatedGoogleContacts}
// @Failure      401  {object} errors.ErrorResponse
// @Failure      429  {object} errors.ErrorResponse
// @Failure      502  {object} errors.ErrorResponse
// @Router       /users/contacts [get]
// @ID GetUserContacts
func (h *UserHandler) GetUserContacts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	contatcs, err := h.service.GetGoogleContacts(r.Context(), query.Get("pageToken"))
	if err != nil {
		h.RespondError(w, r, errors.ErrExternalService(err))
		return
	}

	h.Respond(w, r, payloads.OK(contatcs))

}
