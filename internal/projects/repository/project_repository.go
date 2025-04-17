package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/projects/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	ListProjects(ctx context.Context, userID uuid.UUID) ([]types.Project, error)
	GetProject(ctx context.Context, userID, projectID uuid.UUID) (types.Project, error)
	CreateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectCreatePayload) (types.Project, error)
	UpdateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectUpdatePayload) (types.Project, error)
	DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error
	GetProjectWallets(ctx context.Context, userID, projectID uuid.UUID) ([]db.Wallet, error)
	ListProjectsPaginated(ctx context.Context, userID uuid.UUID, cursor time.Time, cursorID uuid.UUID, limit int32) ([]types.Project, error)
	SearchProjects(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]types.Project, error)
}

type projectRepository struct {
	queries *db.Queries
}

func NewProjectRepository(queries *db.Queries) ProjectRepository {
	return &projectRepository{queries: queries}
}

func (p *projectRepository) CreateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectCreatePayload) (types.Project, error) {
	params := db.CreateProjectParams{
		UserID:        userID,
		Name:          projectData.Name,
		Description:   utils.ToNullableText(projectData.Description),
		Status:        db.ProjectsStatus(projectData.Status),
		StartDate:     utils.ToNullableTimestamp(projectData.StartDate),
		EndDate:       utils.ToNullableTimestamp(projectData.EndDate),
		Budget:        utils.ToNullableNumeric(projectData.Budget),
		AddressLine1:  utils.ToNullableText(projectData.AddressLine1),
		AddressLine2:  utils.ToNullableText(projectData.AddressLine2),
		Country:       utils.ToNullableText(projectData.Country),
		City:          utils.ToNullableText(projectData.City),
		StateProvince: utils.ToNullableText(projectData.StateProvince),
		ZipPostalCode: utils.ToNullableText(projectData.ZipPostalCode),
		Website:       utils.ToNullableText(projectData.Website),
		Tags:          projectData.Tags,
	}

	project, err := p.queries.CreateProject(ctx, params)
	if err != nil {
		return types.Project{}, errors.HandleRepositoryError(err, "create", "project(s)")
	}

	return toProject(project), nil
}

func (p *projectRepository) ListProjects(ctx context.Context, userID uuid.UUID) ([]types.Project, error) {
	projects, err := p.queries.ListProjects(ctx, userID)
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "list", "project(s)")
	}

	return toProjects(projects), nil
}

func (p *projectRepository) GetProject(ctx context.Context, userID, projectID uuid.UUID) (types.Project, error) {
	project, err := p.queries.GetProject(ctx, db.GetProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return types.Project{}, errors.HandleRepositoryError(err, "get", "project(s)")
	}

	return toProject(project), nil
}

// toNullableProjectStatus converts a string to NullProjectsStatus, setting Valid to true
// only for valid enum values
func toNullableProjectStatus(status string) db.NullProjectsStatus {
	projectStatus := db.ProjectsStatus(status)
	result := db.NullProjectsStatus{
		ProjectsStatus: projectStatus,
		Valid:          false,
	}

	switch projectStatus {
	case db.ProjectsStatusOngoing,
		db.ProjectsStatusCompleted,
		db.ProjectsStatusCanceled:
		result.Valid = true
	}

	return result
}

func (p *projectRepository) UpdateProject(ctx context.Context, userID uuid.UUID, projectData types.ProjectUpdatePayload) (types.Project, error) {
	if projectData.ProjectID == uuid.Nil || userID == uuid.Nil {
		return types.Project{}, fmt.Errorf("invalid project id or user id")
	}

	params := db.UpdateProjectParams{
		ProjectID:     projectData.ProjectID,
		UserID:        userID,
		Name:          utils.ToNullableText(&projectData.Name),
		Description:   utils.ToNullableText(projectData.Description),
		Status:        toNullableProjectStatus(projectData.Status),
		StartDate:     utils.ToNullableTimestamp(projectData.StartDate),
		EndDate:       utils.ToNullableTimestamp(projectData.EndDate),
		Budget:        utils.ToNullableNumeric(projectData.Budget),
		AddressLine1:  utils.ToNullableText(projectData.AddressLine1),
		AddressLine2:  utils.ToNullableText(projectData.AddressLine2),
		Country:       utils.ToNullableText(projectData.Country),
		City:          utils.ToNullableText(projectData.City),
		StateProvince: utils.ToNullableText(projectData.StateProvince),
		ZipPostalCode: utils.ToNullableText(projectData.ZipPostalCode),
		Website:       utils.ToNullableText(projectData.Website),
		Tags:          projectData.Tags,
	}

	project, err := p.queries.UpdateProject(ctx, params)
	if err != nil {
		return types.Project{}, errors.HandleRepositoryError(err, "update", "project(s)")
	}

	return toProject(project), nil
}

func (p *projectRepository) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	err := p.queries.DeleteProject(ctx, db.DeleteProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return errors.HandleRepositoryError(err, "delete", "project(s)")
	}
	return nil
}

func (p *projectRepository) GetProjectWallets(ctx context.Context, userID, projectID uuid.UUID) ([]db.Wallet, error) {
	wallets, err := p.queries.GetProjectWallets(ctx, db.GetProjectWalletsParams{
		ProjectID: utils.ToNullableUUID(projectID),
		UserID:    userID,
	})
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "get wallets for", "project(s)")
	}
	return wallets, nil
}

func (p *projectRepository) ListProjectsPaginated(ctx context.Context, userID uuid.UUID, cursor time.Time, cursorID uuid.UUID, limit int32) ([]types.Project, error) {
	projects, err := p.queries.ListProjectsPaginated(ctx, db.ListProjectsPaginatedParams{
		UserID:    userID,
		CreatedAt: utils.ToNullableTimestamp(&cursor),
		ProjectID: cursorID,
		Limit:     limit,
	})
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "list paginated", "project(s)")
	}

	return toProjects(projects), nil
}

func (p *projectRepository) SearchProjects(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]types.Project, error) {
	projects, err := p.queries.SearchProjects(ctx, db.SearchProjectsParams{
		UserID: userID,
		Name:   query,
		Limit:  limit,
	})
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "search", "project(s)")
	}

	return toProjects(projects), nil
}

// Helper functions to convert between domain and database types
func toProject(p db.Project) types.Project {
	return types.Project{
		ProjectID:     p.ProjectID,
		Name:          p.Name,
		Description:   utils.PgtextToStringPtr(p.Description),
		Status:        string(p.Status),
		StartDate:     utils.GetTimePtr(p.StartDate),
		EndDate:       utils.GetTimePtr(p.EndDate),
		Budget:        utils.GetFloat64Ptr(p.Budget),
		AddressLine1:  utils.PgtextToStringPtr(p.AddressLine1),
		AddressLine2:  utils.PgtextToStringPtr(p.AddressLine2),
		Country:       utils.PgtextToStringPtr(p.Country),
		City:          utils.PgtextToStringPtr(p.City),
		StateProvince: utils.PgtextToStringPtr(p.StateProvince),
		ZipPostalCode: utils.PgtextToStringPtr(p.ZipPostalCode),
		Website:       utils.PgtextToStringPtr(p.Website),
		Tags:          p.Tags,
		CreatedAt:     p.CreatedAt.Time,
		UpdatedAt:     p.UpdatedAt.Time,
	}
}

func toProjects(projects []db.Project) []types.Project {
	result := make([]types.Project, len(projects))
	for i, p := range projects {
		result[i] = toProject(p)
	}
	return result
}
