# Errors Package Documentation

## Overview

The `pkg/errors` package provides a comprehensive, structured error handling system for the API Orchestration Framework. It offers standardized error types, HTTP status code mapping, error categorization, and integration with observability systems for better error tracking and debugging.

## Purpose

The errors package serves as the error management foundation that:
- Provides standardized error types with consistent structure
- Maps errors to appropriate HTTP status codes
- Categorizes errors for better handling and routing
- Integrates with logging and metrics systems
- Supports error wrapping and context preservation
- Enables retry logic based on error types
- Provides user-friendly error messages

## Core Architecture

### Error Interface

The foundation error interface with enhanced capabilities:
```go
type FrameworkError interface {
    error
    
    // Error type classification
    Type() ErrorType
    Code() string
    
    // HTTP integration
    HTTPStatus() int
    
    // Context and metadata
    Context() map[string]interface{}
    WithContext(key string, value interface{}) FrameworkError
    
    // Error chaining
    Cause() error
    Unwrap() error
    
    // Retry logic
    IsRetryable() bool
    
    // User-facing messages
    UserMessage() string
}
```

### Error Types

Comprehensive error categorization:
```go
type ErrorType string

const (
    // Client errors (4xx)
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypeAuthentication ErrorType = "authentication"
    ErrorTypeAuthorization  ErrorType = "authorization"
    ErrorTypeNotFound      ErrorType = "not_found"
    ErrorTypeConflict      ErrorType = "conflict"
    ErrorTypeRateLimit     ErrorType = "rate_limit"
    ErrorTypeBadRequest    ErrorType = "bad_request"
    
    // Server errors (5xx)
    ErrorTypeInternal      ErrorType = "internal"
    ErrorTypeTimeout       ErrorType = "timeout"
    ErrorTypeUnavailable   ErrorType = "unavailable"
    ErrorTypeCircuitBreaker ErrorType = "circuit_breaker"
    ErrorTypeDatabase      ErrorType = "database"
    ErrorTypeExternal      ErrorType = "external"
    
    // Framework errors
    ErrorTypeConfiguration ErrorType = "configuration"
    ErrorTypeFlow          ErrorType = "flow"
    ErrorTypeStep          ErrorType = "step"
    ErrorTypeTransform     ErrorType = "transform"
    ErrorTypeCache         ErrorType = "cache"
)
```

### BaseError Implementation

Core error implementation with full feature support:
```go
type BaseError struct {
    errorType    ErrorType
    code         string
    message      string
    userMessage  string
    httpStatus   int
    retryable    bool
    cause        error
    context      map[string]interface{}
    timestamp    time.Time
    stackTrace   []string
}

func NewBaseError(errorType ErrorType, code, message string) *BaseError {
    return &BaseError{
        errorType:   errorType,
        code:        code,
        message:     message,
        httpStatus:  getDefaultHTTPStatus(errorType),
        retryable:   getDefaultRetryable(errorType),
        context:     make(map[string]interface{}),
        timestamp:   time.Now(),
        stackTrace:  captureStackTrace(),
    }
}
```

## Error Creation Functions

### Validation Errors
```go
// Field validation errors
validationErr := errors.NewValidationError("INVALID_EMAIL", "Invalid email format").
    WithField("email", "invalid@").
    WithUserMessage("Please provide a valid email address")

// Required field errors
requiredErr := errors.NewRequiredFieldError("name", "Name is required")

// Type validation errors
typeErr := errors.NewTypeValidationError("age", "integer", "string")

// Range validation errors
rangeErr := errors.NewRangeValidationError("age", 18, 120, 15)
```

### Authentication Errors
```go
// Invalid credentials
authErr := errors.NewAuthenticationError("INVALID_CREDENTIALS", "Invalid username or password").
    WithUserMessage("Invalid login credentials")

// Token errors
tokenErr := errors.NewTokenError("EXPIRED_TOKEN", "JWT token has expired").
    WithRetryable(false)

// Missing authentication
missingAuthErr := errors.NewMissingAuthenticationError()
```

