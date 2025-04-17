package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
)

func (r *contactRepository) CreateContact(ctx context.Context, payload types.ContactCreatePayload, userID uuid.UUID) (types.Contact, error) {
	if userID == uuid.Nil {
		return types.Contact{}, fmt.Errorf("invalid user id")
	}

	params := createContactParamsFromPayload(payload, userID)
	contact, err := r.q.CreateContact(ctx, params)
	if err != nil {
		return types.Contact{}, errors.HandleRepositoryError(err, "create", "contact")
	}

	return toContact(contact), nil
}
