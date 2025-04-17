package repository

import (
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
)

type contactRepository struct {
	q *db.Queries
}

// New creates a new contact repository
func New(q *db.Queries) Repository {
	return &contactRepository{q: q}
}
