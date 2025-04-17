package repository

import (
	"github.com/google/uuid"

	"github.com/Abdelrahman-habib/expense-tracker/internal/contacts/types"
	"github.com/Abdelrahman-habib/expense-tracker/internal/db"
	"github.com/Abdelrahman-habib/expense-tracker/internal/utils"
)

// toContact converts a db.Contact to domain types.Contact
func toContact(c db.Contact) types.Contact {
	return types.Contact{
		ContactID:     c.ContactID,
		UserID:        c.UserID,
		Name:          c.Name,
		Phone:         utils.PgtextToStringPtr(c.Phone),
		Email:         utils.PgtextToStringPtr(c.Email),
		AddressLine1:  utils.PgtextToStringPtr(c.AddressLine1),
		AddressLine2:  utils.PgtextToStringPtr(c.AddressLine2),
		Country:       utils.PgtextToStringPtr(c.Country),
		City:          utils.PgtextToStringPtr(c.City),
		StateProvince: utils.PgtextToStringPtr(c.StateProvince),
		ZipPostalCode: utils.PgtextToStringPtr(c.ZipPostalCode),
		Tags:          c.Tags,
		CreatedAt:     c.CreatedAt.Time,
		UpdatedAt:     c.UpdatedAt.Time,
	}
}

// toContacts converts a slice of db.Contact to a slice of domain types.Contact
func toContacts(contacts []db.Contact) []types.Contact {
	result := make([]types.Contact, len(contacts))
	for i, c := range contacts {
		result[i] = toContact(c)
	}
	return result
}

// createContactParamsFromPayload converts ContactCreatePayload to db.CreateContactParams
func createContactParamsFromPayload(payload types.ContactCreatePayload, userID uuid.UUID) db.CreateContactParams {
	return db.CreateContactParams{
		UserID:        userID,
		Name:          payload.Name,
		Phone:         utils.ToNullableText(payload.Phone),
		Email:         utils.ToNullableText(payload.Email),
		AddressLine1:  utils.ToNullableText(payload.AddressLine1),
		AddressLine2:  utils.ToNullableText(payload.AddressLine2),
		Country:       utils.ToNullableText(payload.Country),
		City:          utils.ToNullableText(payload.City),
		StateProvince: utils.ToNullableText(payload.StateProvince),
		ZipPostalCode: utils.ToNullableText(payload.ZipPostalCode),
		Tags:          payload.Tags,
	}
}

// updateContactParamsFromPayload converts ContactUpdatePayload to db.UpdateContactParams
func updateContactParamsFromPayload(payload types.ContactUpdatePayload, userID uuid.UUID) db.UpdateContactParams {
	return db.UpdateContactParams{
		ContactID:     payload.ContactID,
		UserID:        userID,
		Name:          utils.ToNullableText(&payload.Name),
		Phone:         utils.ToNullableText(payload.Phone),
		Email:         utils.ToNullableText(payload.Email),
		AddressLine1:  utils.ToNullableText(payload.AddressLine1),
		AddressLine2:  utils.ToNullableText(payload.AddressLine2),
		Country:       utils.ToNullableText(payload.Country),
		City:          utils.ToNullableText(payload.City),
		StateProvince: utils.ToNullableText(payload.StateProvince),
		ZipPostalCode: utils.ToNullableText(payload.ZipPostalCode),
		Tags:          payload.Tags,
	}
}
