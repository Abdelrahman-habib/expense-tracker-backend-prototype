package repository

import (
	"context"
	"fmt"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"go.uber.org/zap"
)

// CreateUser creates a new user from OAuth data
func (r *authRepository) CreateUser(ctx context.Context, userData types.OAuthUserData) (*types.AuthUser, error) {
	r.logger.Debug("creating user",
		zap.String("name", userData.Name),
		zap.String("email", userData.Email),
		zap.String("provider", userData.Provider),
	)

	user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Name:       userData.Name,
		Email:      userData.Email,
		ExternalID: userData.ExternalID,
		Provider:   userData.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &types.AuthUser{
		ID:       user.UserID,
		Name:     user.Name,
		Email:    user.Email,
		Provider: user.Provider,
	}, nil
}
