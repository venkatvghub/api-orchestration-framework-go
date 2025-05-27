package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestFrameworkError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *FrameworkError
		expected string
	}{
		{
			name: "error with details",
			err: &FrameworkError{
				Code:    "TEST_CODE",
				Message: "test message",
				Details: "additional details",
			},
			expected: "TEST_CODE: test message - additional details",
		},
		{
			name: "error without details",
			err: &FrameworkError{
				Code:    "TEST_CODE",
				Message: "test message",
			},
			expected: "TEST_CODE: test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("FrameworkError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFrameworkError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &FrameworkError{
		Code:    "TEST_CODE",
		Message: "test message",
		Cause:   cause,
	}

	if got := err.Unwrap(); got != cause {
		t.Errorf("FrameworkError.Unwrap() = %v, want %v", got, cause)
	}
}

func TestFrameworkError_WithContext(t *testing.T) {
	err := &FrameworkError{
		Code:    "TEST_CODE",
		Message: "test message",
	}

	result := err.WithContext("key", "value")

	if result != err {
		t.Error("WithContext should return the same error instance")
	}

	if err.Context == nil {
		t.Error("Context should be initialized")
	}

	if err.Context["key"] != "value" {
		t.Errorf("Context['key'] = %v, want 'value'", err.Context["key"])
	}
}

func TestFrameworkError_WithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := &FrameworkError{
		Code:    "TEST_CODE",
		Message: "test message",
	}

	result := err.WithCause(cause)

	if result != err {
		t.Error("WithCause should return the same error instance")
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("VALIDATION_CODE", "validation message")

	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeValidation)
	}
	if err.Code != "VALIDATION_CODE" {
		t.Errorf("Code = %v, want 'VALIDATION_CODE'", err.Code)
	}
	if err.Message != "validation message" {
		t.Errorf("Message = %v, want 'validation message'", err.Message)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusBadRequest)
	}
	if err.Retryable {
		t.Error("Retryable should be false for validation errors")
	}
}

func TestNewAuthenticationError(t *testing.T) {
	err := NewAuthenticationError("AUTH_CODE", "auth message")

	if err.Type != ErrorTypeAuthentication {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuthentication)
	}
	if err.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusUnauthorized)
	}
	if err.Retryable {
		t.Error("Retryable should be false for authentication errors")
	}
}

func TestNewAuthorizationError(t *testing.T) {
	err := NewAuthorizationError("AUTHZ_CODE", "authz message")

	if err.Type != ErrorTypeAuthorization {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuthorization)
	}
	if err.HTTPStatus != http.StatusForbidden {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusForbidden)
	}
	if err.Retryable {
		t.Error("Retryable should be false for authorization errors")
	}
}

func TestNewNetworkError(t *testing.T) {
	err := NewNetworkError("NETWORK_CODE", "network message")

	if err.Type != ErrorTypeNetwork {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeNetwork)
	}
	if err.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusServiceUnavailable)
	}
	if !err.Retryable {
		t.Error("Retryable should be true for network errors")
	}
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("TIMEOUT_CODE", "timeout message")

	if err.Type != ErrorTypeTimeout {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeTimeout)
	}
	if err.HTTPStatus != http.StatusRequestTimeout {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusRequestTimeout)
	}
	if !err.Retryable {
		t.Error("Retryable should be true for timeout errors")
	}
}

func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError("RATE_LIMIT_CODE", "rate limit message")

	if err.Type != ErrorTypeRateLimit {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeRateLimit)
	}
	if err.HTTPStatus != http.StatusTooManyRequests {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusTooManyRequests)
	}
	if !err.Retryable {
		t.Error("Retryable should be true for rate limit errors")
	}
}

func TestNewTransformationError(t *testing.T) {
	err := NewTransformationError("TRANSFORM_CODE", "transform message")

	if err.Type != ErrorTypeTransformation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeTransformation)
	}
	if err.HTTPStatus != http.StatusUnprocessableEntity {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusUnprocessableEntity)
	}
	if err.Retryable {
		t.Error("Retryable should be false for transformation errors")
	}
}

