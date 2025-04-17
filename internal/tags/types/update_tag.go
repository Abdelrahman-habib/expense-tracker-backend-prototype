package types

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
)

// TagUpdatePayload represents the payload for updating an existing tag
// @Description Payload for updating an existing tag's name and color
type TagUpdatePayload struct {
	TagID uuid.UUID `json:"-" example:"123e4567-e89b-12d3-a456-426614174000" format:"uuid"` // Set from URL parameter
	Name  string    `json:"name" example:"Important" minLength:"1" maxLength:"255"`
	Color *string   `json:"color,omitempty" example:"#FF5733" format:"hex-color"`
}

func (u *TagUpdatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":  validation.Validate(u.Name, validation.Required, validation.Length(1, 255)),
		"color": validation.Validate(u.Color, validation.When(u.Color != nil, is.HexColor)),
	}.Filter()
}
