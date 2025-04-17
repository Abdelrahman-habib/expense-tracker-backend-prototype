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
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	coreTypes "github.com/Abdelrahman-habib/expense-tracker/internal/core/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
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

type ContactIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	service   db.Service
	pool      *pgxpool.Pool
	handler   *handlers.ContactHandler
	router    *chi.Mux
	userID    uuid.UUID
	ctx       context.Context
}

func TestContactIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ContactIntegrationTestSuite))
}

func (s *ContactIntegrationTestSuite) SetupSuite() {
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
	s.clearContacts()

	// Create test user
	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, $2, 'cit_Test User', 'cit_test@example.com')
	`, s.userID, s.userID.String())
	require.NoError(s.T(), err)

	// Initialize components
	logger := zap.NewNop()
	repo := repository.New(dbService.Queries())
	contactService := service.NewContactService(repo, logger)
	s.handler = handlers.NewContactHandler(contactService, logger)

	// Setup router
	router := chi.NewRouter()
	router.Route("/contacts", func(r chi.Router) {
		r.Get("/search", s.handler.SearchContacts)
		r.Get("/paginated", s.handler.ListContactsPaginated)
		r.Post("/", s.handler.CreateContact)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handler.GetContact)
			r.Put("/", s.handler.UpdateContact)
			r.Delete("/", s.handler.DeleteContact)
		})
	})
	s.router = router
}

func (s *ContactIntegrationTestSuite) TearDownSuite() {
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

func (s *ContactIntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	s.clearContacts()
}

func (s *ContactIntegrationTestSuite) runMigrations() error {
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

func stringPtr(v string) *string {
	return &v
}

func (s *ContactIntegrationTestSuite) clearContacts() {
	_, err := s.pool.Exec(s.ctx, `DELETE FROM contacts WHERE user_id = $1`, s.userID)
	require.NoError(s.T(), err)
}

// Helper method to create a test contact
func (s *ContactIntegrationTestSuite) createTestContact() types.Contact {
	createPayload := types.ContactCreatePayload{
		Name:          "Integration Test Contact",
		Phone:         stringPtr("+1-555-123-4567"),
		Email:         stringPtr("test@example.com"),
		AddressLine1:  stringPtr("123 Main St"),
		AddressLine2:  stringPtr("Apt 4B"),
		Country:       stringPtr("US"),
		City:          stringPtr("New York"),
		StateProvince: stringPtr("NY"),
		ZipPostalCode: stringPtr("10001"),
		Tags:          []uuid.UUID{uuid.New(), uuid.New()},
	}

	payloadBytes, err := json.Marshal(createPayload)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)

	contactData := response["data"].(map[string]interface{})
	return types.Contact{
		ContactID: uuid.MustParse(contactData["contactId"].(string)),
		Name:      contactData["name"].(string),
		Phone:     stringPtr(contactData["phone"].(string)),
	}
}

// Helper method for making authenticated requests
func (s *ContactIntegrationTestSuite) newAuthenticatedRequest(method, path string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, path, body)
	return req.WithContext(context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID))
}

// Helper method to verify contact state
func (s *ContactIntegrationTestSuite) verifyContactState(contactID uuid.UUID, expectedName string, expectedPhone *string) {
	req := s.newAuthenticatedRequest(http.MethodGet, "/contacts/"+contactID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", contactID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)
	getData := response["data"].(map[string]interface{})
	s.Equal(expectedName, getData["name"])
	if expectedPhone != nil {
		s.Equal(*expectedPhone, getData["phone"])
	}
}

func (s *ContactIntegrationTestSuite) TestContactLifecycle() {
	// Create a contact and use it across all tests
	contact := &types.Contact{}
	*contact = s.createTestContact()

	s.Run("get contact", func() {
		s.testGetContact(contact)
	})
	s.Run("update contact name", func() {
		s.testUpdateContactName(contact)
	})
	s.Run("update contact phone", func() {
		s.testUpdateContactPhone(contact)
	})
	s.Run("delete contact", func() {
		s.testDeleteContact(contact)
	})
}

func (s *ContactIntegrationTestSuite) testGetContact(contact *types.Contact) {
	req := s.newAuthenticatedRequest(http.MethodGet, "/contacts/"+contact.ContactID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", contact.ContactID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)

	getData := response["data"].(map[string]interface{})
	s.Equal(contact.Name, getData["name"])
	s.Equal(*contact.Phone, getData["phone"])
}

func (s *ContactIntegrationTestSuite) testUpdateContactName(contact *types.Contact) {
	updatePayload := types.ContactUpdatePayload{
		ContactID: contact.ContactID,
		Name:      "Updated Contact Name",
		Phone:     contact.Phone,
	}

	payloadBytes, err := json.Marshal(updatePayload)
	s.Require().NoError(err)

	req := s.newAuthenticatedRequest(http.MethodPut, "/contacts/"+contact.ContactID.String(), bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", contact.ContactID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	contact.Name = updatePayload.Name
	s.verifyContactState(contact.ContactID, contact.Name, contact.Phone)
}

func (s *ContactIntegrationTestSuite) testUpdateContactPhone(contact *types.Contact) {
	updatePayload := types.ContactUpdatePayload{
		ContactID: contact.ContactID,
		Name:      contact.Name,
		Phone:     stringPtr("+1-555-987-6543"), // should be 15559876543
	}

	payloadBytes, err := json.Marshal(updatePayload)
	s.Require().NoError(err)

	req := s.newAuthenticatedRequest(http.MethodPut, "/contacts/"+contact.ContactID.String(), bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", contact.ContactID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	contact.Phone = stringPtr("15559876543")
	s.verifyContactState(contact.ContactID, contact.Name, contact.Phone)
}

func (s *ContactIntegrationTestSuite) testDeleteContact(contact *types.Contact) {
	req := s.newAuthenticatedRequest(http.MethodDelete, "/contacts/"+contact.ContactID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", contact.ContactID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Verify contact is deleted
	req = s.newAuthenticatedRequest(http.MethodGet, "/contacts/"+contact.ContactID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusNotFound, w.Code)
}

// Helper method to create multiple test contacts
func (s *ContactIntegrationTestSuite) createTestContacts(count int) []types.Contact {
	contacts := make([]types.Contact, count)
	// Create contacts in order (1 to count)
	for i := 0; i < count; i++ {
		createPayload := types.ContactCreatePayload{
			Name:  fmt.Sprintf("Test Contact %d", i+1), // Start from 1 and increment
			Phone: stringPtr(fmt.Sprintf("+1-555-%03d-%04d", i+1, i+1)),
			Email: stringPtr(fmt.Sprintf("contact%d@example.com", i+1)),
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		s.Require().Equal(http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		s.Require().NoError(err)

		contactData := response["data"].(map[string]interface{})
		createdAt, err := time.Parse(time.RFC3339, contactData["createdAt"].(string))
		s.Require().NoError(err)

		contacts[count-1-i] = types.Contact{ // Store in reverse order
			ContactID: uuid.MustParse(contactData["contactId"].(string)),
			Name:      contactData["name"].(string),
			Phone:     stringPtr(contactData["phone"].(string)),
			Email:     stringPtr(contactData["email"].(string)),
			CreatedAt: createdAt,
		}
		time.Sleep(time.Millisecond * 10) // Ensure distinct timestamps
	}
	return contacts
}

func (s *ContactIntegrationTestSuite) TestListContactsPaginated() {
	// Clear contacts table
	s.clearContacts()

	// Create 10 test contacts
	contacts := s.createTestContacts(10)

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
				"next_token": coreTypes.EncodeCursor(contacts[4].CreatedAt, contacts[4].ContactID),
			},
			expectedStatus:  http.StatusOK,
			expectedLen:     5,
			expectedLimit:   "5",
			expectNextToken: true,
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
			urlPath := "/contacts/paginated"
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

				contacts := response["data"].([]interface{})
				s.Len(contacts, tt.expectedLen)

				meta := response["meta"].(map[string]interface{})
				if tt.expectedLimit != "" {
					s.Equal(tt.expectedLimit, fmt.Sprint(meta["limit"]))
				}

				if tt.expectNextToken {
					s.NotEmpty(meta["next_token"])
				} else {
					nextToken, exists := meta["next_token"]
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

func (s *ContactIntegrationTestSuite) TestPaginationEdgeCases() {
	// Test extreme pagination cases
	s.Run("pagination edge cases", func() {
		// Create 20 contacts for testing
		_ = s.createTestContacts(20)

		tests := []struct {
			name          string
			limit         int32
			expectedCode  int
			expectedErr   string
			expectedLimit float64
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
				expectedLimit: float64(coreTypes.MaxLimit),
			},
		}

		for _, tt := range tests {
			s.Run(tt.name, func() {
				urlPath := fmt.Sprintf("/contacts/paginated?limit=%d", tt.limit)
				req := httptest.NewRequest(http.MethodGet, urlPath, nil)
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
				req = req.WithContext(ctx)

				w := httptest.NewRecorder()
				s.router.ServeHTTP(w, req)

				s.Equal(tt.expectedCode, w.Code)

				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				s.Require().NoError(err)

				if tt.expectedCode != http.StatusOK {

					s.Contains(response["error"].(string), tt.expectedErr)
				} else {
					if tt.expectedLimit != 0 {
						metadata := response["meta"].(map[string]interface{})
						s.Equal(tt.expectedLimit, metadata["limit"])
					}
				}
			})
		}
	})
}

func (s *ContactIntegrationTestSuite) TestSearchContacts() {
	// Create test contacts with more distinct names and phone numbers
	contacts := []types.ContactCreatePayload{
		{Name: "Contact Alpha", Phone: stringPtr("+1-555-111-0001")},
		{Name: "Beta Person", Phone: stringPtr("+1-555-222-0002")},
		{Name: "Gamma Contact", Phone: stringPtr("+1-555-333-0003")},
		{Name: "Delta Personal", Phone: stringPtr("+1-555-444-0004")},
		{Name: "Contact Management", Phone: stringPtr("+1-555-555-0005")},
		{Name: "Contact Mnagement", Phone: stringPtr("+1-555-666-0006")},    // Misspelling of "Management"
		{Name: "Cntact Management", Phone: stringPtr("+1-555-777-0007")},    // Missing 'o'
		{Name: "Personal Contact", Phone: stringPtr("+1-555-888-0008")},     // Different type
		{Name: "Private Person", Phone: stringPtr("+1-555-999-0009")},       // Completely different
		{Name: "Business Contact", Phone: stringPtr("+1-555-000-0010")},     // Completely different
		{Name: "Contact #123", Phone: stringPtr("+1-555-123-0011")},         // With special characters
		{Name: "Alpha (Beta) Contact", Phone: stringPtr("+1-555-456-0012")}, // With parentheses
	}

	// Create all test contacts
	for _, c := range contacts {
		payloadBytes, err := json.Marshal(c)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewReader(payloadBytes))
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
		byPhone        bool
		limit          string
		expectedStatus int
		expectedCount  int
		expectedNames  []string // Expected contact names in order
		expectedError  string   // Expected error message
	}{
		{
			name:           "case insensitive search",
			query:          "contact",
			expectedStatus: http.StatusOK,
			expectedCount:  9,
			expectedNames:  []string{"Contact #123", "Gamma Contact", "Contact Alpha", "Personal Contact", "Business Contact", "Contact Mnagement", "Contact Management", "Alpha (Beta) Contact", "Cntact Management"},
		},
		{
			name:           "similarity search - management misspelling",
			query:          "Management",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			expectedNames:  []string{"Cntact Management", "Contact Management", "Contact Mnagement"},
		},
		{
			name:           "similarity search - missing letter",
			query:          "Cntact",
			expectedStatus: http.StatusOK,
			expectedCount:  9,
			expectedNames:  []string{"Cntact Management", "Contact #123", "Contact Alpha", "Gamma Contact", "Personal Contact", "Business Contact", "Contact Mnagement", "Contact Management", "Alpha (Beta) Contact"},
		},
		{
			name:           "person search",
			query:          "Person",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedNames:  []string{"Beta Person", "Private Person", "Delta Personal", "Personal Contact"},
		},
		{
			name:           "with custom limit",
			query:          "Person",
			limit:          "1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedNames:  []string{"Beta Person"},
		},
		{
			name:           "phone search",
			query:          "1-555",
			byPhone:        true,
			limit:          "30",
			expectedStatus: http.StatusOK,
			expectedCount:  len(contacts), // Should find all contacts
		},
		{
			name:           "specific phone search",
			query:          "+1-555-111-0001",
			byPhone:        true,
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedNames:  []string{"Contact Alpha"},
		},
		{
			name:           "empty query",
			query:          "",
			limit:          fmt.Sprint(len(contacts)),
			expectedStatus: http.StatusOK,
			expectedCount:  len(contacts),
		},
		{
			name:           "query too long",
			query:          strings.Repeat("a", 101), // Exceeds maxQueryLength
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit",
			query:          "Contact",
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
			expectedNames:  []string{"Contact #123"},
		},
		{
			name:           "parentheses",
			query:          "(",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedNames:  []string{"Alpha (Beta) Contact"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Build URL with query parameters
			urlPath := fmt.Sprintf("/contacts/search?q=%s", url.QueryEscape(tt.query))
			if tt.limit != "" {
				urlPath += fmt.Sprintf("&limit=%s", tt.limit)
			}
			if tt.byPhone {
				urlPath += "&by_phone=true"
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

				contacts := response["data"].([]interface{})
				s.Len(contacts, tt.expectedCount)
				metadata := response["meta"].(map[string]interface{})
				if tt.query != "" {
					s.Equal(tt.query, metadata["query"])
				}
				if tt.limit != "" {
					limit, _ := strconv.ParseFloat(tt.limit, 64)
					s.Equal(limit, metadata["limit"])
				}

				// Verify contact names if expected names are provided
				if len(tt.expectedNames) > 0 {
					actualNames := make([]string, len(contacts))
					for i, c := range contacts {
						contact := c.(map[string]interface{})
						actualNames[i] = contact["name"].(string)
					}
					s.Equal(tt.expectedNames, actualNames)
				}
			}
		})
	}
}

func (s *ContactIntegrationTestSuite) TestConcurrentUpdates() {
	// Create a contact
	contact := s.createTestContact()

	// Try to update the same contact concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			updatePayload := types.ContactUpdatePayload{
				ContactID: contact.ContactID,
				Name:      fmt.Sprintf("Updated Name %d", i),
				Phone:     stringPtr(fmt.Sprintf("+1-555-%03d-%04d", i+1, i+1)),
			}

			payloadBytes, err := json.Marshal(updatePayload)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/contacts/"+contact.ContactID.String(), bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", contact.ContactID.String())
			req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)

			// All updates should succeed
			s.Equal(http.StatusOK, w.Code)
		}(i)
	}
	wg.Wait()
}

func (s *ContactIntegrationTestSuite) TestDatabaseConstraintsAndValidation() {
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
					"phone": "+1-555-123-4567",
					// name missing
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "name: cannot be blank",
				errorMessage:  "Invalid request",
			},
			{
				name: "name too long",
				payload: map[string]interface{}{
					"name":  strings.Repeat("a", 256),
					"phone": "+1-555-123-4567",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "name: the length must be between 1 and 255",
				errorMessage:  "Invalid request",
			},
			{
				name: "invalid phone format",
				payload: map[string]interface{}{
					"name":  "Test Contact",
					"phone": "not-a-phone",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "phone: invalid phone number format",
				errorMessage:  "Invalid request",
			},
			{
				name: "invalid email format",
				payload: map[string]interface{}{
					"name":  "Test Contact",
					"phone": "+1-555-123-4567",
					"email": "not-an-email",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "email: must be a valid email address",
				errorMessage:  "Invalid request",
			},
			{
				name: "too many tags",
				payload: map[string]interface{}{
					"name":  "Test Contact",
					"phone": "+1-555-123-4567",
					"tags":  []string{uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()}, // Exceeds MaxTagsCount
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "tags: the length must be no more than 10",
				errorMessage:  "Invalid request",
			},
			{
				name: "invalid tag UUID",
				payload: map[string]interface{}{
					"name":  "Test Contact",
					"phone": "+1-555-123-4567",
					"tags":  []string{"not-a-uuid"},
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "invalid UUID",
				errorMessage:  "Invalid request",
			},
			{
				name: "address line too long",
				payload: map[string]interface{}{
					"name":         "Test Contact",
					"phone":        "+1-555-123-4567",
					"addressLine1": strings.Repeat("a", 256),
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "address_line1: the length must be between 1 and 255",
				errorMessage:  "Invalid request",
			},
		}

		for _, tt := range tests {
			s.Run(tt.name, func() {
				payloadBytes, err := json.Marshal(tt.payload)
				s.Require().NoError(err)

				req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewReader(payloadBytes))
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

func (s *ContactIntegrationTestSuite) TestUnauthorizedAccess() {
	// Create a contact first
	contact := s.createTestContacts(1)[0]

	// Create another user
	otherUserID := uuid.New()
	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, $2, 'cit_Other User', 'cit_other@example.com')
	`, otherUserID, otherUserID.String())
	s.Require().NoError(err)

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		expectedCode int
	}{
		{
			name: "access without user ID",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/contacts/"+contact.ContactID.String(), nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", contact.ContactID.String())
				return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "access with wrong user",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/contacts/"+contact.ContactID.String(), nil)
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, otherUserID)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", contact.ContactID.String())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "update with wrong user",
			setupRequest: func() *http.Request {
				payload := types.ContactUpdatePayload{
					ContactID: contact.ContactID,
					Name:      "Unauthorized Update",
					Phone:     stringPtr("+1-555-999-9999"),
				}
				payloadBytes, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/contacts/"+contact.ContactID.String(), bytes.NewReader(payloadBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, otherUserID)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", contact.ContactID.String())
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

