package transformers

import (
	"fmt"
	"strings"

	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
)

// FlattenTransformer flattens nested map structures into a single level
type FlattenTransformer struct {
	*BaseTransformer
	Prefix      string
	Separator   string
	MaxDepth    int
	IncludeMeta bool
}

// NewFlattenTransformer creates a new flatten transformer
func NewFlattenTransformer(name, prefix string) *FlattenTransformer {
	return &FlattenTransformer{
		BaseTransformer: NewBaseTransformer(name),
		Prefix:          prefix,
		Separator:       ".",
		MaxDepth:        10, // Prevent infinite recursion
		IncludeMeta:     true,
	}
}

// WithSeparator sets the separator for flattened keys
func (ft *FlattenTransformer) WithSeparator(separator string) *FlattenTransformer {
	ft.Separator = separator
	return ft
}

// WithMaxDepth sets the maximum depth for flattening
func (ft *FlattenTransformer) WithMaxDepth(depth int) *FlattenTransformer {
	ft.MaxDepth = depth
	return ft
}

// WithMeta controls whether to include metadata fields
func (ft *FlattenTransformer) WithMeta(include bool) *FlattenTransformer {
	ft.IncludeMeta = include
	return ft
}

func (ft *FlattenTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("flatten transformer validation failed: %w", err)
	}

	result := make(map[string]interface{})
	ft.flattenRecursive(data, ft.Prefix, result, 0)

	// Filter metadata if not included
	if !ft.IncludeMeta {
		filtered := make(map[string]interface{})
		for key, value := range result {
			if !ft.isMetaField(key) {
				filtered[key] = value
			}
		}
		return filtered, nil
	}

	return result, nil
}

// flattenRecursive recursively flattens nested structures
func (ft *FlattenTransformer) flattenRecursive(data map[string]interface{}, prefix string, result map[string]interface{}, depth int) {
	if depth >= ft.MaxDepth {
		// If max depth reached, store as-is
		if prefix != "" {
			result[prefix] = data
		} else {
			for key, value := range data {
				result[key] = value
			}
		}
		return
	}

	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + ft.Separator + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively flatten nested maps
			ft.flattenRecursive(v, newKey, result, depth+1)
		case map[interface{}]interface{}:
			// Convert and flatten interface{} maps
			converted := make(map[string]interface{})
			for k, val := range v {
				if strKey, ok := k.(string); ok {
					converted[strKey] = val
				}
			}
			ft.flattenRecursive(converted, newKey, result, depth+1)
		case []interface{}:
			// Handle arrays by flattening each element if it's a map
			ft.flattenArray(v, newKey, result, depth)
		default:
			// Store primitive values directly
			result[newKey] = value
		}
	}
}

// flattenArray flattens array elements that are maps
func (ft *FlattenTransformer) flattenArray(arr []interface{}, prefix string, result map[string]interface{}, depth int) {
	for i, item := range arr {
		itemKey := fmt.Sprintf("%s%s%d", prefix, ft.Separator, i)

		switch v := item.(type) {
		case map[string]interface{}:
			ft.flattenRecursive(v, itemKey, result, depth+1)
		case map[interface{}]interface{}:
			converted := make(map[string]interface{})
			for k, val := range v {
				if strKey, ok := k.(string); ok {
					converted[strKey] = val
				}
			}
			ft.flattenRecursive(converted, itemKey, result, depth+1)
		default:
			result[itemKey] = item
		}
	}
}

// isMetaField checks if a field is considered metadata
func (ft *FlattenTransformer) isMetaField(key string) bool {
	metaFields := []string{"_meta", "_metadata", "_info", "_debug"}
	for _, meta := range metaFields {
		if strings.HasPrefix(key, meta) {
			return true
		}
	}
	return false
}

// UnflattenTransformer reverses the flattening process
type UnflattenTransformer struct {
	*BaseTransformer
	Separator string
}

// NewUnflattenTransformer creates a new unflatten transformer
func NewUnflattenTransformer(name string) *UnflattenTransformer {
	return &UnflattenTransformer{
		BaseTransformer: NewBaseTransformer(name),
		Separator:       ".",
	}
}

// WithSeparator sets the separator used in flattened keys
func (ut *UnflattenTransformer) WithSeparator(separator string) *UnflattenTransformer {
	ut.Separator = separator
	return ut
}

func (ut *UnflattenTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("unflatten transformer validation failed: %w", err)
	}

	result := make(map[string]interface{})

	for key, value := range data {
		if err := utils.SetNestedValue(result, strings.ReplaceAll(key, ut.Separator, "."), value); err != nil {
			// If setting nested value fails, use the key as-is
			result[key] = value
		}
	}

	return result, nil
}

// PrefixTransformer adds a prefix to all field names
type PrefixTransformer struct {
	*BaseTransformer
	Prefix    string
	Separator string
}

// NewPrefixTransformer creates a new prefix transformer
func NewPrefixTransformer(name, prefix string) *PrefixTransformer {
	return &PrefixTransformer{
		BaseTransformer: NewBaseTransformer(name),
		Prefix:          prefix,
		Separator:       "_",
	}
}

// WithSeparator sets the separator between prefix and field name
func (pt *PrefixTransformer) WithSeparator(separator string) *PrefixTransformer {
	pt.Separator = separator
	return pt
}

func (pt *PrefixTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("prefix transformer validation failed: %w", err)
	}

	result := make(map[string]interface{})

	for key, value := range data {
		newKey := pt.Prefix + pt.Separator + key
		result[newKey] = value
	}

	return result, nil
}

// RemovePrefixTransformer removes a prefix from field names
type RemovePrefixTransformer struct {
	*BaseTransformer
	Prefix    string
	Separator string
}

// NewRemovePrefixTransformer creates a new remove prefix transformer
func NewRemovePrefixTransformer(name, prefix string) *RemovePrefixTransformer {
	return &RemovePrefixTransformer{
		BaseTransformer: NewBaseTransformer(name),
		Prefix:          prefix,
		Separator:       "_",
	}
}

// WithSeparator sets the separator between prefix and field name
func (rpt *RemovePrefixTransformer) WithSeparator(separator string) *RemovePrefixTransformer {
	rpt.Separator = separator
	return rpt
}

func (rpt *RemovePrefixTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("remove prefix transformer validation failed: %w", err)
	}

	result := make(map[string]interface{})
	prefixWithSeparator := rpt.Prefix + rpt.Separator

	for key, value := range data {
		newKey := key
		if strings.HasPrefix(key, prefixWithSeparator) {
			newKey = strings.TrimPrefix(key, prefixWithSeparator)
		}
		result[newKey] = value
	}

	return result, nil
}
