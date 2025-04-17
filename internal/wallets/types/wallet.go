package types

import (
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
)

const (
	MaxNameLength = 255
	MaxTagsCount  = 10
)

// Wallet represents the domain model for a wallet
// @Description A wallet entity
type Wallet struct {
	WalletID  uuid.UUID   `json:"walletId" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    uuid.UUID   `json:"userId" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProjectID *uuid.UUID  `json:"projectId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name      string      `json:"name" example:"My Wallet"`
	Balance   *float64    `json:"balance,omitempty" example:"100.50"`
	Currency  string      `json:"currency" example:"USD"`
	Tags      []uuid.UUID `json:"tags,omitempty"`
	CreatedAt time.Time   `json:"createdAt" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time   `json:"updatedAt" example:"2023-01-01T00:00:00Z"`
}

// WalletCreatePayload represents the payload for creating a new wallet
// @Description Request payload for creating a new wallet
type WalletCreatePayload struct {
	ProjectID *uuid.UUID  `json:"projectId,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name      string      `json:"name" example:"My Wallet" binding:"required"`
	Balance   *float64    `json:"balance,omitempty" example:"100.50"`
	Currency  string      `json:"currency" example:"USD" binding:"required"`
	Tags      []uuid.UUID `json:"tags,omitempty"`
}

// Bind implements render.Binder interface and validates the create wallet payload
func (c *WalletCreatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":     validation.Validate(c.Name, validation.Required, validation.Length(1, MaxNameLength)),
		"currency": validation.Validate(c.Currency, validation.Required, is.CurrencyCode), // ISO 4217 currency codes are 3 characters
		"balance":  validation.Validate(c.Balance, validation.When(c.Balance != nil, validation.Min(0.0).Error("balance must be non-negative"))),
		"tags":     validation.Validate(c.Tags, validation.Length(0, MaxTagsCount)),
	}.Filter()
}

// WalletUpdatePayload represents the payload for updating an existing wallet
type WalletUpdatePayload struct {
	WalletID  uuid.UUID   `json:"-"` // Not part of JSON, set from URL
	ProjectID *uuid.UUID  `json:"projectId,omitempty"`
	Name      string      `json:"name"`
	Balance   *float64    `json:"balance,omitempty"`
	Currency  string      `json:"currency"`
	Tags      []uuid.UUID `json:"tags,omitempty"`
}

// Bind implements render.Binder interface and validates the update wallet payload
func (u *WalletUpdatePayload) Bind(r *http.Request) error {
	return validation.Errors{
		"name":     validation.Validate(u.Name, validation.Required, validation.Length(1, MaxNameLength)),
		"currency": validation.Validate(u.Currency, validation.Required, is.CurrencyCode),
		"balance":  validation.Validate(u.Balance, validation.When(u.Balance != nil, validation.Min(0.0).Error("balance must be non-negative"))),
		"tags":     validation.Validate(u.Tags, validation.Length(0, MaxTagsCount)),
	}.Filter()
}

// ToUpdatePayload converts a Wallet to WalletUpdatePayload
func (w *Wallet) ToUpdatePayload() WalletUpdatePayload {
	return WalletUpdatePayload{
		WalletID:  w.WalletID,
		ProjectID: w.ProjectID,
		Name:      w.Name,
		Balance:   w.Balance,
		Currency:  w.Currency,
		Tags:      w.Tags,
	}
}
