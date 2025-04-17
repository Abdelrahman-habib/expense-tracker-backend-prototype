package validate

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	// ErrZipcode is the error that returns in case of an invalid zipcode.
	ErrZipCode = validation.NewError("validation_is_zipcode", "invalid zip code format")
	reZipcode  = regexp.MustCompile(`^[A-Za-z0-9\s\-]{3,10}$`)
	// Zipcode validates if a string is a valid Zipcode
	Zipcode = validation.NewStringRuleWithError(isZipcode, ErrZipCode)
)

func isZipcode(value string) bool {
	return reZipcode.MatchString(value)
}
