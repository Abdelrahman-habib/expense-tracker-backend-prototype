package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock repository
type mockWalletRepository struct {
	mock.Mock
}

func (m *mockWalletRepository) GetWallet(ctx context.Context, walletID, userID uuid.UUID) (types.Wallet, error) {
	args := m.Called(ctx, walletID, userID)
	return args.Get(0).(types.Wallet), args.Error(1)
}

func (m *mockWalletRepository) ListWallets(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Wallet, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func (m *mockWalletRepository) ListWalletsPaginated(ctx context.Context, userID uuid.UUID, createdAt time.Time, walletID uuid.UUID, limit int32) ([]types.Wallet, error) {
	args := m.Called(ctx, userID, createdAt, walletID, limit)
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func (m *mockWalletRepository) CreateWallet(ctx context.Context, payload types.WalletCreatePayload, userID uuid.UUID) (types.Wallet, error) {
	args := m.Called(ctx, payload, userID)
	return args.Get(0).(types.Wallet), args.Error(1)
}

func (m *mockWalletRepository) UpdateWallet(ctx context.Context, payload types.WalletUpdatePayload, userID uuid.UUID) (types.Wallet, error) {
	args := m.Called(ctx, payload, userID)
	return args.Get(0).(types.Wallet), args.Error(1)
}

func (m *mockWalletRepository) DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	args := m.Called(ctx, walletID, userID)
	return args.Error(0)
}

func (m *mockWalletRepository) GetProjectWallets(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]types.Wallet, error) {
	args := m.Called(ctx, projectID, userID)
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func (m *mockWalletRepository) SearchWallets(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Wallet, error) {
	args := m.Called(ctx, userID, name, limit)
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func setupTest(t *testing.T) (*mockWalletRepository, WalletService) {
	mockRepo := new(mockWalletRepository)
	logger := zap.NewNop()
	service := NewWalletService(mockRepo, logger)
	return mockRepo, service
}

func TestWalletService_CreateWallet(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		payload types.WalletCreatePayload
		mock    func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful create",
			payload: types.WalletCreatePayload{
				Name:     "New Wallet",
				Currency: "USD",
			},
			mock: func() {
				mockRepo.On("CreateWallet", ctx, mock.AnythingOfType("types.WalletCreatePayload"), userID).
					Return(types.Wallet{Name: "New Wallet"}, nil)
			},
			wantErr: false,
		},
		{
			name: "empty name",
			payload: types.WalletCreatePayload{
				Name:     "",
				Currency: "USD",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "wallet name is required",
		},
		{
			name: "invalid currency length",
			payload: types.WalletCreatePayload{
				Name:     "Test Wallet",
				Currency: "USDD",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "currency must be a 3-letter ISO code",
		},
		{
			name: "negative balance",
			payload: types.WalletCreatePayload{
				Name:     "Test Wallet",
				Currency: "USD",
				Balance:  float64Ptr(-100.0),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "balance cannot be negative",
		},
		{
			name: "too many tags",
			payload: types.WalletCreatePayload{
				Name:     "Test Wallet",
				Currency: "USD",
				Tags:     make([]uuid.UUID, types.MaxTagsCount+1),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "number of tags exceeds maximum allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			wallet, err := service.CreateWallet(ctx, tt.payload, userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, wallet)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_GetWallet(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful retrieval",
			mock: func() {
				expectedWallet := types.Wallet{
					WalletID: walletID,
					Name:     "Test Wallet",
					Currency: "USD",
				}
				mockRepo.On("GetWallet", ctx, walletID, userID).Return(expectedWallet, nil)
			},
			wantErr: false,
		},
		{
			name: "not found error",
			mock: func() {
				mockRepo.On("GetWallet", ctx, walletID, userID).Return(types.Wallet{}, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			wallet, err := service.GetWallet(ctx, walletID, userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, walletID, wallet.WalletID)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_ListWallets(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		limit   int32
		offset  int32
		mock    func()
		wantErr bool
		wantLen int
	}{
		{
			name:   "successful list",
			limit:  10,
			offset: 0,
			mock: func() {
				wallets := []types.Wallet{
					{
						WalletID: uuid.New(),
						Name:     "Wallet 1",
						Currency: "USD",
					},
					{
						WalletID: uuid.New(),
						Name:     "Wallet 2",
						Currency: "EUR",
					},
				}
				mockRepo.On("ListWallets", ctx, userID, int32(10), int32(0)).Return(wallets, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:   "empty list",
			limit:  10,
			offset: 0,
			mock: func() {
				mockRepo.On("ListWallets", ctx, userID, int32(10), int32(0)).Return([]types.Wallet{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:   "repository error",
			limit:  10,
			offset: 0,
			mock: func() {
				mockRepo.On("ListWallets", ctx, userID, int32(10), int32(0)).Return([]types.Wallet{}, errors.New("database error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			wallets, err := service.ListWallets(ctx, userID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, wallets, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_UpdateWallet(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()

	tests := []struct {
		name    string
		payload types.WalletUpdatePayload
		mock    func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful update",
			payload: types.WalletUpdatePayload{
				WalletID: walletID,
				Name:     "Updated Wallet",
				Currency: "EUR",
			},
			mock: func() {
				mockRepo.On("UpdateWallet", ctx, mock.AnythingOfType("types.WalletUpdatePayload"), userID).
					Return(types.Wallet{Name: "Updated Wallet"}, nil)
			},
			wantErr: false,
		},
		{
			name: "empty name",
			payload: types.WalletUpdatePayload{
				WalletID: walletID,
				Name:     "",
				Currency: "USD",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "wallet name is required",
		},
		{
			name: "invalid currency",
			payload: types.WalletUpdatePayload{
				WalletID: walletID,
				Name:     "Test Wallet",
				Currency: "INVALID",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "currency must be a 3-letter ISO code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			wallet, err := service.UpdateWallet(ctx, tt.payload, userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, wallet)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_ListWalletsPaginated(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()
	cursorID := uuid.New()

	tests := []struct {
		name     string
		cursor   time.Time
		cursorID uuid.UUID
		limit    int32
		mock     func()
		wantErr  bool
		wantLen  int
	}{
		{
			name:     "successful pagination",
			cursor:   now,
			cursorID: cursorID,
			limit:    10,
			mock: func() {
				wallets := []types.Wallet{
					{
						WalletID:  uuid.New(),
						Name:      "Wallet 1",
						Currency:  "USD",
						CreatedAt: now.Add(-1 * time.Hour),
					},
					{
						WalletID:  uuid.New(),
						Name:      "Wallet 2",
						Currency:  "EUR",
						CreatedAt: now.Add(-2 * time.Hour),
					},
				}
				mockRepo.On("ListWalletsPaginated", ctx, userID, now, cursorID, int32(10)).
					Return(wallets, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:     "invalid limit",
			cursor:   now,
			cursorID: cursorID,
			limit:    -1,
			mock:     func() {},
			wantErr:  true,
		},
		{
			name:     "empty result",
			cursor:   now,
			cursorID: cursorID,
			limit:    10,
			mock: func() {
				mockRepo.On("ListWalletsPaginated", ctx, userID, now, cursorID, int32(10)).
					Return([]types.Wallet{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			wallets, err := service.ListWalletsPaginated(ctx, userID, tt.cursor, tt.cursorID, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, wallets, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_DeleteWallet(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful delete",
			mock: func() {
				mockRepo.On("DeleteWallet", ctx, walletID, userID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found error",
			mock: func() {
				mockRepo.On("DeleteWallet", ctx, walletID, userID).Return(errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			err := service.DeleteWallet(ctx, walletID, userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_GetProjectWallets(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
		wantLen int
	}{
		{
			name: "successful retrieval",
			mock: func() {
				wallets := []types.Wallet{
					{
						WalletID:  uuid.New(),
						Name:      "Project Wallet 1",
						Currency:  "USD",
						CreatedAt: time.Now(),
					},
					{
						WalletID:  uuid.New(),
						Name:      "Project Wallet 2",
						Currency:  "EUR",
						CreatedAt: time.Now(),
					},
				}
				mockRepo.On("GetProjectWallets", ctx, projectID, userID).Return(wallets, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "empty result",
			mock: func() {
				mockRepo.On("GetProjectWallets", ctx, projectID, userID).Return([]types.Wallet{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "repository error",
			mock: func() {
				mockRepo.On("GetProjectWallets", ctx, projectID, userID).Return([]types.Wallet{}, errors.New("database error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			wallets, err := service.GetProjectWallets(ctx, projectID, userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, wallets, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWalletService_SearchWallets(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		query   string
		limit   int32
		mock    func()
		wantErr bool
		wantLen int
		errMsg  string
	}{
		{
			name:  "successful search",
			query: "test",
			limit: 10,
			mock: func() {
				wallets := []types.Wallet{
					{
						WalletID:  uuid.New(),
						Name:      "Test Wallet 1",
						Currency:  "USD",
						CreatedAt: time.Now(),
					},
					{
						WalletID:  uuid.New(),
						Name:      "Test Wallet 2",
						Currency:  "EUR",
						CreatedAt: time.Now(),
					},
				}
				mockRepo.On("SearchWallets", ctx, userID, "test", int32(10)).Return(wallets, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "invalid limit",
			query:   "test",
			limit:   -1,
			mock:    func() {},
			wantErr: true,
			errMsg:  "limit must be positive",
		},
		{
			name:  "empty result",
			query: "nonexistent",
			limit: 10,
			mock: func() {
				mockRepo.On("SearchWallets", ctx, userID, "nonexistent", int32(10)).Return([]types.Wallet{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:  "repository error",
			query: "test",
			limit: 10,
			mock: func() {
				mockRepo.On("SearchWallets", ctx, userID, "test", int32(10)).Return([]types.Wallet{}, errors.New("database error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			wallets, err := service.SearchWallets(ctx, userID, tt.query, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, wallets, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create float64 pointer
func float64Ptr(v float64) *float64 {
	return &v
}
