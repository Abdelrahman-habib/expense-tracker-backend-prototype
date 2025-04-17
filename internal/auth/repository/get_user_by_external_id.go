package repository

import (
	"context"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"go.uber.org/zap"
)

// GetUserByExternalID retrieves a user by their external OAuth ID
func (r *authRepository) GetUserByExternalID(ctx context.Context, externalID, provider string) (*types.AuthUser, error) {
	r.logger.Debug("getting user by external ID",
		zap.String("external_id", externalID),
		zap.String("provider", provider),
	)

	user, err := r.queries.GetUserByExternalID(ctx, db.GetUserByExternalIDParams{
		ExternalID: externalID,
		Provider:   provider,
	})
	if err != nil {
		return nil, err
	}

	return &types.AuthUser{
		ID:       user.UserID,
		Name:     user.Name,
		Email:    user.Email,
		Provider: user.Provider,
	}, nil
}
