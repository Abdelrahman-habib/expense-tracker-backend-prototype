package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"go.uber.org/zap"
)

// OAuthService handles OAuth provider operations
type OAuthService interface {
	ConfigureProviders() error
	BeginAuth(w http.ResponseWriter, r *http.Request, provider string, scopes []string) error
	CompleteAuth(w http.ResponseWriter, r *http.Request) (*types.OAuthUserData, error)
	GetGoogleToken(ctx context.Context) (types.GoogleOauthToken, error)
}

type oauthService struct {
	config  *types.Config
	repo    repository.Repository
	logger  *zap.Logger
	session SessionService
	token   TokenService
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(cfg *types.Config, repo repository.Repository, session SessionService, token TokenService, logger *zap.Logger) OAuthService {
	return &oauthService{
		config:  cfg,
		repo:    repo,
		session: session,
		token:   token,
		logger:  logger,
	}
}

// ConfigureProviders sets up OAuth providers
func (s *oauthService) ConfigureProviders() error {
	goth.UseProviders(
		google.New(
			s.config.OAuth.Google.ClientID,
			s.config.OAuth.Google.ClientSecret,
			s.config.OAuth.Google.RedirectURL,
			s.config.OAuth.Google.DefaultScopes...,
		),
		github.New(
			s.config.OAuth.GitHub.ClientID,
			s.config.OAuth.GitHub.ClientSecret,
			s.config.OAuth.GitHub.RedirectURL,
			"user:email",
		),
	)

	gothic.Store = s.session.GetStore()
	return nil
}

// BeginAuth initiates the OAuth flow
func (s *oauthService) BeginAuth(w http.ResponseWriter, r *http.Request, provider string, scopes []string) error {
	if provider != "google" && provider != "github" {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	// Handle additional scopes for Google OAuth
	if provider == "google" && len(scopes) > 0 {
		// Store additional scopes in session
		if err := s.session.Set(r, w, GoogleScopesKey, scopes); err != nil {
			return fmt.Errorf("failed to store scopes: %w", err)
		}

		// Get the existing Google provider
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
			allScopes := append(s.config.OAuth.Google.DefaultScopes, scopes...)
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

	gothic.BeginAuthHandler(w, r)
	return nil
}

// CompleteAuth handles the OAuth callback
func (s *oauthService) CompleteAuth(w http.ResponseWriter, r *http.Request) (*types.OAuthUserData, error) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return nil, fmt.Errorf("failed to complete auth: %w", err)
	}

	// Store token in session for Google OAuth
	if user.Provider == "google" {
		if err := s.session.Set(r, w, GoogleTokenKey, user.AccessToken); err != nil {
			s.logger.Warn("failed to store Google token", zap.Error(err))
		}
	}

	return &types.OAuthUserData{
		ExternalID: user.UserID,
		Name:       user.Name,
		Email:      user.Email,
		Provider:   user.Provider,
	}, nil
}

// GetGoogleToken retrieves the Google OAuth token
func (s *oauthService) GetGoogleToken(ctx context.Context) (types.GoogleOauthToken, error) {
	// Get user claims from context
	claims, ok := s.token.GetUserClaims(ctx)
	if !ok {
		return types.GoogleOauthToken{}, fmt.Errorf("no user claims found")
	}

	// Get user from database
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return types.GoogleOauthToken{}, fmt.Errorf("invalid user ID in claims")
	}

	// Get token from session
	token, err := s.session.Get(ctx, GoogleTokenKey)
	if err != nil {
		return types.GoogleOauthToken{}, fmt.Errorf("failed to get Google token: %w", err)
	}

	// Get scopes from session
	scopes, err := s.session.Get(ctx, GoogleScopesKey)
	if err != nil {
		// Fall back to default scopes
		scopes = s.config.OAuth.Google.DefaultScopes
	}

	return types.GoogleOauthToken{
		ExternalAccountID: userIDStr,
		Token:             token.(string),
		Scopes:            scopes.([]string),
	}, nil
}
