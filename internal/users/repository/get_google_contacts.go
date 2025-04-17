package repository

import (
	"context"
	"fmt"

	"github.com/Abdelrahman-habib/expense-tracker/internal/users/types"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

func convertToGoogleContact(contact *people.Person) types.GoogleContact {
	googleContact := types.GoogleContact{}

	// Handle Name
	if len(contact.Names) > 0 {
		googleContact.Name = contact.Names[0].DisplayName
	}

	// Handle Phone Numbers
	for _, num := range contact.PhoneNumbers {
		if num.Value != "" {
			googleContact.PhoneNumbers = append(googleContact.PhoneNumbers, num.Value)
		}
	}

	// Handle Email Addresses
	for _, email := range contact.EmailAddresses {
		if email.Value != "" {
			googleContact.EmailAddresses = append(googleContact.EmailAddresses, email.Value)
		}
	}

	// Handle Address (using the first address if available)
	if len(contact.Addresses) > 0 {
		addr := contact.Addresses[0]
		googleContact.StreetAddress = addr.StreetAddress
		googleContact.ExtendedAddress = addr.ExtendedAddress
		googleContact.Country = addr.Country
		googleContact.CountryCode = addr.CountryCode
		googleContact.City = addr.City
		googleContact.Region = addr.Region
		googleContact.PostalCode = addr.PostalCode
	}

	return googleContact
}

func (r *usersRepository) GetGoogleContacts(ctx context.Context, token string, pageToken string) (*types.PaginatedGoogleContacts, error) {
	oauth2Token := &oauth2.Token{AccessToken: token}
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(oauth2Token))

	peopleService, err := people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create people service: %w", err)
	}

	fields := "names,phoneNumbers,emailAddresses,addresses"

	// Execute single page request
	call := peopleService.People.Connections.List("people/me").
		PageSize(100).        // Adjust page size as needed
		PersonFields(fields). // Requested data fields
		PageToken(pageToken)  // Pagination token

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contacts: %w", err)
	}

	var contacts []types.GoogleContact
	for _, contact := range resp.Connections {
		googleContact := convertToGoogleContact(contact)
		contacts = append(contacts, googleContact)
	}

	return &types.PaginatedGoogleContacts{
		Contacts:      contacts,
		NextPageToken: resp.NextPageToken,
		TotalSize:     int(resp.TotalPeople),
	}, nil
}
