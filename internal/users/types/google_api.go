package types

import (
	"encoding/json"
)

type GoogleContact struct {
	Name            string   `json:"name"`
	PhoneNumbers    []string `json:"phone_numbers"`
	EmailAddresses  []string `json:"email_addresses"`
	StreetAddress   string   `json:"street_address"`
	ExtendedAddress string   `json:"extended_address"`
	Country         string   `json:"country"`
	CountryCode     string   `json:"country_code"`
	City            string   `json:"city"`
	Region          string   `json:"region"`
	PostalCode      string   `json:"postal_code"`
}

type GoogleOauthToken struct {
	ExternalAccountID string          `json:"external_account_id"`
	Object            string          `json:"object"`
	Token             string          `json:"token"`
	PublicMetadata    json.RawMessage `json:"public_metadata"`
	Label             *string         `json:"label"`
	Scopes            []string        `json:"scopes,omitempty"`
	TokenSecret       *string         `json:"token_secret,omitempty"`
}

type PaginatedGoogleContacts struct {
	Contacts      []GoogleContact `json:"contacts"`
	NextPageToken string          `json:"next_page_token"`
	TotalSize     int             `json:"total_size"`
}
