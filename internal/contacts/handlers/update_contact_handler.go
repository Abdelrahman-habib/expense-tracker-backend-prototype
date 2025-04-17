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

// UpdateContact godoc
// @Summary Update a Contact
// @Description Updates an existing Contact
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID" format(uuid)
// @Param request body types.ContactUpdatePayload true "Contact update request"
// @Success 200 {object} payloads.Response{data=types.Contact}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /contacts/{id} [put]
// @ID UpdateContact
func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
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

	// Get existing contact first
	existingContact, err := h.service.GetContact(r.Context(), contactID, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	// Create update payload from existing contact
	updatePayload := existingContact.ToUpdatePayload()

	// Use render.Bind to decode and validate
	if err := render.Bind(r, &updatePayload); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	contact, err := h.service.UpdateContact(r.Context(), updatePayload, userID)
	if err != nil {
		if errors.IsErrorType(err, errors.ErrorTypeNotFound) {
			h.RespondError(w, r, errors.ErrNotFound())
			return
		}
		h.RespondError(w, r, errors.ErrDatabase(err))
		return
	}

	h.Respond(w, r, payloads.Updated(contact))
}
