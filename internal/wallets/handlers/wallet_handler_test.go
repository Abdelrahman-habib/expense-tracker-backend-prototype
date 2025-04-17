package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	coreTypes "github.com/Abdelrahman-habib/expense-tracker/internal/core/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock service
type mockWalletService struct {
	mock.Mock
}

func (m *mockWalletService) GetWallet(ctx context.Context, walletID, userID uuid.UUID) (types.Wallet, error) {
	args := m.Called(ctx, walletID, userID)
	return args.Get(0).(types.Wallet), args.Error(1)
}

func (m *mockWalletService) ListWallets(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Wallet, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func (m *mockWalletService) ListWalletsPaginated(ctx context.Context, userID uuid.UUID, createdAt time.Time, walletID uuid.UUID, limit int32) ([]types.Wallet, error) {
	args := m.Called(ctx, userID, createdAt, walletID, limit)
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func (m *mockWalletService) CreateWallet(ctx context.Context, payload types.WalletCreatePayload, userID uuid.UUID) (types.Wallet, error) {
	args := m.Called(ctx, payload, userID)
	return args.Get(0).(types.Wallet), args.Error(1)
}

func (m *mockWalletService) UpdateWallet(ctx context.Context, payload types.WalletUpdatePayload, userID uuid.UUID) (types.Wallet, error) {
	args := m.Called(ctx, payload, userID)
	return args.Get(0).(types.Wallet), args.Error(1)
}

func (m *mockWalletService) DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	args := m.Called(ctx, walletID, userID)
	return args.Error(0)
}

func (m *mockWalletService) GetProjectWallets(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]types.Wallet, error) {
	args := m.Called(ctx, projectID, userID)
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func (m *mockWalletService) SearchWallets(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Wallet, error) {
	args := m.Called(ctx, userID, name, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Wallet), args.Error(1)
}

func setupTest(t *testing.T) (*mockWalletService, *WalletHandler) {
	mockService := new(mockWalletService)
	logger := zap.NewNop()
	handler := NewWalletHandler(mockService, logger)
	return mockService, handler
}

// Helper function to create float64 pointer
func float64Ptr(v float64) *float64 {
	return &v
}

func TestWalletHandler_CreateWallet(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		payload        string
		setupAuth      bool
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "successful creation",
			payload: `{
				"name": "Test Wallet",
				"currency": "USD",
				"balance": 100.50
			}`,
			setupAuth: true,
			setupMock: func() {
				expectedWallet := types.Wallet{
					WalletID: uuid.New(),
					Name:     "Test Wallet",
					Currency: "USD",
					Balance:  float64Ptr(100.50),
				}
				mockService.On("CreateWallet", mock.Anything, mock.AnythingOfType("types.WalletCreatePayload"), userID).
					Return(expectedWallet, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid payload",
			payload: `{
				"name": "",
				"currency": "INVALID"
			}`,
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing auth",
			payload:        `{}`,
			setupAuth:      false,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.CreateWallet(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(http.StatusCreated), response["status"])
				assert.NotNil(t, response["data"])
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestWalletHandler_GetWallet(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	walletID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		walletID       string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:      "successful retrieval",
			setupAuth: true,
			walletID:  walletID.String(),
			setupMock: func() {
				expectedWallet := types.Wallet{
					WalletID: walletID,
					Name:     "Test Wallet",
					Currency: "USD",
				}
				mockService.On("GetWallet", mock.Anything, walletID, userID).
					Return(expectedWallet, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid wallet ID",
			setupAuth:      true,
			walletID:       "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing auth",
			setupAuth:      false,
			walletID:       walletID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodGet, "/wallets/"+tt.walletID, nil)

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.walletID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.GetWallet(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(http.StatusOK), response["status"])
				assert.NotNil(t, response["data"])
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestWalletHandler_ListWalletsPaginated(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	now := time.Now().UTC()
	cursorID := uuid.New()

	tests := []struct {
		name            string
		setupAuth       bool
		queryParams     map[string]string
		setupMock       func()
		expectedStatus  int
		expectedLen     int
		expectedLimit   string
		expectNextToken bool
		expectedError   string
	}{
		{
			name:        "first page with default values",
			setupAuth:   true,
			queryParams: map[string]string{},
			setupMock: func() {
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
				mockService.On("ListWalletsPaginated",
					mock.Anything,
					userID,
					mock.MatchedBy(func(t time.Time) bool {
						return time.Since(t) < time.Minute
					}),
					mock.MatchedBy(func(id uuid.UUID) bool {
						return id == uuid.Nil
					}),
					int32(coreTypes.DefaultLimit),
				).Return(wallets, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
			expectedLimit:  fmt.Sprint(coreTypes.DefaultLimit),
		},
		{
			name:      "first page with custom limit",
			setupAuth: true,
			queryParams: map[string]string{
				"limit": "5",
			},
			setupMock: func() {
				wallets := []types.Wallet{
					{
						WalletID:  uuid.New(),
						Name:      "Wallet 1",
						Currency:  "USD",
						CreatedAt: now.Add(-1 * time.Hour),
					},
				}
				mockService.On("ListWalletsPaginated",
					mock.Anything,
					userID,
					mock.MatchedBy(func(t time.Time) bool {
						return time.Since(t) < time.Minute
					}),
					mock.MatchedBy(func(id uuid.UUID) bool {
						return id == uuid.Nil
					}),
					int32(5),
				).Return(wallets, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    1,
			expectedLimit:  "5",
		},
		{
			name:      "second page with next_token",
			setupAuth: true,
			queryParams: map[string]string{
				"next_token": coreTypes.EncodeCursor(now, cursorID),
			},
			setupMock: func() {
				wallets := []types.Wallet{
					{
						WalletID:  uuid.New(),
						Name:      "Wallet 3",
						Currency:  "USD",
						CreatedAt: now.Add(-3 * time.Hour),
					},
				}
				mockService.On("ListWalletsPaginated",
					mock.Anything,
					userID,
					mock.MatchedBy(func(t time.Time) bool {
						return t.Truncate(time.Second).Equal(now.Truncate(time.Second))
					}),
					cursorID,
					int32(coreTypes.DefaultLimit),
				).Return(wallets, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    1,
			expectedLimit:  fmt.Sprint(coreTypes.DefaultLimit),
		},
		{
			name:      "invalid next_token format",
			setupAuth: true,
			queryParams: map[string]string{
				"next_token": "invalid-token",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid token",
		},
		{
			name:      "limit below minimum",
			setupAuth: true,
			queryParams: map[string]string{
				"limit": "0",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "limit: must be no less than 1",
		},
		{
			name:      "limit above maximum gets capped",
			setupAuth: true,
			queryParams: map[string]string{
				"limit": fmt.Sprintf("%d", coreTypes.MaxLimit+1),
			},
			setupMock: func() {
				wallets := []types.Wallet{
					{
						WalletID:  uuid.New(),
						Name:      "Wallet 1",
						Currency:  "USD",
						CreatedAt: now.Add(-1 * time.Hour),
					},
				}
				mockService.On("ListWalletsPaginated",
					mock.Anything,
					userID,
					mock.Anything,
					mock.Anything,
					int32(coreTypes.MaxLimit),
				).Return(wallets, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    1,
			expectedLimit:  fmt.Sprint(coreTypes.MaxLimit),
		},
		{
			name:           "missing auth",
			setupAuth:      false,
			queryParams:    map[string]string{},
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			reqURL := "/wallets/paginated"
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for k, v := range tt.queryParams {
					values.Add(k, v)
				}
				reqURL += "?" + values.Encode()
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.ListWalletsPaginated(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, float64(http.StatusOK), response["status"])
				assert.Equal(t, "Success", response["message"])

				wallets := response["data"].([]interface{})
				assert.Len(t, wallets, tt.expectedLen)

				meta := response["meta"].(map[string]interface{})
				if tt.expectedLimit != "" {
					assert.Equal(t, tt.expectedLimit, fmt.Sprint(meta["limit"]))
				}

				if tt.expectNextToken {
					assert.NotEmpty(t, meta["next_token"])
				} else {
					nextToken, exists := meta["next_token"]
					if exists {
						assert.Empty(t, nextToken)
					}
				}
			} else if tt.expectedError != "" {
				errMsg, ok := response["error"].(string)
				assert.True(t, ok)
				assert.Contains(t, errMsg, tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestWalletHandler_SearchWallets(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		queryParams    map[string]string
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:      "valid request with all parameters",
			setupAuth: true,
			queryParams: map[string]string{
				"q":     "test",
				"limit": "20",
			},
			setupMock: func() {
				wallets := []types.Wallet{
					{WalletID: uuid.New(), Name: "Test Wallet"},
					{WalletID: uuid.New(), Name: "Testing Account"},
				}
				mockService.On("SearchWallets", mock.Anything, userID, "test", int32(20)).
					Return(wallets, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				metadata := response["meta"].(map[string]interface{})
				assert.Equal(t, "test", metadata["query"])
				assert.Equal(t, float64(20), metadata["limit"])
				assert.Equal(t, float64(2), metadata["count"])
			},
		},
		{
			name:      "limit exceeds maximum will be capped to maximum",
			setupAuth: true,
			queryParams: map[string]string{
				"q":     "test",
				"limit": fmt.Sprint(coreTypes.MaxSearchLimit), // > maxSearchLimit
			},
			setupMock: func() {
				wallets := []types.Wallet{}
				mockService.On("SearchWallets", mock.Anything, userID, "test", int32(coreTypes.MaxSearchLimit)).
					Return(wallets, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				metadata := response["meta"].(map[string]interface{})
				assert.Equal(t, float64(coreTypes.MaxSearchLimit), metadata["limit"])
			},
		},
		{
			name:      "query too long",
			setupAuth: true,
			queryParams: map[string]string{
				"q": strings.Repeat("a", coreTypes.MaxQueryLength+1),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "negative limit",
			setupAuth: true,
			queryParams: map[string]string{
				"q":     "test",
				"limit": "-1",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "service returns error",
			setupAuth: true,
			queryParams: map[string]string{
				"q": "test",
			},
			setupMock: func() {
				mockService.On("SearchWallets", mock.Anything, userID, "test", int32(coreTypes.DefaultSearchLimit)).
					Return([]types.Wallet(nil), fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			reqURL := "/wallets/search"
			if len(tt.queryParams) > 0 {
				params := make([]string, 0, len(tt.queryParams))
				for k, v := range tt.queryParams {
					params = append(params, k+"="+url.QueryEscape(v))
				}
				reqURL += "?" + strings.Join(params, "&")
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.SearchWallets(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK && tt.checkResponse != nil {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestWalletHandler_GetProjectWallets(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		projectID      string
		setupMock      func()
		expectedStatus int
		expectedLen    int
	}{
		{
			name:      "successful retrieval",
			setupAuth: true,
			projectID: projectID.String(),
			setupMock: func() {
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
				mockService.On("GetProjectWallets", mock.Anything, projectID, userID).
					Return(wallets, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:           "invalid project ID",
			setupAuth:      true,
			projectID:      "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing auth",
			setupAuth:      false,
			projectID:      projectID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodGet, "/projects/"+tt.projectID+"/wallets", nil)

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("project_id", tt.projectID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.GetProjectWallets(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(http.StatusOK), response["status"])
				assert.Equal(t, "Success", response["message"])

				wallets := response["data"].([]interface{})
				assert.Len(t, wallets, tt.expectedLen)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestWalletHandler_UpdateWallet(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	walletID := uuid.New()

	tests := []struct {
		name           string
		walletID       string
		payload        string
		setupAuth      bool
		setupMock      func()
		expectedStatus int
	}{
		{
			name:     "successful update",
			walletID: walletID.String(),
			payload: `{
				"name": "Updated Wallet",
				"currency": "EUR",
				"balance": 200.50
			}`,
			setupAuth: true,
			setupMock: func() {
				existingWallet := types.Wallet{
					WalletID: walletID,
					Name:     "Original Wallet",
					Currency: "USD",
					Balance:  float64Ptr(100.50),
				}
				updatedWallet := types.Wallet{
					WalletID: walletID,
					Name:     "Updated Wallet",
					Currency: "EUR",
					Balance:  float64Ptr(200.50),
				}
				mockService.On("GetWallet", mock.Anything, walletID, userID).
					Return(existingWallet, nil)
				mockService.On("UpdateWallet", mock.Anything, mock.AnythingOfType("types.WalletUpdatePayload"), userID).
					Return(updatedWallet, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid wallet ID",
			walletID:       "invalid-uuid",
			payload:        `{}`,
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "wallet not found",
			walletID:  uuid.New().String(),
			payload:   `{}`,
			setupAuth: true,
			setupMock: func() {
				mockService.On("GetWallet", mock.Anything, mock.AnythingOfType("uuid.UUID"), userID).
					Return(types.Wallet{}, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing auth",
			walletID:       walletID.String(),
			payload:        `{}`,
			setupAuth:      false,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPut, "/wallets/"+tt.walletID, strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.walletID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.UpdateWallet(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(http.StatusOK), response["status"])
				assert.NotNil(t, response["data"])
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestWalletHandler_DeleteWallet(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	walletID := uuid.New()

	tests := []struct {
		name           string
		walletID       string
		setupAuth      bool
		setupMock      func()
		expectedStatus int
	}{
		{
			name:      "successful deletion",
			walletID:  walletID.String(),
			setupAuth: true,
			setupMock: func() {
				mockService.On("DeleteWallet", mock.Anything, walletID, userID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid wallet ID",
			walletID:       "invalid-uuid",
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing auth",
			walletID:       walletID.String(),
			setupAuth:      false,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodDelete, "/wallets/"+tt.walletID, nil)

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.walletID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.DeleteWallet(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(http.StatusOK), response["status"])
			}
			mockService.AssertExpectations(t)
		})
	}
}
