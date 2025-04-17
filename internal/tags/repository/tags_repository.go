package repository

import (
	"context"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/tags/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TagRepository interface {
	ListTags(ctx context.Context, userID uuid.UUID) ([]types.Tag, error)
	GetTag(ctx context.Context, userID, tagID uuid.UUID) (types.Tag, error)
	CreateTag(ctx context.Context, userID uuid.UUID, tagData types.TagCreatePayload) (types.Tag, error)
	UpdateTag(ctx context.Context, userID uuid.UUID, tagData types.TagUpdatePayload) (types.Tag, error)
	DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error
	DeleteUserTags(ctx context.Context, userID uuid.UUID) error
}

type tagRepository struct {
	queries *db.Queries
}

func NewTagRepository(queries *db.Queries) TagRepository {
	return &tagRepository{queries: queries}
}

func (t *tagRepository) CreateTag(ctx context.Context, userID uuid.UUID, tagData types.TagCreatePayload) (types.Tag, error) {

	tag := db.CreateTagParams{
		UserID: userID,
		Name:   tagData.Name,
		Color: pgtype.Text{
			String: utils.StringPtrToString(tagData.Color),
			Valid:  tagData.Color != nil,
		},
	}

	createdtag, err := t.queries.CreateTag(ctx, tag)
	if err != nil {
		return types.Tag{}, errors.HandleRepositoryError(err, "create", "tag")
	}

	return types.Tag{
		TagID:     createdtag.TagID,
		Name:      createdtag.Name,
		Color:     &createdtag.Color.String,
		CreatedAt: createdtag.CreatedAt.Time,
		UpdatedAt: createdtag.UpdatedAt.Time,
	}, nil
}

func (t *tagRepository) ListTags(ctx context.Context, userID uuid.UUID) ([]types.Tag, error) {
	tags, err := t.queries.ListTags(ctx, userID)
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "list", "tags")
	}

	var result []types.Tag
	for _, tag := range tags {
		result = append(result, types.Tag{
			TagID:     tag.TagID,
			Name:      tag.Name,
			Color:     &tag.Color.String,
			CreatedAt: tag.CreatedAt.Time,
			UpdatedAt: tag.UpdatedAt.Time,
		})
	}
	return result, nil
}

func (t *tagRepository) GetTag(ctx context.Context, userID, tagID uuid.UUID) (types.Tag, error) {
	tag, err := t.queries.GetTag(ctx, db.GetTagParams{
		UserID: userID,
		TagID:  tagID,
	})
	if err != nil {
		return types.Tag{}, errors.HandleRepositoryError(err, "get", "tag")
	}

	return types.Tag{
		TagID:     tag.TagID,
		Name:      tag.Name,
		Color:     &tag.Color.String,
		CreatedAt: tag.CreatedAt.Time,
		UpdatedAt: tag.UpdatedAt.Time,
	}, nil
}

func (t *tagRepository) UpdateTag(ctx context.Context, userID uuid.UUID, tagData types.TagUpdatePayload) (types.Tag, error) {
	params := db.UpdateTagParams{
		UserID: userID,
		TagID:  tagData.TagID,
		Name:   tagData.Name,
		Color: pgtype.Text{
			String: *tagData.Color,
			Valid:  tagData.Color != nil,
		},
	}

	updatedTag, err := t.queries.UpdateTag(ctx, params)
	if err != nil {
		return types.Tag{}, errors.HandleRepositoryError(err, "update", "tag")
	}

	return types.Tag{
		TagID:     updatedTag.TagID,
		Name:      updatedTag.Name,
		Color:     &updatedTag.Color.String,
		CreatedAt: updatedTag.CreatedAt.Time,
		UpdatedAt: updatedTag.UpdatedAt.Time,
	}, nil
}

func (t *tagRepository) DeleteTag(ctx context.Context, userID, tagID uuid.UUID) error {
	err := t.queries.DeleteTag(ctx, db.DeleteTagParams{
		UserID: userID,
		TagID:  tagID,
	})
	if err != nil {
		return errors.HandleRepositoryError(err, "delete", "tag")
	}
	return err
}

func (t *tagRepository) DeleteUserTags(ctx context.Context, userID uuid.UUID) error {
	err := t.queries.DeleteUserTags(ctx, userID)
	if err != nil {
		return errors.HandleRepositoryError(err, "delete", "tags")
	}
	return err
}
