package core

import (
	"strings"
	"testing"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"go.uber.org/zap"
)

func TestNewValueStep(t *testing.T) {
	step := NewValueStep("test_value", "set")

	if step.Name() != "test_value" {
		t.Errorf("Name() = %v, want 'test_value'", step.Name())
	}

	if step.Description() != "Value operation" {
		t.Errorf("Description() = %v, want 'Value operation'", step.Description())
	}

	if step.operation != "set" {
		t.Errorf("operation = %v, want 'set'", step.operation)
	}

	if step.valueType != "string" {
		t.Errorf("valueType = %v, want 'string'", step.valueType)
	}
}

func TestValueStep_WithSource(t *testing.T) {
	step := NewValueStep("test_value", "copy")
	result := step.WithSource("source_field")

	if result != step {
		t.Error("WithSource should return the same instance")
	}

	if step.sourceField != "source_field" {
		t.Errorf("sourceField = %v, want 'source_field'", step.sourceField)
	}
}

func TestValueStep_WithTarget(t *testing.T) {
	step := NewValueStep("test_value", "set")
	result := step.WithTarget("target_field")

	if result != step {
		t.Error("WithTarget should return the same instance")
	}

	if step.targetField != "target_field" {
		t.Errorf("targetField = %v, want 'target_field'", step.targetField)
	}
}

func TestValueStep_WithValue(t *testing.T) {
	step := NewValueStep("test_value", "set")
	result := step.WithValue("test_value")

	if result != step {
		t.Error("WithValue should return the same instance")
	}

	if step.value != "test_value" {
		t.Errorf("value = %v, want 'test_value'", step.value)
	}
}

func TestValueStep_WithValueType(t *testing.T) {
	step := NewValueStep("test_value", "set")
	result := step.WithValueType("int")

	if result != step {
		t.Error("WithValueType should return the same instance")
	}

	if step.valueType != "int" {
		t.Errorf("valueType = %v, want 'int'", step.valueType)
	}
}

func TestValueStep_WithTransform(t *testing.T) {
	step := NewValueStep("test_value", "transform")
	transform := func(v interface{}) interface{} {
		return v
	}
	result := step.WithTransform(transform)

	if result != step {
		t.Error("WithTransform should return the same instance")
	}

	if step.transform == nil {
		t.Error("transform should be set")
	}
}

func TestValueStep_Run_UnsupportedOperation(t *testing.T) {
	step := NewValueStep("test_value", "invalid_operation")
	ctx := flow.NewContext()

	err := step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail with unsupported operation")
	}

	if err.Error() != "unsupported value operation: invalid_operation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestValueStep_HandleSet_Success(t *testing.T) {
	step := NewValueStep("test_value", "set").
		WithTarget("test_field").
		WithValue("test_value")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	value, ok := ctx.Get("test_field")
	if !ok || value != "test_value" {
		t.Error("Value should be set in context")
	}
}

func TestValueStep_HandleSet_NoTarget(t *testing.T) {
	step := NewValueStep("test_value", "set").
		WithValue("test_value")

	ctx := flow.NewContext()
	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail when target field is missing")
	}

	if err.Error() != "target field is required for set operation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestValueStep_HandleSet_WithTypeConversion(t *testing.T) {
	step := NewValueStep("test_value", "set").
		WithTarget("int_field").
		WithValue("123").
		WithValueType("int")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	value, ok := ctx.Get("int_field")
	if !ok {
		t.Error("Value should be set in context")
	}

	intValue, ok := value.(int)
	if !ok || intValue != 123 {
		t.Errorf("Value should be converted to int 123, got %v", value)
	}
}

func TestValueStep_HandleCopy_Success(t *testing.T) {
	step := NewValueStep("test_value", "copy").
		WithSource("source_field").
		WithTarget("target_field")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("source_field", "source_value")

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	value, ok := ctx.Get("target_field")
	if !ok || value != "source_value" {
		t.Error("Value should be copied to target field")
	}

	// Source should still exist
	sourceValue, ok := ctx.Get("source_field")
	if !ok || sourceValue != "source_value" {
		t.Error("Source value should still exist after copy")
	}
}

func TestValueStep_HandleCopy_MissingFields(t *testing.T) {
	step := NewValueStep("test_value", "copy").
		WithSource("source_field")

	ctx := flow.NewContext()
	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail when target field is missing")
	}

	if err.Error() != "both source and target fields are required for copy operation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestValueStep_HandleCopy_MissingSourceValue(t *testing.T) {
	step := NewValueStep("test_value", "copy").
		WithSource("nonexistent_field").
		WithTarget("target_field")

	ctx := flow.NewContext()
	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail when source field doesn't exist")
	}

	if err.Error() != "field 'nonexistent_field' not found in context" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestValueStep_HandleMove_Success(t *testing.T) {
	step := NewValueStep("test_value", "move").
		WithSource("source_field").
		WithTarget("target_field")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("source_field", "source_value")

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	value, ok := ctx.Get("target_field")
	if !ok || value != "source_value" {
		t.Error("Value should be moved to target field")
	}

	// Source should be deleted
	_, ok = ctx.Get("source_field")
	if ok {
		t.Error("Source value should be deleted after move")
	}
}

