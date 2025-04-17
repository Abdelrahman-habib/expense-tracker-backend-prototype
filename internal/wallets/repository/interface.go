package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
)

// WalletRepository defines the interface for wallet data access operations
type WalletRepository interface {
	// GetWallet retrieves a wallet by its ID and user ID
	GetWallet(ctx context.Context, walletID, userID uuid.UUID) (types.Wallet, error)

	// ListWallets retrieves a paginated list of wallets for a user
	ListWallets(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Wallet, error)

	// ListWalletsPaginated retrieves a cursor-based paginated list of wallets
	ListWalletsPaginated(ctx context.Context, userID uuid.UUID, createdAt time.Time, walletID uuid.UUID, limit int32) ([]types.Wallet, error)

	// CreateWallet creates a new wallet
	CreateWallet(ctx context.Context, payload types.WalletCreatePayload, userID uuid.UUID) (types.Wallet, error)

	// UpdateWallet updates an existing wallet
	UpdateWallet(ctx context.Context, payload types.WalletUpdatePayload, userID uuid.UUID) (types.Wallet, error)

	// DeleteWallet deletes a wallet
	DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error

	// GetProjectWallets retrieves all wallets associated with a project
	GetProjectWallets(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]types.Wallet, error)

	// SearchWallets searches for wallets by name
	SearchWallets(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Wallet, error)
}
