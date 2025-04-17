package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

func (r *contactRepository) GetContact(ctx context.Context, contactID, userID uuid.UUID) (types.Contact, error) {
	if contactID == uuid.Nil || userID == uuid.Nil {
		return types.Contact{}, fmt.Errorf("invalid contact id or user id")
	}

	contact, err := r.q.GetContact(ctx, db.GetContactParams{
		ContactID: contactID,
		UserID:    userID,
	})
	if err != nil {
		return types.Contact{}, errors.HandleRepositoryError(err, "get", "contact")
	}

	return toContact(contact), nil
}
