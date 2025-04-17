package repository

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

// WalletRepositoryImpl implements WalletRepository interface
type WalletRepositoryImpl struct {
	db *db.Queries
}

// NewWalletRepository creates a new instance of WalletRepository
func NewWalletRepository(queries *db.Queries) WalletRepository {
	return &WalletRepositoryImpl{
		db: queries,
	}
}
