package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
	"go.uber.org/zap"
)

// ConditionStep evaluates conditions and controls flow execution
type ConditionStep struct {
	*base.BaseStep
	field     string
	operator  string
	value     interface{}
	condition func(*flow.Context) bool
}

// NewConditionStep creates a new condition step
func NewConditionStep(name, field, operator string, value interface{}) *ConditionStep {
	return &ConditionStep{
		BaseStep: base.NewBaseStep(name, "Condition evaluation"),
		field:    field,
		operator: operator,
		value:    value,
	}
}

// WithCustomCondition sets a custom condition function
func (cs *ConditionStep) WithCustomCondition(condition func(*flow.Context) bool) *ConditionStep {
	cs.condition = condition
	return cs
}

func (cs *ConditionStep) Run(ctx *flow.Context) error {
	var result bool
	var err error

	if cs.condition != nil {
		// Use custom condition
		result = cs.condition(ctx)
		ctx.Logger().Info("Custom condition evaluated",
			zap.String("step", cs.Name()),
			zap.Bool("result", result))
	} else {
		// Use field-based condition
		result, err = cs.evaluateFieldCondition(ctx)
		if err != nil {
			ctx.Logger().Error("Condition evaluation failed",
				zap.String("step", cs.Name()),
				zap.String("field", cs.field),
				zap.String("operator", cs.operator),
				zap.Error(err))
			return err
		}

		ctx.Logger().Info("Field condition evaluated",
			zap.String("step", cs.Name()),
			zap.String("field", cs.field),
			zap.String("operator", cs.operator),
			zap.Any("expected_value", cs.value),
			zap.Bool("result", result))
	}

	// Store condition result
	ctx.Set("condition_result", result)
	ctx.Set(fmt.Sprintf("condition_%s", cs.Name()), result)

	return nil
}

func (cs *ConditionStep) evaluateFieldCondition(ctx *flow.Context) (bool, error) {
	// Get field value from context
	var fieldValue interface{}
	var exists bool

	if strings.Contains(cs.field, ".") {
		// Handle nested field access
		fieldValueStr := utils.GetNestedValue(cs.field, ctx)
		// If it returns the placeholder format, field doesn't exist
		if strings.HasPrefix(fieldValueStr, "${") && strings.HasSuffix(fieldValueStr, "}") {
			exists = false
		} else {
			exists = true
			fieldValue = fieldValueStr
		}
	} else {
		fieldValue, exists = ctx.Get(cs.field)
	}

	// Evaluate based on operator
	switch strings.ToLower(cs.operator) {
	case "exists":
		return exists, nil
	case "not_exists", "!exists":
		return !exists, nil
	case "empty":
		return !exists || cs.isEmpty(fieldValue), nil
	case "not_empty", "!empty":
		return exists && !cs.isEmpty(fieldValue), nil
	case "equals", "eq", "==":
		return exists && cs.compareValues(fieldValue, cs.value, "eq"), nil
	case "not_equals", "ne", "!=":
		return !exists || !cs.compareValues(fieldValue, cs.value, "eq"), nil
	case "greater_than", "gt", ">":
		return exists && cs.compareValues(fieldValue, cs.value, "gt"), nil
	case "greater_equal", "gte", ">=":
		return exists && cs.compareValues(fieldValue, cs.value, "gte"), nil
	case "less_than", "lt", "<":
		return exists && cs.compareValues(fieldValue, cs.value, "lt"), nil
	case "less_equal", "lte", "<=":
		return exists && cs.compareValues(fieldValue, cs.value, "lte"), nil
	case "contains":
		return exists && cs.stringContains(fieldValue, cs.value), nil
	case "not_contains", "!contains":
		return !exists || !cs.stringContains(fieldValue, cs.value), nil
	case "starts_with":
		return exists && cs.stringStartsWith(fieldValue, cs.value), nil
	case "ends_with":
		return exists && cs.stringEndsWith(fieldValue, cs.value), nil
	case "in":
		return exists && cs.valueInList(fieldValue, cs.value), nil
	case "not_in", "!in":
		return !exists || !cs.valueInList(fieldValue, cs.value), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", cs.operator)
	}
}

