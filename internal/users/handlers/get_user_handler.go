package handlers

import (
	"net/http"

	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
)

// GetUser godoc
// @Summary      Get authenticated user profile
// @Description  Retrieves the profile of the currently authenticated user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  payloads.Response{data=types.User}
// @Failure      401  {object} errors.ErrorResponse
// @Failure      404  {object} errors.ErrorResponse
// @Failure      429  {object} errors.ErrorResponse
// @Failure      500  {object} errors.ErrorResponse
// @Router       /users/me [get]
// @ID GetUser
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := requestcontext.GetUserIDFromContext(r.Context())
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		h.HandleServiceError(w, r, err)
		return
	}

	h.Respond(w, r, payloads.OK(user))

}
