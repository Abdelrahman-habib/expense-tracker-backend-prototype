package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Abdelrahman-habib/expense-tracker/internal/core/errors"
	requestcontext "github.com/Abdelrahman-habib/expense-tracker/pkg/context"
	"github.com/google/uuid"
)

// TODO: consider another appraoch that doesn't envolve fetching from the db, maybe look into clerk metadata, can we assign the user id into the meta from the webhook?
func (u *UserHandler) WithUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := u.auth.GetUserClaims(r.Context())
		if !ok {
			u.RespondError(w, r, errors.ErrAuthorization(fmt.Errorf("user claims missing")))
			return
		}

		// Get user ID from claims
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			u.RespondError(w, r, errors.ErrAuthorization(fmt.Errorf("invalid user ID in claims")))
			return
		}

		// Parse UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			u.RespondError(w, r, errors.ErrAuthorization(fmt.Errorf("invalid user ID format")))
			return
		}

		// Get user from database
		user, err := u.service.GetUser(r.Context(), userID)
		if err != nil {
			u.RespondError(w, r, errors.ErrInternal(fmt.Errorf("user not found")))
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), requestcontext.UserIDKey, user.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
