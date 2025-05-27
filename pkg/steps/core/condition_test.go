package core

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
)

func TestNewConditionStep(t *testing.T) {
	step := NewConditionStep("test_condition", "user.status", "equals", "active")

	assert.NotNil(t, step)
	assert.Equal(t, "test_condition", step.Name())
	assert.Equal(t, "Condition evaluation", step.Description())
	assert.Equal(t, "user.status", step.field)
	assert.Equal(t, "equals", step.operator)
	assert.Equal(t, "active", step.value)
	assert.Nil(t, step.condition)
}

func TestConditionStep_WithCustomCondition(t *testing.T) {
	customCondition := func(ctx *flow.Context) bool {
		return true
	}

	step := NewConditionStep("test", "field", "equals", "value").
		WithCustomCondition(customCondition)

	assert.NotNil(t, step.condition)
}

func TestConditionStep_Run_CustomCondition(t *testing.T) {
	customCondition := func(ctx *flow.Context) bool {
		value, _ := ctx.Get("test_field")
		return value == "expected"
	}

	step := NewConditionStep("test_condition", "", "", nil).
		WithCustomCondition(customCondition)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("test_field", "expected")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))

	namedResult, exists := ctx.Get("condition_test_condition")
	assert.True(t, exists)
	assert.True(t, namedResult.(bool))
}

func TestConditionStep_Run_ExistsOperator(t *testing.T) {
	step := NewConditionStep("exists_test", "user_id", "exists", nil)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", 123)

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))
}

func TestConditionStep_Run_NotExistsOperator(t *testing.T) {
	step := NewConditionStep("not_exists_test", "missing_field", "not_exists", nil)

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))
}

func TestConditionStep_Run_EmptyOperator(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil value", nil, true},
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"non-empty string", "hello", false},
		{"empty slice", []interface{}{}, true},
		{"non-empty slice", []interface{}{1, 2}, false},
		{"empty map", map[string]interface{}{}, true},
		{"non-empty map", map[string]interface{}{"key": "value"}, false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero float", 0.0, true},
		{"non-zero float", 3.14, false},
		{"false bool", false, true},
		{"true bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := NewConditionStep("empty_test", "test_field", "empty", nil)
			ctx := flow.NewContext().WithFlowName("test_flow")

			if tt.value != nil {
				ctx.Set("test_field", tt.value)
			}

			err := step.Run(ctx)
			assert.NoError(t, err)

			result, exists := ctx.Get("condition_result")
			assert.True(t, exists)
			assert.Equal(t, tt.expected, result.(bool))
		})
	}
}

func TestConditionStep_Run_NotEmptyOperator(t *testing.T) {
	step := NewConditionStep("not_empty_test", "test_field", "not_empty", nil)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("test_field", "hello")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))
}

func TestConditionStep_Run_EqualsOperator(t *testing.T) {
	tests := []struct {
		name      string
		fieldVal  interface{}
		expectVal interface{}
		expected  bool
	}{
		{"string equals", "hello", "hello", true},
		{"string not equals", "hello", "world", false},
		{"int equals", 42, 42, true},
		{"int not equals", 42, 24, false},
		{"float equals", 3.14, 3.14, true},
		{"bool equals", true, true, true},
		{"bool not equals", true, false, false},
		{"string to number", "42", 42, true},
		{"number to string", 42, "42", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := NewConditionStep("equals_test", "test_field", "equals", tt.expectVal)
			ctx := flow.NewContext().WithFlowName("test_flow")
			ctx.Set("test_field", tt.fieldVal)

			err := step.Run(ctx)
			assert.NoError(t, err)

			result, exists := ctx.Get("condition_result")
			assert.True(t, exists)
			assert.Equal(t, tt.expected, result.(bool))
		})
	}
}

func TestConditionStep_Run_NotEqualsOperator(t *testing.T) {
	step := NewConditionStep("not_equals_test", "test_field", "not_equals", "hello")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("test_field", "world")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))
}

