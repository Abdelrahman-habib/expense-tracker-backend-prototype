package service

import (
	"context"

	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TagService interface {
	ListTags(ctx context.Context, userID uuid.UUID) ([]types.Tag, error)
	GetTag(ctx context.Context, userID, tagID uuid.UUID) (types.Tag, error)
	CreateTag(ctx context.Context, userID uuid.UUID, tagData types.TagCreatePayload) (types.Tag, error)
	UpdateTag(ctx context.Context, userID uuid.UUID, tagData types.TagUpdatePayload) (types.Tag, error)
	DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error
	DeleteUserTags(ctx context.Context, userID uuid.UUID) error
}

type tagService struct {
	repo   repository.TagRepository
	logger *zap.Logger
}

func NewTagService(repo repository.TagRepository, logger *zap.Logger) TagService {
	return &tagService{
		repo:   repo,
		logger: logger,
	}
}

// ListTags returns all tags for a user
func (s *tagService) ListTags(ctx context.Context, userID uuid.UUID) ([]types.Tag, error) {
	return s.repo.ListTags(ctx, userID)
}

// GetTag returns a specific tag by ID
func (s *tagService) GetTag(ctx context.Context, userID, tagID uuid.UUID) (types.Tag, error) {
	return s.repo.GetTag(ctx, userID, tagID)
}

// CreateTag creates a new tag for a user
func (s *tagService) CreateTag(ctx context.Context, userID uuid.UUID, tagData types.TagCreatePayload) (types.Tag, error) {
	return s.repo.CreateTag(ctx, userID, tagData)
}

// UpdateTag updates an existing tag
func (s *tagService) UpdateTag(ctx context.Context, userID uuid.UUID, tagData types.TagUpdatePayload) (types.Tag, error) {
	return s.repo.UpdateTag(ctx, userID, tagData)
}

// DeleteTag deletes a specific tag
func (s *tagService) DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error {
	return s.repo.DeleteTag(ctx, userID, tagID)
}

// DeleteUserTags deletes all tags for a user
func (s *tagService) DeleteUserTags(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteUserTags(ctx, userID)
}
