package errors

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ApiError represents a structured API error response
type ApiError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	RequestID  string      `json:"request_id"`
	Timestamp  string      `json:"timestamp"`
	StatusCode int         `json:"-"`
}

// Error implements the error interface
func (e *ApiError) Error() string {
	return e.Message
}

// NewApiError creates a new API error with the given parameters
func NewApiError(code, message string, statusCode int, details interface{}) *ApiError {
	return &ApiError{
		Code:       code,
		Message:    message,
		Details:    details,
		RequestID:  uuid.New().String(),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		StatusCode: statusCode,
	}
}

// Authentication errors
func AuthenticationFailed(message string) *ApiError {
	if message == "" {
		message = "Authentication failed"
	}
	return NewApiError("AUTHENTICATION_FAILED", message, http.StatusUnauthorized, nil)
}

func TokenExpired() *ApiError {
	return NewApiError("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized, nil)
}

func TokenRevoked() *ApiError {
	return NewApiError("TOKEN_REVOKED", "Token has been revoked", http.StatusUnauthorized, nil)
}

// Authorization errors
func AuthorizationFailed(resource, action string) *ApiError {
	return NewApiError("AUTHORIZATION_FAILED", "Not authorized to perform this action", http.StatusForbidden, map[string]string{
		"resource": resource,
		"action":   action,
	})
}

// Resource errors
func NotFound(resource string, id interface{}) *ApiError {
	return NewApiError("RESOURCE_NOT_FOUND", "Resource not found", http.StatusNotFound, map[string]interface{}{
		"resource": resource,
		"id":       id,
	})
}

func DuplicateResource(resource, field string) *ApiError {
	return NewApiError("DUPLICATE_RESOURCE", "Resource already exists", http.StatusConflict, map[string]string{
		"resource": resource,
		"field":    field,
	})
}

// Validation errors
func ValidationFailed(errors interface{}) *ApiError {
	return NewApiError("VALIDATION_FAILED", "Validation failed. Please check your input.", http.StatusUnprocessableEntity, map[string]interface{}{
		"validation_errors": errors,
	})
}

func ParameterMissing(parameter string) *ApiError {
	return NewApiError("PARAMETER_MISSING", "Required parameter is missing", http.StatusBadRequest, map[string]string{
		"parameter": parameter,
	})
}

// Business logic errors
func InvalidStateTransition(from, to string) *ApiError {
	return NewApiError("INVALID_STATE_TRANSITION", "Invalid state transition", http.StatusUnprocessableEntity, map[string]string{
		"from": from,
		"to":   to,
	})
}

func EditTimeExpired() *ApiError {
	return NewApiError("EDIT_TIME_EXPIRED", "Edit time limit has expired", http.StatusForbidden, nil)
}

// System errors
func InternalError() *ApiError {
	return NewApiError("INTERNAL_ERROR", "An unexpected error occurred", http.StatusInternalServerError, nil)
}

func RateLimitExceeded() *ApiError {
	return NewApiError("RATE_LIMIT_EXCEEDED", "Too many requests", http.StatusTooManyRequests, nil)
}

// ErrorHandler is a custom error handler for Echo
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// Check if it's our ApiError
	if apiErr, ok := err.(*ApiError); ok {
		// Use request ID from context if available
		if reqID := c.Request().Header.Get(echo.HeaderXRequestID); reqID != "" {
			apiErr.RequestID = reqID
		}
		c.JSON(apiErr.StatusCode, map[string]interface{}{
			"error": apiErr,
		})
		return
	}

	// Handle Echo's HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		apiErr := NewApiError("HTTP_ERROR", he.Error(), he.Code, nil)
		if reqID := c.Request().Header.Get(echo.HeaderXRequestID); reqID != "" {
			apiErr.RequestID = reqID
		}
		c.JSON(he.Code, map[string]interface{}{
			"error": apiErr,
		})
		return
	}

	// Handle unexpected errors
	apiErr := InternalError()
	if reqID := c.Request().Header.Get(echo.HeaderXRequestID); reqID != "" {
		apiErr.RequestID = reqID
	}
	c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error": apiErr,
	})
}
