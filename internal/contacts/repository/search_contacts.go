package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

func (r *contactRepository) SearchContacts(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Contact, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user id")
	}

	contacts, err := r.q.SearchContacts(ctx, db.SearchContactsParams{
		UserID: userID,
		Name:   name,
		Limit:  limit,
	})
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "search", "contacts")
	}

	return toContacts(contacts), nil
}
