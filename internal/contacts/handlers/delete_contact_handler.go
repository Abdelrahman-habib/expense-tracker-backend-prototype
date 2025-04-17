package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DeleteContact godoc
// @Summary Delete a Contact
// @Description Deletes a Contact by ID
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID" format(uuid)
// @Success 200 {object} payloads.Response
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /contacts/{id} [delete]
// @ID DeleteContact
func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	// Check if contact exists and belongs to user
	_, err = h.service.GetContact(r.Context(), contactID, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	err = h.service.DeleteContact(r.Context(), contactID, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.Deleted())
}
