package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/repository"
	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ContactService interface {
	GetContact(ctx context.Context, contactID, userID uuid.UUID) (types.Contact, error)
	ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Contact, error)
	CreateContact(ctx context.Context, payload types.ContactCreatePayload, userID uuid.UUID) (types.Contact, error)
	UpdateContact(ctx context.Context, payload types.ContactUpdatePayload, userID uuid.UUID) (types.Contact, error)
	DeleteContact(ctx context.Context, contactID, userID uuid.UUID) error
	ListContactsPaginated(ctx context.Context, userID uuid.UUID, cursor *time.Time, cursorID *uuid.UUID, limit int32) ([]types.Contact, error)
	SearchContacts(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Contact, error)
	SearchContactsByPhone(ctx context.Context, userID uuid.UUID, phone string, limit int32) ([]types.Contact, error)
}

type contactService struct {
	repo   repository.Repository
	logger *zap.Logger
}

func NewContactService(repo repository.Repository, logger *zap.Logger) ContactService {
	return &contactService{
		repo:   repo,
		logger: logger.With(zap.String("component", "contact_service")),
	}
}

// cleanPhoneNumber removes any '+' or '-' characters from the phone number
func cleanPhoneNumber(phone string) string {
	phone = strings.ReplaceAll(phone, "+", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, " ", "")
	return phone
}

// Common validation function
func validateContact(name string, tags []uuid.UUID) error {
	// Validate required fields
	if name == "" {
		return fmt.Errorf("contact name is required")
	}

	// Validate text field lengths
	if len(name) > types.MaxNameLength {
		return fmt.Errorf("name exceeds maximum length of %d characters", types.MaxNameLength)
	}

	// Validate tags
	if tags != nil && len(tags) > types.MaxTagsCount {
		return fmt.Errorf("number of tags exceeds maximum allowed of %d", types.MaxTagsCount)
	}

	// Validate for duplicate tags
	if tags != nil {
		seen := make(map[uuid.UUID]bool)
		for _, tag := range tags {
			if seen[tag] {
				return fmt.Errorf("duplicate tag found: %s", tag)
			}
			seen[tag] = true
		}
	}

	return nil
}

func (s *contactService) CreateContact(ctx context.Context, payload types.ContactCreatePayload, userID uuid.UUID) (types.Contact, error) {
	s.logger.Info("creating contact",
		zap.String("user_id", userID.String()),
		zap.String("name", payload.Name))

	if err := validateContact(payload.Name, payload.Tags); err != nil {
		return types.Contact{}, err
	}

	// Clean phone number if provided
	if payload.Phone != nil {
		cleaned := cleanPhoneNumber(*payload.Phone)
		payload.Phone = &cleaned
	}

	return s.repo.CreateContact(ctx, payload, userID)
}

func (s *contactService) GetContact(ctx context.Context, contactID, userID uuid.UUID) (types.Contact, error) {
	s.logger.Info("getting contact",
		zap.String("contact_id", contactID.String()),
		zap.String("user_id", userID.String()))
	return s.repo.GetContact(ctx, contactID, userID)
}

func (s *contactService) ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Contact, error) {
	s.logger.Info("listing contacts",
		zap.String("user_id", userID.String()),
		zap.Int32("limit", limit),
		zap.Int32("offset", offset))

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	return s.repo.ListContacts(ctx, userID, limit, offset)
}

func (s *contactService) UpdateContact(ctx context.Context, payload types.ContactUpdatePayload, userID uuid.UUID) (types.Contact, error) {
	s.logger.Info("updating contact",
		zap.String("contact_id", payload.ContactID.String()),
		zap.String("user_id", userID.String()))

	if err := validateContact(payload.Name, payload.Tags); err != nil {
		return types.Contact{}, err
	}

	// Clean phone number if provided
	if payload.Phone != nil {
		cleaned := cleanPhoneNumber(*payload.Phone)
		payload.Phone = &cleaned
	}

	return s.repo.UpdateContact(ctx, payload, userID)
}

func (s *contactService) DeleteContact(ctx context.Context, contactID, userID uuid.UUID) error {
	s.logger.Info("deleting contact",
		zap.String("contact_id", contactID.String()),
		zap.String("user_id", userID.String()))
	return s.repo.DeleteContact(ctx, contactID, userID)
}

func (s *contactService) ListContactsPaginated(ctx context.Context, userID uuid.UUID, cursor *time.Time, cursorID *uuid.UUID, limit int32) ([]types.Contact, error) {
	s.logger.Info("listing paginated contacts",
		zap.String("user_id", userID.String()),
		zap.Any("cursor", cursor),
		zap.Any("cursor_id", cursorID),
		zap.Int32("limit", limit))

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	return s.repo.ListContactsPaginated(ctx, userID, cursor, cursorID, limit)
}

func (s *contactService) SearchContacts(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Contact, error) {
	s.logger.Info("searching contacts by name",
		zap.String("user_id", userID.String()),
		zap.String("name", name),
		zap.Int32("limit", limit))

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	return s.repo.SearchContacts(ctx, userID, name, limit)
}

func (s *contactService) SearchContactsByPhone(ctx context.Context, userID uuid.UUID, phone string, limit int32) ([]types.Contact, error) {
	s.logger.Info("searching contacts by phone",
		zap.String("user_id", userID.String()),
		zap.String("phone", phone),
		zap.Int32("limit", limit))

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	// Clean the phone number query
	cleanedPhone := cleanPhoneNumber(phone)

	return s.repo.SearchContactsByPhone(ctx, userID, cleanedPhone, limit)
}
