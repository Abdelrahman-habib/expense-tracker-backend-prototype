package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// GetProjectWallets retrieves all wallets associated with a project
func (r *WalletRepositoryImpl) GetProjectWallets(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]types.Wallet, error) {
	wallets, err := r.db.GetProjectWallets(ctx, db.GetProjectWalletsParams{
		ProjectID: utils.ToNullableUUID(projectID),
		UserID:    userID,
	})
	if err != nil {
		return []types.Wallet{}, errors.HandleRepositoryError(err, "get project", "wallet(s)")
	}

	return toWallets(wallets), nil
}