func TestConditionStep_Run_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name      string
		operator  string
		fieldVal  interface{}
		expectVal interface{}
		expected  bool
	}{
		{"greater than int", "greater_than", 10, 5, true},
		{"greater than false", "greater_than", 5, 10, false},
		{"greater equal true", "greater_equal", 10, 10, true},
		{"greater equal false", "greater_equal", 5, 10, false},
		{"less than true", "less_than", 5, 10, true},
		{"less than false", "less_than", 10, 5, false},
		{"less equal true", "less_equal", 5, 5, true},
		{"less equal false", "less_equal", 10, 5, false},
		{"string greater", "greater_than", "b", "a", true},
		{"string less", "less_than", "a", "b", true},
		{"float comparison", "greater_than", 3.14, 2.71, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := NewConditionStep("comparison_test", "test_field", tt.operator, tt.expectVal)
			ctx := flow.NewContext().WithFlowName("test_flow")
			ctx.Set("test_field", tt.fieldVal)

			err := step.Run(ctx)
			assert.NoError(t, err)

			result, exists := ctx.Get("condition_result")
			assert.True(t, exists)
			assert.Equal(t, tt.expected, result.(bool), "operator: %s, field: %v, expect: %v", tt.operator, tt.fieldVal, tt.expectVal)
		})
	}
}

func TestConditionStep_Run_StringOperators(t *testing.T) {
	tests := []struct {
		name      string
		operator  string
		fieldVal  string
		expectVal string
		expected  bool
	}{
		{"contains true", "contains", "hello world", "world", true},
		{"contains false", "contains", "hello world", "foo", false},
		{"not contains true", "not_contains", "hello world", "foo", true},
		{"not contains false", "not_contains", "hello world", "world", false},
		{"starts with true", "starts_with", "hello world", "hello", true},
		{"starts with false", "starts_with", "hello world", "world", false},
		{"ends with true", "ends_with", "hello world", "world", true},
		{"ends with false", "ends_with", "hello world", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := NewConditionStep("string_test", "test_field", tt.operator, tt.expectVal)
			ctx := flow.NewContext().WithFlowName("test_flow")
			ctx.Set("test_field", tt.fieldVal)

			err := step.Run(ctx)
			assert.NoError(t, err)

			result, exists := ctx.Get("condition_result")
			assert.True(t, exists)
			assert.Equal(t, tt.expected, result.(bool))
		})
	}
}

func TestConditionStep_Run_InOperator(t *testing.T) {
	step := NewConditionStep("in_test", "status", "in", []interface{}{"active", "pending", "completed"})

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("status", "active")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))
}

func TestConditionStep_Run_NotInOperator(t *testing.T) {
	step := NewConditionStep("not_in_test", "status", "not_in", []interface{}{"inactive", "deleted"})

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("status", "active")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))
}

func TestConditionStep_Run_NestedField(t *testing.T) {
	step := NewConditionStep("nested_test", "user.profile.status", "equals", "active")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user", map[string]interface{}{
		"profile": map[string]interface{}{
			"status": "active",
		},
	})

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.True(t, result.(bool))
}

func TestConditionStep_Run_NestedFieldMissing(t *testing.T) {
	step := NewConditionStep("nested_missing_test", "user.profile.missing", "exists", nil)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user", map[string]interface{}{
		"profile": map[string]interface{}{
			"status": "active",
		},
	})

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.False(t, result.(bool))
}

func TestConditionStep_Run_UnsupportedOperator(t *testing.T) {
	step := NewConditionStep("unsupported_test", "test_field", "unsupported_op", "value")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("test_field", "value")

	err := step.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported operator")
}

func TestConditionStep_Run_MissingFieldForNonExistenceCheck(t *testing.T) {
	step := NewConditionStep("missing_field_test", "missing_field", "equals", "value")

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	assert.NoError(t, err)

	result, exists := ctx.Get("condition_result")
	assert.True(t, exists)
	assert.False(t, result.(bool))
}

