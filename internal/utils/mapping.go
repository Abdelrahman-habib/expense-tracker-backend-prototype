package utils

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func StringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func PgtextToStringPtr(text pgtype.Text) *string {
	if !text.Valid {
		return nil
	}
	return &text.String
}

func ToNullableText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func ToNullableTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

func ToNullableNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	var n pgtype.Numeric
	// Convert float64 to string to avoid precision issues
	err := n.Scan(fmt.Sprintf("%.10f", *f))
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}

func GetTimePtr(t pgtype.Timestamp) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func GetFloat64Ptr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	val, err := n.Float64Value()
	if err != nil || val.Float64 > math.MaxFloat64 || val.Float64 < -math.MaxFloat64 {
		return nil
	}
	return &val.Float64
}

func ToNullableUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

func UUIDToNullableUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: *id,
		Valid: true,
	}
}

// Common test helper functions
func StringPtr(s string) *string {
	return &s
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

// Helper function to create pgtype.Numeric from float64
func MustScanNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	err := n.Scan(f)
	if err != nil {
		panic(err)
	}
	return n
}

// Helper function to create a slice of UUIDs
func CreateUUIDSlice(count int) []uuid.UUID {
	result := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		result[i] = uuid.New()
	}
	return result
}

func GetUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	result := uuid.UUID(id.Bytes)
	return &result
}
