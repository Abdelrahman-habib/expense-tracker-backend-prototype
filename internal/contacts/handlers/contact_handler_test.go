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

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	coreTypes "github.com/Abdelrahman-habib/expense-tracker/internal/core/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock service
type mockContactService struct {
	mock.Mock
}

func (m *mockContactService) GetContact(ctx context.Context, contactID, userID uuid.UUID) (types.Contact, error) {
	args := m.Called(ctx, contactID, userID)
	if args.Get(0) == nil {
		return types.Contact{}, args.Error(1)
	}
	return args.Get(0).(types.Contact), args.Error(1)
}

func (m *mockContactService) ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]types.Contact), args.Error(1)
}

func (m *mockContactService) ListContactsPaginated(ctx context.Context, userID uuid.UUID, cursor *time.Time, cursorID *uuid.UUID, limit int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, cursor, cursorID, limit)
	return args.Get(0).([]types.Contact), args.Error(1)
}

func (m *mockContactService) CreateContact(ctx context.Context, payload types.ContactCreatePayload, userID uuid.UUID) (types.Contact, error) {
	args := m.Called(ctx, payload, userID)
	return args.Get(0).(types.Contact), args.Error(1)
}

func (m *mockContactService) UpdateContact(ctx context.Context, payload types.ContactUpdatePayload, userID uuid.UUID) (types.Contact, error) {
	args := m.Called(ctx, payload, userID)
	return args.Get(0).(types.Contact), args.Error(1)
}

func (m *mockContactService) DeleteContact(ctx context.Context, contactID, userID uuid.UUID) error {
	args := m.Called(ctx, contactID, userID)
	return args.Error(0)
}

func (m *mockContactService) SearchContacts(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Contact), args.Error(1)
}

func (m *mockContactService) SearchContactsByPhone(ctx context.Context, userID uuid.UUID, phone string, limit int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, phone, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Contact), args.Error(1)
}

func setupTest(t *testing.T) (*mockContactService, *ContactHandler) {
	mockService := new(mockContactService)
	logger := zap.NewNop()
	handler := NewContactHandler(mockService, logger)
	return mockService, handler
}

// Helper function to create string pointer
func stringPtr(v string) *string {
	return &v
}