### Authorization Errors
```go
// Insufficient permissions
authzErr := errors.NewAuthorizationError("INSUFFICIENT_PERMISSIONS", "User lacks required permissions").
    WithRequiredPermission("admin").
    WithUserMessage("You don't have permission to perform this action")

// Resource access denied
accessErr := errors.NewAccessDeniedError("resource_id", "users")
```

### HTTP Errors
```go
// Not found errors
notFoundErr := errors.NewNotFoundError("USER_NOT_FOUND", "User not found").
    WithResourceID("123").
    WithResourceType("user")

// Conflict errors
conflictErr := errors.NewConflictError("EMAIL_EXISTS", "Email already exists").
    WithConflictingField("email", "john@example.com")

// Rate limit errors
rateLimitErr := errors.NewRateLimitError("RATE_LIMIT_EXCEEDED", "Too many requests").
    WithRetryAfter(60 * time.Second)
```

### Server Errors
```go
// Internal server errors
internalErr := errors.NewInternalError("DATABASE_CONNECTION", "Failed to connect to database").
    WithRetryable(true)

// Timeout errors
timeoutErr := errors.NewTimeoutError("REQUEST_TIMEOUT", "Request timed out after 30s").
    WithTimeout(30 * time.Second).
    WithRetryable(true)

// Service unavailable
unavailableErr := errors.NewUnavailableError("SERVICE_DOWN", "External service is unavailable").
    WithService("payment-service").
    WithRetryable(true)

// Circuit breaker errors
circuitErr := errors.NewCircuitBreakerError("CIRCUIT_OPEN", "Circuit breaker is open").
    WithService("user-service").
    WithRetryable(false)
```

### Framework Errors
```go
// Configuration errors
configErr := errors.NewConfigurationError("INVALID_CONFIG", "Invalid HTTP timeout configuration").
    WithConfigField("http.timeout").
    WithRetryable(false)

// Flow execution errors
flowErr := errors.NewFlowError("FLOW_EXECUTION_FAILED", "Flow execution failed at step 'validate'").
    WithFlowName("UserRegistration").
    WithStepName("validate").
    WithRetryable(false)

// Step execution errors
stepErr := errors.NewStepError("STEP_FAILED", "HTTP step failed with 500 status").
    WithStepName("fetchUser").
    WithStepType("http").
    WithRetryable(true)

// Transform errors
transformErr := errors.NewTransformError("TRANSFORM_FAILED", "Failed to transform user data").
    WithTransformerName("mobileTransformer").
    WithRetryable(false)
```

## Error Wrapping and Chaining

### Wrapping Errors
```go
// Wrap existing errors with additional context
originalErr := fmt.Errorf("database connection failed")
wrappedErr := errors.Wrap(originalErr, "FETCH_USER_FAILED", "Failed to fetch user data").
    WithContext("user_id", "123").
    WithRetryable(true)

// Wrap with specific error type
dbErr := errors.WrapAsDatabase(originalErr, "DB_QUERY_FAILED", "User query failed").
    WithQuery("SELECT * FROM users WHERE id = ?").
    WithTable("users")
```

### Error Chaining
```go
// Chain multiple errors
validationErr := errors.NewValidationError("INVALID_DATA", "Validation failed")
stepErr := errors.NewStepError("VALIDATION_STEP_FAILED", "Validation step failed").
    WithCause(validationErr)
flowErr := errors.NewFlowError("USER_FLOW_FAILED", "User registration flow failed").
    WithCause(stepErr)

// Access error chain
rootCause := errors.RootCause(flowErr)
allErrors := errors.ErrorChain(flowErr)
```

## Error Context and Metadata

### Adding Context
```go
err := errors.NewInternalError("API_CALL_FAILED", "External API call failed").
    WithContext("url", "https://api.example.com/users").
    WithContext("method", "GET").
    WithContext("status_code", 500).
    WithContext("response_time", "2.5s").
    WithContext("retry_count", 2)
```

