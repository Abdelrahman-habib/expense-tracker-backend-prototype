package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/wallets/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

/************************************************
*                Test Suite Setup                 *
************************************************/

// WalletRepositoryTestSuite defines the test suite
type WalletRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	pool      *pgxpool.Pool
	queries   *db.Queries
	repo      repository.WalletRepository
	ctx       context.Context
	testUser  uuid.UUID
}

// TestWalletRepository is the single entry point for the test suite
func TestWalletRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(WalletRepositoryTestSuite))
}

/************************************************
*            Setup and Teardown                  *
************************************************/

func (s *WalletRepositoryTestSuite) SetupSuite() {
	fmt.Println("Starting test suite setup...")
	s.ctx = context.Background()

	var host, port string
	var err error

	if os.Getenv("CI") == "true" {
		// Running in GitHub Actions, use service-based PostgreSQL
		fmt.Println("Running in CI, using GitHub Actions PostgreSQL service...")
		host = "localhost"
		port = "5432"
	} else {
		// Running locally, use TestContainers
		fmt.Println("Running locally, creating PostgreSQL container...")

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
		s.Require().NoError(err)
		s.container = container

		// Get container host and port
		host, err = container.Host(s.ctx)
		s.Require().NoError(err)
		portMapped, err := container.MappedPort(s.ctx, "5432")
		s.Require().NoError(err)
		port = portMapped.Port()
	}

	// Create connection string
	connString := fmt.Sprintf(
		"postgres://test:test@%s:%s/testdb?sslmode=disable",
		host, port,
	)

	// Connect to database
	fmt.Println("Connecting to database...")
	s.pool, err = pgxpool.New(s.ctx, connString)
	s.Require().NoError(err)

	// Run migrations
	fmt.Println("Running migrations...")
	err = s.runMigrations()
	s.Require().NoError(err)

	// Create queries and repository
	fmt.Println("Creating repository...")
	s.queries = db.New(s.pool)
	s.repo = repository.NewWalletRepository(s.queries)

	// Create test user
	fmt.Println("Creating test user...")
	s.testUser = uuid.New()
	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, $2, 'wrt_Test User', 'wrt_test@example.com')
	`, s.testUser, s.testUser.String())
	s.Require().NoError(err)
	fmt.Println("Test suite setup completed successfully")
}

func (s *WalletRepositoryTestSuite) TearDownSuite() {
	fmt.Println("Tearing down test suite...")

	if s.pool != nil {
		s.pool.Close()
		fmt.Println("Database pool closed.")
	}

	if s.container != nil && os.Getenv("CI") != "true" {
		fmt.Println("Terminating TestContainers PostgreSQL instance...")
		err := s.container.Terminate(s.ctx)
		s.Require().NoError(err)
		fmt.Println("Test container terminated.")
	}

	fmt.Println("Test suite teardown complete.")
}

func (s *WalletRepositoryTestSuite) SetupTest() {
	// Clean up tables before each test
	s.clearWallets()
}

func (s *WalletRepositoryTestSuite) TearDownTest() {
	// Clean up tables after each test
	s.clearWallets()
}

func (s *WalletRepositoryTestSuite) clearWallets() {
	_, err := s.pool.Exec(s.ctx, "DELETE FROM wallets WHERE user_id = $1", s.testUser)
	require.NoError(s.T(), err)
	_, err = s.pool.Exec(s.ctx, "DELETE FROM projects WHERE user_id = $1", s.testUser)
	require.NoError(s.T(), err)
}

/************************************************
*              Test Cases                        *
************************************************/

func (s *WalletRepositoryTestSuite) TestCreateWallet() {
	// Create a test project first
	projectID := s.createTestProject("Test Project for Wallet Creation")

	tests := []struct {
		name    string
		payload types.WalletCreatePayload
		wantErr bool
	}{
		{
			name: "valid wallet",
			payload: types.WalletCreatePayload{
				Name:     "Test Wallet",
				Currency: "USD",
			},
			wantErr: false,
		},
		{
			name: "wallet with all fields",
			payload: types.WalletCreatePayload{
				Name:      "Full Wallet",
				Balance:   utils.Float64Ptr(1000.50),
				Currency:  "EUR",
				ProjectID: &projectID,
				Tags:      []uuid.UUID{uuid.New(), uuid.New()},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wallet, err := s.repo.CreateWallet(s.ctx, tt.payload, s.testUser)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.NotEmpty(wallet.WalletID)
			s.Equal(tt.payload.Name, wallet.Name)
			s.Equal(tt.payload.Currency, wallet.Currency)

			// Check optional fields only if they are provided in the payload
			if tt.payload.Balance != nil {
				s.NotNil(wallet.Balance)
				s.Equal(*tt.payload.Balance, *wallet.Balance)
			}
			if tt.payload.ProjectID != nil {
				s.NotNil(wallet.ProjectID)
				s.Equal(*tt.payload.ProjectID, *wallet.ProjectID)
			}
			if tt.payload.Tags != nil {
				s.Equal(tt.payload.Tags, wallet.Tags)
			}

			s.NotZero(wallet.CreatedAt)
			s.NotZero(wallet.UpdatedAt)
		})
	}
}

func (s *WalletRepositoryTestSuite) TestGetWallet() {
	// Create a test wallet first
	createPayload := types.WalletCreatePayload{
		Name:     "Test Wallet",
		Currency: "USD",
		Balance:  utils.Float64Ptr(100.00),
	}
	created, err := s.repo.CreateWallet(s.ctx, createPayload, s.testUser)
	require.NoError(s.T(), err)

	tests := []struct {
		name     string
		userID   uuid.UUID
		walletID uuid.UUID
		wantErr  bool
	}{
		{
			name:     "existing wallet",
			userID:   s.testUser,
			walletID: created.WalletID,
			wantErr:  false,
		},
		{
			name:     "non-existent wallet",
			userID:   s.testUser,
			walletID: uuid.New(),
			wantErr:  true,
		},
		{
			name:     "wrong user",
			userID:   uuid.New(),
			walletID: created.WalletID,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wallet, err := s.repo.GetWallet(s.ctx, tt.walletID, tt.userID)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Equal(created.WalletID, wallet.WalletID)
			s.Equal(created.Name, wallet.Name)
			s.Equal(created.Currency, wallet.Currency)
			s.Equal(*created.Balance, *wallet.Balance)
		})
	}
}

func (s *WalletRepositoryTestSuite) TestUpdateWallet() {
	// Create a test wallet first
	createPayload := types.WalletCreatePayload{
		Name:     "Test Wallet",
		Currency: "USD",
		Balance:  utils.Float64Ptr(100.00),
		Tags:     []uuid.UUID{uuid.New(), uuid.New()},
	}
	created, err := s.repo.CreateWallet(s.ctx, createPayload, s.testUser)
	require.NoError(s.T(), err)

	tests := []struct {
		name    string
		payload types.WalletUpdatePayload
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name: "valid update",
			payload: types.WalletUpdatePayload{
				WalletID: created.WalletID,
				Name:     "Updated Wallet",
				Currency: "EUR",
				Balance:  utils.Float64Ptr(200.00),
			},
			userID:  s.testUser,
			wantErr: false,
		},
		{
			name: "update with wrong user",
			payload: types.WalletUpdatePayload{
				WalletID: created.WalletID,
				Name:     "Should Not Update",
				Currency: "GBP",
			},
			userID:  uuid.New(),
			wantErr: true,
		},
		{
			name: "update non-existent wallet",
			payload: types.WalletUpdatePayload{
				WalletID: uuid.New(),
				Name:     "Non-existent",
				Currency: "USD",
			},
			userID:  s.testUser,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wallet, err := s.repo.UpdateWallet(s.ctx, tt.payload, tt.userID)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Equal(tt.payload.WalletID, wallet.WalletID)
			s.Equal(tt.payload.Name, wallet.Name)
			s.Equal(tt.payload.Currency, wallet.Currency)
			if tt.payload.Balance != nil {
				s.Equal(*tt.payload.Balance, *wallet.Balance)
			}
		})
	}
}

func (s *WalletRepositoryTestSuite) TestListWallets() {
	// Create test wallets
	wallets := []types.WalletCreatePayload{
		{Name: "Wallet 1", Currency: "USD", Balance: utils.Float64Ptr(100.00)},
		{Name: "Wallet 2", Currency: "EUR", Balance: utils.Float64Ptr(200.00)},
		{Name: "Wallet 3", Currency: "GBP", Balance: utils.Float64Ptr(300.00)},
	}

	for _, w := range wallets {
		_, err := s.repo.CreateWallet(s.ctx, w, s.testUser)
		s.Require().NoError(err)
	}

	tests := []struct {
		name    string
		userID  uuid.UUID
		limit   int32
		offset  int32
		want    int
		wantErr bool
	}{
		{
			name:   "list all wallets",
			userID: s.testUser,
			limit:  10,
			offset: 0,
			want:   3,
		},
		{
			name:   "list with limit",
			userID: s.testUser,
			limit:  2,
			offset: 0,
			want:   2,
		},
		{
			name:   "list with offset",
			userID: s.testUser,
			limit:  10,
			offset: 1,
			want:   2,
		},
		{
			name:   "list for different user",
			userID: uuid.New(),
			limit:  10,
			offset: 0,
			want:   0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wallets, err := s.repo.ListWallets(s.ctx, tt.userID, tt.limit, tt.offset)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(wallets, tt.want)
		})
	}
}

func (s *WalletRepositoryTestSuite) TestListWalletsPaginated() {
	// Create test wallets in order from oldest to newest
	wallets := []types.WalletCreatePayload{
		{Name: "Wallet 1", Currency: "USD", Balance: utils.Float64Ptr(100.00)}, // Oldest
		{Name: "Wallet 2", Currency: "EUR", Balance: utils.Float64Ptr(200.00)},
		{Name: "Wallet 3", Currency: "GBP", Balance: utils.Float64Ptr(300.00)},
		{Name: "Wallet 4", Currency: "JPY", Balance: utils.Float64Ptr(400.00)}, // Newest
	}

	var createdWallets []types.Wallet
	for _, w := range wallets {
		time.Sleep(time.Millisecond * 100) // Ensure distinct timestamps
		wallet, err := s.repo.CreateWallet(s.ctx, w, s.testUser)
		s.Require().NoError(err)
		createdWallets = append(createdWallets, wallet)
	}

	tests := []struct {
		name      string
		cursor    time.Time
		cursorID  uuid.UUID
		limit     int32
		wantLen   int
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "get first page",
			cursor:    time.Now().UTC(),
			cursorID:  uuid.Nil,
			limit:     2,
			wantLen:   2,
			wantNames: []string{"Wallet 4", "Wallet 3"},
			wantErr:   false,
		},
		{
			name:      "get second page",
			cursor:    createdWallets[2].CreatedAt,
			cursorID:  createdWallets[2].WalletID,
			limit:     2,
			wantLen:   2,
			wantNames: []string{"Wallet 2", "Wallet 1"},
			wantErr:   false,
		},
		{
			name:      "get empty page",
			cursor:    createdWallets[0].CreatedAt,
			cursorID:  createdWallets[0].WalletID,
			limit:     2,
			wantLen:   0,
			wantNames: []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wallets, err := s.repo.ListWalletsPaginated(s.ctx, s.testUser, tt.cursor, tt.cursorID, tt.limit)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(wallets, tt.wantLen)

			if len(tt.wantNames) > 0 {
				actualNames := make([]string, len(wallets))
				for i, w := range wallets {
					actualNames[i] = w.Name
				}
				s.Equal(tt.wantNames, actualNames)
			}

			// Verify ordering for non-empty results
			if len(wallets) > 1 {
				for i := 1; i < len(wallets); i++ {
					isCorrectOrder := wallets[i-1].CreatedAt.After(wallets[i].CreatedAt) ||
						(wallets[i-1].CreatedAt.Equal(wallets[i].CreatedAt) &&
							wallets[i-1].WalletID.String() > wallets[i].WalletID.String())
					s.True(isCorrectOrder, "Wallets should be ordered by created_at DESC and then by wallet_id DESC")
				}
			}
		})
	}
}

func (s *WalletRepositoryTestSuite) TestSearchWallets() {
	// Create test wallets with various names
	wallets := []types.WalletCreatePayload{
		{Name: "Savings Wallet", Currency: "USD"},
		{Name: "My Savings", Currency: "EUR"},
		{Name: "Travel Fund", Currency: "GBP"},
		{Name: "Emergency Savings", Currency: "USD"},
		{Name: "Investment Portfolio", Currency: "USD"},
		{Name: "Holiday Budget", Currency: "EUR"},
		{Name: "Svings Account", Currency: "USD"}, // Misspelling
	}

	for _, w := range wallets {
		_, err := s.repo.CreateWallet(s.ctx, w, s.testUser)
		s.Require().NoError(err)
		time.Sleep(time.Millisecond * 100) // Ensure distinct timestamps
	}

	tests := []struct {
		name      string
		query     string
		limit     int32
		wantLen   int
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "search for savings",
			query:     "Savings",
			limit:     10,
			wantLen:   4,
			wantNames: []string{"My Savings", "Savings Wallet", "Emergency Savings", "Svings Account"},
			wantErr:   false,
		},
		{
			name:      "search with limit",
			query:     "Savings",
			limit:     2,
			wantLen:   2,
			wantNames: []string{"My Savings", "Savings Wallet"},
			wantErr:   false,
		},
		{
			name:      "search with similar word",
			query:     "Svings",
			limit:     10,
			wantLen:   4, // Should find the misspelled one and similar ones
			wantNames: []string{"Svings Account", "My Savings", "Savings Wallet", "Emergency Savings"},
			wantErr:   false,
		},
		{
			name:      "no results",
			query:     "NonExistent",
			limit:     10,
			wantLen:   0,
			wantNames: []string{},
			wantErr:   false,
		},
		{
			name:    "invalid limit",
			query:   "test",
			limit:   -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wallets, err := s.repo.SearchWallets(s.ctx, s.testUser, tt.query, tt.limit)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(wallets, tt.wantLen)

			if len(tt.wantNames) > 0 {
				actualNames := make([]string, len(wallets))
				for i, w := range wallets {
					actualNames[i] = w.Name
				}
				s.Equal(tt.wantNames, actualNames)
			}
		})
	}
}

func (s *WalletRepositoryTestSuite) TestGetProjectWallets() {
	// Create test project first
	projectID := s.createTestProject("Test Project for GetProjectWallets")

	// Create test wallets
	wallets := []types.WalletCreatePayload{
		{Name: "Project Wallet 1", Currency: "USD", ProjectID: &projectID},
		{Name: "Project Wallet 2", Currency: "EUR", ProjectID: &projectID},
		{Name: "Personal Wallet", Currency: "GBP"}, // No project ID
	}

	for _, w := range wallets {
		_, err := s.repo.CreateWallet(s.ctx, w, s.testUser)
		s.Require().NoError(err)
	}

	tests := []struct {
		name      string
		projectID uuid.UUID
		userID    uuid.UUID
		want      int
		wantErr   bool
	}{
		{
			name:      "get project wallets",
			projectID: projectID,
			userID:    s.testUser,
			want:      2,
			wantErr:   false,
		},
		{
			name:      "get wallets for non-existent project",
			projectID: uuid.New(),
			userID:    s.testUser,
			want:      0,
			wantErr:   false,
		},
		{
			name:      "get wallets with wrong user",
			projectID: projectID,
			userID:    uuid.New(),
			want:      0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			wallets, err := s.repo.GetProjectWallets(s.ctx, tt.projectID, tt.userID)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(wallets, tt.want)

			for _, w := range wallets {
				if tt.want > 0 {
					s.Equal(tt.projectID, *w.ProjectID)
				}
			}
		})
	}
}

/************************************************
*              Helper Functions                  *
************************************************/

func (s *WalletRepositoryTestSuite) runMigrations() error {
	migrationsDir := "../../db/sql/migrations"

	// Convert pool to *sql.DB for goose
	db := stdlib.OpenDBFromPool(s.pool)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (s *WalletRepositoryTestSuite) createTestProject(name string) uuid.UUID {
	var projectID uuid.UUID
	err := s.pool.QueryRow(s.ctx, `
		INSERT INTO projects (user_id, name, status)
		VALUES ($1, $2, 'ongoing')
		RETURNING project_id
	`, s.testUser, name).Scan(&projectID)
	s.Require().NoError(err)
	return projectID
}
