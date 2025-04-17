package types

import (
	"net/http"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
)

const (
	MaxDescriptionLength = 1000
	MaxNameLength        = 255
	MaxAddressLength     = 255
	MaxTagsCount         = 10
)

// Project represents a project entity
// @Description Project information including details, status, dates, location and tags
type Project struct {
	ProjectID     uuid.UUID   `json:"projectId" example:"123e4567-e89b-12d3-a456-426614174000" format:"uuid"`
	Name          string      `json:"name" example:"My Project" minLength:"1" maxLength:"255"`
	Description   *string     `json:"description,omitempty" example:"Detailed project description" maxLength:"1000"`
	Status        string      `json:"status" example:"ongoing" enums:"ongoing,completed,canceled"`
	StartDate     *time.Time  `json:"startDate,omitempty" example:"2024-01-01T00:00:00Z" format:"date-time"`
	EndDate       *time.Time  `json:"endDate,omitempty" example:"2024-12-31T00:00:00Z" format:"date-time"`
	Budget        *float64    `json:"budget,omitempty" example:"10000.50" minimum:"0"`
	AddressLine1  *string     `json:"addressLine1,omitempty" example:"123 Main St" maxLength:"255"`
	AddressLine2  *string     `json:"addressLine2,omitempty" example:"Suite 100" maxLength:"255"`
	Country       *string     `json:"country,omitempty" example:"US" format:"iso-3166-1-alpha-2" pattern:"^[A-Z]{2}$"`
	City          *string     `json:"city,omitempty" example:"New York" maxLength:"255"`
	StateProvince *string     `json:"stateProvince,omitempty" example:"NY" maxLength:"255"`
	ZipPostalCode *string     `json:"zipPostalCode,omitempty" example:"10001" format:"zip-code" pattern:"^\\d{5}(?:[-\\s]\\d{4})?$"`
	Website       *string     `json:"website,omitempty" example:"https://example.com" format:"uri"`
	Tags          []uuid.UUID `json:"tags,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001" format:"uuid" validate:"unique,max=10"`
	CreatedAt     time.Time   `json:"createdAt" example:"2024-01-01T00:00:00Z" format:"date-time"`
	UpdatedAt     time.Time   `json:"updatedAt" example:"2024-01-01T00:00:00Z" format:"date-time"`
}

// ProjectCreatePayload represents the payload for creating a new project
// @Description Payload for creating a new project
type ProjectCreatePayload struct {
	Name          string      `json:"name" example:"My Project" minLength:"1" maxLength:"255" validate:"required"`
	Description   *string     `json:"description" extensions:"x-nullable" example:"Detailed project description" maxLength:"1000"`
	Status        string      `json:"status" example:"ongoing" enums:"ongoing,completed,canceled" validate:"required" default:"ongoing"`
	StartDate     *time.Time  `json:"startDate" extensions:"x-nullable" example:"2024-01-01T00:00:00Z" format:"date-time"`
	EndDate       *time.Time  `json:"endDate" extensions:"x-nullable" example:"2024-12-31T00:00:00Z" format:"date-time"`
	Budget        *float64    `json:"budget" extensions:"x-nullable" example:"10000.50" minimum:"0"`
	AddressLine1  *string     `json:"addressLine1" extensions:"x-nullable" example:"123 Main St" maxLength:"255"`
	AddressLine2  *string     `json:"addressLine2" extensions:"x-nullable" example:"Suite 100" maxLength:"255"`
	Country       *string     `json:"country" extensions:"x-nullable" example:"US" format:"iso-3166-1-alpha-2" pattern:"^[A-Z]{2}$"`
	City          *string     `json:"city" extensions:"x-nullable" example:"New York" maxLength:"255"`
	StateProvince *string     `json:"stateProvince" extensions:"x-nullable" example:"NY" maxLength:"255"`
	ZipPostalCode *string     `json:"zipPostalCode" extensions:"x-nullable" example:"10001" format:"zip-code" pattern:"^\\d{5}(?:[-\\s]\\d{4})?$"`
	Website       *string     `json:"website" extensions:"x-nullable" example:"https://example.com" format:"uri"`
	Tags          []uuid.UUID `json:"tags" items:"uuid"  example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001" format:"uuid" validate:"unique,max=10"`
}

// Bind implements render.Binder interface
func (c *ProjectCreatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":          validation.Validate(c.Name, validation.Required, validation.Length(1, MaxNameLength)),
		"description":   validation.Validate(c.Description, validation.When(c.Description != nil, validation.Length(0, MaxDescriptionLength))),
		"status":        validation.Validate(c.Status, validation.Required, validation.In(string(db.ProjectsStatusOngoing), string(db.ProjectsStatusCompleted), string(db.ProjectsStatusCanceled))),
		"end_date":      validation.Validate(c.EndDate, validation.When(c.StartDate != nil && c.EndDate != nil, validation.Min(c.StartDate).Error("end date must be after start date"))),
		"country":       validation.Validate(c.Country, validation.When(c.Country != nil, is.CountryCode2)),
		"zip_code":      validation.Validate(c.ZipPostalCode, validation.When(c.ZipPostalCode != nil, validate.Zipcode)),
		"website":       validation.Validate(c.Website, validation.When(c.Website != nil, is.URL)),
		"address_line1": validation.Validate(c.AddressLine1, validation.When(c.AddressLine1 != nil, validation.Length(0, MaxAddressLength))),
		"address_line2": validation.Validate(c.AddressLine2, validation.When(c.AddressLine2 != nil, validation.Length(0, MaxAddressLength))),
		"city":          validation.Validate(c.City, validation.When(c.City != nil, validation.Length(0, MaxAddressLength))),
		"tags":          validation.Validate(c.Tags, validation.Length(0, MaxTagsCount), validation.Each(is.UUID)),
		"budget":        validation.Validate(c.Budget, validation.When(c.Budget != nil, validation.Min(0.0).Error("budget must be bigger than 0"))),
	}.Filter()
}

