package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// SearchWallets searches for wallets by name
func (r *WalletRepositoryImpl) SearchWallets(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Wallet, error) {
	wallets, err := r.db.SearchWallets(ctx, db.SearchWalletsParams{
		UserID: userID,
		Name:   name,
		Limit:  limit,
	})
	if err != nil {
		return []types.Wallet{}, errors.HandleRepositoryError(err, "search", "wallet(s)")
	}

	return toWallets(wallets), nil
}