func TestValueStep_HandleMove_MissingFields(t *testing.T) {
	step := NewValueStep("test_value", "move").
		WithSource("source_field")

	ctx := flow.NewContext()
	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail when target field is missing")
	}

	if err.Error() != "both source and target fields are required for move operation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestValueStep_HandleDelete_Success(t *testing.T) {
	step := NewValueStep("test_value", "delete").
		WithTarget("test_field")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("test_field", "test_value")

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	_, ok := ctx.Get("test_field")
	if ok {
		t.Error("Field should be deleted from context")
	}
}

func TestValueStep_HandleDelete_WithSourceField(t *testing.T) {
	step := NewValueStep("test_value", "delete").
		WithSource("test_field")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("test_field", "test_value")

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	_, ok := ctx.Get("test_field")
	if ok {
		t.Error("Field should be deleted from context")
	}
}

func TestValueStep_HandleDelete_NoField(t *testing.T) {
	step := NewValueStep("test_value", "delete")

	ctx := flow.NewContext()
	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail when no field is specified")
	}

	if err.Error() != "field is required for delete operation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestValueStep_HandleTransform_Success(t *testing.T) {
	transform := func(v interface{}) interface{} {
		if str, ok := v.(string); ok {
			return strings.ToUpper(str)
		}
		return v
	}

	step := NewValueStep("test_value", "transform").
		WithSource("test_field").
		WithTransform(transform)

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("test_field", "hello world")

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	value, ok := ctx.Get("test_field")
	if !ok || value != "HELLO WORLD" {
		t.Error("Value should be transformed")
	}
}

func TestValueStep_HandleTransform_NoTransform(t *testing.T) {
	step := NewValueStep("test_value", "transform").
		WithSource("test_field")

	ctx := flow.NewContext()
	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail when transform function is missing")
	}

	if err.Error() != "transform function is required for transform operation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestValueStep_ConvertValueType_String(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("string")

	result, err := step.convertValueType(123)
	if err != nil {
		t.Errorf("convertValueType failed: %v", err)
	}

	if result != "123" {
		t.Errorf("convertValueType = %v, want '123'", result)
	}
}

func TestValueStep_ConvertValueType_Int(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("int")

	result, err := step.convertValueType("123")
	if err != nil {
		t.Errorf("convertValueType failed: %v", err)
	}

	if result != 123 {
		t.Errorf("convertValueType = %v, want 123", result)
	}
}

func TestValueStep_ConvertValueType_Float(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("float")

	result, err := step.convertValueType("123.45")
	if err != nil {
		t.Errorf("convertValueType failed: %v", err)
	}

	if result != 123.45 {
		t.Errorf("convertValueType = %v, want 123.45", result)
	}
}

func TestValueStep_ConvertValueType_Bool(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("bool")

	result, err := step.convertValueType("true")
	if err != nil {
		t.Errorf("convertValueType failed: %v", err)
	}

	if result != true {
		t.Errorf("convertValueType = %v, want true", result)
	}
}

func TestValueStep_ConvertValueType_InvalidInt(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("int")

	_, err := step.convertValueType("invalid")
	if err == nil {
		t.Error("convertValueType should fail with invalid int")
	}
}

func TestValueStep_ConvertValueType_InvalidFloat(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("float")

	_, err := step.convertValueType("invalid")
	if err == nil {
		t.Error("convertValueType should fail with invalid float")
	}
}

func TestValueStep_ConvertValueType_InvalidBool(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("bool")

	_, err := step.convertValueType("invalid")
	if err == nil {
		t.Error("convertValueType should fail with invalid bool")
	}
}

