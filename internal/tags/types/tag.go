package types

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a tag entity
// @Description Tag information including name, color and metadata
type Tag struct {
	TagID     uuid.UUID `json:"tagId" example:"123e4567-e89b-12d3-a456-426614174000" format:"uuid"`
	Name      string    `json:"name" example:"Important" minLength:"1" maxLength:"255"`
	Color     *string   `json:"color,omitempty" example:"#FF5733" format:"hex-color"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt time.Time `json:"updatedAt" example:"2024-01-01T00:00:00Z" format:"date-time"`
}

func (t *Tag) ToUpdatePayload() TagUpdatePayload {
	return TagUpdatePayload{
		TagID: t.TagID,
		Name:  t.Name,
		Color: t.Color,
	}
}