### Structured Context
```go
// Add structured context for better observability
err := errors.NewStepError("HTTP_STEP_FAILED", "HTTP request failed").
    WithHTTPContext(map[string]interface{}{
        "url":          "https://api.example.com/users/123",
        "method":       "GET",
        "status_code":  500,
        "headers":      map[string]string{"Content-Type": "application/json"},
        "response_time": 2500, // milliseconds
    }).
    WithUserContext(map[string]interface{}{
        "user_id":   "123",
        "operation": "fetch_profile",
        "client_ip": "192.168.1.1",
    })
```

## HTTP Status Code Mapping

### Automatic Mapping
```go
// Errors automatically map to appropriate HTTP status codes
validationErr := errors.NewValidationError("INVALID_EMAIL", "Invalid email")
// validationErr.HTTPStatus() returns 400

authErr := errors.NewAuthenticationError("INVALID_TOKEN", "Invalid token")
// authErr.HTTPStatus() returns 401

notFoundErr := errors.NewNotFoundError("USER_NOT_FOUND", "User not found")
// notFoundErr.HTTPStatus() returns 404

internalErr := errors.NewInternalError("DB_ERROR", "Database error")
// internalErr.HTTPStatus() returns 500
```

### Custom Status Codes
```go
// Override default status codes when needed
customErr := errors.NewValidationError("CUSTOM_VALIDATION", "Custom validation error").
    WithHTTPStatus(422) // Unprocessable Entity instead of 400
```

## Retry Logic Integration

### Retry Classification
```go
// Check if error is retryable
if errors.IsRetryable(err) {
    // Implement retry logic
    time.Sleep(retryDelay)
    return retryOperation()
}

// Check specific retry conditions
if errors.IsTimeoutError(err) || errors.IsUnavailableError(err) {
    // Retry with exponential backoff
    return retryWithBackoff()
}
```

### Retry Configuration
```go
// Configure retry behavior
retryableErr := errors.NewTimeoutError("REQUEST_TIMEOUT", "Request timed out").
    WithRetryable(true).
    WithRetryAfter(5 * time.Second).
    WithMaxRetries(3)
```

## Error Aggregation

### Multiple Error Collection
```go
// Collect multiple errors
errorCollector := errors.NewErrorCollector()

// Add individual errors
errorCollector.Add(errors.NewValidationError("INVALID_EMAIL", "Invalid email"))
errorCollector.Add(errors.NewValidationError("INVALID_AGE", "Invalid age"))

// Check if any errors occurred
if errorCollector.HasErrors() {
    // Get aggregated error
    aggregatedErr := errorCollector.Error()
    // aggregatedErr contains all validation errors
}
```

### Validation Error Aggregation
```go
// Aggregate validation errors with field context
validationErrors := errors.NewValidationErrorCollector()
validationErrors.AddFieldError("email", "Invalid email format")
validationErrors.AddFieldError("age", "Age must be between 18 and 120")
validationErrors.AddFieldError("name", "Name is required")

if validationErrors.HasErrors() {
    return validationErrors.Error() // Returns structured validation error
}
```

## Integration with Framework Components

### Flow Integration
```go
// Flows automatically handle framework errors
flow := flow.NewFlow("UserRegistration").
    Step("validate", validationStep).
    Step("create", createUserStep).
    WithErrorHandler(func(err error) error {
        if frameworkErr, ok := err.(errors.FrameworkError); ok {
            // Log structured error
            log.Error("Flow step failed",
                zap.String("error_type", string(frameworkErr.Type())),
                zap.String("error_code", frameworkErr.Code()),
                zap.Int("http_status", frameworkErr.HTTPStatus()),
                zap.Any("context", frameworkErr.Context()))
        }
        return err
    })
```

