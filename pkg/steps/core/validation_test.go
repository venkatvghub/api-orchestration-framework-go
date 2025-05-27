package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/validators"
)

// MockValidator for testing
type MockValidator struct {
	mock.Mock
	name string
}

func NewMockValidator(name string) *MockValidator {
	return &MockValidator{name: name}
}

func (m *MockValidator) Name() string {
	return m.name
}

func (m *MockValidator) Validate(data map[string]interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

func TestNewValidationStep(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := NewValidationStep("test_validation", validator)

	assert.NotNil(t, step)
	assert.Equal(t, "test_validation", step.Name())
	assert.Equal(t, "Data validation", step.Description())
	assert.Equal(t, validator, step.validator)
	assert.Empty(t, step.dataField)
	assert.False(t, step.continueOnErr)
	assert.True(t, step.storeResult)
	assert.Equal(t, "validation_result", step.resultField)
}

func TestValidationStep_WithDataField(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := NewValidationStep("test", validator)

	result := step.WithDataField("user_data")

	assert.Equal(t, step, result) // Fluent API
	assert.Equal(t, "user_data", step.dataField)
}

func TestValidationStep_WithContinueOnError(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := NewValidationStep("test", validator)

	result := step.WithContinueOnError(true)

	assert.Equal(t, step, result) // Fluent API
	assert.True(t, step.continueOnErr)
}

func TestValidationStep_WithResultStorage(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := NewValidationStep("test", validator)

	result := step.WithResultStorage(false, "custom_result")

	assert.Equal(t, step, result) // Fluent API
	assert.False(t, step.storeResult)
	assert.Equal(t, "custom_result", step.resultField)
}

func TestValidationStep_Run_Success_EntireContext(t *testing.T) {
	validator := NewMockValidator("test_validator")
	validator.On("Validate", mock.MatchedBy(func(data map[string]interface{}) bool {
		return data["user_id"] == "123" && data["email"] == "test@example.com"
	})).Return(nil)

	step := NewValidationStep("test_validation", validator)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("email", "test@example.com")

	err := step.Run(ctx)

	assert.NoError(t, err)

	// Check validation result
	result, exists := ctx.Get("validation_result")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.True(t, resultMap["valid"].(bool))
	assert.Equal(t, "test_validator", resultMap["validator"])
	assert.NotContains(t, resultMap, "error")

	validator.AssertExpectations(t)
}

func TestValidationStep_Run_Success_SpecificField(t *testing.T) {
	validator := NewMockValidator("test_validator")
	userData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	validator.On("Validate", userData).Return(nil)

	step := NewValidationStep("test_validation", validator).
		WithDataField("user_data")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_data", userData)
	ctx.Set("other_field", "should_not_be_validated")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("validation_result")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.True(t, resultMap["valid"].(bool))

	validator.AssertExpectations(t)
}

func TestValidationStep_Run_Success_NonMapField(t *testing.T) {
	validator := NewMockValidator("test_validator")
	validator.On("Validate", map[string]interface{}{
		"user_id": "123",
	}).Return(nil)

	step := NewValidationStep("test_validation", validator).
		WithDataField("user_id")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.NoError(t, err)
	validator.AssertExpectations(t)
}

func TestValidationStep_Run_ValidationFailure_StopOnError(t *testing.T) {
	validator := NewMockValidator("test_validator")
	validationError := fmt.Errorf("validation failed: missing required field")
	validator.On("Validate", mock.Anything).Return(validationError)

	step := NewValidationStep("test_validation", validator).
		WithContinueOnError(false)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")

	// Check validation result is still stored
	result, exists := ctx.Get("validation_result")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.False(t, resultMap["valid"].(bool))
	assert.Equal(t, "validation failed: missing required field", resultMap["error"])

	validator.AssertExpectations(t)
}

func TestValidationStep_Run_ValidationFailure_ContinueOnError(t *testing.T) {
	validator := NewMockValidator("test_validator")
	validationError := fmt.Errorf("validation failed")
	validator.On("Validate", mock.Anything).Return(validationError)

	step := NewValidationStep("test_validation", validator).
		WithContinueOnError(true)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.NoError(t, err) // Should not fail when continue on error is true

	result, exists := ctx.Get("validation_result")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.False(t, resultMap["valid"].(bool))

	validator.AssertExpectations(t)
}

func TestValidationStep_Run_MissingDataField(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := NewValidationStep("test_validation", validator).
		WithDataField("missing_field")

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data field 'missing_field' not found")

	validator.AssertNotCalled(t, "Validate")
}

func TestValidationStep_Run_NoResultStorage(t *testing.T) {
	validator := NewMockValidator("test_validator")
	validator.On("Validate", mock.Anything).Return(nil)

	step := NewValidationStep("test_validation", validator).
		WithResultStorage(false, "")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.NoError(t, err)

	// Should not store result
	_, exists := ctx.Get("validation_result")
	assert.False(t, exists)

	validator.AssertExpectations(t)
}

func TestValidationStep_ExtractValidationDetails(t *testing.T) {
	step := &ValidationStep{}

	tests := []struct {
		name     string
		err      error
		expected interface{}
	}{
		{
			name: "ValidationError",
			err: &validators.ValidationError{
				Field:     "email",
				Value:     "invalid-email",
				Message:   "invalid email format",
				Validator: "email_validator",
			},
			expected: map[string]interface{}{
				"field":     "email",
				"value":     "invalid-email",
				"message":   "invalid email format",
				"validator": "email_validator",
			},
		},
		{
			name: "MultiValidationError",
			err: &validators.MultiValidationError{
				Errors: []error{
					&validators.ValidationError{Field: "name", Message: "required"},
					&validators.ValidationError{Field: "email", Message: "invalid"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"field":     "name",
					"value":     nil,
					"message":   "required",
					"validator": "",
				},
				map[string]interface{}{
					"field":     "email",
					"value":     nil,
					"message":   "invalid",
					"validator": "",
				},
			},
		},
		{
			name:     "Generic error",
			err:      fmt.Errorf("generic error"),
			expected: "generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := step.extractValidationDetails(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ValidationChainStep tests

func TestNewValidationChainStep(t *testing.T) {
	validator1 := NewMockValidator("validator1")
	validator2 := NewMockValidator("validator2")

	step := NewValidationChainStep("test_chain", validator1, validator2)

	assert.NotNil(t, step)
	assert.Equal(t, "test_chain", step.Name())
	assert.Equal(t, "Validation chain", step.Description())
	assert.Len(t, step.validators, 2)
	assert.Equal(t, validator1, step.validators[0])
	assert.Equal(t, validator2, step.validators[1])
	assert.Empty(t, step.dataField)
	assert.True(t, step.stopOnFirst)
	assert.False(t, step.continueOnErr)
	assert.True(t, step.storeResults)
	assert.Equal(t, "validation_results", step.resultField)
}

func TestValidationChainStep_Configuration(t *testing.T) {
	validator := NewMockValidator("validator")
	step := NewValidationChainStep("test", validator)

	result := step.
		WithDataField("user_data").
		WithStopOnFirst(false).
		WithContinueOnError(true).
		WithResultStorage(false, "custom_results")

	assert.Equal(t, step, result) // Fluent API
	assert.Equal(t, "user_data", step.dataField)
	assert.False(t, step.stopOnFirst)
	assert.True(t, step.continueOnErr)
	assert.False(t, step.storeResults)
	assert.Equal(t, "custom_results", step.resultField)
}

func TestValidationChainStep_Run_AllSuccess(t *testing.T) {
	validator1 := NewMockValidator("validator1")
	validator2 := NewMockValidator("validator2")
	validator3 := NewMockValidator("validator3")

	validator1.On("Validate", mock.Anything).Return(nil)
	validator2.On("Validate", mock.Anything).Return(nil)
	validator3.On("Validate", mock.Anything).Return(nil)

	step := NewValidationChainStep("test_chain", validator1, validator2, validator3)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.NoError(t, err)

	// Check results
	result, exists := ctx.Get("validation_results")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.True(t, resultMap["valid"].(bool))
	assert.Equal(t, 3, resultMap["total_validators"])
	assert.Equal(t, 3, resultMap["valid_count"])
	assert.Equal(t, 0, resultMap["failed_count"])

	results, ok := resultMap["results"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, results, 3)

	validator1.AssertExpectations(t)
	validator2.AssertExpectations(t)
	validator3.AssertExpectations(t)
}

func TestValidationChainStep_Run_StopOnFirstFailure(t *testing.T) {
	validator1 := NewMockValidator("validator1")
	validator2 := NewMockValidator("validator2")
	validator3 := NewMockValidator("validator3")

	validator1.On("Validate", mock.Anything).Return(nil)
	validator2.On("Validate", mock.Anything).Return(fmt.Errorf("validation failed"))
	// validator3 should not be called due to stopOnFirst

	step := NewValidationChainStep("test_chain", validator1, validator2, validator3).
		WithStopOnFirst(true)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation chain failed")

	// Check results
	result, exists := ctx.Get("validation_results")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.False(t, resultMap["valid"].(bool))
	assert.Equal(t, 3, resultMap["total_validators"])
	assert.Equal(t, 1, resultMap["valid_count"])
	assert.Equal(t, 2, resultMap["failed_count"])
	assert.Equal(t, "validation failed", resultMap["first_error"])

	results, ok := resultMap["results"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, results, 2) // Only first two validators

	validator1.AssertExpectations(t)
	validator2.AssertExpectations(t)
	validator3.AssertNotCalled(t, "Validate")
}

func TestValidationChainStep_Run_ContinueOnFailure(t *testing.T) {
	validator1 := NewMockValidator("validator1")
	validator2 := NewMockValidator("validator2")
	validator3 := NewMockValidator("validator3")

	validator1.On("Validate", mock.Anything).Return(nil)
	validator2.On("Validate", mock.Anything).Return(fmt.Errorf("validation failed"))
	validator3.On("Validate", mock.Anything).Return(nil)

	step := NewValidationChainStep("test_chain", validator1, validator2, validator3).
		WithStopOnFirst(false)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.Error(t, err) // Should still fail overall

	// Check results
	result, exists := ctx.Get("validation_results")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.False(t, resultMap["valid"].(bool))
	assert.Equal(t, 3, resultMap["total_validators"])
	assert.Equal(t, 2, resultMap["valid_count"])
	assert.Equal(t, 1, resultMap["failed_count"])

	results, ok := resultMap["results"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, results, 3) // All validators ran

	validator1.AssertExpectations(t)
	validator2.AssertExpectations(t)
	validator3.AssertExpectations(t)
}

func TestValidationChainStep_Run_ContinueOnError(t *testing.T) {
	validator1 := NewMockValidator("validator1")
	validator2 := NewMockValidator("validator2")

	validator1.On("Validate", mock.Anything).Return(nil)
	validator2.On("Validate", mock.Anything).Return(fmt.Errorf("validation failed"))

	step := NewValidationChainStep("test_chain", validator1, validator2).
		WithContinueOnError(true)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.NoError(t, err) // Should not fail when continue on error is true

	result, exists := ctx.Get("validation_results")
	assert.True(t, exists)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.False(t, resultMap["valid"].(bool))

	validator1.AssertExpectations(t)
	validator2.AssertExpectations(t)
}

func TestValidationChainStep_Run_MissingDataField(t *testing.T) {
	validator := NewMockValidator("validator")
	step := NewValidationChainStep("test_chain", validator).
		WithDataField("missing_field")

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data field 'missing_field' not found")

	validator.AssertNotCalled(t, "Validate")
}

func TestValidationChainStep_Run_NoResultStorage(t *testing.T) {
	validator := NewMockValidator("validator")
	validator.On("Validate", mock.Anything).Return(nil)

	step := NewValidationChainStep("test_chain", validator).
		WithResultStorage(false, "")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.NoError(t, err)

	// Should not store results
	_, exists := ctx.Get("validation_results")
	assert.False(t, exists)

	validator.AssertExpectations(t)
}

// Test helper functions

func TestNewRequiredFieldsValidationStep(t *testing.T) {
	step := NewRequiredFieldsValidationStep("required_test", "name", "email")

	assert.Equal(t, "required_test", step.Name())
	assert.NotNil(t, step.validator)
	assert.Equal(t, "required_fields", step.validator.Name())
}

func TestNewRequiredStringFieldsValidationStep(t *testing.T) {
	step := NewRequiredStringFieldsValidationStep("string_test", "name", "description")

	assert.Equal(t, "string_test", step.Name())
	assert.NotNil(t, step.validator)
	assert.Equal(t, "required_string_fields", step.validator.Name())
}

func TestNewEmailValidationStep(t *testing.T) {
	step := NewEmailValidationStep("email_test", "email", "contact_email")

	assert.Equal(t, "email_test", step.Name())
	assert.NotNil(t, step.validator)
	assert.Equal(t, "email_required", step.validator.Name())
}

func TestNewIDValidationStep(t *testing.T) {
	step := NewIDValidationStep("id_test", "user_id", "product_id")

	assert.Equal(t, "id_test", step.Name())
	assert.NotNil(t, step.validator)
	assert.Equal(t, "id_required", step.validator.Name())
}

func TestNewCustomValidationStep(t *testing.T) {
	validationFunc := func(data map[string]interface{}) error {
		if data["age"].(int) < 18 {
			return fmt.Errorf("age must be 18 or older")
		}
		return nil
	}

	step := NewCustomValidationStep("custom_test", validationFunc)

	assert.Equal(t, "custom_test", step.Name())
	assert.NotNil(t, step.validator)
	assert.Equal(t, "custom_test_validator", step.validator.Name())
}

func TestValidationStep_GetDataToValidate_EntireContext(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := &ValidationStep{
		validator: validator,
		dataField: "", // Empty means entire context
	}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("email", "test@example.com")

	data, err := step.getDataToValidate(ctx)

	assert.NoError(t, err)
	assert.Len(t, data, 2)
	assert.Equal(t, "123", data["user_id"])
	assert.Equal(t, "test@example.com", data["email"])
}

func TestValidationStep_GetDataToValidate_SpecificMapField(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := &ValidationStep{
		validator: validator,
		dataField: "user_data",
	}

	userData := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_data", userData)

	data, err := step.getDataToValidate(ctx)

	assert.NoError(t, err)
	assert.Equal(t, userData, data)
}

func TestValidationStep_GetDataToValidate_SpecificNonMapField(t *testing.T) {
	validator := NewMockValidator("test_validator")
	step := &ValidationStep{
		validator: validator,
		dataField: "user_id",
	}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	data, err := step.getDataToValidate(ctx)

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"user_id": "123"}, data)
}

func TestValidationChainStep_GetDataToValidate(t *testing.T) {
	step := &ValidationChainStep{
		dataField: "test_field",
	}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("test_field", "test_value")

	data, err := step.getDataToValidate(ctx)

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test_field": "test_value"}, data)
}

func TestValidationChainStep_ExtractValidationDetails(t *testing.T) {
	step := &ValidationChainStep{}

	validationErr := &validators.ValidationError{
		Field:   "email",
		Message: "invalid format",
	}

	result := step.extractValidationDetails(validationErr)

	expected := map[string]interface{}{
		"field":     "email",
		"value":     nil,
		"message":   "invalid format",
		"validator": "",
	}

	assert.Equal(t, expected, result)
}