func TestContactHandler_CreateContact(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		payload        string
		setupAuth      bool
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful creation",
			payload: fmt.Sprintf(`{
				"name": "John Doe",
				"phone": "+1-555-123-4567",
				"email": "john@example.com",
				"addressLine1": "123 Main St",
				"addressLine2": "Apt 4B",
				"country": "US",
				"city": "New York",
				"stateProvince": "NY",
				"zipPostalCode": "10001",
				"tags": [%q, %q]
			}`, uuid.New(), uuid.New()),
			setupAuth: true,
			setupMock: func() {
				expectedContact := types.Contact{
					ContactID: uuid.New(),
					Name:      "John Doe",
					Phone:     stringPtr("15551234567"),
					Tags:      []uuid.UUID{uuid.New(), uuid.New()},
				}
				mockService.On("CreateContact", mock.Anything, mock.AnythingOfType("types.ContactCreatePayload"), userID).
					Return(expectedContact, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty payload",
			payload:        `{}`,
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name: cannot be blank",
		},
		{
			name: "name too long",
			payload: fmt.Sprintf(`{
				"name": "%s"
			}`, strings.Repeat("a", types.MaxNameLength+1)),
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name: the length must be between 1 and 255.",
		},
		{
			name: "invalid email format",
			payload: `{
				"name": "John Doe",
				"email": "not-an-email"
			}`,
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email: must be a valid email address.",
		},
		{
			name: "too many tags",
			payload: fmt.Sprintf(`{
				"name": "John Doe",
				"tags": [%q, %q, %q, %q, %q, %q, %q, %q, %q, %q, %q]
			}`, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(),
				uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()),
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "tags: the length must be no more than 10.",
		},
		{
			name: "duplicate tags",
			payload: func() string {
				tagID := uuid.New()
				return fmt.Sprintf(`{
					"name": "John Doe",
					"tags": [%q, %q]
				}`, tagID, tagID)
			}(),
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "tags: contains duplicate elements.",
		},
		{
			name: "invalid phone format",
			payload: `{
				"name": "John Doe",
				"phone": "not-a-phone"
			}`,
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "phone: invalid phone number format.",
		},
		{
			name: "address line too long",
			payload: fmt.Sprintf(`{
				"name": "John Doe",
				"addressLine1": "%s"
			}`, strings.Repeat("a", types.MaxAddressLength+1)),
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "address_line1: the length must be between 1 and 255.",
		},
		{
			name: "invalid json syntax",
			payload: `{
				"name": "John Doe",
				invalid json here
			}`,
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid character",
		},
		{
			name:           "missing auth",
			payload:        `{"name": "John Doe"}`,
			setupAuth:      false,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing user ID",
		},
		{
			name: "service error",
			payload: `{
				"name": "John Doe"
			}`,
			setupAuth: true,
			setupMock: func() {
				mockService.On("CreateContact", mock.Anything, mock.AnythingOfType("types.ContactCreatePayload"), userID).
					Return(types.Contact{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPost, "/contacts", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.CreateContact(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, float64(http.StatusCreated), response["status"])
				assert.NotNil(t, response["data"])
			} else {
				if tt.expectedError != "" {
					errMsg, ok := response["error"].(string)
					assert.True(t, ok)
					assert.Contains(t, errMsg, tt.expectedError)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestContactHandler_GetContact(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		contactID      string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:      "successful retrieval",
			setupAuth: true,
			contactID: contactID.String(),
			setupMock: func() {
				expectedContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
					Phone:     stringPtr("15551234567"),
					Tags:      []uuid.UUID{uuid.New()},
				}
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(expectedContact, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid contact ID",
			setupAuth:      true,
			contactID:      "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing auth",
			setupAuth:      false,
			contactID:      contactID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodGet, "/contacts/"+tt.contactID, nil)

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.contactID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.GetContact(w, req)

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

func TestContactHandler_ListContactsPaginated(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	now := time.Now()
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
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "John Doe",
						CreatedAt: now.Add(-1 * time.Hour),
					},
					{
						ContactID: uuid.New(),
						Name:      "Jane Smith",
						CreatedAt: now.Add(-2 * time.Hour),
					},
				}
				mockService.On("ListContactsPaginated",
					mock.Anything,
					userID,
					mock.Anything,
					mock.MatchedBy(func(id *uuid.UUID) bool {
						return id == nil
					}),
					int32(coreTypes.DefaultLimit),
				).Return(contacts, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
			expectedLimit:  "10",
		},
		{
			name:        "first page with partial results",
			setupAuth:   true,
			queryParams: map[string]string{"limit": "5"},
			setupMock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "John Doe",
						CreatedAt: now.Add(-1 * time.Hour),
					},
				}
				mockService.On("ListContactsPaginated",
					mock.Anything,
					userID,
					mock.Anything,
					mock.MatchedBy(func(id *uuid.UUID) bool {
						return id == nil
					}),
					int32(5),
				).Return(contacts, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    1,
			expectedLimit:  "5",
		},
		{
			name:      "successful pagination with next_token",
			setupAuth: true,
			queryParams: map[string]string{
				"next_token": coreTypes.EncodeCursor(now, cursorID),
			},
			setupMock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "John Doe",
						CreatedAt: now.Add(-1 * time.Hour),
					},
					{
						ContactID: uuid.New(),
						Name:      "Jane Smith",
						CreatedAt: now.Add(-2 * time.Hour),
					},
				}
				mockService.On("ListContactsPaginated",
					mock.Anything,
					userID,
					mock.MatchedBy(func(t *time.Time) bool {
						return t.Truncate(time.Second).Equal(now.Truncate(time.Second))
					}),
					mock.MatchedBy(func(id *uuid.UUID) bool {
						return *id == cursorID
					}),
					int32(10),
				).Return(contacts, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:        "successful pagination with default values",
			setupAuth:   true,
			queryParams: map[string]string{},
			setupMock: func() {
				mockService.On("ListContactsPaginated",
					mock.Anything,
					userID,
					mock.Anything,
					mock.Anything,
					int32(coreTypes.DefaultLimit),
				).Return([]types.Contact{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name:      "invalid next_token format",
			setupAuth: true,
			queryParams: map[string]string{
				"next_token": "invalid-base64-token",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid token",
		},
		{
			name:      "invalid limit value",
			setupAuth: true,
			queryParams: map[string]string{
				"limit": "not-a-number",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid limit format",
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
			name:      "limit above maximum",
			setupAuth: true,
			queryParams: map[string]string{
				"limit": fmt.Sprintf("%d", coreTypes.MaxLimit+1),
			},
			setupMock: func() {
				mockService.On("ListContactsPaginated",
					mock.Anything,
					userID,
					mock.Anything,
					mock.Anything,
					int32(coreTypes.MaxLimit),
				).Return([]types.Contact{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLimit:  fmt.Sprint(coreTypes.MaxLimit),
			expectedLen:    0,
		},
		{
			name:           "missing auth",
			setupAuth:      false,
			queryParams:    map[string]string{},
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing user ID",
		},
		{
			name:      "service error",
			setupAuth: true,
			queryParams: map[string]string{
				"limit": "10",
			},
			setupMock: func() {
				mockService.On("ListContactsPaginated",
					mock.Anything,
					userID,
					mock.Anything,
					mock.Anything,
					int32(10),
				).Return([]types.Contact{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			reqURL := "/contacts/paginated"
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
			handler.ListContactsPaginated(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, float64(http.StatusOK), response["status"])

				contacts := response["data"].([]interface{})
				assert.Len(t, contacts, tt.expectedLen)

				meta := response["meta"].(map[string]interface{})
				if tt.expectedLimit != "" {
					assert.Equal(t, tt.expectedLimit, fmt.Sprint(meta["limit"]))
				} else {
					if tt.queryParams["limit"] != "" {
						assert.Equal(t, tt.queryParams["limit"], fmt.Sprint(meta["limit"]))
					}
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

func TestContactHandler_SearchContacts(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		queryParams    map[string]string
		setupMock      func()
		expectedStatus int
		expectedError  string
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:      "successful search by name",
			setupAuth: true,
			queryParams: map[string]string{
				"q":     "John",
				"limit": "20",
			},
			setupMock: func() {
				contacts := []types.Contact{
					{ContactID: uuid.New(), Name: "John Doe"},
					{ContactID: uuid.New(), Name: "Johnny Smith"},
				}
				mockService.On("SearchContacts", mock.Anything, userID, "John", int32(20)).
					Return(contacts, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				data := response["data"].([]interface{})
				assert.Len(t, data, 2)

				meta := response["meta"].(map[string]interface{})
				assert.Equal(t, "John", meta["query"])
				assert.Equal(t, float64(20), meta["limit"])
				assert.Equal(t, float64(2), meta["count"])
			},
		},
		{
			name:      "successful search by phone",
			setupAuth: true,
			queryParams: map[string]string{
				"q":        "555",
				"by_phone": "true",
				"limit":    "20",
			},
			setupMock: func() {
				contacts := []types.Contact{
					{ContactID: uuid.New(), Name: "John Doe", Phone: stringPtr("15551234567")},
				}
				mockService.On("SearchContactsByPhone", mock.Anything, userID, "555", int32(20)).
					Return(contacts, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				metadata := response["meta"].(map[string]interface{})
				assert.Equal(t, "555", metadata["query"])
				assert.Equal(t, float64(20), metadata["limit"])
				assert.Equal(t, float64(1), metadata["count"])
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
				mockService.On("SearchContacts", mock.Anything, userID, "test", int32(coreTypes.DefaultSearchLimit)).
					Return([]types.Contact(nil), fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "empty query parameter returns all contacts",
			setupAuth: true,
			queryParams: map[string]string{
				"q": "",
			},
			setupMock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "Recent Contact",
						CreatedAt: time.Now().Add(-1 * time.Hour),
					},
					{
						ContactID: uuid.New(),
						Name:      "Older Contact",
						CreatedAt: time.Now().Add(-2 * time.Hour),
					},
				}
				mockService.On("SearchContacts", mock.Anything, userID, "", int32(coreTypes.DefaultSearchLimit)).
					Return(contacts, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				data := response["data"].([]interface{})
				assert.Len(t, data, 2)

				meta := response["meta"].(map[string]interface{})
				assert.Equal(t, nil, meta["query"])
				assert.Equal(t, float64(coreTypes.DefaultSearchLimit), meta["limit"])
				assert.Equal(t, float64(2), meta["count"])
			},
		},
		{
			name:      "whitespace query parameter returns all contacts",
			setupAuth: true,
			queryParams: map[string]string{
				"q": "   ",
			},
			setupMock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "Recent Contact",
						CreatedAt: time.Now().Add(-1 * time.Hour),
					},
					{
						ContactID: uuid.New(),
						Name:      "Older Contact",
						CreatedAt: time.Now().Add(-2 * time.Hour),
					},
				}
				mockService.On("SearchContacts", mock.Anything, userID, "", int32(coreTypes.DefaultSearchLimit)).
					Return(contacts, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				data := response["data"].([]interface{})
				assert.Len(t, data, 2)

				meta := response["meta"].(map[string]interface{})
				assert.Equal(t, nil, meta["query"])
				assert.Equal(t, float64(coreTypes.DefaultSearchLimit), meta["limit"])
				assert.Equal(t, float64(2), meta["count"])
			},
		},
		{
			name:      "limit exceeds maximum",
			setupAuth: true,
			queryParams: map[string]string{
				"q":     "John",
				"limit": "1001",
			},
			setupMock: func() {
				mockService.On("SearchContacts", mock.Anything, userID, "John", int32(coreTypes.MaxSearchLimit)).
					Return([]types.Contact{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				meta := response["meta"].(map[string]interface{})
				assert.Equal(t, float64(coreTypes.MaxSearchLimit), meta["limit"])
			},
		},
		{
			name:      "invalid phone format for phone search",
			setupAuth: true,
			queryParams: map[string]string{
				"q":        "abc",
				"by_phone": "true",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "query: invalid phone number format.",
		},
		{
			name:      "empty result set",
			setupAuth: true,
			queryParams: map[string]string{
				"q": "NonexistentName",
			},
			setupMock: func() {
				mockService.On("SearchContacts", mock.Anything, userID, "NonexistentName", int32(coreTypes.DefaultSearchLimit)).
					Return([]types.Contact{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				items := response["data"].([]interface{})
				assert.Len(t, items, 0)
				meta := response["meta"].(map[string]interface{})
				assert.Equal(t, nil, meta["count"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			reqURL := "/contacts/search"
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
			handler.SearchContacts(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(http.StatusOK), response["status"])

				if tt.checkResponse != nil {
					tt.checkResponse(t, response)
				}
			} else if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				errMsg, ok := response["error"].(string)
				assert.True(t, ok)
				assert.Contains(t, errMsg, tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestContactHandler_DeleteContact(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name           string
		contactID      string
		setupAuth      bool
		setupMock      func()
		expectedStatus int
	}{
		{
			name:      "successful deletion",
			contactID: contactID.String(),
			setupAuth: true,
			setupMock: func() {
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(types.Contact{ContactID: contactID}, nil)
				mockService.On("DeleteContact", mock.Anything, contactID, userID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid contact ID",
			contactID:      "invalid-uuid",
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "contact not found",
			contactID: uuid.New().String(),
			setupAuth: true,
			setupMock: func() {
				mockService.On("GetContact", mock.Anything, mock.AnythingOfType("uuid.UUID"), userID).
					Return(types.Contact{}, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing auth",
			contactID:      contactID.String(),
			setupAuth:      false,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodDelete, "/contacts/"+tt.contactID, nil)

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.contactID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.DeleteContact(w, req)

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

func TestContactHandler_UpdateContact(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name           string
		contactID      string
		payload        string
		setupAuth      bool
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful update",
			contactID: contactID.String(),
			payload: fmt.Sprintf(`{
				"name": "John Doe Updated",
				"phone": "+1-555-123-4567",
				"email": "john.updated@example.com",
				"addressLine1": "456 Main St",
				"addressLine2": "Apt 5C",
				"country": "US",
				"city": "New York",
				"stateProvince": "NY",
				"zipPostalCode": "10002",
				"tags": [%q, %q, %q]
			}`, uuid.New(), uuid.New(), uuid.New()),
			setupAuth: true,
			setupMock: func() {
				existingContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
					Phone:     stringPtr("15551234567"),
				}
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(existingContact, nil)

				updatedContact := types.Contact{
					ContactID:     contactID,
					Name:          "John Doe Updated",
					Phone:         stringPtr("15551234567"),
					Email:         stringPtr("john.updated@example.com"),
					AddressLine1:  stringPtr("456 Main St"),
					AddressLine2:  stringPtr("Apt 5C"),
					Country:       stringPtr("US"),
					City:          stringPtr("New York"),
					StateProvince: stringPtr("NY"),
					ZipPostalCode: stringPtr("10002"),
				}
				mockService.On("UpdateContact", mock.Anything, mock.AnythingOfType("types.ContactUpdatePayload"), userID).
					Return(updatedContact, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "empty payload (uses existing data)",
			contactID: contactID.String(),
			payload:   `{}`,
			setupAuth: true,
			setupMock: func() {
				existingContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
					Phone:     stringPtr("15551234567"),
				}
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(existingContact, nil)

				// Should use existing contact data for update
				mockService.On("UpdateContact", mock.Anything, mock.AnythingOfType("types.ContactUpdatePayload"), userID).
					Return(existingContact, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "name too long",
			contactID: contactID.String(),
			payload: fmt.Sprintf(`{
				"name": "%s"
			}`, strings.Repeat("a", types.MaxNameLength+1)),
			setupAuth: true,
			setupMock: func() {
				existingContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
				}
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(existingContact, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name: the length must be between 1 and 255.",
		},
		{
			name:      "invalid email format",
			contactID: contactID.String(),
			payload: `{
				"name": "John Doe",
				"email": "not-an-email"
			}`,
			setupAuth: true,
			setupMock: func() {
				existingContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
				}
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(existingContact, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email: must be a valid email address.",
		},
		{
			name:      "too many tags",
			contactID: contactID.String(),
			payload: fmt.Sprintf(`{
				"name": "John Doe",
				"tags": [%q, %q, %q, %q, %q, %q, %q, %q, %q, %q, %q]
			}`, uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(),
				uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()),
			setupAuth: true,
			setupMock: func() {
				existingContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
				}
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(existingContact, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "tags: the length must be no more than 10.",
		},
		{
			name:      "invalid contact ID",
			contactID: "not-a-uuid",
			payload: `{
				"name": "John Doe"
			}`,
			setupAuth:      true,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid UUID",
		},
		{
			name:      "contact not found",
			contactID: uuid.New().String(),
			payload: `{
				"name": "John Doe"
			}`,
			setupAuth: true,
			setupMock: func() {
				mockService.On("GetContact", mock.Anything, mock.AnythingOfType("uuid.UUID"), userID).
					Return(types.Contact{}, fmt.Errorf("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "service error",
			contactID: contactID.String(),
			payload: `{
				"name": "John Doe"
			}`,
			setupAuth: true,
			setupMock: func() {
				existingContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
				}
				mockService.On("GetContact", mock.Anything, contactID, userID).
					Return(existingContact, nil)
				mockService.On("UpdateContact", mock.Anything, mock.AnythingOfType("types.ContactUpdatePayload"), userID).
					Return(types.Contact{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database error",
		},
		{
			name:           "missing auth",
			contactID:      contactID.String(),
			payload:        `{"name": "John Doe"}`,
			setupAuth:      false,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPut, "/contacts/"+tt.contactID, strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.contactID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.UpdateContact(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, float64(http.StatusOK), response["status"])
				assert.NotNil(t, response["data"])
			} else {
				if tt.expectedError != "" {
					errMsg, ok := response["error"].(string)
					assert.True(t, ok)
					assert.Contains(t, errMsg, tt.expectedError)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
