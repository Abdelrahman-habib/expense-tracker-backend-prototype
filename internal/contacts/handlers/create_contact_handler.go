package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/render"
)

// CreateContact godoc
// @Summary Create a new Contact
// @Description Creates a new Contact for the authenticated user
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body types.ContactCreatePayload true "Contact creation request"
// @Success 201 {object} payloads.Response{data=types.Contact}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 429 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /contacts [post]
// @ID CreateContact
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	var req types.ContactCreatePayload
	if err := render.Bind(r, &req); err != nil {
		h.RespondError(w, r, errors.ErrInvalidRequest(err))
		return
	}

	contact, err := h.service.CreateContact(r.Context(), req, userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
	}

	h.Respond(w, r, payloads.Created(contact))
}
