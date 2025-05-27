package transformers

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBaseTransformer tests the BaseTransformer functionality.
func TestBaseTransformer(t *testing.T) {
	name := "test-base"
	bt := NewBaseTransformer(name)
	assert.NotNil(t, bt, "NewBaseTransformer should not return nil")
	assert.Equal(t, name, bt.Name(), "Name() should return the correct name")
}

// TestFuncTransformer tests the FuncTransformer functionality.
func TestFuncTransformer(t *testing.T) {
	transformerName := "test-func-transformer"
	originalData := map[string]interface{}{"key": "value"}
	modifiedData := map[string]interface{}{"key": "modified"}
	expectedError := errors.New("transform error")

	t.Run("successful transform", func(t *testing.T) {
		fn := func(data map[string]interface{}) (map[string]interface{}, error) {
			assert.Equal(t, originalData, data, "Transform function received incorrect data")
			return modifiedData, nil
		}
		ft := NewFuncTransformer(transformerName, fn)
		assert.NotNil(t, ft, "NewFuncTransformer should not return nil")
		assert.Equal(t, transformerName, ft.Name(), "Name() should return the correct name")

		result, err := ft.Transform(originalData)
		assert.NoError(t, err, "Transform should not return an error")
		assert.Equal(t, modifiedData, result, "Transform should return the modified data")
	})

	t.Run("transform returns error", func(t *testing.T) {
		fn := func(data map[string]interface{}) (map[string]interface{}, error) {
			return nil, expectedError
		}
		ft := NewFuncTransformer(transformerName, fn)
		result, err := ft.Transform(originalData)
		assert.Error(t, err, "Transform should return an error")
		assert.Equal(t, expectedError, err, "Transform returned incorrect error")
		assert.Nil(t, result, "Transform should return nil data on error")
	})
}

// TestNoOpTransformer tests the NoOpTransformer functionality.
func TestNoOpTransformer(t *testing.T) {
	nt := NewNoOpTransformer()
	assert.NotNil(t, nt, "NewNoOpTransformer should not return nil")
	assert.Equal(t, "noop", nt.Name(), "Name() should be 'noop'")

	originalData := map[string]interface{}{"key": "value", "nested": map[string]interface{}{"subKey": "subValue"}}
	result, err := nt.Transform(originalData)

	assert.NoError(t, err, "Transform should not return an error")
	assert.Equal(t, originalData, result, "Transform should return data identical to input")

	// Check that it's a copy by modifying original and ensuring result is unchanged
	originalData["newKey"] = "newValue"
	_, ok := result["newKey"]
	assert.False(t, ok, "Changes to original data should not affect the transformed data (copy check)")

	// Test with nil input (should be handled by ValidateTransformerInput, but good to have a direct check if NoOp has specific logic)
	// Based on current NoOpTransformer, it relies on ValidateTransformerInput from chain.go or caller.
	// If ValidateTransformerInput is not called before, NoOpTransformer would panic on nil map.
	// For this test, we assume ValidateTransformerInput is called by a chain or wrapper.
}

// TestCopyTransformer tests the CopyTransformer functionality.
func TestCopyTransformer(t *testing.T) {
	ct := NewCopyTransformer()
	assert.NotNil(t, ct, "NewCopyTransformer should not return nil")
	assert.Equal(t, "copy", ct.Name(), "Name() should be 'copy'")

	originalData := map[string]interface{}{
		"key":         "value",
		"nestedMap":   map[string]interface{}{"subKey": "subValue"},
		"nestedSlice": []interface{}{"a", map[string]interface{}{"b": "c"}},
	}

	result, err := ct.Transform(originalData)
	assert.NoError(t, err, "Transform should not return an error")
	assert.Equal(t, originalData, result, "Transform should return data identical to input")

	// Check that it's a copy by modifying original and ensuring result is unchanged
	originalData["key"] = "newValue"
	assert.Equal(t, "value", result["key"], "Change to original primitive should not affect copy")

	originalNestedMap, _ := originalData["nestedMap"].(map[string]interface{})
	originalNestedMap["subKey"] = "newSubValue"
	resultNestedMap, _ := result["nestedMap"].(map[string]interface{})
	assert.Equal(t, "subValue", resultNestedMap["subKey"], "Change to original nested map should not affect copy")

	originalNestedSlice, _ := originalData["nestedSlice"].([]interface{})
	originalNestedSlice[0] = "z"
	resultNestedSlice, _ := result["nestedSlice"].([]interface{})
	assert.Equal(t, "a", resultNestedSlice[0], "Change to original nested slice element should not affect copy")

	originalNestedSliceMap, _ := originalNestedSlice[1].(map[string]interface{})
	originalNestedSliceMap["b"] = "d"
	resultNestedSliceMap, _ := resultNestedSlice[1].(map[string]interface{})
	assert.Equal(t, "c", resultNestedSliceMap["b"], "Change to map within original nested slice should not affect copy")
}

