package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

func (r *contactRepository) ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Contact, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user id")
	}

	contacts, err := r.q.ListContacts(ctx, db.ListContactsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "list", "contacts")
	}

	return toContacts(contacts), nil
}