func TestConditionStep_IsEmpty(t *testing.T) {
	step := &ConditionStep{}

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"whitespace", "   ", true},
		{"non-empty string", "hello", false},
		{"empty slice", []interface{}{}, true},
		{"non-empty slice", []interface{}{1}, false},
		{"empty map", map[string]interface{}{}, true},
		{"non-empty map", map[string]interface{}{"key": "value"}, false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero float", 0.0, true},
		{"non-zero float", 3.14, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"other type", struct{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := step.isEmpty(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionStep_NormalizeValue(t *testing.T) {
	step := &ConditionStep{}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"nil", nil, nil},
		{"string number", "42", float64(42)},
		{"string float", "3.14", 3.14},
		{"string bool true", "true", true},
		{"string bool false", "false", false},
		{"regular string", "hello", "hello"},
		{"int", 42, 42},
		{"float", 3.14, 3.14},
		{"bool", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := step.normalizeValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionStep_ToFloat64(t *testing.T) {
	step := &ConditionStep{}

	tests := []struct {
		name     string
		input    interface{}
		expected float64
		shouldOk bool
	}{
		{"int", 42, 42.0, true},
		{"int32", int32(42), 42.0, true},
		{"int64", int64(42), 42.0, true},
		{"float32", float32(3.14), float64(float32(3.14)), true},
		{"float64", 3.14, 3.14, true},
		{"string number", "42.5", 42.5, true},
		{"string non-number", "hello", 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := step.toFloat64(tt.input)
			assert.Equal(t, tt.shouldOk, ok)
			if tt.shouldOk {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConditionStep_ValueInList(t *testing.T) {
	step := &ConditionStep{}

	tests := []struct {
		name     string
		fieldVal interface{}
		listVal  interface{}
		expected bool
	}{
		{"string in list", "apple", []interface{}{"apple", "banana", "cherry"}, true},
		{"string not in list", "grape", []interface{}{"apple", "banana", "cherry"}, false},
		{"int in list", 2, []interface{}{1, 2, 3}, true},
		{"int not in list", 4, []interface{}{1, 2, 3}, false},
		{"not a list", "apple", "not a list", false},
		{"empty list", "apple", []interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := step.valueInList(tt.fieldVal, tt.listVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test helper functions

func TestNewExistsCondition(t *testing.T) {
	step := NewExistsCondition("exists_test", "user_id")

	assert.Equal(t, "exists_test", step.Name())
	assert.Equal(t, "user_id", step.field)
	assert.Equal(t, "exists", step.operator)
	assert.Nil(t, step.value)
}

func TestNewEqualsCondition(t *testing.T) {
	step := NewEqualsCondition("equals_test", "status", "active")

	assert.Equal(t, "equals_test", step.Name())
	assert.Equal(t, "status", step.field)
	assert.Equal(t, "equals", step.operator)
	assert.Equal(t, "active", step.value)
}

func TestNewNotEmptyCondition(t *testing.T) {
	step := NewNotEmptyCondition("not_empty_test", "name")

	assert.Equal(t, "not_empty_test", step.Name())
	assert.Equal(t, "name", step.field)
	assert.Equal(t, "not_empty", step.operator)
	assert.Nil(t, step.value)
}

func TestNewContainsCondition(t *testing.T) {
	step := NewContainsCondition("contains_test", "description", "important")

	assert.Equal(t, "contains_test", step.Name())
	assert.Equal(t, "description", step.field)
	assert.Equal(t, "contains", step.operator)
	assert.Equal(t, "important", step.value)
}

func TestNewInCondition(t *testing.T) {
	values := []interface{}{"active", "pending", "completed"}
	step := NewInCondition("in_test", "status", values)

	assert.Equal(t, "in_test", step.Name())
	assert.Equal(t, "status", step.field)
	assert.Equal(t, "in", step.operator)
	assert.Equal(t, values, step.value)
}

func TestConditionStep_Run_AlternativeOperatorNames(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected bool
	}{
		{"eq operator", "eq", true},
		{"== operator", "==", true},
		{"ne operator", "ne", false},
		{"!= operator", "!=", false},
		{"gt operator", "gt", false},
		{"> operator", ">", false},
		{"gte operator", "gte", true},
		{">= operator", ">=", true},
		{"lt operator", "lt", false},
		{"< operator", "<", false},
		{"lte operator", "lte", true},
		{"<= operator", "<=", true},
		{"!exists operator", "!exists", false},
		{"!empty operator", "!empty", true},
		{"!contains operator", "!contains", false},
		{"!in operator", "!in", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var step *ConditionStep
			if tt.operator == "!in" {
				step = NewConditionStep("test", "test_field", tt.operator, []interface{}{"other"})
			} else {
				step = NewConditionStep("test", "test_field", tt.operator, "expected")
			}

			ctx := flow.NewContext().WithFlowName("test_flow")
			ctx.Set("test_field", "expected")

			err := step.Run(ctx)
			assert.NoError(t, err)

			result, exists := ctx.Get("condition_result")
			assert.True(t, exists)
			assert.Equal(t, tt.expected, result.(bool), "operator: %s", tt.operator)
		})
	}
}
