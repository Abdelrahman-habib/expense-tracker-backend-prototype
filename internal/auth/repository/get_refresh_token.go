package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetRefreshToken retrieves a stored refresh token for a user
func (r *authRepository) GetRefreshToken(ctx context.Context, userID uuid.UUID) (*types.StoredToken, error) {
	r.logger.Debug("getting refresh token", zap.String("user_id", userID.String()))

	user, err := r.queries.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.RefreshTokenHash.Valid {
		return nil, fmt.Errorf("no refresh token found")
	}

	return &types.StoredToken{
		UserID: userID,
		Hash:   user.RefreshTokenHash.String,
		// Note: We might want to add expires_at to the database schema
		ExpiresAt: time.Now().Add(24 * time.Hour * 7), // 7 days, should match config
	}, nil
}