### HTTP Steps Integration
```go
// HTTP steps automatically create appropriate errors
httpStep := http.GET("/api/users/123").
    WithErrorMapping(map[int]func(string) error{
        404: func(body string) error {
            return errors.NewNotFoundError("USER_NOT_FOUND", "User not found").
                WithContext("response_body", body)
        },
        500: func(body string) error {
            return errors.NewExternalError("UPSTREAM_ERROR", "Upstream service error").
                WithContext("response_body", body).
                WithRetryable(true)
        },
    })
```

### Validation Integration
```go
// Validators create structured validation errors
validator := validators.NewRequiredFieldsValidator("email", "name").
    WithErrorFactory(func(field string) error {
        return errors.NewRequiredFieldError(field, fmt.Sprintf("%s is required", field)).
            WithUserMessage(fmt.Sprintf("Please provide a %s", field))
    })
```

### Metrics Integration
```go
// Errors automatically generate metrics
err := errors.NewTimeoutError("API_TIMEOUT", "API request timed out").
    WithMetrics(true) // Enable automatic metrics collection

// Metrics include:
// - error_count{type="timeout", code="API_TIMEOUT"}
// - error_duration{type="timeout", code="API_TIMEOUT"}
// - error_http_status{status="408"}
```

## Error Serialization

### JSON Serialization
```go
// Errors can be serialized to JSON for API responses
err := errors.NewValidationError("INVALID_DATA", "Validation failed").
    WithField("email", "invalid@").
    WithUserMessage("Please check your input")

jsonBytes, _ := json.Marshal(err)
// Result: {
//   "type": "validation",
//   "code": "INVALID_DATA",
//   "message": "Validation failed",
//   "user_message": "Please check your input",
//   "http_status": 400,
//   "context": {
//     "field": "email",
//     "value": "invalid@"
//   },
//   "timestamp": "2023-12-01T10:30:00Z"
// }
```

### API Response Format
```go
// Standard API error response format
type APIErrorResponse struct {
    Error struct {
        Type        string                 `json:"type"`
        Code        string                 `json:"code"`
        Message     string                 `json:"message"`
        UserMessage string                 `json:"user_message,omitempty"`
        Context     map[string]interface{} `json:"context,omitempty"`
        Timestamp   time.Time              `json:"timestamp"`
        TraceID     string                 `json:"trace_id,omitempty"`
    } `json:"error"`
}

func ToAPIResponse(err error) APIErrorResponse {
    if frameworkErr, ok := err.(errors.FrameworkError); ok {
        return APIErrorResponse{
            Error: struct{...}{
                Type:        string(frameworkErr.Type()),
                Code:        frameworkErr.Code(),
                Message:     frameworkErr.Error(),
                UserMessage: frameworkErr.UserMessage(),
                Context:     frameworkErr.Context(),
                Timestamp:   time.Now(),
                TraceID:     getTraceID(),
            },
        }
    }
    // Handle non-framework errors
    return createGenericErrorResponse(err)
}
```

## Error Recovery and Fallbacks

### Graceful Degradation
```go
// Implement graceful degradation based on error types
func handleUserFetch(userID string) (User, error) {
    user, err := fetchUserFromPrimary(userID)
    if err != nil {
        if errors.IsUnavailableError(err) || errors.IsTimeoutError(err) {
            // Try fallback source
            user, fallbackErr := fetchUserFromCache(userID)
            if fallbackErr == nil {
                log.Warn("Using cached user data due to primary service error",
                    zap.String("user_id", userID),
                    zap.Error(err))
                return user, nil
            }
        }
        return User{}, err
    }
    return user, nil
}
```

### Circuit Breaker Integration
```go
// Circuit breaker with error-based decisions
circuitBreaker := NewCircuitBreaker().
    WithFailureCondition(func(err error) bool {
        // Open circuit on server errors, not client errors
        if frameworkErr, ok := err.(errors.FrameworkError); ok {
            return frameworkErr.HTTPStatus() >= 500
        }
        return true
    }).
    WithRecoveryCondition(func(err error) bool {
        // Recover on successful responses or client errors
        if frameworkErr, ok := err.(errors.FrameworkError); ok {
            return frameworkErr.HTTPStatus() < 500
        }
        return false
    })
```

