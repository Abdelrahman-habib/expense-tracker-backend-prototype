package repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

// AuthRepositoryTestSuite defines the test suite for auth repository
type AuthRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	pool      *pgxpool.Pool
	queries   *db.Queries
	repo      repository.Repository
	ctx       context.Context
	testUser  uuid.UUID
	logger    *zap.Logger
}

// TestAuthRepository runs the test suite
func TestAuthRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthRepositoryTestSuite))
}

// SetupSuite sets up the test suite
func (s *AuthRepositoryTestSuite) SetupSuite() {
	fmt.Println("Starting test suite setup...")
	var err error
	s.ctx = context.Background()

	// Create a logger for testing
	s.logger, err = zap.NewDevelopment()
	require.NoError(s.T(), err)

	var host, port string

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
		port = portMapped.Port() // Extract numeric port
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
	s.queries = db.New(s.pool)
	s.repo = repository.NewAuthRepository(s.queries, s.logger)

	// Create a test user
	user, err := s.queries.CreateUser(s.ctx, db.CreateUserParams{
		Name:       "Test User",
		Email:      "test@example.com",
		ExternalID: "test-external-id",
		Provider:   "google",
	})
	require.NoError(s.T(), err)
	s.testUser = user.UserID
}

// TearDownSuite tears down the test suite
func (s *AuthRepositoryTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		err := s.container.Terminate(s.ctx)
		if err != nil {
			s.T().Fatalf("failed to terminate container: %v", err)
		}
	}
}

// cleanSessionsTable cleans the sessions table
func (s *AuthRepositoryTestSuite) cleanSessionsTable() {
	_, err := s.pool.Exec(s.ctx, "DELETE FROM sessions")
	require.NoError(s.T(), err)
}

// SetupTest sets up each test
func (s *AuthRepositoryTestSuite) SetupTest() {
	s.cleanSessionsTable()
}

// TearDownTest tears down each test
func (s *AuthRepositoryTestSuite) TearDownTest() {
	// Reset user's refresh token
	_, err := s.pool.Exec(s.ctx, "UPDATE users SET refresh_token_hash = NULL WHERE user_id = $1", s.testUser)
	require.NoError(s.T(), err)
}

// TestStoreRefreshToken tests the StoreRefreshToken method
func (s *AuthRepositoryTestSuite) TestStoreRefreshToken() {
	// Test storing a refresh token
	token := "test-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := s.repo.StoreRefreshToken(s.ctx, s.testUser, token, expiresAt)
	require.NoError(s.T(), err)

	// Verify token was stored
	var storedToken string
	err = s.pool.QueryRow(s.ctx, "SELECT refresh_token_hash FROM users WHERE user_id = $1", s.testUser).Scan(&storedToken)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), token, storedToken)

	// Test updating an existing token
	newToken := "new-refresh-token"
	err = s.repo.StoreRefreshToken(s.ctx, s.testUser, newToken, expiresAt)
	require.NoError(s.T(), err)

	// Verify token was updated
	err = s.pool.QueryRow(s.ctx, "SELECT refresh_token_hash FROM users WHERE user_id = $1", s.testUser).Scan(&storedToken)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), newToken, storedToken)
}

// TestGetRefreshToken tests the GetRefreshToken method
func (s *AuthRepositoryTestSuite) TestGetRefreshToken() {
	// Test getting a non-existent token
	_, err := s.repo.GetRefreshToken(s.ctx, s.testUser)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "no refresh token found")

	// Store a token
	token := "test-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)
	err = s.repo.StoreRefreshToken(s.ctx, s.testUser, token, expiresAt)
	require.NoError(s.T(), err)

	// Test getting the token
	storedToken, err := s.repo.GetRefreshToken(s.ctx, s.testUser)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testUser, storedToken.UserID)
	assert.Equal(s.T(), token, storedToken.Hash)
	assert.True(s.T(), storedToken.ExpiresAt.After(time.Now()))
}

