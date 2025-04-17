package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/config"
	coreTypes "github.com/Abdelrahman-habib/expense-tracker/internal/core/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type WalletIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	service   db.Service
	pool      *pgxpool.Pool
	handler   *handlers.WalletHandler
	router    *chi.Mux
	userID    uuid.UUID
	ctx       context.Context
}

func TestWalletIntegrationSuite(t *testing.T) {
	suite.Run(t, new(WalletIntegrationTestSuite))
}

func (s *WalletIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.userID = uuid.New()

	var host, port string

	if os.Getenv("CI") == "true" {
		// Running in GitHub Actions, use service-based PostgreSQL
		host = "localhost"
		port = "5432"
	} else {
		// Running locally, use TestContainers
		req := testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
			Env: map[string]string{
				"POSTGRES_DB":       "testdb",
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
			},
			NetworkMode: "bridge",
		}

		container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(s.T(), err)
		s.container = container

		// Get container host and port
		host, err = container.Host(s.ctx)
		require.NoError(s.T(), err)
		mappedPort, err := container.MappedPort(s.ctx, "5432")
		require.NoError(s.T(), err)
		port = mappedPort.Port() // Extract numeric port
	}

	// Create database config
	cfg := config.DatabaseConfig{
		Host:        host,
		Port:        port,
		Username:    "test",
		Password:    "test",
		Database:    "testdb",
		Schema:      "public",
		MaxConns:    5,
		MinConns:    1,
		MaxLifetime: time.Hour,
		MaxIdleTime: time.Minute * 30,
		HealthCheck: time.Minute,
		SSLMode:     "disable",
		SearchPath:  "public",
	}

	// Initialize DB service
	dbService := db.NewService(cfg)
	s.service = dbService

	// Get connection pool
	pool, err := pgxpool.New(s.ctx, cfg.GetDSN())
	require.NoError(s.T(), err)
	s.pool = pool

	// Run migrations
	err = s.runMigrations()
	require.NoError(s.T(), err)

	// clear any previous runs data
	s.clearWallets()

	// Create test user
	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, 'wit_test_clerk_id', 'wit_Test User', 'wit_test@example.com')
	`, s.userID)
	require.NoError(s.T(), err)

	// Initialize components
	logger := zap.NewNop()
	repo := repository.NewWalletRepository(dbService.Queries())
	walletService := service.NewWalletService(repo, logger)
	s.handler = handlers.NewWalletHandler(walletService, logger)

	// Setup router
	router := chi.NewRouter()
	router.Route("/wallets", func(r chi.Router) {
		r.Get("/search", s.handler.SearchWallets)
		r.Get("/paginated", s.handler.ListWalletsPaginated)
		r.Post("/", s.handler.CreateWallet)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handler.GetWallet)
			r.Put("/", s.handler.UpdateWallet)
			r.Delete("/", s.handler.DeleteWallet)
		})
	})
	s.router = router
}

func (s *WalletIntegrationTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.service != nil {
		s.service.Close()
	}
	if s.container != nil && os.Getenv("CI") != "true" {
		err := s.container.Terminate(s.ctx)
		require.NoError(s.T(), err)
	}
}

func (s *WalletIntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	s.clearWallets()
}

func (s *WalletIntegrationTestSuite) runMigrations() error {
	migrationsDir := "../../db/sql/migrations"

	// Convert pool to *sql.DB for goose
	sqlDB := stdlib.OpenDBFromPool(s.pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return err
	}

	return nil
}

func float64Ptr(f float64) *float64 {
	return &f
}

func (s *WalletIntegrationTestSuite) clearWallets() {
	_, err := s.pool.Exec(s.ctx, "DELETE FROM wallets WHERE user_id = $1", s.userID)
	require.NoError(s.T(), err)
	_, err = s.pool.Exec(s.ctx, "DELETE FROM projects WHERE user_id = $1", s.userID)
	require.NoError(s.T(), err)
}

// Helper method to create a test wallet
func (s *WalletIntegrationTestSuite) createTestWallet() types.Wallet {
	createPayload := types.WalletCreatePayload{
		Name:     "Integration Test Wallet",
		Currency: "USD",
		Balance:  float64Ptr(1000.50),
	}

	payloadBytes, err := json.Marshal(createPayload)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)

	walletData := response["data"].(map[string]interface{})
	return types.Wallet{
		WalletID: uuid.MustParse(walletData["walletId"].(string)),
		Name:     walletData["name"].(string),
		Currency: walletData["currency"].(string),
	}
}

// Helper method for making authenticated requests
func (s *WalletIntegrationTestSuite) newAuthenticatedRequest(method, path string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, path, body)
	return req.WithContext(context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID))
}

// Helper method to verify wallet state
func (s *WalletIntegrationTestSuite) verifyWalletState(walletID uuid.UUID, expectedName, expectedCurrency string) {
	req := s.newAuthenticatedRequest(http.MethodGet, "/wallets/"+walletID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", walletID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)
	getData := response["data"].(map[string]interface{})
	s.Equal(expectedName, getData["name"])
	s.Equal(expectedCurrency, getData["currency"])
}

func (s *WalletIntegrationTestSuite) TestWalletLifecycle() {
	// Create a wallet and use it across all tests
	wallet := &types.Wallet{}
	*wallet = s.createTestWallet()

	s.Run("get wallet", func() {
		s.testGetWallet(wallet)
	})
	s.Run("update wallet name", func() {
		s.testUpdateWalletName(wallet)
	})
	s.Run("update wallet currency", func() {
		s.testUpdateWalletCurrency(wallet)
	})
	s.Run("delete wallet", func() {
		s.testDeleteWallet(wallet)
	})
}

func (s *WalletIntegrationTestSuite) testGetWallet(wallet *types.Wallet) {
	req := s.newAuthenticatedRequest(http.MethodGet, "/wallets/"+wallet.WalletID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", wallet.WalletID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)

	getData := response["data"].(map[string]interface{})
	s.Equal(wallet.Name, getData["name"])
	s.Equal(wallet.Currency, getData["currency"])
}

func (s *WalletIntegrationTestSuite) testUpdateWalletName(wallet *types.Wallet) {
	updatePayload := types.WalletUpdatePayload{
		WalletID: wallet.WalletID,
		Name:     "Updated Wallet Name",
		Currency: wallet.Currency,
	}

	payloadBytes, err := json.Marshal(updatePayload)
	s.Require().NoError(err)

	req := s.newAuthenticatedRequest(http.MethodPut, "/wallets/"+wallet.WalletID.String(), bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", wallet.WalletID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	wallet.Name = updatePayload.Name
	s.verifyWalletState(wallet.WalletID, wallet.Name, wallet.Currency)
}

func (s *WalletIntegrationTestSuite) testUpdateWalletCurrency(wallet *types.Wallet) {
	updatePayload := types.WalletUpdatePayload{
		WalletID: wallet.WalletID,
		Name:     wallet.Name,
		Currency: "EUR",
	}

	payloadBytes, err := json.Marshal(updatePayload)
	s.Require().NoError(err)

	req := s.newAuthenticatedRequest(http.MethodPut, "/wallets/"+wallet.WalletID.String(), bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", wallet.WalletID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	wallet.Currency = updatePayload.Currency
	s.verifyWalletState(wallet.WalletID, wallet.Name, wallet.Currency)
}

func (s *WalletIntegrationTestSuite) testDeleteWallet(wallet *types.Wallet) {
	req := s.newAuthenticatedRequest(http.MethodDelete, "/wallets/"+wallet.WalletID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", wallet.WalletID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Verify wallet is deleted
	req = s.newAuthenticatedRequest(http.MethodGet, "/wallets/"+wallet.WalletID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusNotFound, w.Code)
}

// Helper method to create multiple test wallets
func (s *WalletIntegrationTestSuite) createTestWallets(count int) []types.Wallet {
	wallets := make([]types.Wallet, count)
	// Create wallets in order (1 to count)
	for i := 0; i < count; i++ {
		createPayload := types.WalletCreatePayload{
			Name:     fmt.Sprintf("Test Wallet %d", i+1), // Start from 1 and increment
			Currency: "USD",
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		s.Require().Equal(http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		s.Require().NoError(err)

		walletData := response["data"].(map[string]interface{})
		createdAt, err := time.Parse(time.RFC3339, walletData["createdAt"].(string))
		s.Require().NoError(err)

		wallets[count-1-i] = types.Wallet{ // Store in reverse order
			WalletID:  uuid.MustParse(walletData["walletId"].(string)),
			Name:      walletData["name"].(string),
			Currency:  walletData["currency"].(string),
			CreatedAt: createdAt,
		}
		time.Sleep(time.Millisecond * 10) // Ensure distinct timestamps
	}
	return wallets
}

func (s *WalletIntegrationTestSuite) TestListWalletsPaginated() {
	// Clear wallets table
	s.clearWallets()

	// Create 10 test wallets
	wallets := s.createTestWallets(10)

	tests := []struct {
		name            string
		queryParams     map[string]string
		expectedStatus  int
		expectedLen     int
		expectedLimit   string
		expectNextToken bool
		expectedError   string
	}{
		{
			name:            "first page with default values",
			queryParams:     map[string]string{},
			expectedStatus:  http.StatusOK,
			expectedLen:     10,
			expectedLimit:   fmt.Sprint(coreTypes.DefaultLimit),
			expectNextToken: true,
		},
		{
			name: "first page with custom limit",
			queryParams: map[string]string{
				"limit": "5",
			},
			expectedStatus:  http.StatusOK,
			expectedLen:     5,
			expectedLimit:   "5",
			expectNextToken: true,
		},
		{
			name: "second page with next_token",
			queryParams: map[string]string{
				"limit":      "5",
				"next_token": coreTypes.EncodeCursor(wallets[4].CreatedAt, wallets[4].WalletID),
			},
			expectedStatus:  http.StatusOK,
			expectedLen:     5,
			expectedLimit:   "5",
			expectNextToken: true, //FIXME: service layer doesn't report on the total amount so no way to till unless the items were below the limit
		},
		{
			name: "invalid next_token format",
			queryParams: map[string]string{
				"next_token": "invalid-token",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid token",
		},
		{
			name: "limit below minimum",
			queryParams: map[string]string{
				"limit": "0",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "limit: must be no less than 1",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			urlPath := "/wallets/paginated"
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for k, v := range tt.queryParams {
					values.Add(k, v)
				}
				urlPath += "?" + values.Encode()
			}

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)

			s.Equal(tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			s.Require().NoError(err)

			if tt.expectedStatus == http.StatusOK {
				s.Equal(float64(http.StatusOK), response["status"])
				s.Equal("Success", response["message"])

				wallets := response["data"].([]interface{})
				s.Len(wallets, tt.expectedLen)

				metadata := response["meta"].(map[string]interface{})
				if tt.expectedLimit != "" {
					s.Equal(tt.expectedLimit, fmt.Sprint(metadata["limit"]))
				}

				if tt.expectNextToken {
					s.NotEmpty(metadata["next_token"])
				} else {
					nextToken, exists := metadata["next_token"]
					if exists {
						s.Empty(nextToken)
					}
				}
			} else if tt.expectedError != "" {
				errMsg, ok := response["error"].(string)
				s.True(ok)
				s.Contains(errMsg, tt.expectedError)
			}
		})
	}
}

func (s *WalletIntegrationTestSuite) TestSearchWallets() {
	// Create test wallets with more distinct names
	wallets := []types.WalletCreatePayload{
		{Name: "Wallet Alpha", Currency: "USD"},
		{Name: "Beta Account", Currency: "EUR"},
		{Name: "Gamma Wallet", Currency: "GBP"},
		{Name: "Delta Savings", Currency: "USD"},
		{Name: "Wallet Management Account", Currency: "EUR"},
		{Name: "Wallet Mnagement", Currency: "USD"},    // Misspelling of "Management"
		{Name: "Wllet Management", Currency: "EUR"},    // Missing 'a'
		{Name: "Savings Account", Currency: "USD"},     // Different type of account
		{Name: "Personal Account", Currency: "EUR"},    // Completely different
		{Name: "Business Wallet", Currency: "USD"},     // Completely different
		{Name: "Wallet #123", Currency: "USD"},         // With special characters
		{Name: "Alpha (Beta) Wallet", Currency: "EUR"}, // With parentheses
	}

	// Create all test wallets
	for _, w := range wallets {
		payloadBytes, err := json.Marshal(w)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Require().Equal(http.StatusCreated, w.Code)
		time.Sleep(time.Millisecond * 10) // Ensure distinct timestamps
	}

	tests := []struct {
		name           string
		query          string
		limit          string
		expectedStatus int
		expectedCount  int
		expectedNames  []string // Expected wallet names in order
	}{
		{
			name:           "case insensitive search",
			query:          "wallet",
			expectedStatus: http.StatusOK,
			expectedCount:  7,
			expectedNames:  []string{"Wallet #123", "Gamma Wallet", "Wallet Alpha", "Business Wallet", "Wallet Mnagement", "Alpha (Beta) Wallet", "Wallet Management Account"},
		},
		{
			name:           "similarity search - management misspelling",
			query:          "Management",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			expectedNames:  []string{"Wllet Management", "Wallet Management Account", "Wallet Mnagement"},
		},
		{
			name:           "similarity search - missing letter",
			query:          "Wllet",
			expectedStatus: http.StatusOK,
			expectedCount:  6,
			expectedNames:  []string{"Wllet Management", "Wallet #123", "Gamma Wallet", "Wallet Alpha", "Business Wallet", "Wallet Mnagement"},
		},
		{
			name:           "account search",
			query:          "Account",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedNames:  []string{"Beta Account", "Savings Account", "Personal Account", "Wallet Management Account"},
		},
		{
			name:           "with custom limit",
			query:          "Account",
			limit:          "2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedNames:  []string{"Beta Account", "Savings Account"},
		},
		{
			name:           "empty query",
			query:          "",
			limit:          fmt.Sprint(len(wallets)),
			expectedStatus: http.StatusOK,
			expectedCount:  len(wallets), // Should return all wallets ordered by created_at DESC
		},
		{
			name:           "query too long",
			query:          strings.Repeat("a", 101), // Exceeds maxQueryLength
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit",
			query:          "Wallet",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no results",
			query:          "NonExistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedNames:  []string{},
		},
		{
			name:           "special characters",
			query:          "#",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedNames:  []string{"Wallet #123"},
		},
		{
			name:           "parentheses",
			query:          "(",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedNames:  []string{"Alpha (Beta) Wallet"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Build URL with query parameters
			urlPath := fmt.Sprintf("/wallets/search?q=%s", url.QueryEscape(tt.query))
			if tt.limit != "" {
				urlPath += fmt.Sprintf("&limit=%s", tt.limit)
			}

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)

			s.Equal(tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				s.Require().NoError(err)
				s.Equal(float64(http.StatusOK), response["status"])

				wallets := response["data"].([]interface{})
				s.Len(wallets, tt.expectedCount)

				metadata := response["meta"].(map[string]interface{})
				if tt.query != "" {
					s.Equal(tt.query, metadata["query"])
				}
				if tt.limit != "" {
					limit, _ := strconv.ParseFloat(tt.limit, 64)
					s.Equal(limit, metadata["limit"])
				}

				// Verify wallet names if expected names are provided
				if len(tt.expectedNames) > 0 {
					actualNames := make([]string, len(wallets))
					for i, w := range wallets {
						wallet := w.(map[string]interface{})
						actualNames[i] = wallet["name"].(string)
					}
					s.Equal(tt.expectedNames, actualNames)
				}
			}
		})
	}
}

func (s *WalletIntegrationTestSuite) TestConcurrentUpdates() {
	// Create a wallet
	wallet := s.createTestWallet()

	// Try to update the same wallet concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			updatePayload := types.WalletUpdatePayload{
				WalletID: wallet.WalletID,
				Name:     fmt.Sprintf("Updated Name %d", i),
				Currency: "USD",
			}

			payloadBytes, err := json.Marshal(updatePayload)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/wallets/"+wallet.WalletID.String(), bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", wallet.WalletID.String())
			req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)

			// All updates should succeed
			s.Equal(http.StatusOK, w.Code)
		}(i)
	}
	wg.Wait()
}

func (s *WalletIntegrationTestSuite) TestUnauthorizedAccess() {
	// Create a wallet first
	wallet := s.createTestWallets(1)[0]

	// Create another user
	otherUserID := uuid.New()
	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, 'wit_other_clerk_id', 'wit_Other User', 'wit_other@example.com')
	`, otherUserID)
	s.Require().NoError(err)

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		expectedCode int
	}{
		{
			name: "access without user ID",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/wallets/"+wallet.WalletID.String(), nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", wallet.WalletID.String())
				return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "access with wrong user",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/wallets/"+wallet.WalletID.String(), nil)
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, otherUserID)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", wallet.WalletID.String())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "update with wrong user",
			setupRequest: func() *http.Request {
				payload := types.WalletUpdatePayload{
					WalletID: wallet.WalletID,
					Name:     "Unauthorized Update",
					Currency: "EUR",
				}
				payloadBytes, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/wallets/"+wallet.WalletID.String(), bytes.NewReader(payloadBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, otherUserID)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", wallet.WalletID.String())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := tt.setupRequest()
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			s.Equal(tt.expectedCode, w.Code)
		})
	}
}

