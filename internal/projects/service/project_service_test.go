package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock repository
type mockProjectRepository struct {
	mock.Mock
}

func (m *mockProjectRepository) ListProjects(ctx context.Context, userID uuid.UUID) ([]types.Project, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]types.Project), args.Error(1)
}

func (m *mockProjectRepository) GetProject(ctx context.Context, userID, projectID uuid.UUID) (types.Project, error) {
	args := m.Called(ctx, userID, projectID)
	return args.Get(0).(types.Project), args.Error(1)
}

func (m *mockProjectRepository) CreateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectCreatePayload) (types.Project, error) {
	args := m.Called(ctx, userID, projectData)
	return args.Get(0).(types.Project), args.Error(1)
}

func (m *mockProjectRepository) UpdateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectUpdatePayload) (types.Project, error) {
	args := m.Called(ctx, userID, projectData)
	return args.Get(0).(types.Project), args.Error(1)
}

func (m *mockProjectRepository) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	args := m.Called(ctx, userID, projectID)
	return args.Error(0)
}

func (m *mockProjectRepository) GetProjectWallets(ctx context.Context, userID, projectID uuid.UUID) ([]db.Wallet, error) {
	args := m.Called(ctx, userID, projectID)
	return args.Get(0).([]db.Wallet), args.Error(1)
}

func (m *mockProjectRepository) ListProjectsPaginated(ctx context.Context, userID uuid.UUID, cursor time.Time, cursorID uuid.UUID, limit int32) ([]types.Project, error) {
	args := m.Called(ctx, userID, cursor, cursorID, limit)
	return args.Get(0).([]types.Project), args.Error(1)
}

func (m *mockProjectRepository) SearchProjects(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]types.Project, error) {
	args := m.Called(ctx, userID, query, limit)
	return args.Get(0).([]types.Project), args.Error(1)
}

func setupTest(t *testing.T) (*mockProjectRepository, ProjectService) {
	mockRepo := new(mockProjectRepository)
	logger := zap.NewNop()
	service := NewProjectService(mockRepo, logger)
	return mockRepo, service
}

