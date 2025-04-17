package repository

import (
	"context"
	"errors"
	"slices"

	"github.com/Abdelrahman-habib/expense-tracker/internal/users/types"
)

func hasRequiredScopes(scopes []string) bool {
	contactScope := "https://www.googleapis.com/auth/contacts.readonly"
	return slices.Contains(scopes, contactScope) // true
}

func (r *usersRepository) GetGoogleToken(ctx context.Context) (types.GoogleOauthToken, error) {
	token, err := r.auth.GetGoogleToken(ctx)
	if err != nil {
		return types.GoogleOauthToken{}, err
	}
	if !hasRequiredScopes(token.Scopes) {
		return types.GoogleOauthToken{}, errors.New("not authorized to read contacts") // TODO: find better error message
	}
	return types.GoogleOauthToken{
		ExternalAccountID: token.ExternalAccountID,
		Token:             token.Token,
		Scopes:            token.Scopes,
	}, nil
}