func TestNewConfigurationError(t *testing.T) {
	err := NewConfigurationError("CONFIG_CODE", "config message")

	if err.Type != ErrorTypeConfiguration {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeConfiguration)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusInternalServerError)
	}
	if err.Retryable {
		t.Error("Retryable should be false for configuration errors")
	}
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("INTERNAL_CODE", "internal message")

	if err.Type != ErrorTypeInternal {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeInternal)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusInternalServerError)
	}
	if err.Retryable {
		t.Error("Retryable should be false for internal errors")
	}
}

func TestNewExternalError(t *testing.T) {
	err := NewExternalError("EXTERNAL_CODE", "external message")

	if err.Type != ErrorTypeExternal {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeExternal)
	}
	if err.HTTPStatus != http.StatusBadGateway {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusBadGateway)
	}
	if !err.Retryable {
		t.Error("Retryable should be true for external errors")
	}
}

func TestInvalidInput(t *testing.T) {
	err := InvalidInput("username", "too short")

	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeValidation)
	}
	if err.Code != ErrCodeInvalidInput {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeInvalidInput)
	}
	if err.Context["field"] != "username" {
		t.Errorf("Context['field'] = %v, want 'username'", err.Context["field"])
	}
	if err.Context["reason"] != "too short" {
		t.Errorf("Context['reason'] = %v, want 'too short'", err.Context["reason"])
	}
}

func TestMissingField(t *testing.T) {
	err := MissingField("email")

	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeValidation)
	}
	if err.Code != ErrCodeMissingField {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeMissingField)
	}
	if err.Context["field"] != "email" {
		t.Errorf("Context['field'] = %v, want 'email'", err.Context["field"])
	}
}

func TestInvalidToken(t *testing.T) {
	err := InvalidToken("malformed")

	if err.Type != ErrorTypeAuthentication {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuthentication)
	}
	if err.Code != ErrCodeInvalidToken {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeInvalidToken)
	}
	if err.Context["reason"] != "malformed" {
		t.Errorf("Context['reason'] = %v, want 'malformed'", err.Context["reason"])
	}
}

func TestExpiredToken(t *testing.T) {
	err := ExpiredToken()

	if err.Type != ErrorTypeAuthentication {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuthentication)
	}
	if err.Code != ErrCodeExpiredToken {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeExpiredToken)
	}
}

func TestInsufficientPermissions(t *testing.T) {
	err := InsufficientPermissions("user", "delete")

	if err.Type != ErrorTypeAuthorization {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuthorization)
	}
	if err.Code != ErrCodeInsufficientPermissions {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeInsufficientPermissions)
	}
	if err.Context["resource"] != "user" {
		t.Errorf("Context['resource'] = %v, want 'user'", err.Context["resource"])
	}
	if err.Context["action"] != "delete" {
		t.Errorf("Context['action'] = %v, want 'delete'", err.Context["action"])
	}
}

func TestConnectionFailed(t *testing.T) {
	cause := errors.New("connection refused")
	err := ConnectionFailed("api.example.com", cause)

	if err.Type != ErrorTypeNetwork {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeNetwork)
	}
	if err.Code != ErrCodeConnectionFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeConnectionFailed)
	}
	if err.Context["target"] != "api.example.com" {
		t.Errorf("Context['target'] = %v, want 'api.example.com'", err.Context["target"])
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestRequestTimeout(t *testing.T) {
	err := RequestTimeout("30s")

	if err.Type != ErrorTypeTimeout {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeTimeout)
	}
	if err.Code != ErrCodeRequestTimeout {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeRequestTimeout)
	}
	if err.Context["duration"] != "30s" {
		t.Errorf("Context['duration'] = %v, want '30s'", err.Context["duration"])
	}
}

