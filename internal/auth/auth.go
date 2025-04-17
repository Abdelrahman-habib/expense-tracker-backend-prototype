package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
	StateSessionName   = "oauth_state"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidProvider = errors.New("invalid provider")
	ErrInvalidRefresh  = errors.New("invalid refresh token")
)

type Service struct {
	config *Config
	db     *db.Queries
	logger *zap.Logger

	accessAuth  *jwtauth.JWTAuth
	refreshAuth *jwtauth.JWTAuth
	sessions    *sessions.CookieStore
}

func NewService(cfg *Config, db *db.Queries, logger *zap.Logger) *Service {
	// Create a secure random key for sessions
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		logger.Fatal("failed to generate session key", zap.Error(err))
	}

	s := &Service{
		config:      cfg,
		db:          db,
		logger:      logger,
		accessAuth:  jwtauth.New("HS256", []byte(cfg.JWT.AccessTokenSecret), nil),
		refreshAuth: jwtauth.New("HS256", []byte(cfg.JWT.RefreshTokenSecret), nil),
		sessions:    sessions.NewCookieStore(sessionKey),
	}

	// Configure session store
	s.sessions.Options = &sessions.Options{
		Path:     cfg.Cookie.Path,
		Domain:   cfg.Cookie.Domain,
		MaxAge:   int(cfg.JWT.RefreshTokenTTL.Seconds()),
		Secure:   cfg.Cookie.Secure,
		HttpOnly: true,
	}

	// Configure Goth providers
	goth.UseProviders(
		google.New(
			cfg.OAuth.Google.ClientID,
			cfg.OAuth.Google.ClientSecret,
			cfg.OAuth.Google.RedirectURL,
			"email", "profile",
		),
		github.New(
			cfg.OAuth.GitHub.ClientID,
			cfg.OAuth.GitHub.ClientSecret,
			cfg.OAuth.GitHub.RedirectURL,
			"user:email",
		),
	)

	gothic.Store = s.sessions
	return s
}

// Middleware returns an http.Handler that authenticates requests using JWT
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AccessTokenCookie)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := s.accessAuth.Decode(cookie.Value)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if token.Expiration().Before(time.Now()) {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := jwtauth.NewContext(r.Context(), token, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserClaims extracts user claims from the context
func (s *Service) GetUserClaims(ctx context.Context) (map[string]interface{}, bool) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, false
	}
	return claims, true
}

// generateTokens creates a new pair of access and refresh tokens
func (s *Service) generateTokens(userID uuid.UUID, claims map[string]interface{}) (string, string, error) {
	// Access token claims
	accessClaims := map[string]interface{}{
		"user_id": userID.String(),
		"exp":     time.Now().Add(s.config.JWT.AccessTokenTTL).Unix(),
	}
	for k, v := range claims {
		accessClaims[k] = v
	}

	// Generate access token
	_, accessToken, err := s.accessAuth.Encode(accessClaims)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Refresh token claims
	refreshClaims := map[string]interface{}{
		"user_id": userID.String(),
		"exp":     time.Now().Add(s.config.JWT.RefreshTokenTTL).Unix(),
	}

	// Generate refresh token
	_, refreshToken, err := s.refreshAuth.Encode(refreshClaims)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token for storage
	hash, err := s.hashToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Store hashed refresh token in database
	if err := s.db.UpdateUserRefreshToken(context.Background(), db.UpdateUserRefreshTokenParams{
		UserID: userID,
		RefreshTokenHash: pgtype.Text{
			String: hash,
			Valid:  true,
		},
	}); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// hashToken creates a secure hash of a token
func (s *Service) hashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash), nil
}

// setCookies sets the access and refresh token cookies
func (s *Service) setCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     s.config.Cookie.Path,
		Domain:   s.config.Cookie.Domain,
		Expires:  time.Now().Add(s.config.JWT.AccessTokenTTL),
		Secure:   s.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     s.config.Cookie.Path,
		Domain:   s.config.Cookie.Domain,
		Expires:  time.Now().Add(s.config.JWT.RefreshTokenTTL),
		Secure:   s.config.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
