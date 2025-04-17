package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

func (r *contactRepository) ListContactsPaginated(ctx context.Context, userID uuid.UUID, cursor *time.Time, cursorID *uuid.UUID, limit int32) ([]types.Contact, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user id")
	}

	if cursor == nil {
		now := time.Now()
		cursor = &now
	}
	if cursorID == nil {
		id := uuid.New()
		cursorID = &id
	}

	contacts, err := r.q.ListContactsPaginated(ctx, db.ListContactsPaginatedParams{
		UserID:    userID,
		CreatedAt: pgtype.Timestamp{Time: *cursor, Valid: true},
		ContactID: *cursorID,
		Limit:     limit,
	})
	if err != nil {
		return nil, errors.HandleRepositoryError(err, "list", "contacts")
	}

	return toContacts(contacts), nil
}
