package handlers

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	"go.uber.org/zap"
)

// RefreshTokenHandler handles token refresh requests
func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie(service.RefreshTokenCookie)
	if err != nil {
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	// Refresh tokens
	result, err := h.service.RefreshTokens(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Error("failed to refresh tokens", zap.Error(err))
		h.RespondError(w, r, errors.ErrAuthorization(err))
		return
	}

	// Set new cookies
	h.service.SetCookies(w, result.AccessToken, result.RefreshToken)

	h.Respond(w, r, payloads.OK(nil))
}

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Logout(r.Context(), w, r); err != nil {
		h.logger.Error("failed to logout", zap.Error(err))
		h.RespondError(w, r, errors.ErrInternal(err))
		return
	}

	h.Respond(w, r, payloads.OK(nil))
}
