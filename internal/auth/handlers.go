package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRoutes registers all auth-related routes
func (s *Service) RegisterRoutes(r chi.Router) {
	r.Get("/auth/{provider}", s.BeginAuthHandler)
	r.Get("/auth/{provider}/callback", s.CallbackHandler)
	r.Post("/auth/refresh", s.RefreshTokenHandler)
	r.Post("/auth/logout", s.LogoutHandler)
}

// BeginAuthHandler initiates the OAuth flow
func (s *Service) BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider != "google" && provider != "github" {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	gothic.BeginAuthHandler(w, r)
}

// CallbackHandler handles the OAuth callback
func (s *Service) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider != "google" && provider != "github" {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		s.logger.Error("failed to complete auth",
			zap.String("provider", provider),
			zap.Error(err),
		)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Check if user exists
	dbUser, err := s.db.GetUserByExternalID(r.Context(), db.GetUserByExternalIDParams{
		ExternalID: user.UserID,
		Provider:   provider,
	})
	if err != nil {
		// Create new user if not found
		dbUser, err = s.db.CreateUser(r.Context(), db.CreateUserParams{
			Name:       user.Name,
			Email:      user.Email,
			ExternalID: user.UserID,
			Provider:   provider,
		})
		if err != nil {
			s.logger.Error("failed to create user",
				zap.String("provider", provider),
				zap.Error(err),
			)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	// Generate tokens
	claims := map[string]interface{}{
		"name":     dbUser.Name,
		"email":    dbUser.Email,
		"provider": dbUser.Provider,
	}

	accessToken, refreshToken, err := s.generateTokens(dbUser.UserID, claims)
	if err != nil {
		s.logger.Error("failed to generate tokens",
			zap.String("user_id", dbUser.UserID.String()),
			zap.Error(err),
		)
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Set cookies
	s.setCookies(w, accessToken, refreshToken)

	// Update last login
	if err := s.db.UpdateUserLastLogin(r.Context(), dbUser.UserID); err != nil {
		s.logger.Error("failed to update last login",
			zap.String("user_id", dbUser.UserID.String()),
			zap.Error(err),
		)
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Authentication successful",
		"user": map[string]interface{}{
			"id":       dbUser.UserID.String(),
			"name":     dbUser.Name,
			"email":    dbUser.Email,
			"provider": dbUser.Provider,
		},
	})
}

// RefreshTokenHandler handles token refresh requests
func (s *Service) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie(RefreshTokenCookie)
	if err != nil {
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}

	// Validate refresh token
	token, err := s.refreshAuth.Decode(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	if token.Expiration().Before(time.Now()) {
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}

	// Extract user ID from claims
	claims := token.PrivateClaims()
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return
	}

	// Parse user ID to UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := s.db.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Verify refresh token hash
	if err := s.verifyRefreshToken(cookie.Value, user.RefreshTokenHash.String); err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	newClaims := map[string]interface{}{
		"name":     user.Name,
		"email":    user.Email,
		"provider": user.Provider,
	}

	accessToken, refreshToken, err := s.generateTokens(userID, newClaims)
	if err != nil {
		s.logger.Error("failed to generate new tokens",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		http.Error(w, "Failed to generate new tokens", http.StatusInternalServerError)
		return
	}

	// Set new cookies
	s.setCookies(w, accessToken, refreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Tokens refreshed successfully",
	})
}

// LogoutHandler handles user logout
func (s *Service) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     s.config.Cookie.Path,
		Domain:   s.config.Cookie.Domain,
		MaxAge:   -1,
		Secure:   s.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     s.config.Cookie.Path,
		Domain:   s.config.Cookie.Domain,
		MaxAge:   -1,
		Secure:   s.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Clear session
	session, _ := gothic.Store.Get(r, StateSessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// verifyRefreshToken verifies a refresh token against its stored hash
func (s *Service) verifyRefreshToken(token, hash string) error {
	if hash == "" {
		return fmt.Errorf("no refresh token hash found")
	}

	hashBytes, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("invalid hash format: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(hashBytes, []byte(token)); err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	return nil
}
