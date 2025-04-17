package service

import (
	"context"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/google/uuid"
	"github.com/markbates/goth"
)

// AuthResult represents the result of an authentication operation
type AuthResult struct {
	User         types.AuthUser `json:"user"`
	AccessToken  string         `json:"-"`
	RefreshToken string         `json:"-"`
}

// Service defines the interface for authentication operations
type Service interface {
	// Middleware returns an http.Handler that authenticates requests
	Middleware(next http.Handler) http.Handler

	// OAuth operations
	BeginAuth(w http.ResponseWriter, r *http.Request, provider string, scopes []string) error
	CompleteAuth(w http.ResponseWriter, r *http.Request) (*AuthResult, error)
	GetGoogleToken(ctx context.Context) (types.GoogleOauthToken, error)

	// Token operations
	RefreshTokens(ctx context.Context, refreshToken string) (*AuthResult, error)
	RevokeTokens(ctx context.Context, userID uuid.UUID) error
	GetUserClaims(ctx context.Context) (map[string]interface{}, bool)
	SetCookies(w http.ResponseWriter, accessToken, refreshToken string)

	// User operations
	AuthenticateUser(ctx context.Context, user goth.User) (*AuthResult, error)

	// Session operations
	Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}
