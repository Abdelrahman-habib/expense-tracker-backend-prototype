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
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/handlers"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/types"
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

type ProjectIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	service   db.Service
	pool      *pgxpool.Pool
	handler   *handlers.ProjectHandler
	router    *chi.Mux
	userID    uuid.UUID
	ctx       context.Context
}

func TestProjectIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ProjectIntegrationTestSuite))
}

func (s *ProjectIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.userID = uuid.New()

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
		}

		// Start container with retry logic
		var container testcontainers.Container
		var err error
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			container, err = testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if err == nil {
				break
			}
			fmt.Printf("Attempt %d: Failed to start container: %v\n", i+1, err)
			time.Sleep(time.Second * 2)
		}
		require.NoError(s.T(), err)
		s.container = container

		// Get container host and port
		host, err = container.Host(s.ctx)
		require.NoError(s.T(), err)
		mappedPort, err := container.MappedPort(s.ctx, "5432")
		require.NoError(s.T(), err)
		port = mappedPort.Port()
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
	fmt.Println(port)
	// Get connection pool
	pool, err := pgxpool.New(s.ctx, cfg.GetDSN())
	require.NoError(s.T(), err)
	s.pool = pool

	// Run migrations
	err = s.runMigrations()
	require.NoError(s.T(), err)

	// clear any previous runs data
	s.clearProjects()

	// Create test user
	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, 'pit_clerk_id', 'pit_Test User', 'pit_test@example.com')
	`, s.userID)
	require.NoError(s.T(), err)

	// Initialize components
	logger := zap.NewNop()
	repo := repository.NewProjectRepository(dbService.Queries())
	projectService := service.NewProjectService(repo, logger)
	s.handler = handlers.NewProjectHandler(projectService, logger)

	// Setup router
	router := chi.NewRouter()
	router.Route("/projects", func(r chi.Router) {
		r.Get("/", s.handler.ListProjects)
		r.Get("/search", s.handler.SearchProjects)
		r.Get("/paginated", s.handler.ListProjectsPaginated)
		r.Post("/", s.handler.CreateProject)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handler.GetProject)
			r.Put("/", s.handler.UpdateProject)
			r.Delete("/", s.handler.DeleteProject)
		})
	})
	s.router = router
}

func (s *ProjectIntegrationTestSuite) TearDownSuite() {
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

func (s *ProjectIntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	s.clearProjects()
}

func (s *ProjectIntegrationTestSuite) runMigrations() error {
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

// Helper function for creating test projects
func (s *ProjectIntegrationTestSuite) createTestProject() types.Project {
	createPayload := types.ProjectCreatePayload{
		Name:        "Integration Test Project",
		Description: stringPtr("Test Description"),
		Status:      "ongoing",
		StartDate:   timePtr(time.Now()),
		Budget:      float64Ptr(1000.50),
	}

	payloadBytes, err := json.Marshal(createPayload)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)

	projectData := response["data"].(map[string]interface{})
	return types.Project{
		ProjectID: uuid.MustParse(projectData["projectId"].(string)),
		Name:      projectData["name"].(string),
		Status:    projectData["status"].(string),
	}
}

func (s *ProjectIntegrationTestSuite) TestProjectLifecycle() {
	// Create a project and use it across all tests
	project := &types.Project{}
	*project = s.createTestProject()

	s.Run("get project", func() {
		s.testGetProject(project)
	})
	s.Run("update project name", func() {
		s.testUpdateProjectName(project)
	})
	s.Run("update project status", func() {
		s.testUpdateProjectStatus(project)
	})
	s.Run("delete project", func() {
		s.testDeleteProject(project)
	})
}

func (s *ProjectIntegrationTestSuite) testGetProject(project *types.Project) {
	req := s.newAuthenticatedRequest(http.MethodGet, "/projects/"+project.ProjectID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", project.ProjectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)

	getData := response["data"].(map[string]interface{})
	s.Equal(project.Name, getData["name"])
	s.Equal(project.Status, getData["status"])
}

func (s *ProjectIntegrationTestSuite) testUpdateProjectName(project *types.Project) {
	updatePayload := types.ProjectUpdatePayload{
		ProjectID: project.ProjectID,
		Name:      "Updated Project Name",
		Status:    project.Status,
	}

	payloadBytes, err := json.Marshal(updatePayload)
	s.Require().NoError(err)

	req := s.newAuthenticatedRequest(http.MethodPut, "/projects/"+project.ProjectID.String(), bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", project.ProjectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	project.Name = updatePayload.Name
	s.verifyProjectState(project.ProjectID, project.Name, project.Status)
}

func (s *ProjectIntegrationTestSuite) testUpdateProjectStatus(project *types.Project) {
	updatePayload := types.ProjectUpdatePayload{
		ProjectID: project.ProjectID,
		Name:      project.Name,
		Status:    "completed",
	}

	payloadBytes, err := json.Marshal(updatePayload)
	s.Require().NoError(err)

	req := s.newAuthenticatedRequest(http.MethodPut, "/projects/"+project.ProjectID.String(), bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", project.ProjectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	project.Status = updatePayload.Status
	s.verifyProjectState(project.ProjectID, project.Name, project.Status)
}

func (s *ProjectIntegrationTestSuite) testDeleteProject(project *types.Project) {
	req := s.newAuthenticatedRequest(http.MethodDelete, "/projects/"+project.ProjectID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", project.ProjectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Verify project is deleted
	req = s.newAuthenticatedRequest(http.MethodGet, "/projects/"+project.ProjectID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusNotFound, w.Code)
}

// Helper methods to reduce duplication
func (s *ProjectIntegrationTestSuite) newAuthenticatedRequest(method, path string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, path, body)
	return req.WithContext(context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID))
}

func (s *ProjectIntegrationTestSuite) verifyProjectState(projectID uuid.UUID, expectedName, expectedStatus string) {
	req := s.newAuthenticatedRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", projectID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	s.Require().NoError(err)
	getData := response["data"].(map[string]interface{})
	s.Equal(expectedName, getData["name"])
	s.Equal(expectedStatus, getData["status"])
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func float64Ptr(f float64) *float64 {
	return &f
}

func (s *ProjectIntegrationTestSuite) clearProjects() {
	_, err := s.pool.Exec(s.ctx, `DELETE FROM projects WHERE user_id = $1`, s.userID)
	require.NoError(s.T(), err)
}

func (s *ProjectIntegrationTestSuite) createTestProjects(count int) []types.Project {
	projects := make([]types.Project, count)

	for i := 0; i < count; i++ {
		createPayload := types.ProjectCreatePayload{
			Name:   fmt.Sprintf("Test Project %d", i+1),
			Status: "ongoing",
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		s.Require().Equal(http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		s.Require().NoError(err)

		projectData := response["data"].(map[string]interface{})

		// Parse time in UTC
		createdAtStr := projectData["createdAt"].(string)
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
		s.Require().NoError(err)

		// Ensure it's in UTC
		createdAt = createdAt.UTC()

		projects[count-1-i] = types.Project{
			ProjectID: uuid.MustParse(projectData["projectId"].(string)),
			Name:      projectData["name"].(string),
			Status:    projectData["status"].(string),
			CreatedAt: createdAt,
		}
		time.Sleep(time.Millisecond * 10)
	}
	return projects
}

func (s *ProjectIntegrationTestSuite) TestListProjectsPaginated() {
	// Clear projects table
	s.clearProjects()

	// Create 10 test projects
	projects := s.createTestProjects(10) // projects[0] = 'test project 10'

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
			name: "default pagination", // First page (no cursor): Gets newest records (10,9,8,7,6)
			queryParams: map[string]string{
				"limit": "5",
			},
			expectedStatus:  http.StatusOK,
			expectedLen:     5,
			expectedLimit:   "5",
			expectNextToken: true,
		},
		{
			name: "with next_token", // Using Project 6's cursor: Gets next newer records (5,4,3)
			queryParams: map[string]string{
				"limit":      "3",
				"next_token": coreTypes.EncodeCursor(projects[4].CreatedAt, projects[4].ProjectID), // Project 6
			},
			expectedStatus:  http.StatusOK,
			expectedLen:     3,
			expectedLimit:   "3",
			expectNextToken: true,
		},
		{
			name: "last page", // Using Project 3's cursor: Gets final records (2,1)
			queryParams: map[string]string{
				"limit":      "5",
				"next_token": coreTypes.EncodeCursor(projects[7].CreatedAt, projects[7].ProjectID), // Project 3
			},
			expectedStatus:  http.StatusOK,
			expectedLen:     2,
			expectedLimit:   "5",
			expectNextToken: false,
		},
		{
			name: "invalid next_token",
			queryParams: map[string]string{
				"limit":      "5",
				"next_token": "invalid_token",
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
			urlPath := "/projects/paginated"
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

				projects := response["data"].([]interface{})
				s.Len(projects, tt.expectedLen)

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

func (s *ProjectIntegrationTestSuite) TestSearchProjects() {
	// Create test projects with more distinct names
	projects := []types.ProjectCreatePayload{
		{Name: "Project Alpha", Status: "ongoing"},
		{Name: "Beta System", Status: "ongoing"},
		{Name: "Gamma Project", Status: "completed"},
		{Name: "Delta Management", Status: "ongoing"},
		{Name: "Project Management System", Status: "completed"},
		{Name: "Project Mnagement", Status: "ongoing"},      // Misspelling of "Management"
		{Name: "Projct Management", Status: "completed"},    // Missing 'e'
		{Name: "Task Tracking System", Status: "ongoing"},   // Different type of system
		{Name: "Customer Portal", Status: "completed"},      // Completely different
		{Name: "Resource Planning", Status: "ongoing"},      // Completely different
		{Name: "Project #123", Status: "ongoing"},           // With special characters
		{Name: "Alpha (Beta) Project", Status: "completed"}, // With parentheses
	}

	// Create all test projects
	for _, p := range projects {
		payloadBytes, err := json.Marshal(p)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(payloadBytes))
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
		expectedNames  []string
		expectedError  string
	}{
		{
			name:           "case insensitive search",
			query:          "project",
			expectedStatus: http.StatusOK,
			expectedCount:  7,
			expectedNames:  []string{"Project #123", "Gamma Project", "Project Alpha", "Project Mnagement", "Alpha (Beta) Project", "Project Management System", "Projct Management"},
		},
		{
			name:           "similarity search - management misspelling",
			query:          "Management",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedNames:  []string{"Delta Management", "Projct Management", "Project Management System", "Project Mnagement"},
		},
		{
			name:           "similarity search - missing letter",
			query:          "Projct",
			expectedStatus: http.StatusOK,
			expectedCount:  6,
			expectedNames:  []string{"Projct Management", "Project #123", "Gamma Project", "Project Alpha", "Project Mnagement", "Alpha (Beta) Project"},
		},
		{
			name:           "system search",
			query:          "System",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			expectedNames:  []string{"Beta System", "Task Tracking System", "Project Management System"},
		},
		{
			name:           "with custom limit",
			query:          "System",
			limit:          "2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedNames:  []string{"Beta System", "Task Tracking System"},
		},
		{
			name:           "empty query",
			query:          "",
			limit:          fmt.Sprint(len(projects)),
			expectedStatus: http.StatusOK,
			expectedCount:  len(projects), // Should return all projects ordered by created_at DESC
		},
		{
			name:           "query too long",
			query:          strings.Repeat("a", coreTypes.MaxQueryLength+1), // Exceeds maxQueryLength
			expectedStatus: http.StatusBadRequest,
			expectedError:  "query: the length must be between 1 and 100.",
		},
		{
			name:           "invalid limit",
			query:          "Project",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "limit: invalid format",
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
			expectedNames:  []string{"Project #123"},
		},
		{
			name:           "parentheses",
			query:          "(",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedNames:  []string{"Alpha (Beta) Project"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			urlPath := fmt.Sprintf("/projects/search?q=%s", url.QueryEscape(tt.query))
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

				projects := response["data"].([]interface{})
				s.Len(projects, tt.expectedCount)
				metadata := response["meta"].(map[string]interface{})
				if tt.query != "" {
					s.Equal(tt.query, metadata["query"])
				}
				if tt.limit != "" {
					limit, _ := strconv.ParseFloat(tt.limit, 64)
					s.Equal(limit, metadata["limit"])
				}

				// Verify project names if expected names are provided
				if len(tt.expectedNames) > 0 {
					actualNames := make([]string, len(projects))
					for i, p := range projects {
						project := p.(map[string]interface{})
						actualNames[i] = project["name"].(string)
					}
					s.Equal(tt.expectedNames, actualNames)
				}
			} else {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				s.Require().NoError(err)
				s.Contains(response["error"].(string), tt.expectedError)
			}
		})
	}
}

func (s *ProjectIntegrationTestSuite) TestConcurrentUpdates() {
	// Create a project
	project := s.createTestProject()

	// Try to update the same project concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			updatePayload := types.ProjectUpdatePayload{
				ProjectID: project.ProjectID,
				Name:      fmt.Sprintf("Updated Name %d", i),
				Status:    "ongoing",
			}

			payloadBytes, err := json.Marshal(updatePayload)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ProjectID.String(), bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", project.ProjectID.String())
			req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)

			// All updates should succeed
			s.Equal(http.StatusOK, w.Code)
		}(i)
	}
	wg.Wait()
}

func (s *ProjectIntegrationTestSuite) TestUnauthorizedAccess() {
	// Create a project first
	project := s.createTestProjects(1)[0]

	// Create another user
	otherUserID := uuid.New()
	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, 'pit_other_clerk_id', 'pit_Other User', 'pit_other@example.com')
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
				req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ProjectID.String(), nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", project.ProjectID.String())
				return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "access with wrong user",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ProjectID.String(), nil)
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, otherUserID)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", project.ProjectID.String())
				return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "update with wrong user",
			setupRequest: func() *http.Request {
				payload := types.ProjectUpdatePayload{
					ProjectID: project.ProjectID,
					Name:      "Unauthorized Update",
					Status:    "ongoing",
				}
				payloadBytes, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ProjectID.String(), bytes.NewReader(payloadBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, otherUserID)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", project.ProjectID.String())
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

func (s *ProjectIntegrationTestSuite) TestComplexProjectLifecycle() {
	// Test the complete lifecycle of a project with multiple operations
	s.Run("full project lifecycle", func() {
		// 1. Create project
		createPayload := types.ProjectCreatePayload{
			Name:      "Lifecycle Project",
			Status:    "ongoing",
			Budget:    float64Ptr(1000),
			StartDate: timePtr(time.Now()),
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		s.Require().NoError(err)
		projectData := response["data"].(map[string]interface{})
		projectID := projectData["projectId"].(string)

		// 2. Update project multiple times with different fields
		updates := []types.ProjectUpdatePayload{
			{
				ProjectID: uuid.MustParse(projectID),
				Name:      "Updated Name",
				Status:    "ongoing",
				Budget:    float64Ptr(2000),
			},
			{
				ProjectID: uuid.MustParse(projectID),
				Name:      "Updated Name",
				Status:    "completed",
				EndDate:   timePtr(time.Now().Add(24 * time.Hour)),
			},
		}

		for _, update := range updates {
			payloadBytes, err = json.Marshal(update)
			s.Require().NoError(err)

			req = httptest.NewRequest(http.MethodPut, "/projects/"+projectID, bytes.NewReader(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			ctx = context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
			req = req.WithContext(ctx)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", projectID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w = httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			s.Equal(http.StatusOK, w.Code)
		}

		// 3. Verify final state
		req = httptest.NewRequest(http.MethodGet, "/projects/"+projectID, nil)
		ctx = context.WithValue(req.Context(), requestcontext.UserIDKey, s.userID)
		req = req.WithContext(ctx)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", projectID)
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
		s.Equal("completed", finalData["status"])
		s.NotNil(finalData["endDate"])
	})
}

func (s *ProjectIntegrationTestSuite) TestPaginationEdgeCases() {
	// Test extreme pagination cases
	s.Run("pagination edge cases", func() {
		// Create 20 projects for testing
		_ = s.createTestProjects(20)

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
				url := fmt.Sprintf("/projects/paginated?limit=%d", tt.limit)
				req := httptest.NewRequest(http.MethodGet, url, nil)
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

func (s *ProjectIntegrationTestSuite) TestDatabaseConstraintsAndValidation() {
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
					"status": "ongoing",
					// name missing
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "name: cannot be blank.",
				errorMessage:  "Invalid request",
			},
			{
				name: "name too long",
				payload: map[string]interface{}{
					"name":   strings.Repeat("a", 256),
					"status": "ongoing",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "name: the length must be between 1 and 255.",
				errorMessage:  "Invalid request",
			},
			{
				name: "invalid UUID format",
				payload: map[string]interface{}{
					"name":   "Test Project",
					"status": "ongoing",
					"tags":   []string{"not-a-uuid"},
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "invalid UUID length",
				errorMessage:  "Invalid request",
			},
			{
				name: "too many tags",
				payload: map[string]interface{}{
					"name":   "Test Project",
					"status": "ongoing",
					"tags":   []string{uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()}, // Exceeds MaxTagsCount
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "tags: the length must be no more than 10",
				errorMessage:  "Invalid request",
			},
			{
				name: "malformed budget",
				payload: map[string]interface{}{
					"name":   "Test Project",
					"status": "ongoing",
					"budget": "not-a-number",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "budget",
				errorMessage:  "Invalid request",
			},
			{
				name: "malformed date",
				payload: map[string]interface{}{
					"name":      "Test Project",
					"status":    "ongoing",
					"startDate": "invalid-date",
				},
				expectedCode:  http.StatusBadRequest,
				errorContains: "invalid-date",
				errorMessage:  "Invalid request",
			},
		}

		for _, tt := range tests {
			s.Run(tt.name, func() {
				payloadBytes, err := json.Marshal(tt.payload)
				s.Require().NoError(err)

				req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(payloadBytes))
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

func (s *ProjectIntegrationTestSuite) TestResponsePayloadStructure() {
	s.Run("response payload structure", func() {
		// Create a project with all fields
		createPayload := types.ProjectCreatePayload{
			Name:        "Response Test Project",
			Description: stringPtr("Test Description"),
			Status:      "ongoing",
			StartDate:   timePtr(time.Now().UTC()),
			Budget:      float64Ptr(1000.50),
			Website:     stringPtr("https://example.com"),
			Tags:        []uuid.UUID{uuid.New(), uuid.New()},
		}

		payloadBytes, err := json.Marshal(createPayload)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(payloadBytes))
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
		s.Require().Contains(data, "projectId")
		s.IsType("", data["projectId"].(string))
		_, err = uuid.Parse(data["projectId"].(string))
		s.NoError(err)

		s.Equal(createPayload.Name, data["name"])
		s.Equal(*createPayload.Description, data["description"])
		s.Equal(createPayload.Status, data["status"])
		s.NotEmpty(data["createdAt"])
		s.NotEmpty(data["updatedAt"])

		// Verify optional fields
		s.Equal(*createPayload.Website, data["website"])
		s.Equal(*createPayload.Budget, data["budget"])

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
