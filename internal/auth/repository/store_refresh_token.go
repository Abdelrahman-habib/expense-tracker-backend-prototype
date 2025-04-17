package repository

import (
	"context"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// StoreRefreshToken stores a hashed refresh token for a user
func (r *authRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, hashedToken string, expiresAt time.Time) error {
	r.logger.Debug("storing refresh token", zap.String("user_id", userID.String()))
	return r.queries.UpdateUserRefreshToken(ctx, db.UpdateUserRefreshTokenParams{
		UserID: userID,
		RefreshTokenHash: pgtype.Text{
			String: hashedToken,
			Valid:  true,
		},
	})
}
