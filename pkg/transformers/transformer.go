package transformers

import (
	"fmt"
)

// Transformer defines the interface for data transformation operations
type Transformer interface {
	// Transform applies the transformation to the input data
	Transform(data map[string]interface{}) (map[string]interface{}, error)

	// Name returns the transformer name for logging and debugging
	Name() string
}

// BaseTransformer provides common functionality for transformers
type BaseTransformer struct {
	name string
}

// NewBaseTransformer creates a new base transformer
func NewBaseTransformer(name string) *BaseTransformer {
	return &BaseTransformer{
		name: name,
	}
}

func (bt *BaseTransformer) Name() string {
	return bt.name
}

// FuncTransformer wraps a function as a transformer
type FuncTransformer struct {
	*BaseTransformer
	transformFunc func(map[string]interface{}) (map[string]interface{}, error)
}

// NewFuncTransformer creates a transformer from a function
func NewFuncTransformer(name string, fn func(map[string]interface{}) (map[string]interface{}, error)) *FuncTransformer {
	return &FuncTransformer{
		BaseTransformer: NewBaseTransformer(name),
		transformFunc:   fn,
	}
}

func (ft *FuncTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	return ft.transformFunc(data)
}

// NoOpTransformer is a transformer that returns data unchanged
type NoOpTransformer struct {
	*BaseTransformer
}

// NewNoOpTransformer creates a no-operation transformer
func NewNoOpTransformer() *NoOpTransformer {
	return &NoOpTransformer{
		BaseTransformer: NewBaseTransformer("noop"),
	}
}

func (nt *NoOpTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	// Return a copy to avoid modifying the original
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	return result, nil
}

// CopyTransformer creates a deep copy of the input data
type CopyTransformer struct {
	*BaseTransformer
}

// NewCopyTransformer creates a copy transformer
func NewCopyTransformer() *CopyTransformer {
	return &CopyTransformer{
		BaseTransformer: NewBaseTransformer("copy"),
	}
}

func (ct *CopyTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	return deepCopyMap(data), nil
}

// deepCopyMap creates a deep copy of a map
func deepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for key, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[key] = deepCopyMap(v)
		case []interface{}:
			copy[key] = deepCopySlice(v)
		default:
			copy[key] = value
		}
	}
	return copy
}

// deepCopySlice creates a deep copy of a slice
func deepCopySlice(original []interface{}) []interface{} {
	copy := make([]interface{}, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]interface{}:
			copy[i] = deepCopyMap(v)
		case []interface{}:
			copy[i] = deepCopySlice(v)
		default:
			copy[i] = value
		}
	}
	return copy
}

// ValidateTransformerInput checks if the input data is valid for transformation
func ValidateTransformerInput(data map[string]interface{}) error {
	if data == nil {
		return fmt.Errorf("transformer input data cannot be nil")
	}
	return nil
}

// MergeTransformerResults merges multiple transformer results
func MergeTransformerResults(results ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	for _, result := range results {
		for key, value := range result {
			merged[key] = value
		}
	}

	return merged
}
