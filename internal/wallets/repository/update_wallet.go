package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// UpdateWallet updates an existing wallet
func (r *WalletRepositoryImpl) UpdateWallet(ctx context.Context, payload types.WalletUpdatePayload, userID uuid.UUID) (types.Wallet, error) {
	if payload.WalletID == uuid.Nil || userID == uuid.Nil {
		return types.Wallet{}, fmt.Errorf("invalid wallet id or user id")
	}

	params := updateWalletParamsFromPayload(payload, userID)
	wallet, err := r.db.UpdateWallet(ctx, params)
	if err != nil {
		return types.Wallet{}, errors.HandleRepositoryError(err, "update", "wallet")
	}

	return toWallet(wallet), nil
}