func TestRateLimitExceeded(t *testing.T) {
	err := RateLimitExceeded(100, "minute")

	if err.Type != ErrorTypeRateLimit {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeRateLimit)
	}
	if err.Code != ErrCodeRateLimitExceeded {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeRateLimitExceeded)
	}
	if err.Context["limit"] != 100 {
		t.Errorf("Context['limit'] = %v, want 100", err.Context["limit"])
	}
	if err.Context["window"] != "minute" {
		t.Errorf("Context['window'] = %v, want 'minute'", err.Context["window"])
	}
}

func TestTransformationFailed(t *testing.T) {
	err := TransformationFailed("json-transformer", "invalid syntax")

	if err.Type != ErrorTypeTransformation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeTransformation)
	}
	if err.Code != ErrCodeTransformationFailed {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeTransformationFailed)
	}
	if err.Context["transformer"] != "json-transformer" {
		t.Errorf("Context['transformer'] = %v, want 'json-transformer'", err.Context["transformer"])
	}
	if err.Context["reason"] != "invalid syntax" {
		t.Errorf("Context['reason'] = %v, want 'invalid syntax'", err.Context["reason"])
	}
}

func TestExternalServiceUnavailable(t *testing.T) {
	cause := errors.New("service down")
	err := ExternalServiceUnavailable("payment-service", cause)

	if err.Type != ErrorTypeExternal {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeExternal)
	}
	if err.Code != ErrCodeExternalServiceUnavailable {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeExternalServiceUnavailable)
	}
	if err.Context["service"] != "payment-service" {
		t.Errorf("Context['service'] = %v, want 'payment-service'", err.Context["service"])
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "retryable framework error",
			err:      NewNetworkError("NET_ERR", "network error"),
			expected: true,
		},
		{
			name:     "non-retryable framework error",
			err:      NewValidationError("VAL_ERR", "validation error"),
			expected: false,
		},
		{
			name:     "non-framework error",
			err:      errors.New("standard error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "framework error with status",
			err:      NewValidationError("VAL_ERR", "validation error"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "non-framework error",
			err:      errors.New("standard error"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHTTPStatus(tt.err); got != tt.expected {
				t.Errorf("GetHTTPStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "framework error with type",
			err:      NewValidationError("VAL_ERR", "validation error"),
			expected: ErrorTypeValidation,
		},
		{
			name:     "non-framework error",
			err:      errors.New("standard error"),
			expected: ErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorType(tt.err); got != tt.expected {
				t.Errorf("GetErrorType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error type constants are defined
	errorTypes := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeAuthentication,
		ErrorTypeAuthorization,
		ErrorTypeNetwork,
		ErrorTypeTimeout,
		ErrorTypeRateLimit,
		ErrorTypeTransformation,
		ErrorTypeConfiguration,
		ErrorTypeInternal,
		ErrorTypeExternal,
	}

	for _, errorType := range errorTypes {
		if string(errorType) == "" {
			t.Errorf("Error type %v is empty", errorType)
		}
	}

	// Test that all error codes are defined
	errorCodes := []string{
		ErrCodeInvalidInput,
		ErrCodeMissingField,
		ErrCodeInvalidFormat,
		ErrCodeValueOutOfRange,
		ErrCodeInvalidToken,
		ErrCodeExpiredToken,
		ErrCodeMissingToken,
		ErrCodeInvalidCredentials,
		ErrCodeInsufficientPermissions,
		ErrCodeAccessDenied,
		ErrCodeConnectionFailed,
		ErrCodeDNSResolution,
		ErrCodeNetworkTimeout,
		ErrCodeRequestTimeout,
		ErrCodeExecutionTimeout,
		ErrCodeRateLimitExceeded,
		ErrCodeQuotaExceeded,
		ErrCodeTransformationFailed,
		ErrCodeInvalidTransformer,
		ErrCodeInvalidConfiguration,
		ErrCodeMissingConfiguration,
		ErrCodeInternalFailure,
		ErrCodeUnexpectedError,
		ErrCodeExternalServiceUnavailable,
		ErrCodeExternalServiceError,
	}

	for _, code := range errorCodes {
		if code == "" {
			t.Errorf("Error code is empty")
		}
	}
}
