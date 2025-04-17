package types

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
// @Description User profile information
type User struct {
	UserID        uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name          string    `json:"name" example:"John Doe"`
	Email         string    `json:"email" example:"john@example.com"`
	ExternalID    string    `json:"external_id" example:"user_123"`
	Provider      string    `json:"provider" example:"provider_name"`
	AddressLine1  *string   `json:"address_line1,omitempty" example:"123 Main St"`
	AddressLine2  *string   `json:"address_line2,omitempty" example:"Apt 4B"`
	Country       *string   `json:"country,omitempty" example:"US"`
	City          *string   `json:"city,omitempty" example:"New York"`
	StateProvince *string   `json:"state_province,omitempty" example:"NY"`
	ZipPostalCode *string   `json:"zip_postal_code,omitempty" example:"10001"`
	CreatedAt     time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}