func (cs *ConditionStep) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
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

func (cs *ConditionStep) compareValues(fieldValue, expectedValue interface{}, operator string) bool {
	// Convert both values to comparable types
	fVal := cs.normalizeValue(fieldValue)
	eVal := cs.normalizeValue(expectedValue)

	// Handle string comparisons
	if fStr, fOk := fVal.(string); fOk {
		if eStr, eOk := eVal.(string); eOk {
			switch operator {
			case "eq":
				return fStr == eStr
			case "gt":
				return fStr > eStr
			case "gte":
				return fStr >= eStr
			case "lt":
				return fStr < eStr
			case "lte":
				return fStr <= eStr
			}
		}
	}

	// Handle numeric comparisons
	if fNum, fOk := cs.toFloat64(fVal); fOk {
		if eNum, eOk := cs.toFloat64(eVal); eOk {
			switch operator {
			case "eq":
				return fNum == eNum
			case "gt":
				return fNum > eNum
			case "gte":
				return fNum >= eNum
			case "lt":
				return fNum < eNum
			case "lte":
				return fNum <= eNum
			}
		}
	}

	// Handle boolean comparisons
	if fBool, fOk := fVal.(bool); fOk {
		if eBool, eOk := eVal.(bool); eOk {
			return fBool == eBool
		}
	}

	// Fallback to reflection-based comparison
	return reflect.DeepEqual(fVal, eVal)
}

func (cs *ConditionStep) normalizeValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		// Try to parse as number
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
		// Try to parse as bool
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
		return v
	default:
		return v
	}
}

func (cs *ConditionStep) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (cs *ConditionStep) stringContains(fieldValue, expectedValue interface{}) bool {
	fStr := fmt.Sprintf("%v", fieldValue)
	eStr := fmt.Sprintf("%v", expectedValue)
	return strings.Contains(fStr, eStr)
}

func (cs *ConditionStep) stringStartsWith(fieldValue, expectedValue interface{}) bool {
	fStr := fmt.Sprintf("%v", fieldValue)
	eStr := fmt.Sprintf("%v", expectedValue)
	return strings.HasPrefix(fStr, eStr)
}

func (cs *ConditionStep) stringEndsWith(fieldValue, expectedValue interface{}) bool {
	fStr := fmt.Sprintf("%v", fieldValue)
	eStr := fmt.Sprintf("%v", expectedValue)
	return strings.HasSuffix(fStr, eStr)
}

func (cs *ConditionStep) valueInList(fieldValue, expectedValue interface{}) bool {
	// Expected value should be a slice
	expectedSlice, ok := expectedValue.([]interface{})
	if !ok {
		return false
	}

	for _, item := range expectedSlice {
		if cs.compareValues(fieldValue, item, "eq") {
			return true
		}
	}

	return false
}

// Helper functions for creating common conditions

// NewExistsCondition creates a condition that checks if a field exists
func NewExistsCondition(name, field string) *ConditionStep {
	return NewConditionStep(name, field, "exists", nil)
}

// NewEqualsCondition creates a condition that checks if a field equals a value
func NewEqualsCondition(name, field string, value interface{}) *ConditionStep {
	return NewConditionStep(name, field, "equals", value)
}

// NewNotEmptyCondition creates a condition that checks if a field is not empty
func NewNotEmptyCondition(name, field string) *ConditionStep {
	return NewConditionStep(name, field, "not_empty", nil)
}

// NewContainsCondition creates a condition that checks if a field contains a value
func NewContainsCondition(name, field string, value interface{}) *ConditionStep {
	return NewConditionStep(name, field, "contains", value)
}

// NewInCondition creates a condition that checks if a field value is in a list
func NewInCondition(name, field string, values []interface{}) *ConditionStep {
	return NewConditionStep(name, field, "in", values)
}
