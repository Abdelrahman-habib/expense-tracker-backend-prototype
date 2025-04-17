package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidProvider = errors.New("invalid provider")
	ErrInvalidRefresh  = errors.New("invalid refresh token")
)

type service struct {
	config  *types.Config
	repo    repository.Repository
	logger  *zap.Logger
	token   TokenService
	oauth   OAuthService
	session SessionService
}

// NewService creates a new auth service
func NewService(cfg *types.Config, repo repository.Repository, logger *zap.Logger) Service {
	// Create session service
	sessionSvc := NewSessionService(cfg, repo, logger)

	// Create token service
	tokenSvc := NewTokenService(cfg, repo, logger)

	// Create OAuth service
	oauthSvc := NewOAuthService(cfg, repo, sessionSvc, tokenSvc, logger)

	// Configure OAuth providers
	if err := oauthSvc.ConfigureProviders(); err != nil {
		logger.Fatal("failed to configure OAuth providers", zap.Error(err))
	}

	return &service{
		config:  cfg,
		repo:    repo,
		logger:  logger,
		token:   tokenSvc,
		oauth:   oauthSvc,
		session: sessionSvc,
	}
}

// Middleware returns an http.Handler that authenticates requests
func (s *service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AccessTokenCookie)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if _, err := s.token.ValidateAccessToken(r.Context(), cookie.Value); err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		next.ServeHTTP(w, r)
	})
}

// BeginAuth initiates the OAuth flow
func (s *service) BeginAuth(w http.ResponseWriter, r *http.Request, provider string, scopes []string) error {
	return s.oauth.BeginAuth(w, r, provider, scopes)
}

// CompleteAuth handles the OAuth callback
func (s *service) CompleteAuth(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
	// Complete OAuth flow
	userData, err := s.oauth.CompleteAuth(w, r)
	if err != nil {
		return nil, fmt.Errorf("failed to complete auth: %w", err)
	}

	// Get or create user
	user, err := s.repo.GetUserByExternalID(r.Context(), userData.ExternalID, userData.Provider)
	if err != nil {
		// Create new user if not found
		user, err = s.repo.CreateUser(r.Context(), *userData)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Generate tokens
	claims := map[string]interface{}{
		"name":     user.Name,
		"email":    user.Email,
		"provider": user.Provider,
	}

	tokenPair, err := s.token.GenerateTokenPair(r.Context(), user.ID, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login
	if err := s.repo.UpdateUserLastLogin(r.Context(), user.ID); err != nil {
		s.logger.Warn("failed to update last login", zap.Error(err))
	}

	return &AuthResult{
		User:         *user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// AuthenticateUser authenticates a user using OAuth data
func (s *service) AuthenticateUser(ctx context.Context, user goth.User) (*AuthResult, error) {
	userData := &types.OAuthUserData{
		ExternalID: user.UserID,
		Name:       user.Name,
		Email:      user.Email,
		Provider:   user.Provider,
	}

	// Get or create user
	authUser, err := s.repo.GetUserByExternalID(ctx, userData.ExternalID, userData.Provider)
	if err != nil {
		// Create new user if not found
		authUser, err = s.repo.CreateUser(ctx, *userData)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Generate tokens
	claims := map[string]interface{}{
		"name":     authUser.Name,
		"email":    authUser.Email,
		"provider": authUser.Provider,
	}

	tokenPair, err := s.token.GenerateTokenPair(ctx, authUser.ID, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login
	if err := s.repo.UpdateUserLastLogin(ctx, authUser.ID); err != nil {
		s.logger.Warn("failed to update last login", zap.Error(err))
	}

	return &AuthResult{
		User:         *authUser,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshTokens refreshes the access and refresh tokens
func (s *service) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Validate refresh token
	token, err := s.token.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user ID from claims
	claims := token.PrivateClaims()
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// Get user from database
	user, err := s.repo.GetUserByExternalID(ctx, userIDStr, claims["provider"].(string))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate new tokens
	newClaims := map[string]interface{}{
		"name":     user.Name,
		"email":    user.Email,
		"provider": user.Provider,
	}

	tokenPair, err := s.token.GenerateTokenPair(ctx, user.ID, newClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResult{
		User:         *user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RevokeTokens revokes all tokens for a user
func (s *service) RevokeTokens(ctx context.Context, userID uuid.UUID) error {
	return s.token.RevokeRefreshToken(ctx, userID)
}

// GetUserClaims extracts user claims from the context
func (s *service) GetUserClaims(ctx context.Context) (map[string]interface{}, bool) {
	return s.token.GetUserClaims(ctx)
}

// SetCookies sets the access and refresh token cookies
func (s *service) SetCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     s.config.Cookie.Path,
		Domain:   s.config.Cookie.Domain,
		MaxAge:   int(s.config.JWT.AccessTokenTTL.Seconds()),
		Secure:   s.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     s.config.Cookie.Path,
		Domain:   s.config.Cookie.Domain,
		MaxAge:   int(s.config.JWT.RefreshTokenTTL.Seconds()),
		Secure:   s.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// GetGoogleToken retrieves the Google OAuth token
func (s *service) GetGoogleToken(ctx context.Context) (types.GoogleOauthToken, error) {
	return s.oauth.GetGoogleToken(ctx)
}

// Logout handles user logout
func (s *service) Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	if err := s.session.Delete(r, w, StateSessionName); err != nil {
		return fmt.Errorf("failed to clear session: %w", err)
	}

	return nil
}
