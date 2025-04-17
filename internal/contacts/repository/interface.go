package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
)

// Repository defines the interface for contact operations
type Repository interface {
	// GetContact retrieves a contact by ID and user ID
	GetContact(ctx context.Context, contactID, userID uuid.UUID) (types.Contact, error)

	// ListContacts retrieves a paginated list of contacts for a user
	ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Contact, error)

	// CreateContact creates a new contact
	CreateContact(ctx context.Context, payload types.ContactCreatePayload, userID uuid.UUID) (types.Contact, error)

	// UpdateContact updates an existing contact
	UpdateContact(ctx context.Context, payload types.ContactUpdatePayload, userID uuid.UUID) (types.Contact, error)

	// DeleteContact deletes a contact
	DeleteContact(ctx context.Context, contactID, userID uuid.UUID) error

	// ListContactsPaginated retrieves a cursor-paginated list of contacts
	ListContactsPaginated(ctx context.Context, userID uuid.UUID, cursor *time.Time, cursorID *uuid.UUID, limit int32) ([]types.Contact, error)

	// SearchContacts searches for contacts by name using trigram similarity
	SearchContacts(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Contact, error)

	// SearchContactsByPhone searches for contacts by phone number
	SearchContactsByPhone(ctx context.Context, userID uuid.UUID, phone string, limit int32) ([]types.Contact, error)
}
