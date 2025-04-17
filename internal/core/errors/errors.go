package errors

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

type ErrorType string

const (
	ErrorTypeValidation      ErrorType = "VALIDATION_ERROR"
	ErrorTypeDatabase        ErrorType = "DATABASE_ERROR"
	ErrorTypeAuthorization   ErrorType = "AUTHORIZATION_ERROR"
	ErrorTypeNotFound        ErrorType = "NOT_FOUND"
	ErrorTypeInternal        ErrorType = "INTERNAL_ERROR"
	ErrorTypeExternalService ErrorType = "EXTERNAL_SERVICE"
	ErrorTypeRender          ErrorType = "RENDER_ERROR"
	ErrorTypeForbidden       ErrorType = "FORBIDDEN"
	ErrorTypeConflict        ErrorType = "CONFLICT"
	ErrorTypeRateLimit       ErrorType = "RATE_LIMIT"
	ErrorTypeUnsupported     ErrorType = "UNSUPPORTED_ERROR"
)

// ErrorResponse represents an application error
// @Description Application error response
type ErrorResponse struct {
	Type      ErrorType `json:"type"`
	Message   string    `json:"message" example:"Invalid request parameters" enums:"Invalid request parameters,Authorization failed,Resource not found,Internal server error,Database error occurred,External service error,Error rendering response,Access forbidden,Resource conflict,Too many requests,Unsupported operation"`
	Err       error     `json:"-"` // Internal error details (not exposed to client)
	Code      int       `json:"code,omitempty" example:"400" enums:"400,401,404,500,502,422,403,409,429,501"`
	ErrorText string    `json:"error,omitempty" example:"field: required"`
}

func (e *ErrorResponse) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.Code)
	return nil
}

func ErrNotFound() render.Renderer {
	return &ErrorResponse{
		Type:    ErrorTypeNotFound,
		Message: "Resource not found",
		Code:    http.StatusNotFound,
		Err:     fmt.Errorf("resource not found"),
	}
}

func ErrValidation(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeValidation,
		Message:   "Invalid request parameters",
		Err:       err,
		Code:      http.StatusBadRequest,
		ErrorText: err.Error(),
	}
}

func ErrDatabase(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeDatabase,
		Message:   "Database error occurred",
		Err:       err,
		Code:      http.StatusInternalServerError,
		ErrorText: err.Error(),
	}
}

func ErrAuthorization(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeAuthorization,
		Message:   "Authorization failed",
		Err:       err,
		Code:      http.StatusUnauthorized,
		ErrorText: err.Error(),
	}
}

func ErrInternal(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeInternal,
		Message:   "Internal server error",
		Err:       err,
		Code:      http.StatusInternalServerError,
		ErrorText: err.Error(),
	}
}

func ErrExternalService(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeExternalService,
		Message:   "External service error",
		Err:       err,
		Code:      http.StatusBadGateway,
		ErrorText: err.Error(),
	}
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeValidation,
		Message:   "Invalid request",
		Err:       err,
		Code:      http.StatusBadRequest,
		ErrorText: err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeRender,
		Message:   "Error rendering response",
		Err:       err,
		Code:      http.StatusUnprocessableEntity,
		ErrorText: err.Error(),
	}
}

func ErrForbidden(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeForbidden,
		Message:   "Access forbidden",
		Err:       err,
		Code:      http.StatusForbidden,
		ErrorText: err.Error(),
	}
}

func ErrConflict(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeConflict,
		Message:   "Resource conflict",
		Err:       err,
		Code:      http.StatusConflict,
		ErrorText: err.Error(),
	}
}

func ErrRateLimit(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeRateLimit,
		Message:   "Too many requests",
		Err:       err,
		Code:      http.StatusTooManyRequests,
		ErrorText: err.Error(),
	}
}

func ErrUnsupported(err error) render.Renderer {
	return &ErrorResponse{
		Type:      ErrorTypeUnsupported,
		Message:   "Unsupported operation",
		Err:       err,
		Code:      http.StatusNotImplemented,
		ErrorText: err.Error(),
	}
}

func IsErrorType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*ErrorResponse); ok {
		return appErr.Type == errorType
	}
	return false
}
