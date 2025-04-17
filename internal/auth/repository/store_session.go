package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// StoreSession stores a session value with expiration
func (r *authRepository) StoreSession(ctx context.Context, key string, value interface{}, expiresAt time.Time) error {
	r.logger.Debug("storing session", zap.String("key", key))

	// Convert value to JSON bytes
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal session value: %w", err)
	}

	_, err = r.queries.UpsertSession(ctx, db.UpsertSessionParams{
		Key:       key,
		Value:     valueBytes,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	return nil
}
