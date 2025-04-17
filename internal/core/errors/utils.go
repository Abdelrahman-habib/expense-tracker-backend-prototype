package errors

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

// handleRepositoryError is a helper function to handle common database errors
func HandleRepositoryError(err error, operation, repoName string) error {
	if err == pgx.ErrNoRows {
		return &ErrorResponse{
			Type:    ErrorTypeNotFound,
			Message: fmt.Sprintf("%s not found", repoName),
			Err:     err,
		}
	}
	return &ErrorResponse{
		Type:    ErrorTypeDatabase,
		Message: fmt.Sprintf("Failed to %s %s", operation, repoName),
		Err:     err,
	}
}
