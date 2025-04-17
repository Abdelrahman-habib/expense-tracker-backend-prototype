package validate

import (
	"fmt"
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	// ErrDuplicateFound is the error that returns when a duplicate element is found
	ErrDuplicateFound = validation.NewError(
		"validation_duplicate_found",
		"contains duplicate elements",
	)
)

// NoDuplicates returns a validation rule that checks if a slice or array contains duplicate elements.
// This rule should only be used for validating slices and arrays.
// An empty value is considered valid. Use the Required rule to make sure a value is not empty.
func NoDuplicates() DuplicateRule {
	return DuplicateRule{
		err: ErrDuplicateFound,
	}
}

// DuplicateRule is a validation rule that checks if a slice or array contains duplicate elements.
type DuplicateRule struct {
	err validation.Error
}

// Validate checks if the given value is valid or not.
func (r DuplicateRule) Validate(value interface{}) error {
	value, isNil := validation.Indirect(value)
	if isNil || validation.IsEmpty(value) {
		return nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return fmt.Errorf("cannot check duplicates on non-slice/array value of type %T", value)
	}

	length := v.Len()
	if length <= 1 {
		return nil
	}

	seen := make(map[interface{}]struct{})
	for i := 0; i < length; i++ {
		item := v.Index(i).Interface()
		if _, exists := seen[item]; exists {
			return r.err
		}
		seen[item] = struct{}{}
	}

	return nil
}

// Error sets the error message for the rule.
func (r DuplicateRule) Error(message string) DuplicateRule {
	r.err = r.err.SetMessage(message)
	return r
}

// ErrorObject sets the error struct for the rule.
func (r DuplicateRule) ErrorObject(err validation.Error) DuplicateRule {
	r.err = err
	return r
}
