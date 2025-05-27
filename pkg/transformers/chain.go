package transformers

import (
	"fmt"
)

// TransformerChain applies multiple transformers in sequence
type TransformerChain struct {
	*BaseTransformer
	transformers []Transformer
}

// NewTransformerChain creates a new transformer chain
func NewTransformerChain(name string, transformers ...Transformer) *TransformerChain {
	return &TransformerChain{
		BaseTransformer: NewBaseTransformer(name),
		transformers:    transformers,
	}
}

// Add appends a transformer to the chain
func (tc *TransformerChain) Add(transformer Transformer) *TransformerChain {
	tc.transformers = append(tc.transformers, transformer)
	return tc
}

// Insert adds a transformer at a specific position in the chain
func (tc *TransformerChain) Insert(index int, transformer Transformer) *TransformerChain {
	if index < 0 || index > len(tc.transformers) {
		tc.transformers = append(tc.transformers, transformer)
		return tc
	}

	// Insert at specific position
	tc.transformers = append(tc.transformers[:index], append([]Transformer{transformer}, tc.transformers[index:]...)...)
	return tc
}

// Remove removes a transformer from the chain by name
func (tc *TransformerChain) Remove(name string) *TransformerChain {
	for i, transformer := range tc.transformers {
		if transformer.Name() == name {
			tc.transformers = append(tc.transformers[:i], tc.transformers[i+1:]...)
			break
		}
	}
	return tc
}

// Clear removes all transformers from the chain
func (tc *TransformerChain) Clear() *TransformerChain {
	tc.transformers = tc.transformers[:0]
	return tc
}

// Len returns the number of transformers in the chain
func (tc *TransformerChain) Len() int {
	return len(tc.transformers)
}

// GetTransformers returns a copy of the transformers slice
func (tc *TransformerChain) GetTransformers() []Transformer {
	result := make([]Transformer, len(tc.transformers))
	copy(result, tc.transformers)
	return result
}

func (tc *TransformerChain) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("transformer chain validation failed: %w", err)
	}

	if len(tc.transformers) == 0 {
		// No transformers, return copy of input
		return deepCopyMap(data), nil
	}

	current := data
	for i, transformer := range tc.transformers {
		result, err := transformer.Transform(current)
		if err != nil {
			return nil, fmt.Errorf("transformer chain failed at step %d (%s): %w", i, transformer.Name(), err)
		}
		current = result
	}

	return current, nil
}

// ParallelTransformerChain applies multiple transformers in parallel and merges results
type ParallelTransformerChain struct {
	*BaseTransformer
	transformers []Transformer
	mergeFunc    func([]map[string]interface{}) map[string]interface{}
}

// NewParallelTransformerChain creates a new parallel transformer chain
func NewParallelTransformerChain(name string, transformers ...Transformer) *ParallelTransformerChain {
	return &ParallelTransformerChain{
		BaseTransformer: NewBaseTransformer(name),
		transformers:    transformers,
		mergeFunc:       defaultMergeFunc,
	}
}

// WithMergeFunc sets a custom merge function for combining parallel results
func (ptc *ParallelTransformerChain) WithMergeFunc(mergeFunc func([]map[string]interface{}) map[string]interface{}) *ParallelTransformerChain {
	ptc.mergeFunc = mergeFunc
	return ptc
}

// Add appends a transformer to the parallel chain
func (ptc *ParallelTransformerChain) Add(transformer Transformer) *ParallelTransformerChain {
	ptc.transformers = append(ptc.transformers, transformer)
	return ptc
}

func (ptc *ParallelTransformerChain) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("parallel transformer chain validation failed: %w", err)
	}

	if len(ptc.transformers) == 0 {
		return deepCopyMap(data), nil
	}

	// Channel to collect results
	type result struct {
		data map[string]interface{}
		err  error
		name string
	}

	resultChan := make(chan result, len(ptc.transformers))

	// Execute transformers in parallel
	for _, transformer := range ptc.transformers {
		go func(t Transformer) {
			transformed, err := t.Transform(data)
			resultChan <- result{
				data: transformed,
				err:  err,
				name: t.Name(),
			}
		}(transformer)
	}

	// Collect results
	results := make([]map[string]interface{}, 0, len(ptc.transformers))
	for i := 0; i < len(ptc.transformers); i++ {
		res := <-resultChan
		if res.err != nil {
			return nil, fmt.Errorf("parallel transformer %s failed: %w", res.name, res.err)
		}
		results = append(results, res.data)
	}

	// Merge results using the configured merge function
	return ptc.mergeFunc(results), nil
}

// ConditionalTransformerChain applies transformers based on conditions
type ConditionalTransformerChain struct {
	*BaseTransformer
	conditions []conditionalTransformer
	fallback   Transformer
}

type conditionalTransformer struct {
	condition   func(map[string]interface{}) bool
	transformer Transformer
}

// NewConditionalTransformerChain creates a new conditional transformer chain
func NewConditionalTransformerChain(name string) *ConditionalTransformerChain {
	return &ConditionalTransformerChain{
		BaseTransformer: NewBaseTransformer(name),
		conditions:      make([]conditionalTransformer, 0),
	}
}

// When adds a conditional transformer
func (ctc *ConditionalTransformerChain) When(condition func(map[string]interface{}) bool, transformer Transformer) *ConditionalTransformerChain {
	ctc.conditions = append(ctc.conditions, conditionalTransformer{
		condition:   condition,
		transformer: transformer,
	})
	return ctc
}

// Otherwise sets a fallback transformer
func (ctc *ConditionalTransformerChain) Otherwise(transformer Transformer) *ConditionalTransformerChain {
	ctc.fallback = transformer
	return ctc
}

func (ctc *ConditionalTransformerChain) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateTransformerInput(data); err != nil {
		return nil, fmt.Errorf("conditional transformer chain validation failed: %w", err)
	}

	// Check conditions in order
	for _, ct := range ctc.conditions {
		if ct.condition(data) {
			return ct.transformer.Transform(data)
		}
	}

	// Use fallback if no condition matched
	if ctc.fallback != nil {
		return ctc.fallback.Transform(data)
	}

	// No condition matched and no fallback, return copy of input
	return deepCopyMap(data), nil
}

// defaultMergeFunc is the default merge function for parallel transformers
func defaultMergeFunc(results []map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	for _, result := range results {
		for key, value := range result {
			merged[key] = value
		}
	}

	return merged
}

// Helper functions for creating common transformer chains

// FieldProcessingChain creates a chain for common field processing operations
func FieldProcessingChain(name string) *TransformerChain {
	return NewTransformerChain(name)
}

// MobileOptimizationChain creates a chain optimized for mobile responses
func MobileOptimizationChain() *TransformerChain {
	return NewTransformerChain("mobile_optimization").
		Add(ExcludeFieldsTransformer("_debug", "_internal", "_temp")).
		Add(NewFieldTransformer("mobile_fields", []string{}).WithMeta(true))
}

// DataCleanupChain creates a chain for cleaning up response data
func DataCleanupChain() *TransformerChain {
	return NewTransformerChain("data_cleanup").
		Add(ExcludeFieldsTransformer("password", "secret", "token")).
		Add(NewFuncTransformer("null_cleanup", func(data map[string]interface{}) (map[string]interface{}, error) {
			result := make(map[string]interface{})
			for key, value := range data {
				if value != nil {
					result[key] = value
				}
			}
			return result, nil
		}))
}
