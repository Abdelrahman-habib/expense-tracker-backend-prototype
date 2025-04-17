package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

func (r *contactRepository) DeleteContact(ctx context.Context, contactID, userID uuid.UUID) error {
	if contactID == uuid.Nil || userID == uuid.Nil {
		return fmt.Errorf("invalid contact id or user id")
	}

	err := r.q.DeleteContact(ctx, db.DeleteContactParams{
		ContactID: contactID,
		UserID:    userID,
	})
	if err != nil {
		return errors.HandleRepositoryError(err, "delete", "contact")
	}

	return nil
}
