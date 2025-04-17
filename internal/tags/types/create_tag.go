package types

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// TagCreatePayload represents the payload for creating a new tag
// @Description Payload for creating a new tag with name and optional color
type TagCreatePayload struct {
	Name  string  `json:"name" binding:"required" example:"Important" minLength:"1" maxLength:"255"`
	Color *string `json:"color,omitempty" example:"#FF5733" format:"hex-color"`
}

func (c *TagCreatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":  validation.Validate(c.Name, validation.Required, validation.Length(1, 255)),
		"color": validation.Validate(c.Color, is.HexColor),
	}.Filter()
}