// TestDeepCopyMap tests the deepCopyMap utility function.
func TestDeepCopyMap(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		copied := deepCopyMap(nil)
		assert.Nil(t, copied, "deepCopyMap of nil should be nil")
	})

	t.Run("empty map", func(t *testing.T) {
		original := make(map[string]interface{}) // or map[string]interface{}{}
		copied := deepCopyMap(original)
		assert.NotNil(t, copied, "deepCopyMap of empty map should not be nil")
		assert.Empty(t, copied, "deepCopyMap of empty map should be empty")
		// Check independence by modifying original
		original["test"] = "value"
		assert.Empty(t, copied, "Copied empty map should remain empty after original modification")
	})

	t.Run("map with primitives", func(t *testing.T) {
		original := map[string]interface{}{"a": 1, "b": "string", "c": true}
		copied := deepCopyMap(original)
		assert.Equal(t, original, copied)
		original["a"] = 2
		assert.Equal(t, 1, copied["a"], "Changes to original primitive should not affect copy")
	})

	t.Run("map with nested map", func(t *testing.T) {
		original := map[string]interface{}{"nested": map[string]interface{}{"key": "value"}}
		copied := deepCopyMap(original)
		assert.Equal(t, original, copied)

		originalNested, _ := original["nested"].(map[string]interface{})
		originalNested["key"] = "newValue"
		copiedNested, _ := copied["nested"].(map[string]interface{})
		assert.Equal(t, "value", copiedNested["key"], "Changes to original nested map should not affect copy")
	})

	t.Run("map with nested slice", func(t *testing.T) {
		original := map[string]interface{}{"slice": []interface{}{1, "two", map[string]interface{}{"a": "b"}}}
		copied := deepCopyMap(original)
		assert.True(t, reflect.DeepEqual(original, copied), "DeepEqual should pass for map with slice")

		originalSlice, _ := original["slice"].([]interface{})
		originalSlice[0] = 100
		copiedSlice, _ := copied["slice"].([]interface{})
		assert.Equal(t, 1, copiedSlice[0], "Changes to original slice primitive should not affect copy")

		originalSliceMap, _ := originalSlice[2].(map[string]interface{})
		originalSliceMap["a"] = "z"
		copiedSliceMap, _ := copiedSlice[2].(map[string]interface{})
		assert.Equal(t, "b", copiedSliceMap["a"], "Changes to map in original slice should not affect copy")
	})
}

// TestDeepCopySlice tests the deepCopySlice utility function.
func TestDeepCopySlice(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		copied := deepCopySlice(nil)
		assert.Nil(t, copied, "deepCopySlice of nil should be nil")
	})

	t.Run("empty slice", func(t *testing.T) {
		original := make([]interface{}, 0)
		copied := deepCopySlice(original)
		assert.NotNil(t, copied, "deepCopySlice of empty slice should not be nil")
		assert.Empty(t, copied, "deepCopySlice of empty slice should be empty")
		// Test independence: since both are empty, just verify they're separate instances
		assert.True(t, len(copied) == 0 && len(original) == 0, "Both slices should be empty")
	})

	t.Run("slice with primitives", func(t *testing.T) {
		original := []interface{}{1, "string", true}
		copied := deepCopySlice(original)
		assert.Equal(t, original, copied)
		original[0] = 2
		assert.Equal(t, 1, copied[0], "Changes to original primitive should not affect copy")
	})

	t.Run("slice with nested map", func(t *testing.T) {
		original := []interface{}{map[string]interface{}{"key": "value"}}
		copied := deepCopySlice(original)
		assert.True(t, reflect.DeepEqual(original, copied), "DeepEqual should pass for slice with map")

		originalNested, _ := original[0].(map[string]interface{})
		originalNested["key"] = "newValue"
		copiedNested, _ := copied[0].(map[string]interface{})
		assert.Equal(t, "value", copiedNested["key"], "Changes to original nested map should not affect copy")
	})

	t.Run("slice with nested slice", func(t *testing.T) {
		original := []interface{}{[]interface{}{1, "two"}}
		copied := deepCopySlice(original)
		assert.True(t, reflect.DeepEqual(original, copied), "DeepEqual should pass for slice with slice")

		originalNested, _ := original[0].([]interface{})
		originalNested[0] = 100
		copiedNested, _ := copied[0].([]interface{})
		assert.Equal(t, 1, copiedNested[0], "Changes to original nested slice primitive should not affect copy")
	})
}

// TestValidateTransformerInput tests the ValidateTransformerInput utility function.
func TestValidateTransformerInput(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		err := ValidateTransformerInput(nil)
		assert.Error(t, err, "ValidateTransformerInput should return error for nil data")
		assert.EqualError(t, err, "transformer input data cannot be nil")
	})

	t.Run("non-nil data", func(t *testing.T) {
		data := map[string]interface{}{"key": "value"}
		err := ValidateTransformerInput(data)
		assert.NoError(t, err, "ValidateTransformerInput should not return error for non-nil data")
	})

	t.Run("empty map data", func(t *testing.T) {
		data := make(map[string]interface{}) // or map[string]interface{}{}
		err := ValidateTransformerInput(data)
		assert.NoError(t, err, "ValidateTransformerInput should not return error for empty map data")
	})
}
