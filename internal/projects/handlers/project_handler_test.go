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
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/types"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock service
type mockProjectService struct {
	mock.Mock
}

func (m *mockProjectService) ListProjects(ctx context.Context, userID uuid.UUID) ([]types.Project, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]types.Project), args.Error(1)
}

func (m *mockProjectService) GetProject(ctx context.Context, userID, projectID uuid.UUID) (types.Project, error) {
	args := m.Called(ctx, userID, projectID)
	return args.Get(0).(types.Project), args.Error(1)
}

func (m *mockProjectService) CreateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectCreatePayload) (types.Project, error) {
	args := m.Called(ctx, userID, projectData)
	return args.Get(0).(types.Project), args.Error(1)
}

func (m *mockProjectService) UpdateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectUpdatePayload) (types.Project, error) {
	args := m.Called(ctx, userID, projectData)
	return args.Get(0).(types.Project), args.Error(1)
}

func (m *mockProjectService) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	args := m.Called(ctx, userID, projectID)
	return args.Error(0)
}

func (m *mockProjectService) GetProjectWallets(ctx context.Context, userID, projectID uuid.UUID) ([]db.Wallet, error) {
	args := m.Called(ctx, userID, projectID)
	return args.Get(0).([]db.Wallet), args.Error(1)
}

func (m *mockProjectService) ListProjectsPaginated(ctx context.Context, userID uuid.UUID, cursor time.Time, cursorID uuid.UUID, limit int32) ([]types.Project, error) {
	args := m.Called(ctx, userID, cursor, cursorID, limit)
	return args.Get(0).([]types.Project), args.Error(1)
}

func (m *mockProjectService) SearchProjects(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]types.Project, error) {
	args := m.Called(ctx, userID, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Project), args.Error(1)
}

func setupTest(t *testing.T) (*mockProjectService, *ProjectHandler) {
	mockService := new(mockProjectService)
	logger := zap.NewNop()
	handler := NewProjectHandler(mockService, logger)
	return mockService, handler
}