// ProjectUpdatePayload represents the payload for updating an existing project
// @Description Payload for updating an existing project
type ProjectUpdatePayload struct {
	ProjectID     uuid.UUID   `json:"-" example:"123e4567-e89b-12d3-a456-426614174000" format:"uuid"`
	Name          string      `json:"name" example:"My Project" minLength:"1" maxLength:"255"`
	Description   *string     `json:"description" extensions:"x-nullable" example:"Detailed project description" maxLength:"1000"`
	Status        string      `json:"status" example:"ongoing" enums:"ongoing,completed,canceled"`
	StartDate     *time.Time  `json:"startDate" extensions:"x-nullable" example:"2024-01-01T00:00:00Z" format:"date-time"`
	EndDate       *time.Time  `json:"endDate" extensions:"x-nullable" example:"2024-12-31T00:00:00Z" format:"date-time"`
	Budget        *float64    `json:"budget" extensions:"x-nullable" example:"10000.50" minimum:"0"`
	AddressLine1  *string     `json:"addressLine1" extensions:"x-nullable" example:"123 Main St" maxLength:"255"`
	AddressLine2  *string     `json:"addressLine2" extensions:"x-nullable" example:"Suite 100" maxLength:"255"`
	Country       *string     `json:"country" extensions:"x-nullable" example:"US" format:"iso-3166-1-alpha-2" pattern:"^[A-Z]{2}$"`
	City          *string     `json:"city" extensions:"x-nullable" example:"New York" maxLength:"255"`
	StateProvince *string     `json:"stateProvince" extensions:"x-nullable" example:"NY" maxLength:"255"`
	ZipPostalCode *string     `json:"zipPostalCode" extensions:"x-nullable" example:"10001" format:"zip-code" pattern:"^\\d{5}(?:[-\\s]\\d{4})?$"`
	Website       *string     `json:"website" extensions:"x-nullable" example:"https://example.com" format:"uri"`
	Tags          []uuid.UUID `json:"tags,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001" format:"uuid" validate:"unique,max=10"`
}

// Bind implements render.Binder interface
func (u *ProjectUpdatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":          validation.Validate(u.Name, validation.Required, validation.Length(1, MaxNameLength)),
		"description":   validation.Validate(u.Description, validation.When(u.Description != nil, validation.Length(0, MaxDescriptionLength))),
		"status":        validation.Validate(u.Status, validation.Required, validation.In(string(db.ProjectsStatusOngoing), string(db.ProjectsStatusCompleted), string(db.ProjectsStatusCanceled))),
		"end_date":      validation.Validate(u.EndDate, validation.When(u.StartDate != nil && u.EndDate != nil, validation.Min(u.StartDate).Error("end date must be after start date"))),
		"country":       validation.Validate(u.Country, validation.When(u.Country != nil, is.CountryCode2)),
		"zip_code":      validation.Validate(u.ZipPostalCode, validation.When(u.ZipPostalCode != nil, validate.Zipcode)),
		"website":       validation.Validate(u.Website, validation.When(u.Website != nil, is.URL)),
		"address_line1": validation.Validate(u.AddressLine1, validation.When(u.AddressLine1 != nil, validation.Length(0, MaxAddressLength))),
		"address_line2": validation.Validate(u.AddressLine2, validation.When(u.AddressLine2 != nil, validation.Length(0, MaxAddressLength))),
		"city":          validation.Validate(u.City, validation.When(u.City != nil, validation.Length(0, MaxAddressLength))),
		"tags":          validation.Validate(u.Tags, validation.Length(0, MaxTagsCount), validation.Each(is.UUID)),
		"budget":        validation.Validate(u.Budget, validation.When(u.Budget != nil, validation.Min(0.0).Error("budget must be bigger than 0"))),
	}.Filter()
}

func (p *Project) ToUpdatePayload() ProjectUpdatePayload {
	return ProjectUpdatePayload{
		ProjectID:     p.ProjectID,
		Name:          p.Name,          // Non-optional
		Description:   p.Description,   // Optional
		Status:        p.Status,        // Non-optional
		StartDate:     p.StartDate,     // Optional
		EndDate:       p.EndDate,       // Optional
		Budget:        p.Budget,        // Optional
		AddressLine1:  p.AddressLine1,  // Optional
		AddressLine2:  p.AddressLine2,  // Optional
		Country:       p.Country,       // Optional
		City:          p.City,          // Optional
		StateProvince: p.StateProvince, // Optional
		ZipPostalCode: p.ZipPostalCode, // Optional
		Website:       p.Website,       // Optional
		Tags:          p.Tags,          // Optional
	}
}
