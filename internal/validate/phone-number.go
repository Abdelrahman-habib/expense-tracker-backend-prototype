package validate

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	// ErrPhoneNumber is the error that returns in case of an invalid PhoneNumber.
	ErrPhoneNumber = validation.NewError("validation_is_PhoneNumber", "invalid phone number format")
	rePhoneNumber  = regexp.MustCompile(`[+]?[\d\s-()]+$`)
	// PhoneNumber validates if a string is a valid PhoneNumber
	PhoneNumber = validation.NewStringRuleWithError(isPhoneNumber, ErrPhoneNumber)
)

func isPhoneNumber(value string) bool {
	return rePhoneNumber.MatchString(value)
}
