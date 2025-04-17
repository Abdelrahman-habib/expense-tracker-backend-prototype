package requestcontext

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// RequestContextKey is a custom type for context keys to avoid collisions
type RequestContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey RequestContextKey = "requestID"
	// StartTimeKey is the context key for request start time
	StartTimeKey RequestContextKey = "startTime"

	// UserIDKey is the context key for db User ID
	UserIDKey RequestContextKey = "userID"
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("missing user ID from context")
	}
	return userID, nil
}

func GetRequestIDFromContext(ctx context.Context) (uuid.UUID, error) {
	requestID, ok := ctx.Value(RequestIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("missing request id from context")
	}
	return requestID, nil
}

func GetStartTimeFromContext(ctx context.Context) (time.Time, error) {
	startTime, ok := ctx.Value(StartTimeKey).(time.Time)
	if !ok {
		return time.Time{}, errors.New("missing start time from context")
	}
	return startTime, nil
}
