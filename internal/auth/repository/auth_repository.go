package repository

import (
	"context"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Repository defines the interface for auth-related storage operations
type Repository interface {
	// Token operations
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, hashedToken string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, userID uuid.UUID) (*types.StoredToken, error)
	DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error

	// Session operations
	StoreSession(ctx context.Context, key string, value interface{}, expiresAt time.Time) error
	GetSession(ctx context.Context, key string) (*types.StoredSession, error)
	DeleteSession(ctx context.Context, key string) error

	// OAuth operations
	GetUserByExternalID(ctx context.Context, externalID, provider string) (*types.AuthUser, error)
	CreateUser(ctx context.Context, userData types.OAuthUserData) (*types.AuthUser, error)
	UpdateUserLastLogin(ctx context.Context, userID uuid.UUID) error
}

// authRepository implements the Repository interface
type authRepository struct {
	queries *db.Queries
	logger  *zap.Logger
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(queries *db.Queries, logger *zap.Logger) Repository {
	return &authRepository{
		queries: queries,
		logger:  logger,
	}
}