func TestProjectService_CreateProject(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		payload types.ProjectCreatePayload
		mock    func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful create",
			payload: types.ProjectCreatePayload{
				Name:   "New Project",
				Status: "ongoing",
			},
			mock: func() {
				mockRepo.On("CreateProject", ctx, userID, mock.AnythingOfType("types.ProjectCreatePayload")).
					Return(types.Project{Name: "New Project"}, nil)
			},
			wantErr: false,
		},
		{
			name: "empty name",
			payload: types.ProjectCreatePayload{
				Name:   "",
				Status: "ongoing",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "project name is required",
		},
		{
			name: "invalid status",
			payload: types.ProjectCreatePayload{
				Name:   "Test Project",
				Status: "invalid_status",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "invalid project status",
		},
		{
			name: "invalid date combination",
			payload: types.ProjectCreatePayload{
				Name:      "Test Project",
				Status:    "ongoing",
				StartDate: utils.TimePtr(time.Now()),
				EndDate:   utils.TimePtr(time.Now().Add(-24 * time.Hour)),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "end date cannot be before start date",
		},
		{
			name: "negative budget",
			payload: types.ProjectCreatePayload{
				Name:   "Test Project",
				Status: "ongoing",
				Budget: utils.Float64Ptr(-1000.0),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "budget cannot be negative",
		},
		{
			name: "name too long",
			payload: types.ProjectCreatePayload{
				Name:   strings.Repeat("a", 256),
				Status: "ongoing",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "name exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			project, err := service.CreateProject(ctx, userID, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, project)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_GetProject(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful retrieval",
			mock: func() {
				expectedProject := types.Project{
					ProjectID: projectID,
					Name:      "Test Project",
					Status:    "ongoing",
				}
				mockRepo.On("GetProject", ctx, userID, projectID).Return(expectedProject, nil)
			},
			wantErr: false,
		},
		{
			name: "not found error",
			mock: func() {
				mockRepo.On("GetProject", ctx, userID, projectID).Return(types.Project{}, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockRepo.ExpectedCalls = nil

			tt.mock()
			project, err := service.GetProject(ctx, userID, projectID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, projectID, project.ProjectID)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_ListProjects(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
		wantLen int
	}{
		{
			name: "successful list",
			mock: func() {
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
				mockRepo.On("ListProjects", ctx, userID).Return(projects, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "empty list",
			mock: func() {
				mockRepo.On("ListProjects", ctx, userID).Return([]types.Project{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "repository error",
			mock: func() {
				mockRepo.On("ListProjects", ctx, userID).Return([]types.Project{}, errors.New("database error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous expectations
			mockRepo.ExpectedCalls = nil

			tt.mock()
			projects, err := service.ListProjects(ctx, userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, projects, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_UpdateProject(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()

	tests := []struct {
		name    string
		payload types.ProjectUpdatePayload
		mock    func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful update",
			payload: types.ProjectUpdatePayload{
				ProjectID: projectID,
				Name:      "Updated Project",
				Status:    "completed",
			},
			mock: func() {
				mockRepo.On("UpdateProject", ctx, userID, mock.AnythingOfType("types.ProjectUpdatePayload")).
					Return(types.Project{ProjectID: projectID, Name: "Updated Project"}, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			payload: types.ProjectUpdatePayload{
				ProjectID: projectID,
				Name:      "Test Project",
				Status:    "invalid_status",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "invalid project status",
		},
		{
			name: "invalid date combination",
			payload: types.ProjectUpdatePayload{
				ProjectID: projectID,
				Name:      "Test Project",
				Status:    "ongoing",
				StartDate: utils.TimePtr(time.Now()),
				EndDate:   utils.TimePtr(time.Now().Add(-24 * time.Hour)),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "end date cannot be before start date",
		},
		{
			name: "negative budget",
			payload: types.ProjectUpdatePayload{
				ProjectID: projectID,
				Name:      "Test Project",
				Status:    "ongoing",
				Budget:    utils.Float64Ptr(-1000.0),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "budget cannot be negative",
		},
		{
			name: "name too long",
			payload: types.ProjectUpdatePayload{
				ProjectID: projectID,
				Name:      strings.Repeat("a", 256), // Exceeds 255 characters
				Status:    "ongoing",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "name exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			project, err := service.UpdateProject(ctx, userID, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, project)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_ListProjectsPaginated(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()
	cursorID := uuid.New()

	tests := []struct {
		name     string
		cursor   time.Time
		cursorID uuid.UUID
		limit    int32
		mock     func()
		wantErr  bool
		wantLen  int
		errMsg   string
	}{
		{
			name:     "successful pagination",
			cursor:   now,
			cursorID: cursorID,
			limit:    10,
			mock: func() {
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
				mockRepo.On("ListProjectsPaginated", ctx, userID, now, cursorID, int32(10)).
					Return(projects, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:     "invalid limit",
			cursor:   now,
			cursorID: cursorID,
			limit:    -1,
			mock:     func() {},
			wantErr:  true,
			errMsg:   "limit must be positive",
		},
		{
			name:     "empty result",
			cursor:   now,
			cursorID: cursorID,
			limit:    10,
			mock: func() {
				mockRepo.On("ListProjectsPaginated", ctx, userID, now, cursorID, int32(10)).
					Return([]types.Project{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:     "repository error",
			cursor:   now,
			cursorID: cursorID,
			limit:    10,
			mock: func() {
				mockRepo.On("ListProjectsPaginated", ctx, userID, now, cursorID, int32(10)).
					Return([]types.Project{}, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			projects, err := service.ListProjectsPaginated(ctx, userID, tt.cursor, tt.cursorID, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, projects, tt.wantLen)
			mockRepo.AssertExpectations(t)

			// Verify ordering for non-empty results
			if len(projects) > 1 {
				for i := 1; i < len(projects); i++ {
					// Check that results are ordered by created_at DESC
					assert.True(t, projects[i-1].CreatedAt.After(projects[i].CreatedAt) ||
						(projects[i-1].CreatedAt.Equal(projects[i].CreatedAt) &&
							projects[i-1].ProjectID.String() > projects[i].ProjectID.String()))
				}
			}
		})
	}
}
