package repository

import (
	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// toWallet converts a db.Wallet to domain types.Wallet
func toWallet(w db.Wallet) types.Wallet {
	return types.Wallet{
		WalletID:  w.WalletID,
		UserID:    w.UserID,
		ProjectID: utils.GetUUIDPtr(w.ProjectID),
		Name:      w.Name,
		Balance:   utils.GetFloat64Ptr(w.Balance),
		Currency:  w.Currency,
		Tags:      w.Tags,
		CreatedAt: w.CreatedAt.Time,
		UpdatedAt: w.UpdatedAt.Time,
	}
}

// toWallets converts a slice of db.Wallet to a slice of domain types.Wallet
func toWallets(wallets []db.Wallet) []types.Wallet {
	result := make([]types.Wallet, len(wallets))
	for i, w := range wallets {
		result[i] = toWallet(w)
	}
	return result
}

// createWalletParamsFromPayload converts WalletCreatePayload to db.CreateWalletParams
func createWalletParamsFromPayload(payload types.WalletCreatePayload, userID uuid.UUID) db.CreateWalletParams {
	return db.CreateWalletParams{
		UserID:    userID,
		ProjectID: utils.UUIDToNullableUUID(payload.ProjectID),
		Name:      payload.Name,
		Balance:   utils.ToNullableNumeric(payload.Balance),
		Currency:  payload.Currency,
		Tags:      payload.Tags,
	}
}

// updateWalletParamsFromPayload converts WalletUpdatePayload to db.UpdateWalletParams
func updateWalletParamsFromPayload(payload types.WalletUpdatePayload, userID uuid.UUID) db.UpdateWalletParams {
	return db.UpdateWalletParams{
		WalletID: payload.WalletID,
		UserID:   userID,
		Name:     utils.ToNullableText(&payload.Name),
		Balance:  utils.ToNullableNumeric(payload.Balance),
		Currency: utils.ToNullableText(&payload.Currency),
		Tags:     payload.Tags,
	}
}
