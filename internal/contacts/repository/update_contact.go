package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
)

func (r *contactRepository) UpdateContact(ctx context.Context, payload types.ContactUpdatePayload, userID uuid.UUID) (types.Contact, error) {
	if payload.ContactID == uuid.Nil || userID == uuid.Nil {
		return types.Contact{}, fmt.Errorf("invalid contact id or user id")
	}

	params := updateContactParamsFromPayload(payload, userID)
	contact, err := r.q.UpdateContact(ctx, params)
	if err != nil {
		return types.Contact{}, errors.HandleRepositoryError(err, "update", "contact")
	}

	return toContact(contact), nil
}
