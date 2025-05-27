package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldTransformer(t *testing.T) {
	baseData := map[string]interface{}{
		"id":          123,
		"name":        "Test Item",
		"description": "A description for testing",
		"price":       99.99,
		"active":      true,
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"_meta": map[string]interface{}{
			"version": "1.0",
			"source":  "test",
		},
	}

	t.Run("NewFieldTransformer", func(t *testing.T) {
		ft := NewFieldTransformer("testField", []string{"id", "name"})
		assert.NotNil(t, ft)
		assert.Equal(t, "testField", ft.Name())
		assert.Equal(t, []string{"id", "name"}, ft.Fields)
		assert.True(t, ft.IncludeMeta, "IncludeMeta should default to true")
		assert.False(t, ft.Flatten, "Flatten should default to false")
		assert.Equal(t, "_", ft.Separator, "Separator should default to '_' ")
	})

	t.Run("WithPrefix", func(t *testing.T) {
		ft := NewFieldTransformer("test", nil).WithPrefix("item")
		assert.Equal(t, "item", ft.Prefix)
	})

	t.Run("WithMeta", func(t *testing.T) {
		ft := NewFieldTransformer("test", nil).WithMeta(false)
		assert.False(t, ft.IncludeMeta)
	})

	t.Run("WithFlatten", func(t *testing.T) {
		ft := NewFieldTransformer("test", nil).WithFlatten(true)
		assert.True(t, ft.Flatten)
	})

	t.Run("WithSeparator", func(t *testing.T) {
		ft := NewFieldTransformer("test", nil).WithSeparator(".")
		assert.Equal(t, ".", ft.Separator)
	})

	t.Run("Transform_NilInput", func(t *testing.T) {
		ft := NewFieldTransformer("test", []string{"id"})
		_, err := ft.Transform(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field transformer validation failed: transformer input data cannot be nil")
	})

	t.Run("Transform_SelectSpecificFields_NoMeta", func(t *testing.T) {
		ft := NewFieldTransformer("selectFields", []string{"id", "name"}).WithMeta(false)
		data := deepCopyMap(baseData) // Use a copy
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		expected := map[string]interface{}{
			"id":   123,
			"name": "Test Item",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Transform_SelectSpecificFields_WithMeta", func(t *testing.T) {
		ft := NewFieldTransformer("selectFieldsMeta", []string{"id", "name"}).WithMeta(true)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		expected := map[string]interface{}{
			"id":   123,
			"name": "Test Item",
			"_meta": map[string]interface{}{
				"version": "1.0",
				"source":  "test",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Transform_SelectNestedFields", func(t *testing.T) {
		ft := NewFieldTransformer("selectNested", []string{"id", "nested.key1"}).WithMeta(false)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		expected := map[string]interface{}{
			"id":          123,
			"nested.key1": "value1", // Note: key remains as specified in Fields
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Transform_SelectNonExistentField", func(t *testing.T) {
		ft := NewFieldTransformer("selectNonExistent", []string{"id", "nonexistent"}).WithMeta(false)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		expected := map[string]interface{}{
			"id": 123,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Transform_SelectNonExistentNestedField", func(t *testing.T) {
		ft := NewFieldTransformer("selectNonExistentNested", []string{"nested.nonexistent"}).WithMeta(false)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		assert.Empty(t, result, "Should be empty if only non-existent nested field is selected")

		ft2 := NewFieldTransformer("selectNonExistentNestedWithPath", []string{"nested.key1.nonexistent"}).WithMeta(false)
		result2, err2 := ft2.Transform(data)
		assert.NoError(t, err2)
		assert.Empty(t, result2, "Should be empty if path to non-existent nested field is invalid")
	})

	t.Run("Transform_NoFieldsSpecified_NoMeta", func(t *testing.T) {
		ft := NewFieldTransformer("noFieldsNoMeta", []string{}).WithMeta(false)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		expected := map[string]interface{}{
			"id":          123,
			"name":        "Test Item",
			"description": "A description for testing",
			"price":       99.99,
			"active":      true,
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Transform_NoFieldsSpecified_WithMeta", func(t *testing.T) {
		ft := NewFieldTransformer("noFieldsWithMeta", []string{}).WithMeta(true)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		// Expect all original data as _meta is included by default when no fields specified
		// and WithMeta(true) is called.
		assert.Equal(t, data, result)
	})

	t.Run("Transform_WithPrefix", func(t *testing.T) {
		ft := NewFieldTransformer("prefixTest", []string{"id", "name"}).WithPrefix("item").WithMeta(false).WithSeparator("-")
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		expected := map[string]interface{}{
			"item-id":   123,
			"item-name": "Test Item",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Transform_WithPrefix_AndMeta", func(t *testing.T) {
		ft := NewFieldTransformer("prefixMetaTest", []string{"id"}).WithPrefix("item").WithMeta(true)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)

		assert.NoError(t, err)
		expected := map[string]interface{}{
			"item_id": 123,
			"item__meta": map[string]interface{}{
				"version": "1.0",
				"source":  "test",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("IncludeFieldsTransformer", func(t *testing.T) {
		tr := IncludeFieldsTransformer("id", "name")
		result, err := tr.Transform(baseData)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"id":   123,
			"name": "Test Item",
			"_meta": map[string]interface{}{
				"version": "1.0",
				"source":  "test",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("ExcludeFieldsTransformer", func(t *testing.T) {
		tr := ExcludeFieldsTransformer("description", "price")
		result, err := tr.Transform(baseData)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"id":     123,
			"name":   "Test Item",
			"active": true,
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			"_meta": map[string]interface{}{
				"version": "1.0",
				"source":  "test",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("RenameFieldsTransformer", func(t *testing.T) {
		tr := RenameFieldsTransformer(map[string]string{"id": "item_id", "name": "item_name"})
		result, err := tr.Transform(baseData)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"item_id":     123,
			"item_name":   "Test Item",
			"description": "A description for testing",
			"price":       99.99,
			"active":      true,
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			"_meta": map[string]interface{}{
				"version": "1.0",
				"source":  "test",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("FieldTransformer_Flatten", func(t *testing.T) {
		ft := NewFieldTransformer("flattenTest", []string{"nested"}).WithFlatten(true)
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		// Should flatten nested map into top-level keys with dot separator
		assert.Contains(t, result, "nested.key1")
		assert.Contains(t, result, "nested.key2")
	})

	t.Run("FieldTransformer_PrefixAndSeparator", func(t *testing.T) {
		ft := NewFieldTransformer("prefixSepTest", []string{"id", "name"}).WithPrefix("pfx").WithSeparator(":")
		data := deepCopyMap(baseData)
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		assert.Contains(t, result, "pfx:id")
		assert.Contains(t, result, "pfx:name")
	})

	t.Run("FieldTransformer_MetaFieldVariants", func(t *testing.T) {
		ft := NewFieldTransformer("metaVariants", []string{"id"})
		data := deepCopyMap(baseData)
		data["_metadata"] = map[string]interface{}{"foo": "bar"}
		data["_info"] = 42
		result, err := ft.Transform(data)
		assert.NoError(t, err)
		assert.Contains(t, result, "_meta")
		assert.Contains(t, result, "_metadata")
		assert.Contains(t, result, "_info")
	})

	t.Run("FieldTransformer_EmptyFieldsAndNoMeta", func(t *testing.T) {
		ft := NewFieldTransformer("emptyFieldsNoMeta", []string{}).WithMeta(false)
		result, err := ft.Transform(map[string]interface{}{})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("FieldTransformer_NonMapInput", func(t *testing.T) {
		ft := NewFieldTransformer("nonMap", []string{"id"})
		_, err := ft.Transform(nil)
		assert.Error(t, err)
	})
}
