package types

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ClerkWebhook struct {
	Data struct {
		Deleted        bool `json:"deleted,omitempty"`
		EmailAddresses []struct {
			EmailAddress string `json:"email_address"`
		} `json:"email_addresses,omitempty"`
		ID       string  `json:"id"`
		ImageURL *string `json:"image_url,omitempty"`
		Username *string `json:"username,omitempty"`
	} `json:"data"`
	Type string `json:"type"`
}

// Bind implements render.Binder interface for ClerkWebhook
func (cw *ClerkWebhook) Bind(r *http.Request) error {
	return validation.Errors{
		"email_addresses": validation.Validate(cw.Data.EmailAddresses,
			validation.When(cw.Type != "user.deleted", validation.Required, validation.Length(1, 0), validation.Each(validation.By(func(value interface{}) error {
				email := value.(struct {
					EmailAddress string `json:"email_address"`
				})
				return validation.Validate(email.EmailAddress,
					validation.Required.Error("email address is required"),
					is.Email.Error("must be a valid email address"),
				)
			})))),
		"type":     validation.Validate(cw.Type, validation.Required),
		"username": validation.Validate(cw.Data.Username, validation.When(cw.Type != "user.deleted", validation.Required)),
		"id":       validation.Validate(cw.Data.ID, validation.Required),
	}.Filter()
}
