package validators

import (
	"fmt"
	"strings"
)

// RequiredFieldsValidator validates that required fields exist and are not empty
type RequiredFieldsValidator struct {
	*BaseValidator
	fields      []string
	allowEmpty  bool
	customNames map[string]string // Custom field names for error messages
}

// NewRequiredFieldsValidator creates a new required fields validator
func NewRequiredFieldsValidator(fields ...string) *RequiredFieldsValidator {
	return &RequiredFieldsValidator{
		BaseValidator: NewBaseValidator("required_fields"),
		fields:        fields,
		allowEmpty:    false,
		customNames:   make(map[string]string),
	}
}

// WithAllowEmpty allows required fields to be empty (just checks existence)
func (rfv *RequiredFieldsValidator) WithAllowEmpty(allow bool) *RequiredFieldsValidator {
	rfv.allowEmpty = allow
	return rfv
}

// WithCustomNames sets custom names for fields in error messages
func (rfv *RequiredFieldsValidator) WithCustomNames(names map[string]string) *RequiredFieldsValidator {
	rfv.customNames = names
	return rfv
}

// AddField adds a field to the required fields list
func (rfv *RequiredFieldsValidator) AddField(field string) *RequiredFieldsValidator {
	rfv.fields = append(rfv.fields, field)
	return rfv
}

// RemoveField removes a field from the required fields list
func (rfv *RequiredFieldsValidator) RemoveField(field string) *RequiredFieldsValidator {
	for i, f := range rfv.fields {
		if f == field {
			rfv.fields = append(rfv.fields[:i], rfv.fields[i+1:]...)
			break
		}
	}
	return rfv
}

func (rfv *RequiredFieldsValidator) Validate(data map[string]interface{}) error {
	if err := ValidateValidatorInput(data); err != nil {
		return err
	}

	multiErr := NewMultiValidationError()

	for _, field := range rfv.fields {
		if err := rfv.validateField(data, field); err != nil {
			multiErr.Add(err)
		}
	}

	if multiErr.HasErrors() {
		return multiErr
	}

	return nil
}

func (rfv *RequiredFieldsValidator) validateField(data map[string]interface{}, field string) error {
	value, exists := GetFieldValue(data, field)

	// Check if field exists
	if !exists {
		fieldName := rfv.getFieldName(field)
		return NewValidationError(field, nil, fmt.Sprintf("required field '%s' is missing", fieldName), rfv.Name())
	}

	// Check if field is empty (if not allowing empty)
	if !rfv.allowEmpty && IsEmpty(value) {
		fieldName := rfv.getFieldName(field)
		return NewValidationError(field, value, fmt.Sprintf("required field '%s' cannot be empty", fieldName), rfv.Name())
	}

	return nil
}

func (rfv *RequiredFieldsValidator) getFieldName(field string) string {
	if customName, exists := rfv.customNames[field]; exists {
		return customName
	}
	return field
}

// RequiredStringFieldsValidator validates that required string fields exist and are not empty
type RequiredStringFieldsValidator struct {
	*BaseValidator
	fields    []string
	minLength int
	maxLength int
}

// NewRequiredStringFieldsValidator creates a validator for required string fields
func NewRequiredStringFieldsValidator(fields ...string) *RequiredStringFieldsValidator {
	return &RequiredStringFieldsValidator{
		BaseValidator: NewBaseValidator("required_string_fields"),
		fields:        fields,
		minLength:     0,
		maxLength:     -1, // No max length by default
	}
}

// WithMinLength sets the minimum length for string fields
func (rsfv *RequiredStringFieldsValidator) WithMinLength(length int) *RequiredStringFieldsValidator {
	rsfv.minLength = length
	return rsfv
}

// WithMaxLength sets the maximum length for string fields
func (rsfv *RequiredStringFieldsValidator) WithMaxLength(length int) *RequiredStringFieldsValidator {
	rsfv.maxLength = length
	return rsfv
}

func (rsfv *RequiredStringFieldsValidator) Validate(data map[string]interface{}) error {
	if err := ValidateValidatorInput(data); err != nil {
		return err
	}

	multiErr := NewMultiValidationError()

	for _, field := range rsfv.fields {
		if err := rsfv.validateStringField(data, field); err != nil {
			multiErr.Add(err)
		}
	}

	if multiErr.HasErrors() {
		return multiErr
	}

	return nil
}

func (rsfv *RequiredStringFieldsValidator) validateStringField(data map[string]interface{}, field string) error {
	value, exists := GetFieldValue(data, field)

	if !exists {
		return NewValidationError(field, nil, fmt.Sprintf("required string field '%s' is missing", field), rsfv.Name())
	}

	str, ok := value.(string)
	if !ok {
		return NewValidationError(field, value, fmt.Sprintf("field '%s' must be a string", field), rsfv.Name())
	}

	if strings.TrimSpace(str) == "" {
		return NewValidationError(field, value, fmt.Sprintf("required string field '%s' cannot be empty", field), rsfv.Name())
	}

	if rsfv.minLength > 0 && len(str) < rsfv.minLength {
		return NewValidationError(field, value, fmt.Sprintf("field '%s' must be at least %d characters long", field, rsfv.minLength), rsfv.Name())
	}

	if rsfv.maxLength > 0 && len(str) > rsfv.maxLength {
		return NewValidationError(field, value, fmt.Sprintf("field '%s' must be at most %d characters long", field, rsfv.maxLength), rsfv.Name())
	}

	return nil
}

