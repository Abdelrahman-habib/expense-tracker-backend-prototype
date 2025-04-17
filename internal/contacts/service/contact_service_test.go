package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock repository
type mockContactRepository struct {
	mock.Mock
}

func (m *mockContactRepository) GetContact(ctx context.Context, contactID, userID uuid.UUID) (types.Contact, error) {
	args := m.Called(ctx, contactID, userID)
	if args.Get(0) == nil {
		return types.Contact{}, args.Error(1)
	}
	return args.Get(0).(types.Contact), args.Error(1)
}

func (m *mockContactRepository) ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]types.Contact), args.Error(1)
}

func (m *mockContactRepository) CreateContact(ctx context.Context, payload types.ContactCreatePayload, userID uuid.UUID) (types.Contact, error) {
	args := m.Called(ctx, payload, userID)
	if args.Get(0) == nil {
		return types.Contact{}, args.Error(1)
	}
	return args.Get(0).(types.Contact), args.Error(1)
}

func (m *mockContactRepository) UpdateContact(ctx context.Context, payload types.ContactUpdatePayload, userID uuid.UUID) (types.Contact, error) {
	args := m.Called(ctx, payload, userID)
	if args.Get(0) == nil {
		return types.Contact{}, args.Error(1)
	}
	return args.Get(0).(types.Contact), args.Error(1)
}

func (m *mockContactRepository) DeleteContact(ctx context.Context, contactID, userID uuid.UUID) error {
	args := m.Called(ctx, contactID, userID)
	return args.Error(0)
}

func (m *mockContactRepository) ListContactsPaginated(ctx context.Context, userID uuid.UUID, cursor *time.Time, cursorID *uuid.UUID, limit int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, cursor, cursorID, limit)
	return args.Get(0).([]types.Contact), args.Error(1)
}

func (m *mockContactRepository) SearchContacts(ctx context.Context, userID uuid.UUID, name string, limit int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, name, limit)
	return args.Get(0).([]types.Contact), args.Error(1)
}

func (m *mockContactRepository) SearchContactsByPhone(ctx context.Context, userID uuid.UUID, phone string, limit int32) ([]types.Contact, error) {
	args := m.Called(ctx, userID, phone, limit)
	return args.Get(0).([]types.Contact), args.Error(1)
}

func setupTest(t *testing.T) (*mockContactRepository, ContactService) {
	mockRepo := new(mockContactRepository)
	logger := zap.NewNop()
	service := NewContactService(mockRepo, logger)
	return mockRepo, service
}