func (s *ContactIntegrationTestSuite) TestComplexContactLifecycle() {
	// Test the complete lifecycle of a contact with multiple operations
	s.Run("full contact lifecycle", func() {
		// 1. Create contact
		createPayload := types.ContactCreatePayload{
			Name:          "Lifecycle Contact",
			Phone:         stringPtr("+1-555-123-4567"),
			Email:         stringPtr("lifecycle@example.com"),
			AddressLine1:  stringPtr("123 Main St"),
			AddressLine2:  stringPtr("Apt 4B"),
			Country:       stringPtr("US"),
			City:          stringPtr("New York"),
			StateProvince: stringPtr("NY"),
			ZipPostalCode: stringPtr("10001"),
			Tags:          []uuid.UUID{uuid.New(), uuid.New()},
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		s.Require().NoError(err)
		contactData := response["data"].(map[string]interface{})
		contactID := contactData["contactId"].(string)

		// 2. Update contact multiple times with different fields
		updates := []types.ContactUpdatePayload{
			{
				ContactID:    uuid.MustParse(contactID),
				Name:         "Updated Name",
				Phone:        stringPtr("+1-555-987-6543"), // will be 15559876543
				Email:        stringPtr("updated@example.com"),
				AddressLine1: stringPtr("456 Main St"),
			},
			{
				ContactID:    uuid.MustParse(contactID),
				Name:         "Updated Name",
				Phone:        stringPtr("+1-555-987-6543"),
				Email:        stringPtr("final@example.com"),
				AddressLine1: stringPtr("789 Main St"),
				Tags:         []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
			},
		}

		for _, update := range updates {
			payloadBytes, err = json.Marshal(update)
			s.Require().NoError(err)

			req = httptest.NewRequest(http.MethodPut, "/contacts/"+contactID, bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx = context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			req = req.WithContext(ctx)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", contactID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w = httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			s.Equal(http.StatusOK, w.Code)
		}

		// 3. Verify final state
		req = httptest.NewRequest(http.MethodGet, "/contacts/"+contactID, nil)
		ctx = context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", contactID)
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
		s.Equal("15559876543", finalData["phone"])
		s.Equal("final@example.com", finalData["email"])
		s.Equal("789 Main St", finalData["addressLine1"])
		tags := finalData["tags"].([]interface{})
		s.Len(tags, 3)
	})
}

func (s *ContactIntegrationTestSuite) TestResponsePayloadStructure() {
	s.Run("response payload structure", func() {
		// Create a contact with all fields
		createPayload := types.ContactCreatePayload{
			Name:          "Response Test Contact",
			Phone:         stringPtr("+1-555-123-4567"), // 15551234567
			Email:         stringPtr("response@example.com"),
			AddressLine1:  stringPtr("123 Main St"),
			AddressLine2:  stringPtr("Apt 4B"),
			Country:       stringPtr("US"),
			City:          stringPtr("New York"),
			StateProvince: stringPtr("NY"),
			ZipPostalCode: stringPtr("10001"),
			Tags:          []uuid.UUID{uuid.New(), uuid.New()},
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewReader(payloadBytes))
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
		s.Require().Contains(data, "contactId")
		s.IsType("", data["contactId"].(string))
		_, err = uuid.Parse(data["contactId"].(string))
		s.NoError(err)

		s.Equal(createPayload.Name, data["name"])
		s.Equal("15551234567", data["phone"])
		s.Equal(*createPayload.Email, data["email"])
		s.Equal(*createPayload.AddressLine1, data["addressLine1"])
		s.Equal(*createPayload.AddressLine2, data["addressLine2"])
		s.Equal(*createPayload.Country, data["country"])
		s.Equal(*createPayload.City, data["city"])
		s.Equal(*createPayload.StateProvince, data["stateProvince"])
		s.Equal(*createPayload.ZipPostalCode, data["zipPostalCode"])

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
	})
}
