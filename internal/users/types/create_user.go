package types

import (
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateUserPayload struct {
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	ExternalID    string  `json:"external_id"`
	Provider      string  `json:"provider"`
	AddressLine1  *string `json:"address_line1,omitempty"`
	AddressLine2  *string `json:"address_line2,omitempty"`
	Country       *string `json:"country,omitempty"`
	City          *string `json:"city,omitempty"`
	StateProvince *string `json:"state_province,omitempty"`
	ZipPostalCode *string `json:"zip_postal_code,omitempty"`
}

func (c *CreateUserPayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":           validation.Validate(c.Name, validation.Required, validation.Length(1, 255)),
		"email":          validation.Validate(c.Email, validation.Required, is.Email),
		"external_id":    validation.Validate(c.ExternalID, validation.Required),
		"provider":       validation.Validate(c.Provider, validation.Required),
		"country":        validation.Validate(c.Country, is.CountryCode2),
		"address_line1":  validation.Validate(c.AddressLine1, validation.Length(0, 255)),
		"address_line2":  validation.Validate(c.AddressLine2, validation.Length(0, 255)),
		"city":           validation.Validate(c.City, validation.Length(0, 255)),
		"state_province": validation.Validate(c.StateProvince, validation.Length(0, 255)),
		"zip_code":       validation.Validate(c.ZipPostalCode, validate.Zipcode),
	}.Filter()
}
