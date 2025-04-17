package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// TokenService handles JWT token operations
type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID uuid.UUID, claims map[string]interface{}) (*types.TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (jwt.Token, error)
	ValidateRefreshToken(ctx context.Context, token string) (jwt.Token, error)
	RevokeRefreshToken(ctx context.Context, userID uuid.UUID) error
	GetUserClaims(ctx context.Context) (map[string]interface{}, bool)
}

type tokenService struct {
	config      *types.Config
	repo        repository.Repository
	logger      *zap.Logger
	accessAuth  *jwtauth.JWTAuth
	refreshAuth *jwtauth.JWTAuth
}

// NewTokenService creates a new token service
func NewTokenService(cfg *types.Config, repo repository.Repository, logger *zap.Logger) TokenService {
	return &tokenService{
		config:      cfg,
		repo:        repo,
		logger:      logger,
		accessAuth:  jwtauth.New("HS256", []byte(cfg.JWT.AccessTokenSecret), nil),
		refreshAuth: jwtauth.New("HS256", []byte(cfg.JWT.RefreshTokenSecret), nil),
	}
}

// GenerateTokenPair creates a new pair of access and refresh tokens
func (s *tokenService) GenerateTokenPair(ctx context.Context, userID uuid.UUID, claims map[string]interface{}) (*types.TokenPair, error) {
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
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Refresh token claims
	refreshClaims := map[string]interface{}{
		"user_id": userID.String(),
		"exp":     time.Now().Add(s.config.JWT.RefreshTokenTTL).Unix(),
	}

	// Generate refresh token
	_, refreshToken, err := s.refreshAuth.Encode(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token for storage
	hash, err := s.hashToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Store hashed refresh token
	expiresAt := time.Now().Add(s.config.JWT.RefreshTokenTTL)
	if err := s.repo.StoreRefreshToken(ctx, userID, hash, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &types.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.config.JWT.AccessTokenTTL),
	}, nil
}

// ValidateAccessToken validates an access token
func (s *tokenService) ValidateAccessToken(ctx context.Context, token string) (jwt.Token, error) {
	parsed, err := s.accessAuth.Decode(token)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	if parsed.Expiration().Before(time.Now()) {
		return nil, fmt.Errorf("access token expired")
	}

	return parsed, nil
}

// ValidateRefreshToken validates a refresh token
func (s *tokenService) ValidateRefreshToken(ctx context.Context, token string) (jwt.Token, error) {
	parsed, err := s.refreshAuth.Decode(token)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if parsed.Expiration().Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Get user ID from claims
	claims := parsed.PrivateClaims()
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Get stored token
	stored, err := s.repo.GetRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored token: %w", err)
	}

	// Verify token hash
	if err := s.verifyTokenHash(token, stored.Hash); err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return parsed, nil
}

// RevokeRefreshToken revokes a refresh token for a user
func (s *tokenService) RevokeRefreshToken(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteRefreshToken(ctx, userID)
}

// GetUserClaims extracts user claims from the context
func (s *tokenService) GetUserClaims(ctx context.Context) (map[string]interface{}, bool) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, false
	}
	return claims, true
}

// hashToken creates a secure hash of a token
func (s *tokenService) hashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash), nil
}

// verifyTokenHash verifies a token against its hash
func (s *tokenService) verifyTokenHash(token, hash string) error {
	hashBytes, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("invalid hash format: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(hashBytes, []byte(token)); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	return nil
}
