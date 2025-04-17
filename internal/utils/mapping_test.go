package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestStringPtrToString(t *testing.T) {
	tests := []struct {
		name string
		s    *string
		want string
	}{
		{
			name: "nil pointer",
			s:    nil,
			want: "",
		},
		{
			name: "empty string",
			s:    stringPtr(""),
			want: "",
		},
		{
			name: "non-empty string",
			s:    stringPtr("test"),
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringPtrToString(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgtextToStringPtr(t *testing.T) {
	tests := []struct {
		name string
		text pgtype.Text
		want *string
	}{
		{
			name: "invalid text",
			text: pgtype.Text{Valid: false},
			want: nil,
		},
		{
			name: "empty string",
			text: pgtype.Text{String: "", Valid: true},
			want: stringPtr(""),
		},
		{
			name: "non-empty string",
			text: pgtype.Text{String: "test", Valid: true},
			want: stringPtr("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PgtextToStringPtr(tt.text)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestToNullableText(t *testing.T) {
	tests := []struct {
		name string
		s    *string
		want pgtype.Text
	}{
		{
			name: "nil string",
			s:    nil,
			want: pgtype.Text{Valid: false},
		},
		{
			name: "empty string",
			s:    stringPtr(""),
			want: pgtype.Text{String: "", Valid: true},
		},
		{
			name: "non-empty string",
			s:    stringPtr("test"),
			want: pgtype.Text{String: "test", Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToNullableText(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToNullableTimestamp(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name string
		t    *time.Time
		want pgtype.Timestamp
	}{
		{
			name: "nil time",
			t:    nil,
			want: pgtype.Timestamp{Valid: false},
		},
		{
			name: "valid time",
			t:    &now,
			want: pgtype.Timestamp{Time: now, Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToNullableTimestamp(tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToNullableNumeric(t *testing.T) {
	tests := []struct {
		name string
		f    *float64
		want bool
	}{
		{
			name: "nil float",
			f:    nil,
			want: false,
		},
		{
			name: "zero",
			f:    float64Ptr(0),
			want: true,
		},
		{
			name: "positive float",
			f:    float64Ptr(123.45),
			want: true,
		},
		{
			name: "negative float",
			f:    float64Ptr(-123.45),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToNullableNumeric(tt.f)
			assert.Equal(t, tt.want, got.Valid)
			if tt.f != nil {
				val, err := got.Float64Value()
				assert.NoError(t, err)
				assert.InDelta(t, *tt.f, val.Float64, 0.000001)
			}
		})
	}
}

func TestGetTimePtr(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name string
		t    pgtype.Timestamp
		want *time.Time
	}{
		{
			name: "invalid timestamp",
			t:    pgtype.Timestamp{Valid: false},
			want: nil,
		},
		{
			name: "valid timestamp",
			t:    pgtype.Timestamp{Time: now, Valid: true},
			want: &now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTimePtr(tt.t)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestGetFloat64Ptr(t *testing.T) {
	tests := []struct {
		name string
		n    pgtype.Numeric
		want *float64
	}{
		{
			name: "invalid numeric",
			n:    pgtype.Numeric{Valid: false},
			want: nil,
		},
	}

	// Add test cases with valid numeric values
	zero := float64(0)
	pos := float64(123.45)
	neg := float64(-123.45)

	// Create valid numeric values
	zeroNum := pgtype.Numeric{}
	posNum := pgtype.Numeric{}
	negNum := pgtype.Numeric{}

	assert.NoError(t, zeroNum.Scan("0"))
	assert.NoError(t, posNum.Scan("123.45"))
	assert.NoError(t, negNum.Scan("-123.45"))

	tests = append(tests, []struct {
		name string
		n    pgtype.Numeric
		want *float64
	}{
		{
			name: "zero value",
			n:    zeroNum,
			want: &zero,
		},
		{
			name: "positive value",
			n:    posNum,
			want: &pos,
		},
		{
			name: "negative value",
			n:    negNum,
			want: &neg,
		},
	}...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFloat64Ptr(tt.n)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.InDelta(t, *tt.want, *got, 0.000001)
			}
		})
	}
}

func TestToNullableUUID(t *testing.T) {
	id := uuid.New()
	got := ToNullableUUID(id)
	assert.True(t, got.Valid)
	// Convert both to strings for comparison to ensure byte order is correct
	assert.Equal(t, id.String(), uuid.UUID(got.Bytes).String())
}

func TestUUIDToNullableUUID(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name string
		id   *uuid.UUID
		want pgtype.UUID
	}{
		{
			name: "nil UUID",
			id:   nil,
			want: pgtype.UUID{Valid: false},
		},
		{
			name: "valid UUID",
			id:   &id,
			want: pgtype.UUID{Bytes: id, Valid: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UUIDToNullableUUID(tt.id)
			assert.Equal(t, tt.want.Valid, got.Valid)
			if tt.want.Valid {
				// Convert both to strings for comparison to ensure byte order is correct
				assert.Equal(t, tt.id.String(), uuid.UUID(got.Bytes).String())
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create pgtype.Numeric from float64
func mustScanNumeric(t *testing.T, f float64) pgtype.Numeric {
	t.Helper()
	var n pgtype.Numeric
	err := n.Scan(f)
	assert.NoError(t, err)
	return n
}
