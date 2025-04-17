package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProjectService interface {
	ListProjects(ctx context.Context, userID uuid.UUID) ([]types.Project, error)
	GetProject(ctx context.Context, userID, projectID uuid.UUID) (types.Project, error)
	CreateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectCreatePayload) (types.Project, error)
	UpdateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectUpdatePayload) (types.Project, error)
	DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error
	GetProjectWallets(ctx context.Context, userID, projectID uuid.UUID) ([]db.Wallet, error)
	ListProjectsPaginated(ctx context.Context, userID uuid.UUID, cursor time.Time, cursorID uuid.UUID, limit int32) ([]types.Project, error)
	SearchProjects(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]types.Project, error)
}

type projectService struct {
	repo   repository.ProjectRepository
	logger *zap.Logger
}

func NewProjectService(repo repository.ProjectRepository, logger *zap.Logger) ProjectService {
	return &projectService{
		repo:   repo,
		logger: logger.With(zap.String("component", "project_service")),
	}
}

func (s *projectService) ListProjects(ctx context.Context, userID uuid.UUID) ([]types.Project, error) {
	s.logger.Info("listing projects for user", zap.String("user_id", userID.String()))
	return s.repo.ListProjects(ctx, userID)
}

func (s *projectService) GetProject(ctx context.Context, userID, projectID uuid.UUID) (types.Project, error) {
	s.logger.Info("getting project",
		zap.String("user_id", userID.String()),
		zap.String("project_id", projectID.String()))
	return s.repo.GetProject(ctx, userID, projectID)
}

// Common validation function
func validateProject(name, status string, startDate, endDate *time.Time, budget *float64, description *string) error {
	// Validate required fields
	if name == "" {
		return fmt.Errorf("project name is required")
	}

	// Validate status
	if !isValidProjectStatus(status) {
		return fmt.Errorf("invalid project status: %s", status)
	}

	// Validate dates
	if startDate != nil && endDate != nil {
		if endDate.Before(*startDate) {
			return fmt.Errorf("end date cannot be before start date")
		}
	}

	// Validate budget
	if budget != nil && *budget < 0 {
		return fmt.Errorf("budget cannot be negative")
	}

	// Validate text field lengths
	if len(name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}
	if description != nil && len(*description) > 1000 {
		return fmt.Errorf("description exceeds maximum length of 1000 characters")
	}

	return nil
}

func (s *projectService) CreateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectCreatePayload) (types.Project, error) {
	// Validate project data
	if err := validateProject(
		projectData.Name,
		projectData.Status,
		projectData.StartDate,
		projectData.EndDate,
		projectData.Budget,
		projectData.Description,
	); err != nil {
		return types.Project{}, err
	}

	s.logger.Info("creating project",
		zap.String("user_id", userID.String()),
		zap.String("name", projectData.Name))
	return s.repo.CreateProject(ctx, userID, projectData)
}

func (s *projectService) UpdateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectUpdatePayload) (types.Project, error) {
	// Validate project data
	if err := validateProject(
		projectData.Name,
		projectData.Status,
		projectData.StartDate,
		projectData.EndDate,
		projectData.Budget,
		projectData.Description,
	); err != nil {
		return types.Project{}, err
	}

	s.logger.Info("updating project",
		zap.String("user_id", userID.String()),
		zap.String("project_id", projectData.ProjectID.String()))

	return s.repo.UpdateProject(ctx, userID, projectData)
}

func (s *projectService) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	s.logger.Info("deleting project",
		zap.String("user_id", userID.String()),
		zap.String("project_id", projectID.String()))
	return s.repo.DeleteProject(ctx, userID, projectID)
}

func (s *projectService) GetProjectWallets(ctx context.Context, userID, projectID uuid.UUID) ([]db.Wallet, error) {
	s.logger.Info("getting project wallets",
		zap.String("user_id", userID.String()),
		zap.String("project_id", projectID.String()))
	return s.repo.GetProjectWallets(ctx, userID, projectID)
}

func (s *projectService) ListProjectsPaginated(ctx context.Context, userID uuid.UUID, cursor time.Time, cursorID uuid.UUID, limit int32) ([]types.Project, error) {
	s.logger.Info("listing paginated projects",
		zap.String("user_id", userID.String()),
		zap.Time("cursor", cursor),
		zap.String("cursor_id", cursorID.String()),
		zap.Int32("limit", limit))

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	return s.repo.ListProjectsPaginated(ctx, userID, cursor, cursorID, limit)
}

func (s *projectService) SearchProjects(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]types.Project, error) {
	s.logger.Info("searching projects",
		zap.String("user_id", userID.String()),
		zap.String("query", query),
		zap.Int32("limit", limit))
	return s.repo.SearchProjects(ctx, userID, query, limit)
}

func isValidProjectStatus(status string) bool {
	validStatuses := []string{"ongoing", "completed", "canceled"}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}
