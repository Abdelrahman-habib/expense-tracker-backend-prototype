package repository

import (
	"context"

	"github.com/Abdelrahman-habib/expense-tracker/internal/auth/service"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/users/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, userData types.CreateUserPayload) (types.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (types.User, error)
	GetUserByExternalID(ctx context.Context, externalID string) (types.User, error)
	ListUsers(ctx context.Context, params types.ListUsersParams) ([]types.User, error)
	SearchUsers(ctx context.Context, params types.SearchUsersParams) ([]types.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, userData types.UpdateUserPayload) (types.User, error)
	GetGoogleToken(ctx context.Context) (types.GoogleOauthToken, error)
	GetGoogleContacts(ctx context.Context, token string, pageToken string) (*types.PaginatedGoogleContacts, error)
}

type usersRepository struct {
	queries *db.Queries
	auth    service.Service
	logger  *zap.Logger
}

func NewUsersRepository(queries *db.Queries, logger *zap.Logger, auth service.Service) UsersRepository {
	return &usersRepository{
		queries: queries,
		logger:  logger,
		auth:    auth,
	}
}

func (r *usersRepository) CreateUser(ctx context.Context, userData types.CreateUserPayload) (types.User, error) {
	r.logger.Debug("creating user", zap.String("name", userData.Name), zap.String("email", userData.Email))

	params := db.CreateUserParams{
		Name:       userData.Name,
		Email:      userData.Email,
		ExternalID: userData.ExternalID,
		Provider:   userData.Provider,
		AddressLine1: pgtype.Text{
			String: utils.StringPtrToString(userData.AddressLine1),
			Valid:  userData.AddressLine1 != nil,
		},
		AddressLine2: pgtype.Text{
			String: utils.StringPtrToString(userData.AddressLine2),
			Valid:  userData.AddressLine2 != nil,
		},
		Country: pgtype.Text{
			String: utils.StringPtrToString(userData.Country),
			Valid:  userData.Country != nil,
		},
		City: pgtype.Text{
			String: utils.StringPtrToString(userData.City),
			Valid:  userData.City != nil,
		},
		StateProvince: pgtype.Text{
			String: utils.StringPtrToString(userData.StateProvince),
			Valid:  userData.StateProvince != nil,
		},
		ZipPostalCode: pgtype.Text{
			String: utils.StringPtrToString(userData.ZipPostalCode),
			Valid:  userData.ZipPostalCode != nil,
		},
	}

	user, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return types.User{}, err
	}

	return mapDBUserToUser(user), nil
}

func (r *usersRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug("deleting user", zap.String("user_id", userID.String()))
	err := r.queries.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *usersRepository) GetUser(ctx context.Context, userID uuid.UUID) (types.User, error) {
	r.logger.Debug("getting user", zap.String("user_id", userID.String()))

	user, err := r.queries.GetUser(ctx, userID)
	if err != nil {
		return types.User{}, err
	}

	return mapDBUserToUser(user), nil
}

func (r *usersRepository) GetUserByExternalID(ctx context.Context, externalID string) (types.User, error) {
	r.logger.Debug("getting user by external ID", zap.String("external_id", externalID))

	params := db.GetUserByExternalIDParams{
		ExternalID: externalID,
		Provider:   "clerk", // Default to clerk for backward compatibility
	}

	user, err := r.queries.GetUserByExternalID(ctx, params)
	if err != nil {
		return types.User{}, err
	}

	return mapDBUserToUser(user), nil
}

func (r *usersRepository) ListUsers(ctx context.Context, params types.ListUsersParams) ([]types.User, error) {
	r.logger.Debug("listing users", zap.Int32("limit", params.Limit), zap.Int32("offset", params.Offset))

	users, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, err
	}

	return mapDBUsersToUsers(users), nil
}

func (r *usersRepository) SearchUsers(ctx context.Context, params types.SearchUsersParams) ([]types.User, error) {
	r.logger.Debug("searching users", zap.String("name", params.Name), zap.Int32("limit", params.Limit))

	users, err := r.queries.SearchUsers(ctx, db.SearchUsersParams{
		Name:  params.Name,
		Limit: params.Limit,
	})
	if err != nil {
		return nil, err
	}

	return mapDBUsersToUsers(users), nil
}

func (r *usersRepository) UpdateUser(ctx context.Context, UserID uuid.UUID, userData types.UpdateUserPayload) (types.User, error) {
	r.logger.Debug("updating user", zap.String("user_id", UserID.String()))

	params := db.UpdateUserParams{
		UserID: UserID,
		Name:   userData.Name,
		Email:  userData.Email,
		AddressLine1: pgtype.Text{
			String: utils.StringPtrToString(userData.AddressLine1),
			Valid:  userData.AddressLine1 != nil,
		},
		AddressLine2: pgtype.Text{
			String: utils.StringPtrToString(userData.AddressLine2),
			Valid:  userData.AddressLine2 != nil,
		},
		Country: pgtype.Text{
			String: utils.StringPtrToString(userData.Country),
			Valid:  userData.Country != nil,
		},
		City: pgtype.Text{
			String: utils.StringPtrToString(userData.City),
			Valid:  userData.City != nil,
		},
		StateProvince: pgtype.Text{
			String: utils.StringPtrToString(userData.StateProvince),
			Valid:  userData.StateProvince != nil,
		},
		ZipPostalCode: pgtype.Text{
			String: utils.StringPtrToString(userData.ZipPostalCode),
			Valid:  userData.ZipPostalCode != nil,
		},
	}

	user, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return types.User{}, err
	}

	return mapDBUserToUser(user), nil
}

// Helper functions for mapping between types
func mapDBUserToUser(dbUser db.User) types.User {
	return types.User{
		UserID:        dbUser.UserID,
		Name:          dbUser.Name,
		Email:         dbUser.Email,
		ExternalID:    dbUser.ExternalID,
		Provider:      dbUser.Provider,
		AddressLine1:  utils.PgtextToStringPtr(dbUser.AddressLine1),
		AddressLine2:  utils.PgtextToStringPtr(dbUser.AddressLine2),
		Country:       utils.PgtextToStringPtr(dbUser.Country),
		City:          utils.PgtextToStringPtr(dbUser.City),
		StateProvince: utils.PgtextToStringPtr(dbUser.StateProvince),
		ZipPostalCode: utils.PgtextToStringPtr(dbUser.ZipPostalCode),
		CreatedAt:     dbUser.CreatedAt.Time,
		UpdatedAt:     dbUser.UpdatedAt.Time,
	}
}

func mapDBUsersToUsers(dbUsers []db.User) []types.User {
	users := make([]types.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = mapDBUserToUser(dbUser)
	}
	return users
}
