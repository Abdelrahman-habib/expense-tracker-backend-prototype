package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContactRepositoryTestSuite defines the test suite
type ContactRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	pool      *pgxpool.Pool
	queries   *db.Queries
	repo      repository.Repository
	ctx       context.Context
	testUser  uuid.UUID
}

// TestContactRepository is the single entry point for the test suite
func TestContactRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(ContactRepositoryTestSuite))
}

func (s *ContactRepositoryTestSuite) SetupSuite() {
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
	fmt.Println("Creating repository...")
	s.queries = db.New(s.pool)
	s.repo = repository.New(s.queries)

	// Create test user
	fmt.Println("Creating test user...")
	s.testUser = uuid.New()
	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, $2, 'crt_Test User', 'crt_test@example.com')
	`, s.testUser, s.testUser.String())
	s.Require().NoError(err)
	fmt.Println("Test suite setup completed successfully")
}

func (s *ContactRepositoryTestSuite) TearDownSuite() {
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

func (s *ContactRepositoryTestSuite) cleanContactTable() {
	// Clean up contacts table after each test
	_, err := s.pool.Exec(s.ctx, `DELETE FROM contacts WHERE user_id = $1`, s.testUser)
	require.NoError(s.T(), err)
}

func (s *ContactRepositoryTestSuite) SetupTest() {
	// Clean up contacts table before each test
	s.cleanContactTable()
}

func (s *ContactRepositoryTestSuite) TearDownTest() {
	// Clean up contacts table after each test
	s.cleanContactTable()
}

func (s *ContactRepositoryTestSuite) TestCreateContact() {
	tests := []struct {
		name    string
		payload types.ContactCreatePayload
		wantErr bool
	}{
		{
			name: "valid contact with minimal fields",
			payload: types.ContactCreatePayload{
				Name: "John Doe",
			},
			wantErr: false,
		},
		{
			name: "contact with all fields",
			payload: types.ContactCreatePayload{
				Name:          "Jane Smith",
				Phone:         utils.StringPtr("+1-555-123-4567"),
				Email:         utils.StringPtr("jane@example.com"),
				AddressLine1:  utils.StringPtr("123 Main St"),
				AddressLine2:  utils.StringPtr("Apt 4B"),
				Country:       utils.StringPtr("US"),
				City:          utils.StringPtr("New York"),
				StateProvince: utils.StringPtr("NY"),
				ZipPostalCode: utils.StringPtr("10001"),
				Tags:          []uuid.UUID{uuid.New(), uuid.New()},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			contact, err := s.repo.CreateContact(s.ctx, tt.payload, s.testUser)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.NotEmpty(contact.ContactID)
			s.Equal(tt.payload.Name, contact.Name)

			// Check optional fields only if they are provided in the payload
			if tt.payload.Phone != nil {
				s.Equal(*tt.payload.Phone, *contact.Phone)
			}
			if tt.payload.Email != nil {
				s.Equal(*tt.payload.Email, *contact.Email)
			}
			if tt.payload.AddressLine1 != nil {
				s.Equal(*tt.payload.AddressLine1, *contact.AddressLine1)
			}
			if tt.payload.AddressLine2 != nil {
				s.Equal(*tt.payload.AddressLine2, *contact.AddressLine2)
			}
			if tt.payload.Country != nil {
				s.Equal(*tt.payload.Country, *contact.Country)
			}
			if tt.payload.City != nil {
				s.Equal(*tt.payload.City, *contact.City)
			}
			if tt.payload.StateProvince != nil {
				s.Equal(*tt.payload.StateProvince, *contact.StateProvince)
			}
			if tt.payload.ZipPostalCode != nil {
				s.Equal(*tt.payload.ZipPostalCode, *contact.ZipPostalCode)
			}
			if tt.payload.Tags != nil {
				s.Equal(tt.payload.Tags, contact.Tags)
			}

			s.NotZero(contact.CreatedAt)
			s.NotZero(contact.UpdatedAt)
		})
	}
}

func (s *ContactRepositoryTestSuite) TestGetContact() {
	// Create a test contact first
	createPayload := types.ContactCreatePayload{
		Name:  "Test Contact",
		Email: utils.StringPtr("test@example.com"),
		Phone: utils.StringPtr("+1-555-123-4567"),
	}
	created, err := s.repo.CreateContact(s.ctx, createPayload, s.testUser)
	require.NoError(s.T(), err)

	tests := []struct {
		name      string
		userID    uuid.UUID
		contactID uuid.UUID
		wantErr   bool
	}{
		{
			name:      "existing contact",
			userID:    s.testUser,
			contactID: created.ContactID,
			wantErr:   false,
		},
		{
			name:      "non-existent contact",
			userID:    s.testUser,
			contactID: uuid.New(),
			wantErr:   true,
		},
		{
			name:      "wrong user",
			userID:    uuid.New(),
			contactID: created.ContactID,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			contact, err := s.repo.GetContact(s.ctx, tt.contactID, tt.userID)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Equal(created.ContactID, contact.ContactID)
			s.Equal(created.Name, contact.Name)
			s.Equal(*created.Email, *contact.Email)
			s.Equal(*created.Phone, *contact.Phone)
		})
	}
}

func (s *ContactRepositoryTestSuite) TestUpdateContact() {
	// Create a test contact first
	createPayload := types.ContactCreatePayload{
		Name:          "Test Contact",
		Email:         utils.StringPtr("test@example.com"),
		Phone:         utils.StringPtr("+1-555-123-4567"),
		AddressLine1:  utils.StringPtr("123 Main St"),
		AddressLine2:  utils.StringPtr("Apt 4B"),
		Country:       utils.StringPtr("US"),
		City:          utils.StringPtr("New York"),
		StateProvince: utils.StringPtr("NY"),
		ZipPostalCode: utils.StringPtr("10001"),
		Tags:          []uuid.UUID{uuid.New(), uuid.New()},
	}
	created, err := s.repo.CreateContact(s.ctx, createPayload, s.testUser)
	require.NoError(s.T(), err)

	tests := []struct {
		name    string
		payload types.ContactUpdatePayload
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name: "valid update",
			payload: types.ContactUpdatePayload{
				ContactID: created.ContactID,
				Name:      "Updated Contact",
				Email:     utils.StringPtr("updated@example.com"),
				Phone:     utils.StringPtr("+1-555-987-6543"),
			},
			userID:  s.testUser,
			wantErr: false,
		},
		{
			name: "update with wrong user",
			payload: types.ContactUpdatePayload{
				ContactID: created.ContactID,
				Name:      "Should Not Update",
				Email:     utils.StringPtr("should.not@example.com"),
			},
			userID:  uuid.New(),
			wantErr: true,
		},
		{
			name: "update non-existent contact",
			payload: types.ContactUpdatePayload{
				ContactID: uuid.New(),
				Name:      "Non-existent",
				Email:     utils.StringPtr("nonexistent@example.com"),
			},
			userID:  s.testUser,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			contact, err := s.repo.UpdateContact(s.ctx, tt.payload, tt.userID)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Equal(tt.payload.ContactID, contact.ContactID)
			s.Equal(tt.payload.Name, contact.Name)
			s.Equal(*tt.payload.Email, *contact.Email)
			s.Equal(*tt.payload.Phone, *contact.Phone)
		})
	}
}

func (s *ContactRepositoryTestSuite) TestListContactsPaginated() {
	// Create test contacts in order from oldest to newest

	s.cleanContactTable()

	contacts := []types.ContactCreatePayload{
		{Name: "Contact 1", Email: utils.StringPtr("contact1@example.com")}, // Oldest
		{Name: "Contact 2", Email: utils.StringPtr("contact2@example.com")},
		{Name: "Contact 3", Email: utils.StringPtr("contact3@example.com")},
		{Name: "Contact 4", Email: utils.StringPtr("contact4@example.com")}, // Newest
	}

	var createdContacts []types.Contact
	for _, c := range contacts {
		time.Sleep(time.Millisecond * 100) // Ensure distinct timestamps
		contact, err := s.repo.CreateContact(s.ctx, c, s.testUser)
		s.Require().NoError(err)
		createdContacts = append(createdContacts, contact)
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
			cursorID:  uuid.New(),
			limit:     2,
			wantLen:   2,
			wantNames: []string{"Contact 4", "Contact 3"},
			wantErr:   false,
		},
		{
			name:      "get second page",
			cursor:    createdContacts[2].CreatedAt,
			cursorID:  createdContacts[2].ContactID,
			limit:     2,
			wantLen:   2,
			wantNames: []string{"Contact 2", "Contact 1"},
			wantErr:   false,
		},
		{
			name:      "get empty page",
			cursor:    createdContacts[0].CreatedAt,
			cursorID:  createdContacts[0].ContactID,
			limit:     2,
			wantLen:   0,
			wantNames: []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			contacts, err := s.repo.ListContactsPaginated(s.ctx, s.testUser, &tt.cursor, &tt.cursorID, tt.limit)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(contacts, tt.wantLen)

			if len(tt.wantNames) > 0 {
				actualNames := make([]string, len(contacts))
				for i, c := range contacts {
					actualNames[i] = c.Name
				}
				s.Equal(tt.wantNames, actualNames)
			}

			// Verify ordering for non-empty results
			if len(contacts) > 1 {
				for i := 1; i < len(contacts); i++ {
					isCorrectOrder := contacts[i-1].CreatedAt.After(contacts[i].CreatedAt) ||
						(contacts[i-1].CreatedAt.Equal(contacts[i].CreatedAt) &&
							contacts[i-1].ContactID.String() > contacts[i].ContactID.String())
					s.True(isCorrectOrder, "Contacts should be ordered by created_at DESC and then by contact_id DESC")
				}
			}
		})
	}
}

func (s *ContactRepositoryTestSuite) TestSearchContacts() {
	// Create test contacts with various names
	contacts := []types.ContactCreatePayload{
		{Name: "John Smith", Email: utils.StringPtr("john@example.com")},
		{Name: "John Doe", Email: utils.StringPtr("doe@example.com")},
		{Name: "Jane Smith", Email: utils.StringPtr("jane@example.com")},
		{Name: "Johnny Walker", Email: utils.StringPtr("walker@example.com")},
		{Name: "Jon Snow", Email: utils.StringPtr("snow@example.com")}, // Similar to "John"
		{Name: "Smith Family", Email: utils.StringPtr("family@example.com")},
		{Name: "Jhn Doe", Email: utils.StringPtr("jhn@example.com")}, // Misspelling
	}

	for _, c := range contacts {
		_, err := s.repo.CreateContact(s.ctx, c, s.testUser)
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
			name:      "search for John",
			query:     "John",
			limit:     10,
			wantLen:   5,
			wantNames: []string{"John Doe", "John Smith", "Johnny Walker", "Jhn Doe", "Jon Snow"},
			wantErr:   false,
		},
		{
			name:      "search for Smith",
			query:     "Smith",
			limit:     10,
			wantLen:   3,
			wantNames: []string{"John Smith", "Jane Smith", "Smith Family"},
			wantErr:   false,
		},
		{
			name:      "search with limit",
			query:     "John",
			limit:     2,
			wantLen:   2,
			wantNames: []string{"John Doe", "John Smith"},
			wantErr:   false,
		},
		{
			name:      "search with similar name",
			query:     "Jhn",
			limit:     10,
			wantLen:   3,
			wantNames: []string{"Jhn Doe", "John Doe", "John Smith"},
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
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			contacts, err := s.repo.SearchContacts(s.ctx, s.testUser, tt.query, tt.limit)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(contacts, tt.wantLen)

			if len(tt.wantNames) > 0 {
				actualNames := make([]string, len(contacts))
				for i, c := range contacts {
					actualNames[i] = c.Name
				}
				s.Equal(tt.wantNames, actualNames)
			}
		})
	}
}

func (s *ContactRepositoryTestSuite) TestSearchContactsByPhone() {
	// Create test contacts with clean phone numbers (no formatting characters)
	contacts := []types.ContactCreatePayload{
		{Name: "John Smith", Phone: utils.StringPtr("15551234567")}, // oldest
		{Name: "Jane Doe", Phone: utils.StringPtr("15551234568")},
		{Name: "Bob Wilson", Phone: utils.StringPtr("15559876543")},
		{Name: "Alice Brown", Phone: utils.StringPtr("5551234569")},
		{Name: "Charlie Davis", Phone: utils.StringPtr("442071234567")}, // UK format
		{Name: "David Miller", Phone: utils.StringPtr("15551234570")},
		{Name: "Eve Wilson", Phone: utils.StringPtr("15551230000")}, // newest
	}

	for _, c := range contacts {
		_, err := s.repo.CreateContact(s.ctx, c, s.testUser)
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
			name:      "search by exact phone",
			query:     "15551234567",
			limit:     10,
			wantLen:   1,
			wantNames: []string{"John Smith"}, // Exact match
			wantErr:   false,
		},
		{
			name:      "search by prefix (area code)",
			query:     "1555",
			limit:     10,
			wantLen:   5, // Only numbers starting with 1555
			wantNames: []string{"Eve Wilson", "David Miller", "Bob Wilson", "Jane Doe", "John Smith"},
			wantErr:   false,
		},
		{
			name:      "search by prefix (longer)",
			query:     "155512345",
			limit:     10,
			wantLen:   3, // Only numbers starting with 155512345
			wantNames: []string{"David Miller", "Jane Doe", "John Smith"},
			wantErr:   false,
		},
		{
			name:      "search with limit",
			query:     "1555",
			limit:     2,
			wantLen:   2,
			wantNames: []string{"Eve Wilson", "David Miller"},
			wantErr:   false,
		},
		{
			name:      "search UK number prefix",
			query:     "4420",
			limit:     10,
			wantLen:   1,
			wantNames: []string{"Charlie Davis"},
			wantErr:   false,
		},
		{
			name:      "no results for non-matching prefix",
			query:     "1999",
			limit:     10,
			wantLen:   0,
			wantNames: []string{},
			wantErr:   false,
		},
		{
			name:      "empty query returns all ordered by created_at",
			query:     "",
			limit:     10,
			wantLen:   7,
			wantNames: []string{"Eve Wilson", "David Miller", "Charlie Davis", "Alice Brown", "Bob Wilson", "Jane Doe", "John Smith"},
			wantErr:   false,
		},
		{
			name:      "search local number prefix",
			query:     "555",
			limit:     10,
			wantLen:   1, // Only the number that starts with 555 (no country code)
			wantNames: []string{"Alice Brown"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			contacts, err := s.repo.SearchContactsByPhone(s.ctx, s.testUser, tt.query, tt.limit)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(contacts, tt.wantLen)

			if len(tt.wantNames) > 0 {
				actualNames := make([]string, len(contacts))
				for i, c := range contacts {
					actualNames[i] = c.Name
				}
				s.Equal(tt.wantNames, actualNames, "Contact names should match in the expected order")
			}
		})
	}
}

func (s *ContactRepositoryTestSuite) runMigrations() error {
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