func TestProjectHandler_CreateProject(t *testing.T) {
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
				"name": "Test Project",
				"status": "ongoing"
			}`,
			setupAuth: true,
			setupMock: func() {
				expectedProject := types.Project{
					ProjectID: uuid.New(),
					Name:      "Test Project",
					Status:    "ongoing",
				}
				mockService.On("CreateProject", mock.Anything, userID, mock.AnythingOfType("types.ProjectCreatePayload")).
					Return(expectedProject, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid payload",
			payload: `{
				"name": "",
				"status": "invalid"
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
			// Clear previous mock expectations
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.CreateProject(w, req)

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

func TestProjectHandler_GetProject(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		projectID      string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:      "successful retrieval",
			setupAuth: true,
			projectID: projectID.String(),
			setupMock: func() {
				expectedProject := types.Project{
					ProjectID: projectID,
					Name:      "Test Project",
					Status:    "ongoing",
				}
				mockService.On("GetProject", mock.Anything, userID, projectID).
					Return(expectedProject, nil)
			},
			expectedStatus: http.StatusOK,
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
			// Clear previous mock expectations
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodGet, "/projects/"+tt.projectID, nil)

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.projectID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.GetProject(w, req)

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

func TestProjectHandler_ListProjects(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		setupMock      func()
		expectedStatus int
		expectedLen    int
	}{
		{
			name:      "successful list",
			setupAuth: true,
			setupMock: func() {
				projects := []types.Project{
					{
						ProjectID: uuid.New(),
						Name:      "Project 1",
						Status:    "ongoing",
					},
					{
						ProjectID: uuid.New(),
						Name:      "Project 2",
						Status:    "completed",
					},
				}
				mockService.On("ListProjects", mock.Anything, userID).Return(projects, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:           "missing auth",
			setupAuth:      false,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous mock expectations
			mockService.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodGet, "/projects", nil)

			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, userID)
				req = req.WithContext(ctx)
			}

			tt.setupMock()
			w := httptest.NewRecorder()
			handler.ListProjects(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, float64(http.StatusOK), response["status"])
				data := response["data"].([]interface{})
				assert.Len(t, data, tt.expectedLen)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestProjectHandler_ListProjectsPaginated(t *testing.T) {
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
				projects := []types.Project{
					{
						ProjectID: uuid.New(),
						Name:      "Project 1",
						Status:    "ongoing",
						CreatedAt: now.Add(-1 * time.Hour),
					},
					{
						ProjectID: uuid.New(),
						Name:      "Project 2",
						Status:    "completed",
						CreatedAt: now.Add(-2 * time.Hour),
					},
				}
				mockService.On("ListProjectsPaginated",
					mock.Anything,
					userID,
					mock.MatchedBy(func(t time.Time) bool {
						return time.Since(t) < time.Minute
					}),
					mock.MatchedBy(func(id uuid.UUID) bool {
						return id == uuid.Nil
					}),
					int32(coreTypes.DefaultLimit),
				).Return(projects, nil)
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
				projects := []types.Project{
					{
						ProjectID: uuid.New(),
						Name:      "Project 1",
						Status:    "ongoing",
						CreatedAt: now.Add(-1 * time.Hour),
					},
				}
				mockService.On("ListProjectsPaginated",
					mock.Anything,
					userID,
					mock.MatchedBy(func(t time.Time) bool {
						return time.Since(t) < time.Minute
					}),
					mock.MatchedBy(func(id uuid.UUID) bool {
						return id == uuid.Nil
					}),
					int32(5),
				).Return(projects, nil)
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
				"limit":      "2",
			},
			setupMock: func() {
				projects := []types.Project{
					{
						ProjectID: uuid.New(),
						Name:      "Project 3",
						Status:    "ongoing",
						CreatedAt: now.Add(-3 * time.Hour),
					},
					{
						ProjectID: uuid.New(),
						Name:      "Project 4",
						Status:    "completed",
						CreatedAt: now.Add(-4 * time.Hour),
					},
				}
				mockService.On("ListProjectsPaginated",
					mock.Anything,
					userID,
					mock.MatchedBy(func(t time.Time) bool {
						return t.Equal(now)
					}),
					mock.MatchedBy(func(id uuid.UUID) bool {
						return id == cursorID
					}),
					int32(2),
				).Return(projects, nil)
			},
			expectedStatus:  http.StatusOK,
			expectedLen:     2,
			expectNextToken: true,
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
				mockService.On("ListProjectsPaginated",
					mock.Anything,
					userID,
					mock.Anything,
					mock.Anything,
					int32(10),
				).Return([]types.Project{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			reqURL := "/projects/paginated"
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
			handler.ListProjectsPaginated(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, float64(http.StatusOK), response["status"])
				assert.Equal(t, "Success", response["message"])

				projects := response["data"].([]interface{})
				assert.Len(t, projects, tt.expectedLen)

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

func TestProjectHandler_SearchProjects(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		setupAuth      bool
		queryParams    map[string]string
		setupMock      func()
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
		expectedError  string
	}{
		{
			name:      "successful search",
			setupAuth: true,
			queryParams: map[string]string{
				"q": "test",
			},
			setupMock: func() {
				projects := []types.Project{
					{
						ProjectID: uuid.New(),
						Name:      "Test Project",
						Status:    "ongoing",
					},
				}
				mockService.On("SearchProjects", mock.Anything, userID, "test", int32(coreTypes.DefaultSearchLimit)).
					Return(projects, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				data := response["data"].([]interface{})
				assert.Len(t, data, 1)
				meta := response["meta"].(map[string]interface{})
				assert.Equal(t, "test", meta["query"])
				assert.Equal(t, float64(coreTypes.DefaultSearchLimit), meta["limit"])
				assert.Equal(t, float64(1), meta["count"])
			},
		},
		{
			name:      "empty query parameter returns all projects",
			setupAuth: true,
			queryParams: map[string]string{
				"q": "",
			},
			setupMock: func() {
				projects := []types.Project{
					{
						ProjectID: uuid.New(),
						Name:      "Recent Project",
						CreatedAt: time.Now().Add(-1 * time.Hour),
					},
					{
						ProjectID: uuid.New(),
						Name:      "Older Project",
						CreatedAt: time.Now().Add(-2 * time.Hour),
					},
				}
				mockService.On("SearchProjects", mock.Anything, userID, "", int32(coreTypes.DefaultSearchLimit)).
					Return(projects, nil)
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
			name:      "query too long",
			setupAuth: true,
			queryParams: map[string]string{
				"q": strings.Repeat("a", coreTypes.MaxQueryLength+1),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "query: the length must be between 1 and 100.",
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
			expectedError:  "limit: must be no less than 1.",
		},
		{
			name:      "service error",
			setupAuth: true,
			queryParams: map[string]string{
				"q": "test",
			},
			setupMock: func() {
				mockService.On("SearchProjects", mock.Anything, userID, "test", int32(coreTypes.DefaultSearchLimit)).
					Return([]types.Project(nil), fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database error",
		},
		{
			name:           "missing auth",
			setupAuth:      false,
			queryParams:    map[string]string{"q": "test"},
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			reqURL := "/projects/search"
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
			handler.SearchProjects(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, float64(http.StatusOK), response["status"])
				if tt.checkResponse != nil {
					tt.checkResponse(t, response)
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