func TestContactService_CreateContact(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		payload types.ContactCreatePayload
		mock    func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful create",
			payload: types.ContactCreatePayload{
				Name:  "John Doe",
				Phone: utils.StringPtr("+1-555-123-4567"),
			},
			mock: func() {
				expectedContact := types.Contact{
					Name:  "John Doe",
					Phone: utils.StringPtr("15551234567"), // Note: phone is cleaned
				}
				mockRepo.On("CreateContact", ctx, mock.AnythingOfType("types.ContactCreatePayload"), userID).
					Return(expectedContact, nil)
			},
			wantErr: false,
		},
		{
			name: "empty name",
			payload: types.ContactCreatePayload{
				Name:  "",
				Phone: utils.StringPtr("+1-555-123-4567"),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "contact name is required",
		},
		{
			name: "name too long",
			payload: types.ContactCreatePayload{
				Name:  strings.Repeat("a", types.MaxNameLength+1),
				Phone: utils.StringPtr("+1-555-123-4567"),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "name exceeds maximum length",
		},
		{
			name: "too many tags",
			payload: types.ContactCreatePayload{
				Name:  "John Doe",
				Phone: utils.StringPtr("+1-555-123-4567"),
				Tags:  make([]uuid.UUID, types.MaxTagsCount+1),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "number of tags exceeds maximum allowed",
		},
		{
			name: "duplicate tags",
			payload: types.ContactCreatePayload{
				Name:  "John Doe",
				Phone: utils.StringPtr("+1-555-123-4567"),
				Tags:  []uuid.UUID{uuid.MustParse("00000000-0000-0000-0000-000000000001"), uuid.MustParse("00000000-0000-0000-0000-000000000001")},
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "duplicate tag found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			contact, err := service.CreateContact(ctx, tt.payload, userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, contact)
			mockRepo.AssertExpectations(t)

			// If phone was provided, verify it was cleaned
			if tt.payload.Phone != nil {
				cleaned := cleanPhoneNumber(*tt.payload.Phone)
				assert.Equal(t, cleaned, *contact.Phone)
			}
		})
	}
}

func TestContactService_GetContact(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful retrieval",
			mock: func() {
				expectedContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe",
					Phone:     utils.StringPtr("15551234567"),
				}
				mockRepo.On("GetContact", ctx, contactID, userID).Return(expectedContact, nil)
			},
			wantErr: false,
		},
		{
			name: "not found error",
			mock: func() {
				mockRepo.On("GetContact", ctx, contactID, userID).Return(types.Contact{}, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			contact, err := service.GetContact(ctx, contactID, userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, contactID, contact.ContactID)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContactService_ListContacts(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		limit   int32
		offset  int32
		mock    func()
		wantErr bool
		wantLen int
		errMsg  string
	}{
		{
			name:   "successful list",
			limit:  10,
			offset: 0,
			mock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "John Doe",
						Phone:     utils.StringPtr("15551234567"),
					},
					{
						ContactID: uuid.New(),
						Name:      "Jane Smith",
						Phone:     utils.StringPtr("15559876543"),
					},
				}
				mockRepo.On("ListContacts", ctx, userID, int32(10), int32(0)).Return(contacts, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "invalid limit",
			limit:   -1,
			offset:  0,
			mock:    func() {},
			wantErr: true,
			errMsg:  "limit must be positive",
		},
		{
			name:   "empty list",
			limit:  10,
			offset: 0,
			mock: func() {
				mockRepo.On("ListContacts", ctx, userID, int32(10), int32(0)).Return([]types.Contact{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			contacts, err := service.ListContacts(ctx, userID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, contacts, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContactService_UpdateContact(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name    string
		payload types.ContactUpdatePayload
		mock    func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful update",
			payload: types.ContactUpdatePayload{
				ContactID: contactID,
				Name:      "John Doe Updated",
				Phone:     utils.StringPtr("+1-555-123-4567"),
			},
			mock: func() {
				expectedContact := types.Contact{
					ContactID: contactID,
					Name:      "John Doe Updated",
					Phone:     utils.StringPtr("15551234567"), // Note: phone is cleaned
				}
				mockRepo.On("UpdateContact", ctx, mock.AnythingOfType("types.ContactUpdatePayload"), userID).
					Return(expectedContact, nil)
			},
			wantErr: false,
		},
		{
			name: "empty name",
			payload: types.ContactUpdatePayload{
				ContactID: contactID,
				Name:      "",
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "contact name is required",
		},
		{
			name: "name too long",
			payload: types.ContactUpdatePayload{
				ContactID: contactID,
				Name:      strings.Repeat("a", types.MaxNameLength+1),
			},
			mock:    func() {},
			wantErr: true,
			errMsg:  "name exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			contact, err := service.UpdateContact(ctx, tt.payload, userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, contact)
			mockRepo.AssertExpectations(t)

			// If phone was provided, verify it was cleaned
			if tt.payload.Phone != nil {
				cleaned := cleanPhoneNumber(*tt.payload.Phone)
				assert.Equal(t, cleaned, *contact.Phone)
			}
		})
	}
}

func TestContactService_DeleteContact(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "successful delete",
			mock: func() {
				mockRepo.On("DeleteContact", ctx, contactID, userID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found error",
			mock: func() {
				mockRepo.On("DeleteContact", ctx, contactID, userID).Return(errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			err := service.DeleteContact(ctx, contactID, userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContactService_ListContactsPaginated(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()
	cursorID := uuid.New()

	tests := []struct {
		name     string
		cursor   *time.Time
		cursorID *uuid.UUID
		limit    int32
		mock     func()
		wantErr  bool
		wantLen  int
		errMsg   string
	}{
		{
			name:     "successful pagination",
			cursor:   &now,
			cursorID: &cursorID,
			limit:    10,
			mock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "John Doe",
						CreatedAt: now.Add(-1 * time.Hour),
					},
					{
						ContactID: uuid.New(),
						Name:      "Jane Smith",
						CreatedAt: now.Add(-2 * time.Hour),
					},
				}
				mockRepo.On("ListContactsPaginated", ctx, userID, &now, &cursorID, int32(10)).
					Return(contacts, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:     "invalid limit",
			cursor:   &now,
			cursorID: &cursorID,
			limit:    -1,
			mock:     func() {},
			wantErr:  true,
			errMsg:   "limit must be positive",
		},
		{
			name:     "repository error",
			cursor:   &now,
			cursorID: &cursorID,
			limit:    10,
			mock: func() {
				mockRepo.On("ListContactsPaginated", ctx, userID, &now, &cursorID, int32(10)).
					Return([]types.Contact{}, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			contacts, err := service.ListContactsPaginated(ctx, userID, tt.cursor, tt.cursorID, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, contacts, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContactService_SearchContacts(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		query   string
		limit   int32
		mock    func()
		wantErr bool
		wantLen int
		errMsg  string
	}{
		{
			name:  "successful search",
			query: "John",
			limit: 10,
			mock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "John Doe",
					},
					{
						ContactID: uuid.New(),
						Name:      "Johnny Smith",
					},
				}
				mockRepo.On("SearchContacts", ctx, userID, "John", int32(10)).Return(contacts, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "invalid limit",
			query:   "John",
			limit:   -1,
			mock:    func() {},
			wantErr: true,
			errMsg:  "limit must be positive",
		},
		{
			name:  "empty result",
			query: "XYZ",
			limit: 10,
			mock: func() {
				mockRepo.On("SearchContacts", ctx, userID, "XYZ", int32(10)).Return([]types.Contact{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			contacts, err := service.SearchContacts(ctx, userID, tt.query, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, contacts, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContactService_SearchContactsByPhone(t *testing.T) {
	mockRepo, service := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		query   string
		limit   int32
		mock    func()
		wantErr bool
		wantLen int
		errMsg  string
	}{
		{
			name:  "successful search with phone cleaning",
			query: "+1-555-123-4567",
			limit: 10,
			mock: func() {
				contacts := []types.Contact{
					{
						ContactID: uuid.New(),
						Name:      "John Doe",
						Phone:     utils.StringPtr("15551234567"),
					},
				}
				// Verify that cleaned phone number is passed to repository
				mockRepo.On("SearchContactsByPhone", ctx, userID, "15551234567", int32(10)).Return(contacts, nil)
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "invalid limit",
			query:   "15551234567",
			limit:   -1,
			mock:    func() {},
			wantErr: true,
			errMsg:  "limit must be positive",
		},
		{
			name:  "repository error",
			query: "15551234567",
			limit: 10,
			mock: func() {
				mockRepo.On("SearchContactsByPhone", ctx, userID, "15551234567", int32(10)).
					Return([]types.Contact{}, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.mock()

			contacts, err := service.SearchContactsByPhone(ctx, userID, tt.query, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, contacts, tt.wantLen)
			mockRepo.AssertExpectations(t)
		})
	}
}
