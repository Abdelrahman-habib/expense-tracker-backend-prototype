package handlers

import (
	"fmt"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/payloads"
	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"go.uber.org/zap"
)

// BeginAuthHandler initiates the OAuth flow
func (h *AuthHandler) BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider != "google" && provider != "github" {
		h.RespondError(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid provider: %s", provider)))
		return
	}

	// Handle additional scopes for Google OAuth
	if provider == "google" {
		// Get additional scopes from query parameters
		additionalScopes := r.URL.Query()["scope"]
		if len(additionalScopes) > 0 {
			// Get the current session
			session, err := gothic.Store.Get(r, StateSessionName)
			if err != nil {
				h.logger.Error("failed to get session",
					zap.Error(err),
				)
				h.RespondError(w, r, errors.ErrInternal(err))
				return
			}

			// Store additional scopes in session
			session.Values[GoogleScopesKey] = additionalScopes
			if err := session.Save(r, w); err != nil {
				h.logger.Error("failed to save session",
					zap.Error(err),
				)
				h.RespondError(w, r, errors.ErrInternal(err))
				return
			}

			// Get the existing Google provider to get its configuration
			var existingProvider *google.Provider
			for _, p := range goth.GetProviders() {
				if p.Name() == "google" {
					if gp, ok := p.(*google.Provider); ok {
						existingProvider = gp
						break
					}
				}
			}

			if existingProvider != nil {
				// Create a new provider with combined scopes
				allScopes := append([]string{"email", "profile"}, additionalScopes...)
				newProvider := google.New(
					existingProvider.ClientKey,
					existingProvider.Secret,
					existingProvider.CallbackURL,
					allScopes...,
				)

				// Replace the existing provider
				goth.UseProviders(newProvider)
			}
		}
	}

	gothic.BeginAuthHandler(w, r)
}

// CallbackHandler handles the OAuth callback
func (h *AuthHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider != "google" && provider != "github" {
		h.RespondError(w, r, errors.ErrInvalidRequest(fmt.Errorf("invalid provider: %s", provider)))
		return
	}

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		h.logger.Error("failed to complete auth",
			zap.String("provider", provider),
			zap.Error(err),
		)
		h.RespondError(w, r, errors.ErrInternal(err))
		return
	}

	// Create or get user and generate tokens
	authResult, err := h.service.AuthenticateUser(r.Context(), user)
	if err != nil {
		h.logger.Error("failed to authenticate user",
			zap.String("provider", provider),
			zap.Error(err),
		)
		h.RespondError(w, r, errors.ErrInternal(err))
		return
	}

	// Set cookies
	h.service.SetCookies(w, authResult.AccessToken, authResult.RefreshToken)

	h.Respond(w, r, payloads.OK(map[string]interface{}{
		"user": authResult.User,
	}))
}
