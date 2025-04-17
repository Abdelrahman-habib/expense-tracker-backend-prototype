package repository

import (
	"context"
	"fmt"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"go.uber.org/zap"
)

// GetSession retrieves a stored session
func (r *authRepository) GetSession(ctx context.Context, key string) (*types.StoredSession, error) {
	r.logger.Debug("getting session", zap.String("key", key))

	session, err := r.queries.GetSession(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &types.StoredSession{
		Key:       session.Key,
		Value:     session.Value,
		ExpiresAt: session.ExpiresAt.Time,
	}, nil
}
