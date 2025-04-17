package service

import (
	"context"
	"fmt"

	"errors"

	"github.com/Abdelrahman-habib/expense-tracker/internal/users/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/users/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UsersService interface {
	CreateUser(ctx context.Context, params types.CreateUserPayload) (types.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (types.User, error)
	GetUserByExternalID(ctx context.Context, clerkExUserID string) (types.User, error)
	ListUsers(ctx context.Context, params types.ListUsersParams) ([]types.User, error)
	SearchUsers(ctx context.Context, params types.SearchUsersParams) ([]types.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, params types.UpdateUserPayload) (types.User, error)
	GetGoogleContacts(ctx context.Context, pageToken string) (*types.PaginatedGoogleContacts, error)
}

type usersService struct {
	repo   repository.UsersRepository
	logger *zap.Logger
}

func NewUsersService(repo repository.UsersRepository, logger *zap.Logger) UsersService {
	return &usersService{
		repo:   repo,
		logger: logger,
	}
}

func (s *usersService) CreateUser(ctx context.Context, params types.CreateUserPayload) (types.User, error) {
	return s.repo.CreateUser(ctx, params)
}

func (s *usersService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteUser(ctx, userID)
}

func (s *usersService) GetUser(ctx context.Context, userID uuid.UUID) (types.User, error) {
	return s.repo.GetUser(ctx, userID)
}

func (s *usersService) GetUserByExternalID(ctx context.Context, clerkExUserID string) (types.User, error) {
	if clerkExUserID == "" {
		return types.User{}, errors.New("clerk external user ID is required")
	}

	return s.repo.GetUserByExternalID(ctx, clerkExUserID)
}

func (s *usersService) ListUsers(ctx context.Context, params types.ListUsersParams) ([]types.User, error) {
	if params.Limit <= 0 {
		params.Limit = 10 // Default limit
	}

	return s.repo.ListUsers(ctx, params)
}

func (s *usersService) SearchUsers(ctx context.Context, params types.SearchUsersParams) ([]types.User, error) {
	if params.Name == "" {
		return nil, errors.New("search name is required")
	}

	if params.Limit <= 0 {
		params.Limit = 10 // Default limit
	}

	return s.repo.SearchUsers(ctx, params)
}

func (s *usersService) UpdateUser(ctx context.Context, userID uuid.UUID, params types.UpdateUserPayload) (types.User, error) {
	return s.repo.UpdateUser(ctx, userID, params)
}

func (s *usersService) GetGoogleContacts(ctx context.Context, pageToken string) (*types.PaginatedGoogleContacts, error) {
	// First, get the Google OAuth token for the user
	token, err := s.repo.GetGoogleToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get google token: %w", err)
	}
	// Use the token to fetch Google contacts
	contacts, err := s.repo.GetGoogleContacts(ctx, token.Token, pageToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch google contacts: %w", err)
	}

	return contacts, nil
}
