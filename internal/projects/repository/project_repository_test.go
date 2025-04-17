package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ProjectRepositoryTestSuite defines the test suite
type ProjectRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	pool      *pgxpool.Pool
	queries   *db.Queries
	repo      repository.ProjectRepository
	ctx       context.Context
	testUser  uuid.UUID
}

// TestProjectRepository is the single entry point for the test suite
func TestProjectRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(ProjectRepositoryTestSuite))
}

func (s *ProjectRepositoryTestSuite) SetupSuite() {
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
	s.repo = repository.NewProjectRepository(s.queries)

	// Create test user
	fmt.Println("Creating test user...")
	s.testUser = uuid.New()
	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO users (user_id, clerk_ex_user_id, name, email)
		VALUES ($1, 'prt_test_clerk_id', 'prt_Test User', 'prt_test@example.com')
	`, s.testUser)
	s.Require().NoError(err)
	fmt.Println("Test suite setup completed successfully")
}

func (s *ProjectRepositoryTestSuite) runMigrations() error {
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

func (s *ProjectRepositoryTestSuite) TearDownSuite() {
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

func (s *ProjectRepositoryTestSuite) SetupTest() {
	// Clean up projects table before each test
	s.clearProjects()
}

func (s *ProjectRepositoryTestSuite) TearDownTest() {
	// Clean up projects table after each test
	s.clearProjects()
}
func (s *ProjectRepositoryTestSuite) clearProjects() {
	_, err := s.pool.Exec(s.ctx, `DELETE FROM projects WHERE user_id = $1`, s.testUser)
	require.NoError(s.T(), err)
}
func (s *ProjectRepositoryTestSuite) TestCreateProject() {
	now := time.Now().UTC()
	tests := []struct {
		name    string
		payload types.ProjectCreatePayload
		wantErr bool
	}{
		{
			name: "valid project",
			payload: types.ProjectCreatePayload{
				Name:   "Test Project",
				Status: "ongoing",
			},
			wantErr: false,
		},
		{
			name: "project with all fields",
			payload: types.ProjectCreatePayload{
				Name:          "Full Project",
				Description:   utils.StringPtr("Test Description"),
				Status:        "ongoing",
				StartDate:     utils.TimePtr(now),
				EndDate:       utils.TimePtr(now.Add(24 * time.Hour)),
				Budget:        utils.Float64Ptr(1000.50),
				Website:       utils.StringPtr("https://test.com"),
				Country:       utils.StringPtr("US"),
				City:          utils.StringPtr("New York"),
				AddressLine1:  utils.StringPtr("123 Main St"),
				AddressLine2:  utils.StringPtr("Apt 4B"),
				StateProvince: utils.StringPtr("NY"),
				ZipPostalCode: utils.StringPtr("10001"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			project, err := s.repo.CreateProject(s.ctx, s.testUser, tt.payload)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.NotEmpty(project.ProjectID)
			s.Equal(tt.payload.Name, project.Name)
			s.Equal(tt.payload.Status, project.Status)

			// Check optional fields only if they are provided in the payload
			if tt.payload.Description != nil {
				s.NotNil(project.Description)
				s.Equal(*tt.payload.Description, *project.Description)
			}
			if tt.payload.StartDate != nil {
				s.NotNil(project.StartDate)
				s.WithinDuration(*tt.payload.StartDate, *project.StartDate, time.Second,
					"StartDate not within expected duration: got %v, want %v",
					project.StartDate, tt.payload.StartDate)
			}
			if tt.payload.EndDate != nil {
				s.NotNil(project.EndDate)
				s.WithinDuration(*tt.payload.EndDate, *project.EndDate, time.Second,
					"EndDate not within expected duration: got %v, want %v",
					project.EndDate, tt.payload.EndDate)
			}
			if tt.payload.Budget != nil {
				s.NotNil(project.Budget)
				s.Equal(*tt.payload.Budget, *project.Budget)
			}
			if tt.payload.Website != nil {
				s.NotNil(project.Website)
				s.Equal(*tt.payload.Website, *project.Website)
			}
			if tt.payload.Country != nil {
				s.NotNil(project.Country)
				s.Equal(*tt.payload.Country, *project.Country)
			}
			if tt.payload.City != nil {
				s.NotNil(project.City)
				s.Equal(*tt.payload.City, *project.City)
			}
			if tt.payload.AddressLine1 != nil {
				s.NotNil(project.AddressLine1)
				s.Equal(*tt.payload.AddressLine1, *project.AddressLine1)
			}
			if tt.payload.AddressLine2 != nil {
				s.NotNil(project.AddressLine2)
				s.Equal(*tt.payload.AddressLine2, *project.AddressLine2)
			}
			if tt.payload.StateProvince != nil {
				s.NotNil(project.StateProvince)
				s.Equal(*tt.payload.StateProvince, *project.StateProvince)
			}
			if tt.payload.ZipPostalCode != nil {
				s.NotNil(project.ZipPostalCode)
				s.Equal(*tt.payload.ZipPostalCode, *project.ZipPostalCode)
			}

			s.NotZero(project.CreatedAt)
			s.NotZero(project.UpdatedAt)
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestGetProject() {
	// Create a test project first
	createPayload := types.ProjectCreatePayload{
		Name:        "Test Project",
		Description: stringPtr("Test Description"),
		Status:      "ongoing",
	}
	created, err := s.repo.CreateProject(s.ctx, s.testUser, createPayload)
	require.NoError(s.T(), err)

	tests := []struct {
		name      string
		userID    uuid.UUID
		projectID uuid.UUID
		wantErr   bool
	}{
		{
			name:      "existing project",
			userID:    s.testUser,
			projectID: created.ProjectID,
			wantErr:   false,
		},
		{
			name:      "non-existent project",
			userID:    s.testUser,
			projectID: uuid.New(),
			wantErr:   true,
		},
		{
			name:      "wrong user",
			userID:    uuid.New(),
			projectID: created.ProjectID,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			project, err := s.repo.GetProject(s.ctx, tt.userID, tt.projectID)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Equal(created.ProjectID, project.ProjectID)
			s.Equal(created.Name, project.Name)
			s.Equal(*created.Description, *project.Description)
			s.Equal(created.Status, project.Status)
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestUpdateProject() {
	// Helper function to create a fresh project for each test case
	createInitialProject := func() types.Project {
		now := time.Now().UTC()
		createPayload := types.ProjectCreatePayload{
			Name:          "Test Project",
			Description:   stringPtr("Initial description"),
			Status:        "ongoing",
			StartDate:     &now,
			EndDate:       timePtr(now.Add(24 * time.Hour)),
			Budget:        float64Ptr(1000.50),
			AddressLine1:  stringPtr("123 Main St"),
			AddressLine2:  stringPtr("Suite 100"),
			Country:       stringPtr("US"),
			City:          stringPtr("Test City"),
			StateProvince: stringPtr("CA"),
			ZipPostalCode: stringPtr("12345"),
			Website:       stringPtr("https://example.com"),
			Tags:          []uuid.UUID{uuid.New(), uuid.New()},
		}

		project, err := s.repo.CreateProject(s.ctx, s.testUser, createPayload)
		s.Require().NoError(err)
		s.Require().NotEmpty(project)
		return project
	}

	testCases := []struct {
		name    string
		setup   func() types.Project
		payload func(types.Project) types.ProjectUpdatePayload
		check   func(*testing.T, types.Project)
		wantErr bool
	}{
		{
			name:  "full_update_preserving_values",
			setup: createInitialProject,
			payload: func(p types.Project) types.ProjectUpdatePayload {
				return types.ProjectUpdatePayload{
					ProjectID:     p.ProjectID,
					Name:          "Updated Name",
					Description:   p.Description,
					Status:        "completed",
					StartDate:     p.StartDate,
					EndDate:       p.EndDate,
					Budget:        p.Budget,
					Website:       p.Website,
					AddressLine1:  p.AddressLine1,
					AddressLine2:  p.AddressLine2,
					Country:       p.Country,
					City:          p.City,
					StateProvince: p.StateProvince,
					ZipPostalCode: p.ZipPostalCode,
					Tags:          p.Tags,
				}
			},
			check: func(t *testing.T, p types.Project) {
				assert.Equal(t, "Updated Name", p.Name)
				assert.Equal(t, "Initial description", *p.Description)
				assert.Equal(t, "completed", p.Status)
				assert.NotNil(t, p.StartDate)
				assert.NotNil(t, p.Budget)
				assert.Equal(t, "123 Main St", *p.AddressLine1)
			},
		},
		{
			name:  "update_with_explicit_null_values",
			setup: createInitialProject,
			payload: func(p types.Project) types.ProjectUpdatePayload {
				emptyStr := ""
				return types.ProjectUpdatePayload{
					ProjectID:   p.ProjectID,
					Name:        "Nullified Fields",
					Status:      "ongoing",
					Description: &emptyStr,
					StartDate:   nil,
					EndDate:     nil,
					Budget:      nil,
					Website:     nil,
				}
			},
			check: func(t *testing.T, p types.Project) {
				assert.Equal(t, "Nullified Fields", p.Name)
				assert.Empty(t, p.Description)
				assert.Equal(t, "ongoing", p.Status)
				assert.Nil(t, p.StartDate)
				assert.Nil(t, p.EndDate)
				assert.Nil(t, p.Budget)
				assert.Nil(t, p.Website)
				// Address fields should be nullified as they weren't included
				assert.Nil(t, p.AddressLine1)
				assert.Nil(t, p.AddressLine2)
			},
		},
		{
			name:  "invalid_project_id",
			setup: createInitialProject,
			payload: func(p types.Project) types.ProjectUpdatePayload {
				return types.ProjectUpdatePayload{
					ProjectID: uuid.New(),
					Name:      "Should Fail",
					Status:    "ongoing",
				}
			},
			wantErr: true,
		},
		{
			name:  "update_with_wrong_user_id",
			setup: createInitialProject,
			payload: func(p types.Project) types.ProjectUpdatePayload {
				return types.ProjectUpdatePayload{
					ProjectID: p.ProjectID,
					Name:      "Valid Name",
					Status:    "ongoing",
				}
			},
			check: func(t *testing.T, p types.Project) {
				// This won't be called as we expect an error
			},
			wantErr: true,
			// Note: We'll need to modify the test execution to use a different userID
		},
		{
			name:  "update_tags_only",
			setup: createInitialProject,
			payload: func(p types.Project) types.ProjectUpdatePayload {
				newTags := []uuid.UUID{uuid.New(), uuid.New()}
				return types.ProjectUpdatePayload{
					ProjectID: p.ProjectID,
					Name:      p.Name,
					Status:    p.Status,
					Tags:      newTags,
				}
			},
			check: func(t *testing.T, p types.Project) {
				assert.Len(t, p.Tags, 2)
				// Other fields should be nullified as they weren't included
				assert.Nil(t, p.Description)
				assert.Nil(t, p.StartDate)
				// ... etc
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Clean up before each subtest
			s.clearProjects()

			// Create a fresh project for each test case
			project := tt.setup()

			// Get the payload using the fresh project
			payload := tt.payload(project)

			// For the wrong user ID test case
			if tt.name == "update_with_wrong_user_id" {
				wrongUserID := uuid.New()
				updated, err := s.repo.UpdateProject(s.ctx, wrongUserID, payload)
				s.Error(err)
				s.Empty(updated)
				return
			}

			updated, err := s.repo.UpdateProject(s.ctx, s.testUser, payload)
			if tt.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)
			s.NotEmpty(updated)
			tt.check(s.T(), updated)

			// Verify persistence by fetching the project again
			fetched, err := s.repo.GetProject(s.ctx, s.testUser, payload.ProjectID)
			s.NoError(err)
			s.Equal(updated, fetched)

			// Clean up after each subtest
			s.clearProjects()
		})
	}
}

// Add these test functions after TestUpdateProject

func (s *ProjectRepositoryTestSuite) TestListProjectsPaginated() {
	// Create test projects in order from oldest to newest
	projects := []types.ProjectCreatePayload{
		{Name: "Project 1", Status: "ongoing"}, // Oldest
		{Name: "Project 2", Status: "completed"},
		{Name: "Project 3", Status: "ongoing"},
		{Name: "Project 4", Status: "ongoing"}, // Newest
	}

	var createdProjects []types.Project
	for _, p := range projects {
		time.Sleep(time.Millisecond * 100) // Ensure distinct timestamps
		project, err := s.repo.CreateProject(s.ctx, s.testUser, p)
		s.Require().NoError(err)
		createdProjects = append(createdProjects, project)
	}

	// Now createdProjects[3] is Project 4 (newest)
	// and createdProjects[0] is Project 1 (oldest)

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
			cursor:    time.Now().UTC(), // Future time to get newest first
			cursorID:  uuid.New(),       // New UUID to ensure it's greater than all projects
			limit:     2,
			wantLen:   2,
			wantNames: []string{"Project 4", "Project 3"}, // Newest first
			wantErr:   false,
		},
		{
			name:      "get second page",
			cursor:    createdProjects[2].CreatedAt, // Use Project 3's timestamp
			cursorID:  createdProjects[2].ProjectID, // Use Project 3's ID
			limit:     2,
			wantLen:   2,
			wantNames: []string{"Project 2", "Project 1"}, // Next oldest pair
			wantErr:   false,
		},
		{
			name:      "get empty page",
			cursor:    createdProjects[0].CreatedAt, // Use oldest project's timestamp
			cursorID:  createdProjects[0].ProjectID, // Use oldest project's ID
			limit:     2,
			wantLen:   0,
			wantNames: []string{},
			wantErr:   false,
		},
		{
			name:     "invalid limit",
			cursor:   time.Now().UTC(),
			cursorID: uuid.New(),
			limit:    -1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			projects, err := s.repo.ListProjectsPaginated(s.ctx, s.testUser, tt.cursor, tt.cursorID, tt.limit)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(projects, tt.wantLen)

			if len(tt.wantNames) > 0 {
				actualNames := make([]string, len(projects))
				for i, p := range projects {
					actualNames[i] = p.Name
				}
				s.Equal(tt.wantNames, actualNames)
			}

			// Verify ordering for non-empty results
			if len(projects) > 1 {
				for i := 1; i < len(projects); i++ {
					isCorrectOrder := projects[i-1].CreatedAt.After(projects[i].CreatedAt) ||
						(projects[i-1].CreatedAt.Equal(projects[i].CreatedAt) &&
							projects[i-1].ProjectID.String() > projects[i].ProjectID.String())
					s.True(isCorrectOrder, "Projects should be ordered by created_at DESC and then by project_id DESC")
				}
			}
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestSearchProjects() {
	// Create test projects with various names to test different search scenarios
	projects := []types.ProjectCreatePayload{
		{Name: "Project Alpha", Status: "ongoing"},             // Exact "Project" prefix
		{Name: "The Project Beta", Status: "ongoing"},          // "Project" as whole word
		{Name: "MyProject Delta", Status: "completed"},         // "Project" as part of word
		{Name: "Project Management System", Status: "ongoing"}, // Multiple words
		{Name: "Simple Proj", Status: "completed"},             // High similarity to "Project"
		{Name: "Task Projct", Status: "ongoing"},               // Misspelling
		{Name: "Project #123", Status: "completed"},            // With special characters
		{Name: "Alpha (Beta) Project", Status: "ongoing"},      // "Project" at end
		{Name: "Management System", Status: "completed"},       // For system search
		{Name: "Task System Pro", Status: "ongoing"},           // For system search
		{Name: "Project Management", Status: "completed"},      // For management search
		{Name: "Project Mnagement", Status: "ongoing"},         // Misspelling of Management
	}

	// Create projects in reverse order to control created_at timestamps
	for i := len(projects) - 1; i >= 0; i-- {
		_, err := s.repo.CreateProject(s.ctx, s.testUser, projects[i])
		s.Require().NoError(err)
		time.Sleep(time.Millisecond * 100) // Increased sleep duration to ensure distinct timestamps
	}

	tests := []struct {
		name      string
		query     string
		limit     int32
		wantLen   int
		wantNames []string // Expected project names in order
		wantErr   bool
	}{
		{
			name:    "management variations",
			query:   "Management",
			limit:   10,
			wantLen: 4,
			wantNames: []string{
				"Management System",         // Contains word
				"Project Management",        // Exact word match, shorter name
				"Project Management System", // Exact word match, longer name
				"Project Mnagement",         // High similarity
			},
			wantErr: false,
		},
		{
			name:    "with custom limit",
			query:   "Project",
			limit:   3,
			wantLen: 3,
			wantNames: []string{
				"Project #123",     // Short name with exact match
				"Project Alpha",    // Short name with exact match
				"The Project Beta", // Longer name with exact match
			},
			wantErr: false,
		},
		{
			name:    "exact and similar matches",
			query:   "Project",
			limit:   30,
			wantLen: 10,
			wantNames: []string{
				"Project #123",              // Short name with exact match shorter
				"Project Alpha",             // Short name with exact match
				"The Project Beta",          // Contains exact word
				"Project Mnagement",         // Exact match shorter
				"Project Management",        // Exact match
				"Alpha (Beta) Project",      // Contains exact word
				"Task Projct",               // High similarity
				"MyProject Delta",           // Part of word
				"Project Management System", // Exact match
				"Simple Proj",               // low similarity
			},
			wantErr: false,
		},
		{
			name:      "similarity matches",
			query:     "Projct",
			limit:     30,
			wantLen:   8,
			wantNames: []string{"Task Projct", "Project #123", "Project Alpha", "Simple Proj", "The Project Beta", "Project Mnagement", "Project Management", "Alpha (Beta) Project"},
			wantErr:   false,
		},
		{
			name:    "whole word matches",
			query:   "system",
			limit:   30,
			wantLen: 3,
			wantNames: []string{
				"Task System Pro",           // Short name with word
				"Management System",         // Contains word
				"Project Management System", // Contains word
			},
			wantErr: false,
		},
		{
			name:    "empty query returns all by created_at",
			query:   "",
			limit:   20,
			wantLen: len(projects),
			wantNames: []string{
				"Project Alpha",
				"The Project Beta",
				"MyProject Delta",
				"Project Management System",
				"Simple Proj",
				"Task Projct",
				"Project #123",
				"Alpha (Beta) Project",
				"Management System",
				"Task System Pro",
				"Project Management",
				"Project Mnagement",
			},
			wantErr: false,
		},
		{
			name:      "special characters",
			query:     "#",
			limit:     10,
			wantLen:   1,
			wantNames: []string{"Project #123"},
			wantErr:   false,
		},
		{
			name:      "parentheses",
			query:     "(",
			limit:     10,
			wantLen:   1,
			wantNames: []string{"Alpha (Beta) Project"},
			wantErr:   false,
		},
		{
			name:    "invalid limit",
			query:   "test",
			limit:   -1,
			wantErr: true,
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
			projects, err := s.repo.SearchProjects(s.ctx, s.testUser, tt.query, tt.limit)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Len(projects, tt.wantLen)

			if tt.name == "empty query returns all" {
				s.T().Log("\nProject ordering details:")
				for i, p := range projects {
					s.T().Logf("%d. Name: %-30s Created At: %v", i+1, p.Name, p.CreatedAt.Format(time.RFC3339Nano))
				}
			}

			if len(tt.wantNames) > 0 {
				actualNames := make([]string, len(projects))
				for i, p := range projects {
					actualNames[i] = p.Name
				}
				s.Equal(tt.wantNames, actualNames, "Project names should match in the exact order based on ranking:\n"+
					"1.0: Exact matches\n"+
					"0.9: Prefix matches\n"+
					"0.8: Whole word matches\n"+
					"0.7+: High similarity matches\n"+
					"0.3: Contains matches\n"+
					"<0.3: Low similarity")
			}
		})
	}
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
