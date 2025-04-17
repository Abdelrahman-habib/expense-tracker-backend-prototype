package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WalletService interface {
	GetWallet(ctx context.Context, walletID, userID uuid.UUID) (types.Wallet, error)
	ListWallets(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Wallet, error)
	ListWalletsPaginated(ctx context.Context, userID uuid.UUID, createdAt time.Time, walletID uuid.UUID, limit int32) ([]types.Wallet, error)
	CreateWallet(ctx context.Context, payload types.WalletCreatePayload, userID uuid.UUID) (types.Wallet, error)
	UpdateWallet(ctx context.Context, payload types.WalletUpdatePayload, userID uuid.UUID) (types.Wallet, error)
	DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error
	GetProjectWallets(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]types.Wallet, error)
	SearchWallets(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Wallet, error)
}

type walletService struct {
	repo   repository.WalletRepository
	logger *zap.Logger
}

func NewWalletService(repo repository.WalletRepository, logger *zap.Logger) WalletService {
	return &walletService{
		repo:   repo,
		logger: logger.With(zap.String("component", "wallet_service")),
	}
}

// Common validation function
func validateWallet(name, currency string, balance *float64, tags []uuid.UUID) error {
	if name == "" {
		return fmt.Errorf("wallet name is required")
	}

	if len(name) > types.MaxNameLength {
		return fmt.Errorf("name exceeds maximum length")
	}

	if currency == "" {
		return fmt.Errorf("currency is required")
	}

	if len(currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter ISO code")
	}

	if balance != nil && *balance < 0 {
		return fmt.Errorf("balance cannot be negative")
	}

	if len(tags) > types.MaxTagsCount {
		return fmt.Errorf("number of tags exceeds maximum allowed")
	}

	return nil
}

func (s *walletService) GetWallet(ctx context.Context, walletID, userID uuid.UUID) (types.Wallet, error) {
	s.logger.Info("getting wallet",
		zap.String("wallet_id", walletID.String()),
		zap.String("user_id", userID.String()))
	return s.repo.GetWallet(ctx, walletID, userID)
}

func (s *walletService) ListWallets(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Wallet, error) {
	s.logger.Info("listing wallets",
		zap.String("user_id", userID.String()),
		zap.Int32("limit", limit),
		zap.Int32("offset", offset))
	return s.repo.ListWallets(ctx, userID, limit, offset)
}

func (s *walletService) ListWalletsPaginated(ctx context.Context, userID uuid.UUID, createdAt time.Time, walletID uuid.UUID, limit int32) ([]types.Wallet, error) {
	s.logger.Info("listing paginated wallets",
		zap.String("user_id", userID.String()),
		zap.Time("cursor", createdAt),
		zap.String("cursor_id", walletID.String()),
		zap.Int32("limit", limit))

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	return s.repo.ListWalletsPaginated(ctx, userID, createdAt, walletID, limit)
}

func (s *walletService) CreateWallet(ctx context.Context, payload types.WalletCreatePayload, userID uuid.UUID) (types.Wallet, error) {
	s.logger.Info("creating wallet",
		zap.String("user_id", userID.String()),
		zap.String("name", payload.Name))

	if err := validateWallet(payload.Name, payload.Currency, payload.Balance, payload.Tags); err != nil {
		return types.Wallet{}, err
	}

	return s.repo.CreateWallet(ctx, payload, userID)
}

func (s *walletService) UpdateWallet(ctx context.Context, payload types.WalletUpdatePayload, userID uuid.UUID) (types.Wallet, error) {
	s.logger.Info("updating wallet",
		zap.String("wallet_id", payload.WalletID.String()),
		zap.String("user_id", userID.String()))

	if err := validateWallet(payload.Name, payload.Currency, payload.Balance, payload.Tags); err != nil {
		return types.Wallet{}, err
	}

	return s.repo.UpdateWallet(ctx, payload, userID)
}

func (s *walletService) DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	s.logger.Info("deleting wallet",
		zap.String("wallet_id", walletID.String()),
		zap.String("user_id", userID.String()))
	return s.repo.DeleteWallet(ctx, walletID, userID)
}

func (s *walletService) GetProjectWallets(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]types.Wallet, error) {
	s.logger.Info("getting project wallets",
		zap.String("project_id", projectID.String()),
		zap.String("user_id", userID.String()))
	return s.repo.GetProjectWallets(ctx, projectID, userID)
}

func (s *walletService) SearchWallets(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Wallet, error) {
	s.logger.Info("searching wallets",
		zap.String("user_id", userID.String()),
		zap.String("query", name),
		zap.Int32("limit", limit))

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	return s.repo.SearchWallets(ctx, userID, name, limit)
}
