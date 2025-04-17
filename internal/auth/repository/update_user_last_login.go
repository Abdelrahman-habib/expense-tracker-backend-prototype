package repository

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateUserLastLogin updates the last login timestamp for a user
func (r *authRepository) UpdateUserLastLogin(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug("updating last login", zap.String("user_id", userID.String()))
	return r.queries.UpdateUserLastLogin(ctx, userID)
}