func (s *WalletIntegrationTestSuite) TestComplexWalletLifecycle() {
	// Test the complete lifecycle of a wallet with multiple operations
	s.Run("full wallet lifecycle", func() {
		// 1. Create wallet
		createPayload := types.WalletCreatePayload{
			Name:     "Lifecycle Wallet",
			Currency: "USD",
			Balance:  float64Ptr(1000),
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		s.Require().NoError(err)
		walletData := response["data"].(map[string]interface{})
		walletID := walletData["walletId"].(string)

		// 2. Update wallet multiple times with different fields
		updates := []types.WalletUpdatePayload{
			{
				WalletID: uuid.MustParse(walletID),
				Name:     "Updated Name",
				Currency: "USD",
				Balance:  float64Ptr(2000),
			},
			{
				WalletID: uuid.MustParse(walletID),
				Name:     "Updated Name",
				Currency: "EUR",
				Balance:  float64Ptr(1500),
			},
		}

		for _, update := range updates {
			payloadBytes, err = json.Marshal(update)
			s.Require().NoError(err)

			req = httptest.NewRequest(http.MethodPut, "/wallets/"+walletID, bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx = context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			req = req.WithContext(ctx)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", walletID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w = httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			s.Equal(http.StatusOK, w.Code)
		}

		// 3. Verify final state
		req = httptest.NewRequest(http.MethodGet, "/wallets/"+walletID, nil)
		ctx = context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", walletID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w = httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusOK, w.Code)

		var finalState map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&finalState)
		s.Require().NoError(err)
		finalData := finalState["data"].(map[string]interface{})

		// Verify final state matches last update
		s.Equal("Updated Name", finalData["name"])
		s.Equal("EUR", finalData["currency"])
		s.Equal(1500.0, finalData["balance"])
	})
}

func (s *WalletIntegrationTestSuite) TestPaginationEdgeCases() {
	// Test extreme pagination cases
	s.Run("pagination edge cases", func() {
		// Create 20 wallets for testing
		_ = s.createTestWallets(20)

		tests := []struct {
			name         string
			limit        int32
			expectedCode int
			expectedErr  string
		}{
			{
				name:         "zero limit",
				limit:        0,
				expectedCode: http.StatusBadRequest,
				expectedErr:  "limit: must be no less than 1.",
			},
			{
				name:         "negative limit",
				limit:        -1,
				expectedCode: http.StatusBadRequest,
				expectedErr:  "limit: must be no less than 1.",
			},
			{
				name:         "very large limit",
				limit:        1000,
				expectedCode: http.StatusOK,
				// Should still work but return max allowed
			},
		}

		for _, tt := range tests {
			s.Run(tt.name, func() {
				urlPath := fmt.Sprintf("/wallets/paginated?limit=%d", tt.limit)
				req := httptest.NewRequest(http.MethodGet, urlPath, nil)
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
				req = req.WithContext(ctx)

				w := httptest.NewRecorder()
				s.router.ServeHTTP(w, req)

				s.Equal(tt.expectedCode, w.Code)
				if tt.expectedCode != http.StatusOK {
					var response map[string]interface{}
					err := json.NewDecoder(w.Body).Decode(&response)
					s.Require().NoError(err)
					s.Contains(response["error"].(string), tt.expectedErr)
				}
			})
		}
	})
}

func (s *WalletIntegrationTestSuite) TestDatabaseConstraintsAndValidation() {
	s.Run("database constraints and validation", func() {
		tests := []struct {
			name          string
			payload       interface{} // Using interface{} to allow malformed JSON
			expectedCode  int
			errorContains string
			errorMessage  string
		}{
			{
				name: "required fields missing",
				payload: map[string]interface{}{
					"balance": 1000.0,
					// name and currency missing
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "currency: cannot be blank; name: cannot be blank",
				errorMessage:  "Invalid request",
			},
			{
				name: "name too long",
				payload: map[string]interface{}{
					"name":     strings.Repeat("a", 256),
					"currency": "USD",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "name: the length must be between 1 and 255",
				errorMessage:  "Invalid request",
			},
			{
				name: "invalid currency",
				payload: map[string]interface{}{
					"name":     "Test Wallet",
					"currency": "INVALID",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "currency: must be valid ISO 4217 currency code",
				errorMessage:  "Invalid request",
			},
			{
				name: "invalid balance format",
				payload: map[string]interface{}{
					"name":     "Test Wallet",
					"currency": "USD",
					"balance":  "not-a-number",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "balance",
				errorMessage:  "Invalid request",
			},
			{
				name: "invalid project ID",
				payload: map[string]interface{}{
					"name":      "Test Wallet",
					"currency":  "USD",
					"projectId": "not-a-uuid",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "invalid UUID length",
				errorMessage:  "Invalid request",
			},
			{
				name: "too many tags",
				payload: map[string]interface{}{
					"name":     "Test Wallet",
					"currency": "USD",
					"tags":     []string{uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()}, // Exceeds MaxTagsCount
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "tags: the length must be no more than 10",
				errorMessage:  "Invalid request",
			},
		}

		for _, tt := range tests {
			s.Run(tt.name, func() {
				payloadBytes, err := json.Marshal(tt.payload)
				s.Require().NoError(err)

				req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(payloadBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
				req = req.WithContext(ctx)

				w := httptest.NewRecorder()
				s.router.ServeHTTP(w, req)

				s.Equal(tt.expectedCode, w.Code)

				var response map[string]interface{}
				err = json.NewDecoder(w.Body).Decode(&response)
				s.Require().NoError(err)
				s.Equal(tt.errorMessage, response["message"])
				s.Contains(response["error"].(string), tt.errorContains)
			})
		}
	})
}

func (s *WalletIntegrationTestSuite) TestResponsePayloadStructure() {
	s.Run("response payload structure", func() {
		// Create a wallet with all fields
		createPayload := types.WalletCreatePayload{
			Name:      "Response Test Wallet",
			Currency:  "USD",
			Balance:   float64Ptr(1000.50),
			ProjectID: nil, // Optional
			Tags:      []uuid.UUID{uuid.New(), uuid.New()},
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		s.Equal(http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		s.Require().NoError(err)

		// Verify response structure
		s.Require().Contains(response, "status")
		s.Require().Contains(response, "data")

		data := response["data"].(map[string]interface{})

		// Verify all expected fields exist and have correct types
		s.Require().Contains(data, "walletId")
		s.IsType("", data["walletId"].(string))
		_, err = uuid.Parse(data["walletId"].(string))
		s.NoError(err)

		s.Equal(createPayload.Name, data["name"])
		s.Equal(createPayload.Currency, data["currency"])
		s.Equal(*createPayload.Balance, data["balance"])
		s.NotEmpty(data["createdAt"])
		s.NotEmpty(data["updatedAt"])

		// Verify timestamps are in correct format
		_, err = time.Parse(time.RFC3339, data["createdAt"].(string))
		s.NoError(err)
		_, err = time.Parse(time.RFC3339, data["updatedAt"].(string))
		s.NoError(err)

		// Verify tags array
		tags := data["tags"].([]interface{})
		s.Len(tags, len(createPayload.Tags))
		for _, tag := range tags {
			_, err := uuid.Parse(tag.(string))
			s.NoError(err)
		}

		// Verify optional fields
		s.Nil(data["projectId"]) // Since we didn't set it
	})
}
