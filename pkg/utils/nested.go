package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// GetNestedValue extracts nested values from context using dot notation (e.g., "user.profile.name")
func GetNestedValue(key string, ctx *flow.Context) string {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		// Single key, try direct lookup
		if val, ok := ctx.Get(key); ok {
			return fmt.Sprintf("%v", val)
		}
		return fmt.Sprintf("${%s}", key)
	}

	rootValue, ok := ctx.Get(parts[0])
	if !ok {
		return fmt.Sprintf("${%s}", key)
	}

	current := rootValue
	for _, part := range parts[1:] {
		if current == nil {
			return fmt.Sprintf("${%s}", key)
		}

		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return fmt.Sprintf("${%s}", key)
			}
		case map[interface{}]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return fmt.Sprintf("${%s}", key)
			}
		case map[string]string:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return fmt.Sprintf("${%s}", key)
			}
		default:
			return fmt.Sprintf("${%s}", key)
		}
	}

	if current == nil {
		return fmt.Sprintf("${%s}", key)
	}

	return fmt.Sprintf("%v", current)
}

// SetNestedValue sets a nested value in a map using dot notation
func SetNestedValue(data map[string]interface{}, key string, value interface{}) error {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		data[key] = value
		return nil
	}

	current := data
	for i, part := range parts[:len(parts)-1] {
		if existing, ok := current[part]; ok {
			if nested, ok := existing.(map[string]interface{}); ok {
				current = nested
			} else {
				return fmt.Errorf("cannot set nested value: %s is not a map at level %d", part, i)
			}
		} else {
			// Create new nested map
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}

	current[parts[len(parts)-1]] = value
	return nil
}

// HasNestedValue checks if a nested value exists in a map
func HasNestedValue(data map[string]interface{}, key string) bool {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		_, ok := data[key]
		return ok
	}

	var current interface{} = data
	for i, part := range parts {
		if current == nil {
			return false
		}

		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				// If this is the last part, we found it
				if i == len(parts)-1 {
					return true
				}
				// Otherwise, continue traversing
				current = val
			} else {
				return false
			}
		default:
			return false
		}
	}

	return true
}

// FlattenMap flattens a nested map using dot notation for keys
func FlattenMap(data map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	flattenMapRecursive(data, prefix, result)
	return result
}

func flattenMapRecursive(data map[string]interface{}, prefix string, result map[string]interface{}) {
	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flattenMapRecursive(v, newKey, result)
		case map[interface{}]interface{}:
			// Convert to map[string]interface{} and flatten
			converted := make(map[string]interface{})
			for k, val := range v {
				if strKey, ok := k.(string); ok {
					converted[strKey] = val
				}
			}
			flattenMapRecursive(converted, newKey, result)
		default:
			result[newKey] = value
		}
	}
}

// UnflattenMap converts a flattened map back to nested structure
func UnflattenMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if err := SetNestedValue(result, key, value); err != nil {
			// If setting nested value fails, use the key as-is
			result[key] = value
		}
	}

	return result
}

// GetNestedValueFromContext extracts nested values using dot notation from ExecutionContext
func GetNestedValueFromContext(path string, ctx interfaces.ExecutionContext) string {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return ""
	}

	// Get the root value from context
	rootValue, ok := ctx.Get(parts[0])
	if !ok {
		return ""
	}

	// If only one part, return the root value
	if len(parts) == 1 {
		return fmt.Sprintf("%v", rootValue)
	}

	// Navigate through nested structure
	current := rootValue
	for _, part := range parts[1:] {
		current = getNestedValueFromInterface(current, part)
		if current == nil {
			return ""
		}
	}

	return fmt.Sprintf("%v", current)
}

// getNestedValueFromInterface extracts a value from an interface using a key
func getNestedValueFromInterface(data interface{}, key string) interface{} {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case map[string]interface{}:
		return v[key]
	case map[interface{}]interface{}:
		return v[key]
	case map[string]string:
		return v[key]
	default:
		// Try to handle as a struct using reflection
		return getValueFromStruct(data, key)
	}
}

// getValueFromStruct extracts a field value from a struct using reflection
func getValueFromStruct(data interface{}, fieldName string) interface{} {
	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		// Try case-insensitive lookup
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			if strings.EqualFold(t.Field(i).Name, fieldName) {
				field = v.Field(i)
				break
			}
		}
	}

	if !field.IsValid() || !field.CanInterface() {
		return nil
	}

	return field.Interface()
}
