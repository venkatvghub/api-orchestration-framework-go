package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
	"go.uber.org/zap"
)

// ValueStep sets or manipulates values in the context
type ValueStep struct {
	*base.BaseStep
	operation   string // set, copy, move, delete, transform
	sourceField string
	targetField string
	value       interface{}
	valueType   string // string, int, float, bool, json
	transform   func(interface{}) interface{}
}

// NewValueStep creates a new value step
func NewValueStep(name, operation string) *ValueStep {
	return &ValueStep{
		BaseStep:  base.NewBaseStep(name, "Value operation"),
		operation: operation,
		valueType: "string",
	}
}

// WithSource sets the source field for copy/move operations
func (vs *ValueStep) WithSource(field string) *ValueStep {
	vs.sourceField = field
	return vs
}

// WithTarget sets the target field for set/copy/move operations
func (vs *ValueStep) WithTarget(field string) *ValueStep {
	vs.targetField = field
	return vs
}

// WithValue sets the value for set operations
func (vs *ValueStep) WithValue(value interface{}) *ValueStep {
	vs.value = value
	return vs
}

// WithValueType sets the type for value conversion
func (vs *ValueStep) WithValueType(valueType string) *ValueStep {
	vs.valueType = valueType
	return vs
}

// WithTransform sets a custom transformation function
func (vs *ValueStep) WithTransform(transform func(interface{}) interface{}) *ValueStep {
	vs.transform = transform
	return vs
}

func (vs *ValueStep) Run(ctx *flow.Context) error {
	switch vs.operation {
	case "set":
		return vs.handleSet(ctx)
	case "copy":
		return vs.handleCopy(ctx)
	case "move":
		return vs.handleMove(ctx)
	case "delete":
		return vs.handleDelete(ctx)
	case "transform":
		return vs.handleTransform(ctx)
	default:
		return fmt.Errorf("unsupported value operation: %s", vs.operation)
	}
}

func (vs *ValueStep) handleSet(ctx *flow.Context) error {
	if vs.targetField == "" {
		return fmt.Errorf("target field is required for set operation")
	}

	// Process the value
	processedValue, err := vs.processValue(ctx, vs.value)
	if err != nil {
		ctx.Logger().Error("Failed to process value",
			zap.String("step", vs.Name()),
			zap.String("target_field", vs.targetField),
			zap.Error(err))
		return err
	}

	// Set the value in context
	if strings.Contains(vs.targetField, ".") {
		// Handle nested field setting
		return vs.setNestedValue(ctx, vs.targetField, processedValue)
	} else {
		ctx.Set(vs.targetField, processedValue)
	}

	ctx.Logger().Info("Value set",
		zap.String("step", vs.Name()),
		zap.String("target_field", vs.targetField),
		zap.String("value_type", vs.valueType),
		zap.Any("value", processedValue))

	return nil
}

func (vs *ValueStep) handleCopy(ctx *flow.Context) error {
	if vs.sourceField == "" || vs.targetField == "" {
		return fmt.Errorf("both source and target fields are required for copy operation")
	}

	// Get source value
	sourceValue, err := vs.getFieldValue(ctx, vs.sourceField)
	if err != nil {
		return err
	}

	// Process the value
	processedValue, err := vs.processValue(ctx, sourceValue)
	if err != nil {
		return err
	}

	// Set target value
	if strings.Contains(vs.targetField, ".") {
		return vs.setNestedValue(ctx, vs.targetField, processedValue)
	} else {
		ctx.Set(vs.targetField, processedValue)
	}

	ctx.Logger().Info("Value copied",
		zap.String("step", vs.Name()),
		zap.String("source_field", vs.sourceField),
		zap.String("target_field", vs.targetField))

	return nil
}

func (vs *ValueStep) handleMove(ctx *flow.Context) error {
	if vs.sourceField == "" || vs.targetField == "" {
		return fmt.Errorf("both source and target fields are required for move operation")
	}

	// Get source value
	sourceValue, err := vs.getFieldValue(ctx, vs.sourceField)
	if err != nil {
		return err
	}

	// Process the value
	processedValue, err := vs.processValue(ctx, sourceValue)
	if err != nil {
		return err
	}

	// Set target value
	if strings.Contains(vs.targetField, ".") {
		if err := vs.setNestedValue(ctx, vs.targetField, processedValue); err != nil {
			return err
		}
	} else {
		ctx.Set(vs.targetField, processedValue)
	}

	// Delete source value
	if strings.Contains(vs.sourceField, ".") {
		// For nested fields, we can't easily delete, so we'll skip
		ctx.Logger().Warn("Cannot delete nested source field",
			zap.String("step", vs.Name()),
			zap.String("source_field", vs.sourceField))
	} else {
		ctx.Delete(vs.sourceField)
	}

	ctx.Logger().Info("Value moved",
		zap.String("step", vs.Name()),
		zap.String("source_field", vs.sourceField),
		zap.String("target_field", vs.targetField))

	return nil
}

func (vs *ValueStep) handleDelete(ctx *flow.Context) error {
	field := vs.targetField
	if field == "" {
		field = vs.sourceField
	}

	if field == "" {
		return fmt.Errorf("field is required for delete operation")
	}

	if strings.Contains(field, ".") {
		ctx.Logger().Warn("Cannot delete nested field",
			zap.String("step", vs.Name()),
			zap.String("field", field))
		return fmt.Errorf("cannot delete nested field: %s", field)
	}

	// Check if field exists before deletion
	_, existed := ctx.Get(field)
	ctx.Delete(field)

	ctx.Logger().Info("Value deleted",
		zap.String("step", vs.Name()),
		zap.String("field", field),
		zap.Bool("existed", existed))

	return nil
}