## Best Practices

### Error Creation
1. **Use Specific Error Types**: Choose the most specific error type for the situation
2. **Provide Clear Messages**: Include both technical and user-friendly messages
3. **Add Context**: Include relevant context for debugging and observability
4. **Set Retry Behavior**: Explicitly set whether errors should trigger retries
5. **Map HTTP Status**: Ensure errors map to appropriate HTTP status codes

### Error Handling
1. **Check Error Types**: Use type assertions to handle specific error types
2. **Preserve Context**: Maintain error context when wrapping or transforming
3. **Log Appropriately**: Log errors with appropriate levels and context
4. **Handle Gracefully**: Implement fallbacks for recoverable errors
5. **Monitor Patterns**: Track error patterns for system health monitoring

### Error Propagation
1. **Wrap Don't Replace**: Wrap errors to preserve the original cause
2. **Add Context at Boundaries**: Add relevant context at service boundaries
3. **Maintain Error Chain**: Preserve the complete error chain for debugging
4. **Convert at Boundaries**: Convert internal errors to appropriate external formats
5. **Sanitize Sensitive Data**: Remove sensitive information from error messages

## Examples

### Complete Error Handling Flow
```go
func ProcessUserRegistration(userData map[string]interface{}) error {
    // Validation with structured errors
    if err := validateUserData(userData); err != nil {
        return errors.Wrap(err, "USER_VALIDATION_FAILED", "User data validation failed").
            WithContext("operation", "registration").
            WithUserMessage("Please check your registration information")
    }
    
    // Database operation with error handling
    userID, err := createUser(userData)
    if err != nil {
        if isDuplicateKeyError(err) {
            return errors.NewConflictError("EMAIL_EXISTS", "Email address already exists").
                WithConflictingField("email", userData["email"]).
                WithUserMessage("An account with this email already exists")
        }
        return errors.WrapAsDatabase(err, "USER_CREATION_FAILED", "Failed to create user").
            WithTable("users").
            WithOperation("INSERT").
            WithRetryable(true)
    }
    
    // External service call with error handling
    if err := sendWelcomeEmail(userID); err != nil {
        // Non-critical error, log but don't fail the registration
        log.Warn("Failed to send welcome email",
            zap.String("user_id", userID),
            zap.Error(err))
        
        // Could also create a warning error for monitoring
        warningErr := errors.NewExternalError("EMAIL_SEND_FAILED", "Failed to send welcome email").
            WithService("email-service").
            WithSeverity("warning").
            WithRetryable(true)
        
        // Queue for retry or manual intervention
        queueEmailRetry(userID, warningErr)
    }
    
    return nil
}
```

### Error-Driven Circuit Breaker
```go
type ErrorAwareCircuitBreaker struct {
    circuitBreaker *CircuitBreaker
    errorAnalyzer  *ErrorAnalyzer
}

func (cb *ErrorAwareCircuitBreaker) Execute(operation func() error) error {
    if cb.circuitBreaker.IsOpen() {
        return errors.NewCircuitBreakerError("CIRCUIT_OPEN", "Circuit breaker is open").
            WithService(cb.circuitBreaker.ServiceName()).
            WithRetryAfter(cb.circuitBreaker.RetryAfter())
    }
    
    err := operation()
    if err != nil {
        // Analyze error to determine circuit breaker action
        if cb.errorAnalyzer.ShouldOpenCircuit(err) {
            cb.circuitBreaker.RecordFailure()
        } else if cb.errorAnalyzer.ShouldIgnoreError(err) {
            // Don't count client errors against circuit breaker
            return err
        }
    } else {
        cb.circuitBreaker.RecordSuccess()
    }
    
    return err
}

type ErrorAnalyzer struct{}

func (ea *ErrorAnalyzer) ShouldOpenCircuit(err error) bool {
    if frameworkErr, ok := err.(errors.FrameworkError); ok {
        // Open circuit on server errors and timeouts
        return frameworkErr.Type() == errors.ErrorTypeInternal ||
               frameworkErr.Type() == errors.ErrorTypeTimeout ||
               frameworkErr.Type() == errors.ErrorTypeUnavailable
    }
    return true // Unknown errors should open circuit
}

func (ea *ErrorAnalyzer) ShouldIgnoreError(err error) bool {
    if frameworkErr, ok := err.(errors.FrameworkError); ok {
        // Ignore client errors for circuit breaker purposes
        return frameworkErr.HTTPStatus() >= 400 && frameworkErr.HTTPStatus() < 500
    }
    return false
}
```

