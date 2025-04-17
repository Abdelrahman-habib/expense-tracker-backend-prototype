package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// GetWallet retrieves a wallet by its ID and user ID
func (r *WalletRepositoryImpl) GetWallet(ctx context.Context, walletID, userID uuid.UUID) (types.Wallet, error) {
	wallet, err := r.db.GetWallet(ctx, db.GetWalletParams{
		WalletID: walletID,
		UserID:   userID,
	})
	if err != nil {
		return types.Wallet{}, errors.HandleRepositoryError(err, "get", "wallet")
	}

	return toWallet(wallet), nil
}
