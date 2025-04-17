package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// ListWallets retrieves a paginated list of wallets for a user
func (r *WalletRepositoryImpl) ListWallets(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Wallet, error) {
	wallets, err := r.db.ListWallets(ctx, db.ListWalletsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return []types.Wallet{}, errors.HandleRepositoryError(err, "list", "wallets")
	}

	return toWallets(wallets), nil
}

// ListWalletsPaginated retrieves a cursor-based paginated list of wallets
func (r *WalletRepositoryImpl) ListWalletsPaginated(ctx context.Context, userID uuid.UUID, createdAt time.Time, walletID uuid.UUID, limit int32) ([]types.Wallet, error) {
	wallets, err := r.db.ListWalletsPaginated(ctx, db.ListWalletsPaginatedParams{
		UserID:    userID,
		CreatedAt: utils.ToNullableTimestamp(&createdAt),
		WalletID:  walletID,
		Limit:     limit,
	})
	if err != nil {
		return []types.Wallet{}, errors.HandleRepositoryError(err, "p-list", "wallets")
	}

	return toWallets(wallets), nil
}
