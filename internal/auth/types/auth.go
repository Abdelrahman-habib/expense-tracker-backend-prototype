package types

import (
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
)

// AuthUser represents minimal user information needed for auth operations
type AuthUser struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Provider string    `json:"provider"`
}

// AuthCallbackResponse represents the response for OAuth callback
type AuthCallbackResponse struct {
	Message string   `json:"message"`
	User    AuthUser `json:"user"`
}

// RefreshTokenResponse represents the response for token refresh
type RefreshTokenResponse struct {
	Message string `json:"message"`
}

// LogoutResponse represents the response for logout
type LogoutResponse struct {
	Message string `json:"message"`
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// BeginAuthRequest represents the request to start OAuth flow
type BeginAuthRequest struct {
	Provider string   `json:"provider"`
	Scopes   []string `json:"scopes,omitempty"`
}

func (r *BeginAuthRequest) Bind(_ *http.Request) error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Provider, validation.Required, validation.In("google", "github")),
		validation.Field(&r.Scopes, validation.Each(validation.Required)),
	)
}

// RefreshTokenRequest represents the request to refresh tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *RefreshTokenRequest) Bind(_ *http.Request) error {
	return validation.ValidateStruct(r,
		validation.Field(&r.RefreshToken, validation.Required),
	)
}

// StoredToken represents a token stored in the database
type StoredToken struct {
	UserID    uuid.UUID `json:"user_id"`
	Hash      string    `json:"hash"`
	ExpiresAt time.Time `json:"expires_at"`
}

// StoredSession represents a session stored in the database
type StoredSession struct {
	Key       string    `json:"key"`
	Value     []byte    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OAuthUserData represents user data received from OAuth providers
type OAuthUserData struct {
	ExternalID string `json:"external_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Provider   string `json:"provider"`
}

func (u *OAuthUserData) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.ExternalID, validation.Required),
		validation.Field(&u.Name, validation.Required),
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Provider, validation.Required, validation.In("google", "github")),
	)
}