func TestValueStep_ConvertValueType_UnsupportedType(t *testing.T) {
	step := NewValueStep("test_value", "set").WithValueType("unsupported")

	_, err := step.convertValueType("value")
	if err == nil {
		t.Error("convertValueType should fail with unsupported type")
	}

	if err.Error() != "unsupported value type: unsupported" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

// Test helper functions

func TestNewSetValueStep(t *testing.T) {
	step := NewSetValueStep("set_test", "target_field", "test_value")

	if step.Name() != "set_test" {
		t.Errorf("Name() = %v, want 'set_test'", step.Name())
	}

	if step.operation != "set" {
		t.Errorf("operation = %v, want 'set'", step.operation)
	}

	if step.targetField != "target_field" {
		t.Errorf("targetField = %v, want 'target_field'", step.targetField)
	}

	if step.value != "test_value" {
		t.Errorf("value = %v, want 'test_value'", step.value)
	}
}

func TestNewCopyValueStep(t *testing.T) {
	step := NewCopyValueStep("copy_test", "source_field", "target_field")

	if step.Name() != "copy_test" {
		t.Errorf("Name() = %v, want 'copy_test'", step.Name())
	}

	if step.operation != "copy" {
		t.Errorf("operation = %v, want 'copy'", step.operation)
	}

	if step.sourceField != "source_field" {
		t.Errorf("sourceField = %v, want 'source_field'", step.sourceField)
	}

	if step.targetField != "target_field" {
		t.Errorf("targetField = %v, want 'target_field'", step.targetField)
	}
}

func TestNewMoveValueStep(t *testing.T) {
	step := NewMoveValueStep("move_test", "source_field", "target_field")

	if step.Name() != "move_test" {
		t.Errorf("Name() = %v, want 'move_test'", step.Name())
	}

	if step.operation != "move" {
		t.Errorf("operation = %v, want 'move'", step.operation)
	}

	if step.sourceField != "source_field" {
		t.Errorf("sourceField = %v, want 'source_field'", step.sourceField)
	}

	if step.targetField != "target_field" {
		t.Errorf("targetField = %v, want 'target_field'", step.targetField)
	}
}

func TestNewDeleteValueStep(t *testing.T) {
	step := NewDeleteValueStep("delete_test", "test_field")

	if step.Name() != "delete_test" {
		t.Errorf("Name() = %v, want 'delete_test'", step.Name())
	}

	if step.operation != "delete" {
		t.Errorf("operation = %v, want 'delete'", step.operation)
	}

	if step.targetField != "test_field" {
		t.Errorf("targetField = %v, want 'test_field'", step.targetField)
	}
}

func TestNewTransformValueStep(t *testing.T) {
	transform := func(v interface{}) interface{} {
		return v
	}
	step := NewTransformValueStep("transform_test", "source_field", transform)

	if step.Name() != "transform_test" {
		t.Errorf("Name() = %v, want 'transform_test'", step.Name())
	}

	if step.operation != "transform" {
		t.Errorf("operation = %v, want 'transform'", step.operation)
	}

	if step.sourceField != "source_field" {
		t.Errorf("sourceField = %v, want 'source_field'", step.sourceField)
	}

	if step.transform == nil {
		t.Error("transform should be set")
	}
}

func TestNewStringValueStep(t *testing.T) {
	step := NewStringValueStep("string_test", "target_field", "test_value")

	if step.Name() != "string_test" {
		t.Errorf("Name() = %v, want 'string_test'", step.Name())
	}

	if step.operation != "set" {
		t.Errorf("operation = %v, want 'set'", step.operation)
	}

	if step.targetField != "target_field" {
		t.Errorf("targetField = %v, want 'target_field'", step.targetField)
	}

	if step.value != "test_value" {
		t.Errorf("value = %v, want 'test_value'", step.value)
	}

	if step.valueType != "string" {
		t.Errorf("valueType = %v, want 'string'", step.valueType)
	}
}

func TestNewIntValueStep(t *testing.T) {
	step := NewIntValueStep("int_test", "target_field", 123)

	if step.Name() != "int_test" {
		t.Errorf("Name() = %v, want 'int_test'", step.Name())
	}

	if step.operation != "set" {
		t.Errorf("operation = %v, want 'set'", step.operation)
	}

	if step.targetField != "target_field" {
		t.Errorf("targetField = %v, want 'target_field'", step.targetField)
	}

	if step.value != 123 {
		t.Errorf("value = %v, want 123", step.value)
	}

	if step.valueType != "int" {
		t.Errorf("valueType = %v, want 'int'", step.valueType)
	}
}

func TestNewBoolValueStep(t *testing.T) {
	step := NewBoolValueStep("bool_test", "target_field", true)

	if step.Name() != "bool_test" {
		t.Errorf("Name() = %v, want 'bool_test'", step.Name())
	}

	if step.operation != "set" {
		t.Errorf("operation = %v, want 'set'", step.operation)
	}

	if step.targetField != "target_field" {
		t.Errorf("targetField = %v, want 'target_field'", step.targetField)
	}

	if step.value != true {
		t.Errorf("value = %v, want true", step.value)
	}

	if step.valueType != "bool" {
		t.Errorf("valueType = %v, want 'bool'", step.valueType)
	}
}
