package repository

import (
	"context"

	"go.uber.org/zap"
)

// DeleteSession removes a stored session
func (r *authRepository) DeleteSession(ctx context.Context, key string) error {
	r.logger.Debug("deleting session", zap.String("key", key))
	return r.queries.DeleteSession(ctx, key)
}
