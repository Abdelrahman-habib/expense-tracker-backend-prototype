package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

// DeleteWallet deletes a wallet
func (r *WalletRepositoryImpl) DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	err := r.db.DeleteWallet(ctx, db.DeleteWalletParams{
		WalletID: walletID,
		UserID:   userID,
	})
	if err != nil {

		return errors.HandleRepositoryError(err, "delete", "wallet")

	}
	return nil
}