### Comprehensive Error Monitoring
```go
type ErrorMonitor struct {
    metrics    metrics.Registry
    logger     *zap.Logger
    alerting   AlertingService
}

func (em *ErrorMonitor) RecordError(err error) {
    if frameworkErr, ok := err.(errors.FrameworkError); ok {
        // Record metrics
        em.metrics.Counter("errors_total").
            WithTag("type", string(frameworkErr.Type())).
            WithTag("code", frameworkErr.Code()).
            WithTag("http_status", strconv.Itoa(frameworkErr.HTTPStatus())).
            Increment()
        
        // Log structured error
        em.logger.Error("Framework error occurred",
            zap.String("error_type", string(frameworkErr.Type())),
            zap.String("error_code", frameworkErr.Code()),
            zap.String("message", frameworkErr.Error()),
            zap.Int("http_status", frameworkErr.HTTPStatus()),
            zap.Bool("retryable", frameworkErr.IsRetryable()),
            zap.Any("context", frameworkErr.Context()),
            zap.Error(frameworkErr.Cause()))
        
        // Check for alerting conditions
        if em.shouldAlert(frameworkErr) {
            em.alerting.SendAlert(AlertFromError(frameworkErr))
        }
    }
}

func (em *ErrorMonitor) shouldAlert(err errors.FrameworkError) bool {
    // Alert on critical errors
    if err.Type() == errors.ErrorTypeInternal ||
       err.Type() == errors.ErrorTypeDatabase ||
       err.Type() == errors.ErrorTypeConfiguration {
        return true
    }
    
    // Alert on high error rates
    errorRate := em.metrics.Counter("errors_total").
        WithTag("type", string(err.Type())).
        Rate()
    
    return errorRate > 0.1 // 10% error rate threshold
}
```

## Troubleshooting

### Common Issues

1. **Error Type Misclassification**
   ```go
   // Wrong: Using generic internal error for validation
   return errors.NewInternalError("INVALID_EMAIL", "Invalid email")
   
   // Correct: Using specific validation error
   return errors.NewValidationError("INVALID_EMAIL", "Invalid email format")
   ```

2. **Missing Context**
   ```go
   // Wrong: Generic error without context
   return errors.NewInternalError("DB_ERROR", "Database error")
   
   // Correct: Error with relevant context
   return errors.NewDatabaseError("QUERY_FAILED", "User query failed").
       WithQuery("SELECT * FROM users WHERE id = ?").
       WithTable("users").
       WithContext("user_id", userID)
   ```

3. **Incorrect Retry Configuration**
   ```go
   // Wrong: Making authentication errors retryable
   return errors.NewAuthenticationError("INVALID_TOKEN", "Invalid token").
       WithRetryable(true) // Don't retry auth errors
   
   // Correct: Non-retryable authentication error
   return errors.NewAuthenticationError("INVALID_TOKEN", "Invalid token").
       WithRetryable(false)
   ```

### Debugging

1. **Enable Error Tracing**: Include stack traces for debugging
2. **Use Error Context**: Add relevant context for troubleshooting
3. **Monitor Error Patterns**: Track error frequencies and patterns
4. **Validate Error Mapping**: Ensure errors map to correct HTTP status codes
5. **Test Error Scenarios**: Include error scenarios in testing 