// TestDeleteRefreshToken tests the DeleteRefreshToken method
func (s *AuthRepositoryTestSuite) TestDeleteRefreshToken() {
	// Store a token
	token := "test-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)
	err := s.repo.StoreRefreshToken(s.ctx, s.testUser, token, expiresAt)
	require.NoError(s.T(), err)

	// Verify token exists
	var storedToken pgtype.Text
	err = s.pool.QueryRow(s.ctx, "SELECT refresh_token_hash FROM users WHERE user_id = $1", s.testUser).Scan(&storedToken)
	require.NoError(s.T(), err)
	assert.True(s.T(), storedToken.Valid)

	// Delete the token
	err = s.repo.DeleteRefreshToken(s.ctx, s.testUser)
	require.NoError(s.T(), err)

	// Verify token was deleted (set to NULL)
	err = s.pool.QueryRow(s.ctx, "SELECT refresh_token_hash FROM users WHERE user_id = $1", s.testUser).Scan(&storedToken)
	require.NoError(s.T(), err)
	assert.False(s.T(), storedToken.Valid)
}

// TestStoreSession tests the StoreSession method
func (s *AuthRepositoryTestSuite) TestStoreSession() {
	// Test storing a session
	key := "test-session-key"
	value := map[string]interface{}{
		"user_id": s.testUser.String(),
		"scopes":  []string{"profile", "email"},
	}
	expiresAt := time.Now().Add(24 * time.Hour)

	err := s.repo.StoreSession(s.ctx, key, value, expiresAt)
	require.NoError(s.T(), err)

	// Verify session was stored
	var storedValue []byte
	var storedExpiresAt time.Time
	err = s.pool.QueryRow(s.ctx, "SELECT value, expires_at FROM sessions WHERE key = $1", key).Scan(&storedValue, &storedExpiresAt)
	require.NoError(s.T(), err)

	// Decode the stored value
	var decodedValue map[string]interface{}
	err = json.Unmarshal(storedValue, &decodedValue)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), s.testUser.String(), decodedValue["user_id"])
	assert.Equal(s.T(), []interface{}{"profile", "email"}, decodedValue["scopes"])
	assert.True(s.T(), storedExpiresAt.Equal(expiresAt.Truncate(time.Microsecond)))

	// Test updating an existing session
	newValue := map[string]interface{}{
		"user_id": s.testUser.String(),
		"scopes":  []string{"profile", "email", "contacts"},
	}
	newExpiresAt := time.Now().Add(48 * time.Hour)

	err = s.repo.StoreSession(s.ctx, key, newValue, newExpiresAt)
	require.NoError(s.T(), err)

	// Verify session was updated
	err = s.pool.QueryRow(s.ctx, "SELECT value, expires_at FROM sessions WHERE key = $1", key).Scan(&storedValue, &storedExpiresAt)
	require.NoError(s.T(), err)

	// Decode the stored value
	err = json.Unmarshal(storedValue, &decodedValue)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), s.testUser.String(), decodedValue["user_id"])
	assert.Equal(s.T(), []interface{}{"profile", "email", "contacts"}, decodedValue["scopes"])
	assert.True(s.T(), storedExpiresAt.Equal(newExpiresAt.Truncate(time.Microsecond)))
}

// TestGetSession tests the GetSession method
func (s *AuthRepositoryTestSuite) TestGetSession() {
	// Test getting a non-existent session
	_, err := s.repo.GetSession(s.ctx, "non-existent-key")
	require.Error(s.T(), err)

	// Store a session
	key := "test-session-key"
	value := map[string]interface{}{
		"user_id": s.testUser.String(),
		"scopes":  []string{"profile", "email"},
	}
	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Microsecond)

	valueBytes, err := json.Marshal(value)
	require.NoError(s.T(), err)

	_, err = s.pool.Exec(s.ctx, "INSERT INTO sessions (key, value, expires_at) VALUES ($1, $2, $3)",
		key, valueBytes, expiresAt)
	require.NoError(s.T(), err)

	// Test getting the session
	session, err := s.repo.GetSession(s.ctx, key)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), key, session.Key)
	assert.Equal(s.T(), valueBytes, session.Value)
	assert.Equal(s.T(), expiresAt, session.ExpiresAt)

	// Decode the value
	var decodedValue map[string]interface{}
	err = json.Unmarshal(session.Value, &decodedValue)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testUser.String(), decodedValue["user_id"])
	assert.Equal(s.T(), []interface{}{"profile", "email"}, decodedValue["scopes"])
}