func (vs *ValueStep) handleTransform(ctx *flow.Context) error {
	if vs.sourceField == "" {
		return fmt.Errorf("source field is required for transform operation")
	}

	if vs.transform == nil {
		return fmt.Errorf("transform function is required for transform operation")
	}

	// Get source value
	sourceValue, err := vs.getFieldValue(ctx, vs.sourceField)
	if err != nil {
		return err
	}

	// Apply transformation
	transformedValue := vs.transform(sourceValue)

	// Process the transformed value
	processedValue, err := vs.processValue(ctx, transformedValue)
	if err != nil {
		return err
	}

	// Set target value (or overwrite source if no target specified)
	targetField := vs.targetField
	if targetField == "" {
		targetField = vs.sourceField
	}

	if strings.Contains(targetField, ".") {
		return vs.setNestedValue(ctx, targetField, processedValue)
	} else {
		ctx.Set(targetField, processedValue)
	}

	ctx.Logger().Info("Value transformed",
		zap.String("step", vs.Name()),
		zap.String("source_field", vs.sourceField),
		zap.String("target_field", targetField))

	return nil
}

func (vs *ValueStep) getFieldValue(ctx *flow.Context, field string) (interface{}, error) {
	if strings.Contains(field, ".") {
		// Handle nested field access
		value := utils.GetNestedValue(field, ctx)
		// Check if it's a placeholder (field not found)
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			return nil, fmt.Errorf("field '%s' not found in context", field)
		}
		return value, nil
	} else {
		value, exists := ctx.Get(field)
		if !exists {
			return nil, fmt.Errorf("field '%s' not found in context", field)
		}
		return value, nil
	}
}

func (vs *ValueStep) setNestedValue(ctx *flow.Context, field string, value interface{}) error {
	// For nested field setting, we need to get the root object and modify it
	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		ctx.Set(field, value)
		return nil
	}

	rootField := parts[0]
	rootValue, exists := ctx.Get(rootField)
	if !exists {
		// Create new nested structure
		rootValue = make(map[string]interface{})
		ctx.Set(rootField, rootValue)
	}

	rootMap, ok := rootValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot set nested value: %s is not a map", rootField)
	}

	// Use utility function to set nested value
	nestedKey := strings.Join(parts[1:], ".")
	return utils.SetNestedValue(rootMap, nestedKey, value)
}

func (vs *ValueStep) processValue(ctx *flow.Context, value interface{}) (interface{}, error) {
	// Handle string interpolation
	if strValue, ok := value.(string); ok {
		interpolated, err := utils.InterpolateString(strValue, ctx)
		if err != nil {
			return nil, err
		}
		value = interpolated
	}

	// Convert to specified type
	return vs.convertValueType(value)
}

func (vs *ValueStep) convertValueType(value interface{}) (interface{}, error) {
	strValue := fmt.Sprintf("%v", value)

	switch vs.valueType {
	case "string":
		return strValue, nil
	case "int":
		return strconv.Atoi(strValue)
	case "float":
		return strconv.ParseFloat(strValue, 64)
	case "bool":
		return strconv.ParseBool(strValue)
	case "json":
		// Keep as-is for JSON (would need JSON parsing in real implementation)
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %s", vs.valueType)
	}
}

// Helper functions for creating common value operations

// NewSetValueStep creates a step that sets a value
func NewSetValueStep(name, targetField string, value interface{}) *ValueStep {
	return NewValueStep(name, "set").
		WithTarget(targetField).
		WithValue(value)
}

// NewCopyValueStep creates a step that copies a value
func NewCopyValueStep(name, sourceField, targetField string) *ValueStep {
	return NewValueStep(name, "copy").
		WithSource(sourceField).
		WithTarget(targetField)
}

// NewMoveValueStep creates a step that moves a value
func NewMoveValueStep(name, sourceField, targetField string) *ValueStep {
	return NewValueStep(name, "move").
		WithSource(sourceField).
		WithTarget(targetField)
}

// NewDeleteValueStep creates a step that deletes a value
func NewDeleteValueStep(name, field string) *ValueStep {
	return NewValueStep(name, "delete").
		WithTarget(field)
}

// NewTransformValueStep creates a step that transforms a value
func NewTransformValueStep(name, sourceField string, transform func(interface{}) interface{}) *ValueStep {
	return NewValueStep(name, "transform").
		WithSource(sourceField).
		WithTransform(transform)
}

// NewStringValueStep creates a step that sets a string value
func NewStringValueStep(name, targetField, value string) *ValueStep {
	return NewValueStep(name, "set").
		WithTarget(targetField).
		WithValue(value).
		WithValueType("string")
}

// NewIntValueStep creates a step that sets an integer value
func NewIntValueStep(name, targetField string, value int) *ValueStep {
	return NewValueStep(name, "set").
		WithTarget(targetField).
		WithValue(value).
		WithValueType("int")
}

// NewBoolValueStep creates a step that sets a boolean value
func NewBoolValueStep(name, targetField string, value bool) *ValueStep {
	return NewValueStep(name, "set").
		WithTarget(targetField).
		WithValue(value).
		WithValueType("bool")
}
