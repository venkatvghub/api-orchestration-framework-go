package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeTransformation ErrorType = "transformation"
	ErrorTypeConfiguration  ErrorType = "configuration"
	ErrorTypeInternal       ErrorType = "internal"
	ErrorTypeExternal       ErrorType = "external"
)

// FrameworkError represents a structured error with context
type FrameworkError struct {
	Type       ErrorType              `json:"type"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Cause      error                  `json:"-"`
	HTTPStatus int                    `json:"http_status,omitempty"`
	Retryable  bool                   `json:"retryable"`
}

func (e *FrameworkError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *FrameworkError) Unwrap() error {
	return e.Cause
}

func (e *FrameworkError) WithContext(key string, value interface{}) *FrameworkError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func (e *FrameworkError) WithCause(cause error) *FrameworkError {
	e.Cause = cause
	return e
}

// Error constructors

func NewValidationError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeValidation,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Retryable:  false,
	}
}

func NewAuthenticationError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeAuthentication,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
		Retryable:  false,
	}
}

func NewAuthorizationError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeAuthorization,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
		Retryable:  false,
	}
}

func NewNetworkError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeNetwork,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
		Retryable:  true,
	}
}

func NewTimeoutError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeTimeout,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusRequestTimeout,
		Retryable:  true,
	}
}

func NewRateLimitError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeRateLimit,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
		Retryable:  true,
	}
}

func NewTransformationError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeTransformation,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnprocessableEntity,
		Retryable:  false,
	}
}

func NewConfigurationError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeConfiguration,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Retryable:  false,
	}
}

func NewInternalError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeInternal,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Retryable:  false,
	}
}

func NewExternalError(code, message string) *FrameworkError {
	return &FrameworkError{
		Type:       ErrorTypeExternal,
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadGateway,
		Retryable:  true,
	}
}

// Common error codes
const (
	// Validation errors
	ErrCodeInvalidInput    = "INVALID_INPUT"
	ErrCodeMissingField    = "MISSING_FIELD"
	ErrCodeInvalidFormat   = "INVALID_FORMAT"
	ErrCodeValueOutOfRange = "VALUE_OUT_OF_RANGE"

	// Authentication errors
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeExpiredToken       = "EXPIRED_TOKEN"
	ErrCodeMissingToken       = "MISSING_TOKEN"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"

	// Authorization errors
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrCodeAccessDenied            = "ACCESS_DENIED"

	// Network errors
	ErrCodeConnectionFailed = "CONNECTION_FAILED"
	ErrCodeDNSResolution    = "DNS_RESOLUTION_FAILED"
	ErrCodeNetworkTimeout   = "NETWORK_TIMEOUT"

	// Timeout errors
	ErrCodeRequestTimeout   = "REQUEST_TIMEOUT"
	ErrCodeExecutionTimeout = "EXECUTION_TIMEOUT"

	// Rate limit errors
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeQuotaExceeded     = "QUOTA_EXCEEDED"

	// Transformation errors
	ErrCodeTransformationFailed = "TRANSFORMATION_FAILED"
	ErrCodeInvalidTransformer   = "INVALID_TRANSFORMER"

	// Configuration errors
	ErrCodeInvalidConfiguration = "INVALID_CONFIGURATION"
	ErrCodeMissingConfiguration = "MISSING_CONFIGURATION"

	// Internal errors
	ErrCodeInternalFailure = "INTERNAL_FAILURE"
	ErrCodeUnexpectedError = "UNEXPECTED_ERROR"

	// External errors
	ErrCodeExternalServiceUnavailable = "EXTERNAL_SERVICE_UNAVAILABLE"
	ErrCodeExternalServiceError       = "EXTERNAL_SERVICE_ERROR"
)

// Helper functions for common errors

func InvalidInput(field, reason string) *FrameworkError {
	return NewValidationError(ErrCodeInvalidInput, fmt.Sprintf("Invalid input for field '%s'", field)).
		WithContext("field", field).
		WithContext("reason", reason)
}

func MissingField(field string) *FrameworkError {
	return NewValidationError(ErrCodeMissingField, fmt.Sprintf("Required field '%s' is missing", field)).
		WithContext("field", field)
}

func InvalidToken(reason string) *FrameworkError {
	return NewAuthenticationError(ErrCodeInvalidToken, "Invalid authentication token").
		WithContext("reason", reason)
}

func ExpiredToken() *FrameworkError {
	return NewAuthenticationError(ErrCodeExpiredToken, "Authentication token has expired")
}

func InsufficientPermissions(resource, action string) *FrameworkError {
	return NewAuthorizationError(ErrCodeInsufficientPermissions,
		fmt.Sprintf("Insufficient permissions to %s %s", action, resource)).
		WithContext("resource", resource).
		WithContext("action", action)
}

func ConnectionFailed(target string, cause error) *FrameworkError {
	return NewNetworkError(ErrCodeConnectionFailed, fmt.Sprintf("Failed to connect to %s", target)).
		WithContext("target", target).
		WithCause(cause)
}

func RequestTimeout(duration string) *FrameworkError {
	return NewTimeoutError(ErrCodeRequestTimeout, fmt.Sprintf("Request timed out after %s", duration)).
		WithContext("duration", duration)
}

func RateLimitExceeded(limit int, window string) *FrameworkError {
	return NewRateLimitError(ErrCodeRateLimitExceeded,
		fmt.Sprintf("Rate limit of %d requests per %s exceeded", limit, window)).
		WithContext("limit", limit).
		WithContext("window", window)
}

func TransformationFailed(transformer, reason string) *FrameworkError {
	return NewTransformationError(ErrCodeTransformationFailed,
		fmt.Sprintf("Transformation failed in %s", transformer)).
		WithContext("transformer", transformer).
		WithContext("reason", reason)
}

func ExternalServiceUnavailable(service string, cause error) *FrameworkError {
	return NewExternalError(ErrCodeExternalServiceUnavailable,
		fmt.Sprintf("External service %s is unavailable", service)).
		WithContext("service", service).
		WithCause(cause)
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if frameworkErr, ok := err.(*FrameworkError); ok {
		return frameworkErr.Retryable
	}
	return false
}

// GetHTTPStatus extracts HTTP status code from error
func GetHTTPStatus(err error) int {
	if frameworkErr, ok := err.(*FrameworkError); ok {
		return frameworkErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorType extracts error type from error
func GetErrorType(err error) ErrorType {
	if frameworkErr, ok := err.(*FrameworkError); ok {
		return frameworkErr.Type
	}
	return ErrorTypeInternal
}