// TestDeleteSession tests the DeleteSession method
func (s *AuthRepositoryTestSuite) TestDeleteSession() {
	// Store a session
	key := "test-session-key"
	value := map[string]interface{}{
		"user_id": s.testUser.String(),
		"scopes":  []string{"profile", "email"},
	}
	expiresAt := time.Now().Add(24 * time.Hour)

	valueBytes, err := json.Marshal(value)
	require.NoError(s.T(), err)

	_, err = s.pool.Exec(s.ctx, "INSERT INTO sessions (key, value, expires_at) VALUES ($1, $2, $3)",
		key, valueBytes, expiresAt)
	require.NoError(s.T(), err)

	// Verify session exists
	var count int
	err = s.pool.QueryRow(s.ctx, "SELECT COUNT(*) FROM sessions WHERE key = $1", key).Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)

	// Delete the session
	err = s.repo.DeleteSession(s.ctx, key)
	require.NoError(s.T(), err)

	// Verify session was deleted
	err = s.pool.QueryRow(s.ctx, "SELECT COUNT(*) FROM sessions WHERE key = $1", key).Scan(&count)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

// TestGetUserByExternalID tests the GetUserByExternalID method
func (s *AuthRepositoryTestSuite) TestGetUserByExternalID() {
	// Test getting a non-existent user
	_, err := s.repo.GetUserByExternalID(s.ctx, "non-existent-id", "google")
	require.Error(s.T(), err)

	// Test getting an existing user
	user, err := s.repo.GetUserByExternalID(s.ctx, "test-external-id", "google")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testUser, user.ID)
	assert.Equal(s.T(), "Test User", user.Name)
	assert.Equal(s.T(), "test@example.com", user.Email)
	assert.Equal(s.T(), "google", user.Provider)
}

// TestCreateUser tests the CreateUser method
func (s *AuthRepositoryTestSuite) TestCreateUser() {
	// Test creating a new user
	userData := types.OAuthUserData{
		ExternalID: "new-external-id",
		Name:       "New User",
		Email:      "new@example.com",
		Provider:   "github",
	}

	user, err := s.repo.CreateUser(s.ctx, userData)
	require.NoError(s.T(), err)
	assert.NotEqual(s.T(), uuid.Nil, user.ID)
	assert.Equal(s.T(), userData.Name, user.Name)
	assert.Equal(s.T(), userData.Email, user.Email)
	assert.Equal(s.T(), userData.Provider, user.Provider)

	// Verify user was created in the database
	var dbUser db.User
	err = s.pool.QueryRow(s.ctx, `
		SELECT user_id, name, email, external_id, provider 
		FROM users 
		WHERE external_id = $1 AND provider = $2
	`, userData.ExternalID, userData.Provider).Scan(
		&dbUser.UserID,
		&dbUser.Name,
		&dbUser.Email,
		&dbUser.ExternalID,
		&dbUser.Provider,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), user.ID, dbUser.UserID)
	assert.Equal(s.T(), userData.Name, dbUser.Name)
	assert.Equal(s.T(), userData.Email, dbUser.Email)
	assert.Equal(s.T(), userData.ExternalID, dbUser.ExternalID)
	assert.Equal(s.T(), userData.Provider, dbUser.Provider)
}

// TestUpdateUserLastLogin tests the UpdateUserLastLogin method
func (s *AuthRepositoryTestSuite) TestUpdateUserLastLogin() {
	// Get current last_login value
	var originalLastLogin pgtype.Timestamp
	err := s.pool.QueryRow(s.ctx, "SELECT last_login FROM users WHERE user_id = $1", s.testUser).Scan(&originalLastLogin)
	require.NoError(s.T(), err)

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Second)

	// Update last login
	err = s.repo.UpdateUserLastLogin(s.ctx, s.testUser)
	require.NoError(s.T(), err)

	// Verify last_login was updated
	var newLastLogin pgtype.Timestamp
	err = s.pool.QueryRow(s.ctx, "SELECT last_login FROM users WHERE user_id = $1", s.testUser).Scan(&newLastLogin)
	require.NoError(s.T(), err)

	assert.True(s.T(), newLastLogin.Time.After(originalLastLogin.Time))
}

// runMigrations runs the database migrations
func (s *AuthRepositoryTestSuite) runMigrations() error {
	// Create a sql.DB connection for goose
	sqlDB := stdlib.OpenDB(*s.pool.Config().ConnConfig)
	if sqlDB == nil {
		return fmt.Errorf("failed to create sql.DB")
	}
	defer sqlDB.Close()

	// Set goose dialect and directory
	goose.SetDialect("postgres")
	migrationDir := "../../db/sql/migrations"

	// Run migrations
	if err := goose.Up(sqlDB, migrationDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
