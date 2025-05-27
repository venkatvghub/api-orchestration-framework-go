package transformers

import (
	"fmt"
	"strings"

	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
)

// FieldTransformer selects and transforms specific fields from input data
type FieldTransformer struct {
	*BaseTransformer
	Fields      []string
	Prefix      string
	IncludeMeta bool
	Flatten     bool
	Separator   string
}

// NewFieldTransformer creates a new field transformer
func NewFieldTransformer(name string, fields []string) *FieldTransformer {
	return &FieldTransformer{
		BaseTransformer: NewBaseTransformer(name),
		Fields:          fields,
		IncludeMeta:     true,
		Flatten:         false,
		Separator:       "_",
	}
}

// WithPrefix sets a prefix for all field names
func (ft *FieldTransformer) WithPrefix(prefix string) *FieldTransformer {
	ft.Prefix = prefix
	return ft
}

// WithMeta controls whether to include metadata fields
func (ft *FieldTransformer) WithMeta(include bool) *FieldTransformer {
	ft.IncludeMeta = include
	return ft
}

// WithFlatten controls whether to flatten nested structures
func (ft *FieldTransformer) WithFlatten(flatten bool) *FieldTransformer {
	ft.Flatten = flatten
	return ft
}

// WithSeparator sets the separator for flattened field names
func (ft *FieldTransformer) WithSeparator(separator string) *FieldTransformer {
	ft.Separator = separator
	return ft
}

func (ft *FieldTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("field transformer validation failed: %w", err)
	}

	result := make(map[string]interface{})

	// If no fields specified, include all non-meta fields
	if len(ft.Fields) == 0 {
		for key, value := range data {
			if !ft.isMetaField(key) {
				ft.addField(result, key, value)
			}
		}
	} else {
		// Include only specified fields
		for _, field := range ft.Fields {
			if value, exists := ft.getFieldValue(data, field); exists {
				ft.addField(result, field, value)
			}
		}
	}

	// Include metadata if requested
	if ft.IncludeMeta {
		for key, value := range data {
			if ft.isMetaField(key) {
				ft.addField(result, key, value)
			}
		}
	}

	return result, nil
}

// getFieldValue retrieves a field value, supporting nested field access
func (ft *FieldTransformer) getFieldValue(data map[string]interface{}, field string) (interface{}, bool) {
	if !strings.Contains(field, ".") {
		// Simple field access
		value, exists := data[field]
		return value, exists
	}

	// Nested field access
	parts := strings.Split(field, ".")
	current := data

	for i, part := range parts {
		if current == nil {
			return nil, false
		}

		value, exists := current[part]
		if !exists {
			return nil, false
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			return value, true
		}

		// Otherwise, continue traversing
		if nested, ok := value.(map[string]interface{}); ok {
			current = nested
		} else {
			return nil, false
		}
	}

	return nil, false
}

// addField adds a field to the result, applying prefix and flattening if configured
func (ft *FieldTransformer) addField(result map[string]interface{}, key string, value interface{}) {
	finalKey := key
	if ft.Prefix != "" {
		finalKey = ft.Prefix + ft.Separator + key
	}

	if ft.Flatten {
		if nested, ok := value.(map[string]interface{}); ok {
			// Flatten nested map
			flattened := utils.FlattenMap(nested, finalKey)
			for flatKey, flatValue := range flattened {
				result[flatKey] = flatValue
			}
			return
		}
	}

	result[finalKey] = value
}

// isMetaField checks if a field is considered metadata
func (ft *FieldTransformer) isMetaField(key string) bool {
	metaFields := []string{"_meta", "_metadata", "_info", "_debug"}
	for _, meta := range metaFields {
		if key == meta || strings.HasPrefix(key, meta+"_") {
			return true
		}
	}
	return false
}

// IncludeFieldsTransformer creates a transformer that includes only specified fields
func IncludeFieldsTransformer(fields ...string) Transformer {
	return NewFieldTransformer("include_fields", fields)
}

// ExcludeFieldsTransformer creates a transformer that excludes specified fields
func ExcludeFieldsTransformer(excludeFields ...string) Transformer {
	return NewFuncTransformer("exclude_fields", func(data map[string]interface{}) (map[string]interface{}, error) {
		if err := ValidateTransformerInput(data); err != nil {
			return nil, err
		}

		result := make(map[string]interface{})
		excludeSet := make(map[string]bool)
		for _, field := range excludeFields {
			excludeSet[field] = true
		}

		for key, value := range data {
			if !excludeSet[key] {
				result[key] = value
			}
		}

		return result, nil
	})
}

// RenameFieldsTransformer creates a transformer that renames fields
func RenameFieldsTransformer(fieldMap map[string]string) Transformer {
	return NewFuncTransformer("rename_fields", func(data map[string]interface{}) (map[string]interface{}, error) {
		if err := ValidateTransformerInput(data); err != nil {
			return nil, err
		}

		result := make(map[string]interface{})

		for key, value := range data {
			newKey := key
			if renamed, exists := fieldMap[key]; exists {
				newKey = renamed
			}
			result[newKey] = value
		}

		return result, nil
	})
}

// AddFieldsTransformer creates a transformer that adds new fields
func AddFieldsTransformer(newFields map[string]interface{}) Transformer {
	return NewFuncTransformer("add_fields", func(data map[string]interface{}) (map[string]interface{}, error) {
		if err := ValidateTransformerInput(data); err != nil {
			return nil, err
		}

		result := make(map[string]interface{})

		// Copy existing fields
		for key, value := range data {
			result[key] = value
		}

		// Add new fields
		for key, value := range newFields {
			result[key] = value
		}

		return result, nil
	})
}
