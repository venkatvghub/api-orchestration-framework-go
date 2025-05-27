package validators

import (
	"fmt"
)

// Validator defines the interface for data validation operations
type Validator interface {
	// Validate checks if the input data meets the validation criteria
	Validate(data map[string]interface{}) error

	// Name returns the validator name for logging and debugging
	Name() string
}

// BaseValidator provides common functionality for validators
type BaseValidator struct {
	name string
}

// NewBaseValidator creates a new base validator
func NewBaseValidator(name string) *BaseValidator {
	return &BaseValidator{
		name: name,
	}
}

func (bv *BaseValidator) Name() string {
	return bv.name
}

// FuncValidator wraps a function as a validator
type FuncValidator struct {
	*BaseValidator
	validateFunc func(map[string]interface{}) error
}

// NewFuncValidator creates a validator from a function
func NewFuncValidator(name string, fn func(map[string]interface{}) error) *FuncValidator {
	return &FuncValidator{
		BaseValidator: NewBaseValidator(name),
		validateFunc:  fn,
	}
}

func (fv *FuncValidator) Validate(data map[string]interface{}) error {
	return fv.validateFunc(data)
}

// NoOpValidator is a validator that always passes
type NoOpValidator struct {
	*BaseValidator
}

// NewNoOpValidator creates a no-operation validator
func NewNoOpValidator() *NoOpValidator {
	return &NoOpValidator{
		BaseValidator: NewBaseValidator("noop"),
	}
}

func (nv *NoOpValidator) Validate(data map[string]interface{}) error {
	return nil
}

// AlwaysFailValidator is a validator that always fails (useful for testing)
type AlwaysFailValidator struct {
	*BaseValidator
	errorMessage string
}

// NewAlwaysFailValidator creates a validator that always fails
func NewAlwaysFailValidator(errorMessage string) *AlwaysFailValidator {
	return &AlwaysFailValidator{
		BaseValidator: NewBaseValidator("always_fail"),
		errorMessage:  errorMessage,
	}
}

func (afv *AlwaysFailValidator) Validate(data map[string]interface{}) error {
	return fmt.Errorf(afv.errorMessage)
}

// ValidateValidatorInput checks if the input data is valid for validation
func ValidateValidatorInput(data map[string]interface{}) error {
	if data == nil {
		return fmt.Errorf("validator input data cannot be nil")
	}
	return nil
}

// ValidationError represents a validation error with additional context
type ValidationError struct {
	Field     string
	Value     interface{}
	Message   string
	Validator string
}

func (ve *ValidationError) Error() string {
	if ve.Field != "" {
		return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
	}
	return fmt.Sprintf("validation failed: %s", ve.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string, validator string) *ValidationError {
	return &ValidationError{
		Field:     field,
		Value:     value,
		Message:   message,
		Validator: validator,
	}
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []error
}

func (mve *MultiValidationError) Error() string {
	if len(mve.Errors) == 0 {
		return "no validation errors"
	}

	if len(mve.Errors) == 1 {
		return mve.Errors[0].Error()
	}

	message := fmt.Sprintf("multiple validation errors (%d):", len(mve.Errors))
	for i, err := range mve.Errors {
		message += fmt.Sprintf("\n  %d. %s", i+1, err.Error())
	}
	return message
}

// Add appends an error to the multi-validation error
func (mve *MultiValidationError) Add(err error) {
	if err != nil {
		mve.Errors = append(mve.Errors, err)
	}
}

// HasErrors returns true if there are any validation errors
func (mve *MultiValidationError) HasErrors() bool {
	return len(mve.Errors) > 0
}

// NewMultiValidationError creates a new multi-validation error
func NewMultiValidationError() *MultiValidationError {
	return &MultiValidationError{
		Errors: make([]error, 0),
	}
}

// Helper functions for common validation patterns

// IsEmpty checks if a value is considered empty
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	case int, int32, int64:
		return v == 0
	case float32, float64:
		return v == 0.0
	case bool:
		return !v
	default:
		return false
	}
}

// GetFieldValue safely extracts a field value from data
func GetFieldValue(data map[string]interface{}, field string) (interface{}, bool) {
	value, exists := data[field]
	return value, exists
}

// GetStringField safely extracts a string field from data
func GetStringField(data map[string]interface{}, field string) (string, error) {
	value, exists := data[field]
	if !exists {
		return "", fmt.Errorf("field '%s' not found", field)
	}

	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("field '%s' is not a string", field)
}

// GetIntField safely extracts an int field from data
func GetIntField(data map[string]interface{}, field string) (int, error) {
	value, exists := data[field]
	if !exists {
		return 0, fmt.Errorf("field '%s' not found", field)
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("field '%s' is not an integer", field)
	}
}

// GetBoolField safely extracts a bool field from data
func GetBoolField(data map[string]interface{}, field string) (bool, error) {
	value, exists := data[field]
	if !exists {
		return false, fmt.Errorf("field '%s' not found", field)
	}

	if b, ok := value.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("field '%s' is not a boolean", field)
}
