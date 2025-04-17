package types

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	MinQueryLength     = 1
	MaxQueryLength     = 100
	MaxSearchLimit     = 50
	DefaultSearchLimit = 10
)

type SearchParams struct {
	Query string
	Limit int32
}

func ParseAndValidateSearchParams(query url.Values) (SearchParams, error) {
	searchQuery := strings.TrimSpace(query.Get("q"))

	// Parse and validate limit
	limit := int32(DefaultSearchLimit)
	if limitStr := query.Get("limit"); limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return SearchParams{}, errors.New("limit: invalid format")
		}
		// Cap the limit
		if l > MaxSearchLimit {
			l = MaxSearchLimit
		}
		limit = int32(l)
	}

	return SearchParams{Query: searchQuery, Limit: limit}, validation.Errors{
		"query": validation.Validate(searchQuery, validation.Length(MinQueryLength, MaxQueryLength)),
		"limit": validation.Validate(limit, validation.Min(1)),
	}.Filter()
}
