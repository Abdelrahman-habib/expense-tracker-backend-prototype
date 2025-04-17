package types

import (
	"net/http"
	"net/url"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
)

const (
	MaxNameLength    = 255
	MaxAddressLength = 255
	MaxTagsCount     = 10
	MaxPhoneLength   = 20
)

// Contact represents the domain model for a contact
// @Description Contact information including personal details, contact methods, address and tags
type Contact struct {
	ContactID     uuid.UUID   `json:"contactId" example:"123e4567-e89b-12d3-a456-426614174000" format:"uuid"`
	UserID        uuid.UUID   `json:"userId" example:"123e4567-e89b-12d3-a456-426614174001" format:"uuid"`
	Name          string      `json:"name" example:"John Doe" minLength:"1" maxLength:"255"`
	Phone         *string     `json:"phone,omitempty" example:"+1-555-123-4567" maxLength:"20" format:"phone"`
	Email         *string     `json:"email,omitempty" example:"john.doe@example.com" format:"email"`
	AddressLine1  *string     `json:"addressLine1,omitempty" example:"123 Main St" maxLength:"255"`
	AddressLine2  *string     `json:"addressLine2,omitempty" example:"Suite 100" maxLength:"255"`
	Country       *string     `json:"country,omitempty" example:"US" format:"iso-3166-1-alpha-2"`
	City          *string     `json:"city,omitempty" example:"New York" maxLength:"255"`
	StateProvince *string     `json:"stateProvince,omitempty" example:"NY" maxLength:"255"`
	ZipPostalCode *string     `json:"zipPostalCode,omitempty" example:"10001" format:"zip-code"`
	Tags          []uuid.UUID `json:"tags,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001"`
	CreatedAt     time.Time   `json:"createdAt" example:"2024-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt     time.Time   `json:"updatedAt" example:"2024-01-01T00:00:00Z" format:"date-time"`
}

// ContactCreatePayload represents the payload for creating a new contact
// @Description Payload for creating a new contact
type ContactCreatePayload struct {
	Name          string      `json:"name" example:"John Doe" minLength:"1" maxLength:"255"`
	Phone         *string     `json:"phone,omitempty" example:"+1-555-123-4567" maxLength:"20" format:"phone"`
	Email         *string     `json:"email,omitempty" example:"john.doe@example.com" format:"email"`
	AddressLine1  *string     `json:"addressLine1,omitempty" example:"123 Main St" maxLength:"255"`
	AddressLine2  *string     `json:"addressLine2,omitempty" example:"Suite 100" maxLength:"255"`
	Country       *string     `json:"country,omitempty" example:"US" format:"iso-3166-1-alpha-2"`
	City          *string     `json:"city,omitempty" example:"New York" maxLength:"255"`
	StateProvince *string     `json:"stateProvince,omitempty" example:"NY" maxLength:"255"`
	ZipPostalCode *string     `json:"zipPostalCode,omitempty" example:"10001" format:"zip-code"`
	Tags          []uuid.UUID `json:"tags,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001"`
}

// Bind implements render.Binder interface and validates the create contact payload
func (c *ContactCreatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":          validation.Validate(c.Name, validation.Required, validation.Length(1, MaxNameLength)),
		"email":         validation.Validate(c.Email, validation.When(c.Email != nil, is.Email)),
		"phone":         validation.Validate(c.Phone, validation.When(c.Phone != nil, validation.Length(1, MaxPhoneLength), validate.PhoneNumber)),
		"country":       validation.Validate(c.Country, validation.When(c.Country != nil, is.CountryCode2)),
		"zip_code":      validation.Validate(c.ZipPostalCode, validation.When(c.ZipPostalCode != nil, validate.Zipcode)),
		"address_line1": validation.Validate(c.AddressLine1, validation.When(c.AddressLine1 != nil, validation.Length(1, MaxAddressLength))),
		"address_line2": validation.Validate(c.AddressLine2, validation.When(c.AddressLine2 != nil, validation.Length(1, MaxAddressLength))),
		"city":          validation.Validate(c.City, validation.When(c.City != nil, validation.Length(1, MaxAddressLength))),
		"tags":          validation.Validate(c.Tags, validation.Length(0, MaxTagsCount), validate.NoDuplicates(), validation.Each(is.UUID)),
	}.Filter()
}

// ContactUpdatePayload represents the payload for updating an existing contact
// @Description Payload for updating an existing contact
type ContactUpdatePayload struct {
	ContactID     uuid.UUID   `json:"-" example:"123e4567-e89b-12d3-a456-426614174000" format:"uuid"`
	Name          string      `json:"name" example:"John Doe" minLength:"1" maxLength:"255"`
	Phone         *string     `json:"phone,omitempty" example:"+1-555-123-4567" maxLength:"20" format:"phone"`
	Email         *string     `json:"email,omitempty" example:"john.doe@example.com" format:"email"`
	AddressLine1  *string     `json:"addressLine1,omitempty" example:"123 Main St" maxLength:"255"`
	AddressLine2  *string     `json:"addressLine2,omitempty" example:"Suite 100" maxLength:"255"`
	Country       *string     `json:"country,omitempty" example:"US" format:"iso-3166-1-alpha-2"`
	City          *string     `json:"city,omitempty" example:"New York" maxLength:"255"`
	StateProvince *string     `json:"stateProvince,omitempty" example:"NY" maxLength:"255"`
	ZipPostalCode *string     `json:"zipPostalCode,omitempty" example:"10001" format:"zip-code"`
	Tags          []uuid.UUID `json:"tags,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001"`
}

// Bind implements render.Binder interface and validates the update contact payload
func (u *ContactUpdatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":          validation.Validate(u.Name, validation.Required, validation.Length(1, MaxNameLength)),
		"email":         validation.Validate(u.Email, validation.When(u.Email != nil, is.Email)),
		"phone":         validation.Validate(u.Phone, validation.When(u.Phone != nil, validation.Length(1, MaxPhoneLength), validate.PhoneNumber)),
		"country":       validation.Validate(u.Country, validation.When(u.Country != nil, is.CountryCode2)),
		"zip_code":      validation.Validate(u.ZipPostalCode, validation.When(u.ZipPostalCode != nil, validate.Zipcode)),
		"address_line1": validation.Validate(u.AddressLine1, validation.When(u.AddressLine1 != nil, validation.Length(1, MaxAddressLength))),
		"address_line2": validation.Validate(u.AddressLine2, validation.When(u.AddressLine2 != nil, validation.Length(1, MaxAddressLength))),
		"city":          validation.Validate(u.City, validation.When(u.City != nil, validation.Length(1, MaxAddressLength))),
		"tags":          validation.Validate(u.Tags, validation.Length(0, MaxTagsCount), validate.NoDuplicates(), validation.Each(is.UUID)),
	}.Filter()
}

// ToUpdatePayload converts a Contact to ContactUpdatePayload
func (c *Contact) ToUpdatePayload() ContactUpdatePayload {
	return ContactUpdatePayload{
		ContactID:     c.ContactID,
		Name:          c.Name,
		Phone:         c.Phone,
		Email:         c.Email,
		AddressLine1:  c.AddressLine1,
		AddressLine2:  c.AddressLine2,
		Country:       c.Country,
		City:          c.City,
		StateProvince: c.StateProvince,
		ZipPostalCode: c.ZipPostalCode,
		Tags:          c.Tags,
	}
}

// SearchParams represents search parameters for contacts
// @Description Search parameters for filtering contacts
type SearchParams struct {
	types.SearchParams
	SearchByPhone bool `json:"searchByPhone" example:"false" description:"Enable phone number search"`
}

func ParseAndValidateSearchParams(query url.Values) (SearchParams, error) {
	var params SearchParams
	searchParams, err := types.ParseAndValidateSearchParams(query)
	if err != nil {
		return SearchParams{}, err
	}
	searchByPhone := query.Get("by_phone") == "true"
	params.Limit = searchParams.Limit
	params.Query = searchParams.Query
	params.SearchByPhone = searchByPhone
	return params, validation.Errors{
		"query": validation.Validate(params.Query, validation.When(searchByPhone, validate.PhoneNumber)),
	}.Filter()
}
