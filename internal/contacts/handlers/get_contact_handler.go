package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetContact godoc
// @Summary Get a Contact
// @Description Retrieves a Contact by ID
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID" format(uuid)
// @Success 200 {object} payloads.Response{data=types.Contact}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /contacts/{id} [get]
// @ID GetContact
func (h *ContactHandler) GetContact(w http.ResponseWriter, r *http.Request) {
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

	contact, err := h.service.GetContact(r.Context(), contactID, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.OK(contact))
}
