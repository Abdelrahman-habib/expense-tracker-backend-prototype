package types

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SearchUsersParams struct {
	Name  string `json:"name"`
	Limit int32  `json:"limit"`
}

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (c *SearchUsersParams) Bind(r *http.Request) error {
	return validation.Errors{
		"name":  validation.Validate(c.Name, validation.Required, validation.Length(1, 255)),
		"limit": validation.Validate(c.Limit, validation.Required),
	}.Filter()
}
func (c *ListUsersParams) Bind(r *http.Request) error {
	return validation.Errors{
		"limit":  validation.Validate(c.Limit, validation.Required),
		"offset": validation.Validate(c.Offset, validation.Required),
	}.Filter()
}
