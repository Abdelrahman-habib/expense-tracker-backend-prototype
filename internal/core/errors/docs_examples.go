package errors

// error types used in the docs only

// ValidationError represents a validation error response
type errInvalidRequest struct {
	Type      ErrorType `json:"type" example:"VALIDATION_ERROR"`
	Message   string    `json:"message" example:"Invalid request parameters"`
	Code      int       `json:"code" example:"400"`
	ErrorText string    `json:"error" example:"invalid request format"`
}

// AuthorizationError represents an authorization error response
type errAuthorization struct {
	Type      ErrorType `json:"type" example:"AUTHORIZATION_ERROR"`
	Message   string    `json:"message" example:"Authorization failed"`
	Code      int       `json:"code" example:"401"`
	ErrorText string    `json:"error" example:"invalid or missing token"`
}

// NotFoundError represents a not found error response
type errNotFound struct {
	Type      ErrorType `json:"type" example:"NOT_FOUND"`
	Message   string    `json:"message" example:"Resource not found"`
	Code      int       `json:"code" example:"404"`
	ErrorText string    `json:"error" example:"Resource not found"`
}

// InternalError represents an internal server error response
type errInternal struct {
	Type      ErrorType `json:"type" example:"INTERNAL_ERROR"`
	Message   string    `json:"message" example:"Internal server error"`
	Code      int       `json:"code" example:"500"`
	ErrorText string    `json:"error" example:"database connection failed"`
}

// DatabaseError represents a database error response
type errDatabase struct {
	Type      ErrorType `json:"type" example:"DATABASE_ERROR"`
	Message   string    `json:"message" example:"Database error occurred"`
	Code      int       `json:"code" example:"500"`
	ErrorText string    `json:"error" example:"failed to execute database query"`
}

// ExternalServiceError represents an external service error response
type errExternalService struct {
	Type      ErrorType `json:"type" example:"EXTERNAL_SERVICE"`
	Message   string    `json:"message" example:"External service error"`
	Code      int       `json:"code" example:"502"`
	ErrorText string    `json:"error" example:"external API request failed"`
}

// RenderError represents a render error response
type errRender struct {
	Type      ErrorType `json:"type" example:"RENDER_ERROR"`
	Message   string    `json:"message" example:"Error rendering response"`
	Code      int       `json:"code" example:"422"`
	ErrorText string    `json:"error" example:"failed to marshal response"`
}

// ForbiddenError represents a forbidden error response
type errForbidden struct {
	Type      ErrorType `json:"type" example:"FORBIDDEN"`
	Message   string    `json:"message" example:"Access forbidden"`
	Code      int       `json:"code" example:"403"`
	ErrorText string    `json:"error" example:"insufficient permissions"`
}

// ConflictError represents a conflict error response
type errConflict struct {
	Type      ErrorType `json:"type" example:"CONFLICT"`
	Message   string    `json:"message" example:"Resource conflict"`
	Code      int       `json:"code" example:"409"`
	ErrorText string    `json:"error" example:"resource already exists"`
}

// RateLimitError represents a rate limit error response
type errRateLimit struct {
	Type      ErrorType `json:"type" example:"RATE_LIMIT"`
	Message   string    `json:"message" example:"Too many requests"`
	Code      int       `json:"code" example:"429"`
	ErrorText string    `json:"error" example:"rate limit exceeded"`
}

// UnsupportedError represents an unsupported operation error response
type errUnsupported struct {
	Type      ErrorType `json:"type" example:"UNSUPPORTED_ERROR"`
	Message   string    `json:"message" example:"Unsupported operation"`
	Code      int       `json:"code" example:"501"`
	ErrorText string    `json:"error" example:"feature not implemented"`
}
