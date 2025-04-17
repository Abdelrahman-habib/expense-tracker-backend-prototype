package repository

import (
	"context"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// DeleteRefreshToken removes a stored refresh token for a user
func (r *authRepository) DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug("deleting refresh token", zap.String("user_id", userID.String()))
	return r.queries.UpdateUserRefreshToken(ctx, db.UpdateUserRefreshTokenParams{
		UserID: userID,
		RefreshTokenHash: pgtype.Text{
			Valid: true,
		},
	})
}
