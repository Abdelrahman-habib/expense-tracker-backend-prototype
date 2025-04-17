package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// CreateWallet creates a new wallet
func (r *WalletRepositoryImpl) CreateWallet(ctx context.Context, payload types.WalletCreatePayload, userID uuid.UUID) (types.Wallet, error) {
	params := createWalletParamsFromPayload(payload, userID)
	wallet, err := r.db.CreateWallet(ctx, params)
	if err != nil {
		return types.Wallet{}, errors.HandleRepositoryError(err, "create", "wallet")
	}

	return toWallet(wallet), nil
}