// RequiredNestedFieldsValidator validates required fields in nested structures
type RequiredNestedFieldsValidator struct {
	*BaseValidator
	fieldPaths []string // Dot-separated paths like "user.profile.name"
}

// NewRequiredNestedFieldsValidator creates a validator for required nested fields
func NewRequiredNestedFieldsValidator(fieldPaths ...string) *RequiredNestedFieldsValidator {
	return &RequiredNestedFieldsValidator{
		BaseValidator: NewBaseValidator("required_nested_fields"),
		fieldPaths:    fieldPaths,
	}
}

func (rnfv *RequiredNestedFieldsValidator) Validate(data map[string]interface{}) error {
	if err := ValidateValidatorInput(data); err != nil {
		return err
	}

	multiErr := NewMultiValidationError()

	for _, fieldPath := range rnfv.fieldPaths {
		if err := rnfv.validateNestedField(data, fieldPath); err != nil {
			multiErr.Add(err)
		}
	}

	if multiErr.HasErrors() {
		return multiErr
	}

	return nil
}

func (rnfv *RequiredNestedFieldsValidator) validateNestedField(data map[string]interface{}, fieldPath string) error {
	parts := strings.Split(fieldPath, ".")
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return NewValidationError(fieldPath, nil, fmt.Sprintf("required nested field '%s' is missing", fieldPath), rnfv.Name())
		}

		// If this is the last part, check if it's empty
		if i == len(parts)-1 {
			if IsEmpty(value) {
				return NewValidationError(fieldPath, value, fmt.Sprintf("required nested field '%s' cannot be empty", fieldPath), rnfv.Name())
			}
			return nil
		}

		// Otherwise, continue traversing
		if nested, ok := value.(map[string]interface{}); ok {
			current = nested
		} else {
			return NewValidationError(fieldPath, value, fmt.Sprintf("nested field path '%s' is invalid at '%s'", fieldPath, part), rnfv.Name())
		}
	}

	return nil
}

// ConditionalRequiredValidator validates fields that are required based on conditions
type ConditionalRequiredValidator struct {
	*BaseValidator
	conditions []conditionalRequirement
}

type conditionalRequirement struct {
	condition func(map[string]interface{}) bool
	fields    []string
	message   string
}

// NewConditionalRequiredValidator creates a validator for conditionally required fields
func NewConditionalRequiredValidator() *ConditionalRequiredValidator {
	return &ConditionalRequiredValidator{
		BaseValidator: NewBaseValidator("conditional_required"),
		conditions:    make([]conditionalRequirement, 0),
	}
}

// When adds a conditional requirement
func (crv *ConditionalRequiredValidator) When(condition func(map[string]interface{}) bool, fields []string, message string) *ConditionalRequiredValidator {
	crv.conditions = append(crv.conditions, conditionalRequirement{
		condition: condition,
		fields:    fields,
		message:   message,
	})
	return crv
}

func (crv *ConditionalRequiredValidator) Validate(data map[string]interface{}) error {
	if err := ValidateValidatorInput(data); err != nil {
		return err
	}

	multiErr := NewMultiValidationError()

	for _, req := range crv.conditions {
		if req.condition(data) {
			for _, field := range req.fields {
				value, exists := GetFieldValue(data, field)
				if !exists || IsEmpty(value) {
					message := req.message
					if message == "" {
						message = fmt.Sprintf("conditionally required field '%s' is missing or empty", field)
					}
					multiErr.Add(NewValidationError(field, value, message, crv.Name()))
				}
			}
		}
	}

	if multiErr.HasErrors() {
		return multiErr
	}

	return nil
}

// Helper functions for creating common required field validators

// EmailRequiredValidator creates a validator for required email fields
func EmailRequiredValidator(fields ...string) Validator {
	return NewFuncValidator("email_required", func(data map[string]interface{}) error {
		multiErr := NewMultiValidationError()

		for _, field := range fields {
			str, err := GetStringField(data, field)
			if err != nil {
				multiErr.Add(NewValidationError(field, nil, fmt.Sprintf("email field '%s' is required", field), "email_required"))
				continue
			}

			if !strings.Contains(str, "@") || !strings.Contains(str, ".") {
				multiErr.Add(NewValidationError(field, str, fmt.Sprintf("field '%s' must be a valid email address", field), "email_required"))
			}
		}

		if multiErr.HasErrors() {
			return multiErr
		}
		return nil
	})
}

// IDRequiredValidator creates a validator for required ID fields
func IDRequiredValidator(fields ...string) Validator {
	return NewFuncValidator("id_required", func(data map[string]interface{}) error {
		multiErr := NewMultiValidationError()

		for _, field := range fields {
			value, exists := GetFieldValue(data, field)
			if !exists {
				multiErr.Add(NewValidationError(field, nil, fmt.Sprintf("ID field '%s' is required", field), "id_required"))
				continue
			}

			// Check if it's a valid ID (string or number, not empty)
			switch v := value.(type) {
			case string:
				if strings.TrimSpace(v) == "" {
					multiErr.Add(NewValidationError(field, value, fmt.Sprintf("ID field '%s' cannot be empty", field), "id_required"))
				}
			case int, int32, int64, float64:
				// Numbers are valid IDs
			default:
				multiErr.Add(NewValidationError(field, value, fmt.Sprintf("ID field '%s' must be a string or number", field), "id_required"))
			}
		}

		if multiErr.HasErrors() {
			return multiErr
		}
		return nil
	})
}
