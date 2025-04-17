package types

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

type Cursor struct {
	Timestamp time.Time
	ID        uuid.UUID
}

type PaginationParams struct {
	Cursor *Cursor
	Limit  int32
}

// ParsePaginationParams parses and validates pagination parameters from URL query
func ParsePaginationParams(query url.Values) (PaginationParams, error) {
	params := PaginationParams{
		Limit: DefaultLimit,
	}

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return params, fmt.Errorf("invalid limit format")
		}
		// cap the limit
		if l > MaxLimit {
			l = MaxLimit
		}
		params.Limit = int32(l)
	}

	// Parse cursor if provided
	if nextToken := query.Get("next_token"); nextToken != "" {
		cursor, err := DecodeCursor(nextToken)
		if err != nil {
			return params, err
		}
		params.Cursor = cursor
	}

	return params, params.Validate()
}

// Validate implements validation for pagination parameters
func (p *PaginationParams) Validate() error {
	return validation.Errors{
		"limit": validation.Validate(p.Limit,
			validation.Required.Error("must be no less than 1"), // we have this because validation package treats 0 as nil so the min and max won't work for limit = 0
			validation.Min(1),
			validation.Max(MaxLimit),
		),
		"cursor": validation.Validate(p.Cursor,
			validation.When(p.Cursor != nil, validation.By(func(value interface{}) error {
				return value.(*Cursor).Validate()
			})),
		),
	}.Filter()
}

// Validate implements validation for cursor parameters
func (c *Cursor) Validate() error {
	return validation.Errors{
		"timestamp": validation.Validate(c.Timestamp,
			validation.Required,
			validation.By(func(value interface{}) error {
				t := value.(time.Time)
				if t.IsZero() {
					return fmt.Errorf("timestamp cannot be zero")
				}
				// Compare with UTC time
				if t.After(time.Now().UTC()) {
					return fmt.Errorf("timestamp cannot be in the future")
				}
				return nil
			}),
		),
		"id": validation.Validate(c.ID,
			validation.Required,
			validation.By(func(value interface{}) error {
				id := value.(uuid.UUID)
				if id == uuid.Nil {
					return fmt.Errorf("ID cannot be nil")
				}
				return nil
			}),
		),
	}.Filter()
}

// EncodeCursor creates a cursor token from timestamp and ID
func EncodeCursor(timestamp time.Time, id uuid.UUID) string {
	cursor := &Cursor{
		Timestamp: timestamp.UTC(), // Ensure UTC
		ID:        id,
	}

	// Validate cursor before encoding
	if err := cursor.Validate(); err != nil {
		return ""
	}

	raw := fmt.Sprintf("%d:%s", timestamp.UTC().UnixNano(), id.String())
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor parses a cursor token into timestamp and ID
func DecodeCursor(token string) (*Cursor, error) {
	if token == "" {
		return nil, nil
	}

	// Decode base64
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token format")
	}

	// Split into parts
	parts := strings.Split(string(raw), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Parse timestamp
	var nanos int64
	if _, err := fmt.Sscanf(parts[0], "%d", &nanos); err != nil {
		return nil, fmt.Errorf("invalid token value")
	}
	timestamp := time.Unix(0, nanos).UTC() // Ensure UTC

	// Parse UUID
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token value")
	}

	cursor := &Cursor{
		Timestamp: timestamp,
		ID:        id,
	}

	// Validate the cursor after decoding
	if err := cursor.Validate(); err != nil {
		return nil, err
	}

	return cursor, nil
}
