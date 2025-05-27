package bff

import (
	"fmt"
	"sync"
	"time"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
	"go.uber.org/zap"
)

// AggregationStep combines multiple API responses for mobile BFF patterns
type AggregationStep struct {
	*base.BaseStep
	steps       []base.Step
	transformer transformers.Transformer
	parallel    bool
	timeout     time.Duration
	failFast    bool
	required    map[string]bool        // Which steps are required vs optional
	fallbacks   map[string]interface{} // Fallback data for failed steps
}

// NewAggregationStep creates a new aggregation step for BFF patterns
func NewAggregationStep(name string) *AggregationStep {
	return &AggregationStep{
		BaseStep:  base.NewBaseStep(name, "BFF Aggregation: "+name),
		steps:     make([]base.Step, 0),
		parallel:  true,
		timeout:   30 * time.Second,
		failFast:  false,
		required:  make(map[string]bool),
		fallbacks: make(map[string]interface{}),
	}
}

// Configuration methods

// AddStep adds a step to the aggregation
func (a *AggregationStep) AddStep(step base.Step) *AggregationStep {
	a.steps = append(a.steps, step)
	return a
}

// AddRequiredStep adds a required step (failure will fail the aggregation)
func (a *AggregationStep) AddRequiredStep(step base.Step) *AggregationStep {
	a.steps = append(a.steps, step)
	a.required[step.Name()] = true
	return a
}

// AddOptionalStep adds an optional step with fallback data
func (a *AggregationStep) AddOptionalStep(step base.Step, fallback interface{}) *AggregationStep {
	a.steps = append(a.steps, step)
	a.required[step.Name()] = false
	a.fallbacks[step.Name()] = fallback
	return a
}

// WithTransformer sets the aggregation transformer
func (a *AggregationStep) WithTransformer(transformer transformers.Transformer) *AggregationStep {
	a.transformer = transformer
	return a
}

// WithParallel controls parallel vs sequential execution
func (a *AggregationStep) WithParallel(parallel bool) *AggregationStep {
	a.parallel = parallel
	return a
}

// WithTimeout sets the overall timeout for aggregation
func (a *AggregationStep) WithTimeout(timeout time.Duration) *AggregationStep {
	a.timeout = timeout
	return a
}

// WithFailFast controls whether to fail immediately on any error
func (a *AggregationStep) WithFailFast(failFast bool) *AggregationStep {
	a.failFast = failFast
	return a
}

// Run executes the aggregation step
func (a *AggregationStep) Run(ctx *flow.Context) error {
	startTime := time.Now()

	ctx.Logger().Info("Starting BFF aggregation",
		zap.String("step", a.Name()),
		zap.Int("step_count", len(a.steps)),
		zap.Bool("parallel", a.parallel),
		zap.Duration("timeout", a.timeout))

	var results map[string]interface{}
	var err error

	if a.parallel {
		results, err = a.runParallel(ctx)
	} else {
		results, err = a.runSequential(ctx)
	}

	if err != nil {
		ctx.Logger().Error("BFF aggregation failed",
			zap.String("step", a.Name()),
			zap.Error(err),
			zap.Duration("duration", time.Since(startTime)))
		return err
	}

	// Apply transformer if configured
	if a.transformer != nil {
		transformedResults, err := a.transformer.Transform(results)
		if err != nil {
			ctx.Logger().Error("BFF aggregation transformation failed",
				zap.String("step", a.Name()),
				zap.String("transformer", a.transformer.Name()),
				zap.Error(err))
			return fmt.Errorf("aggregation transformation failed: %w", err)
		}
		results = transformedResults

		ctx.Logger().Debug("BFF aggregation transformed",
			zap.String("step", a.Name()),
			zap.String("transformer", a.transformer.Name()))
	}

	// Store aggregated results
	ctx.Set("bff_aggregation", results)
	ctx.Set("bff_"+a.Name(), results)

	ctx.Logger().Info("BFF aggregation completed",
		zap.String("step", a.Name()),
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("result_count", len(results)))

	return nil
}

// runParallel executes steps in parallel with timeout
func (a *AggregationStep) runParallel(ctx *flow.Context) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	errors := make(map[string]error)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create a context with timeout
	timeoutCtx := ctx.WithTimeout(a.timeout)
	defer func() {
		// Context cleanup - no Cancel method needed
	}()

	// Execute steps in parallel
	for _, step := range a.steps {
		wg.Add(1)
		go func(s base.Step) {
			defer wg.Done()

			clonedCtx := timeoutCtx.Clone()
			stepCtx, ok := clonedCtx.(*flow.Context)
			if !ok {
				// Fallback: create a new context with the same data
				stepCtx = flow.NewContext()
				for _, key := range timeoutCtx.Keys() {
					if val, exists := timeoutCtx.Get(key); exists {
						stepCtx.Set(key, val)
					}
				}
				stepCtx.WithFlowName(timeoutCtx.FlowName()).WithLogger(timeoutCtx.Logger())
			}

			err := s.Run(stepCtx)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors[s.Name()] = err

				// Use fallback if available
				if fallback, hasFallback := a.fallbacks[s.Name()]; hasFallback {
					results[s.Name()] = fallback
					ctx.Logger().Warn("Using fallback for failed step",
						zap.String("aggregation", a.Name()),
						zap.String("step", s.Name()),
						zap.Error(err))
				} else if a.required[s.Name()] {
					// Required step failed without fallback
					ctx.Logger().Error("Required step failed in aggregation",
						zap.String("aggregation", a.Name()),
						zap.String("step", s.Name()),
						zap.Error(err))
				}
			} else {
				// Collect step results from context
				if stepResult, exists := stepCtx.Get(s.Name()); exists {
					results[s.Name()] = stepResult
				} else if stepResult, exists := stepCtx.Get("http_response"); exists {
					results[s.Name()] = stepResult
				}
			}
		}(step)
	}

	// Wait for all steps to complete
	wg.Wait()

	// Check for required step failures
	for stepName, isRequired := range a.required {
		if isRequired {
			if err, hasError := errors[stepName]; hasError {
				if _, hasFallback := a.fallbacks[stepName]; !hasFallback {
					return nil, fmt.Errorf("required step %s failed: %w", stepName, err)
				}
			}
		}
	}

	// Fail fast if enabled and any error occurred
	if a.failFast && len(errors) > 0 {
		for stepName, err := range errors {
			return nil, fmt.Errorf("step %s failed (fail-fast enabled): %w", stepName, err)
		}
	}

	return results, nil
}

// runSequential executes steps sequentially
func (a *AggregationStep) runSequential(ctx *flow.Context) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for _, step := range a.steps {
		clonedCtx := ctx.Clone()
		stepCtx, ok := clonedCtx.(*flow.Context)
		if !ok {
			// Fallback: create a new context with the same data
			stepCtx = flow.NewContext()
			for _, key := range ctx.Keys() {
				if val, exists := ctx.Get(key); exists {
					stepCtx.Set(key, val)
				}
			}
			stepCtx.WithFlowName(ctx.FlowName()).WithLogger(ctx.Logger())
		}

		err := step.Run(stepCtx)

		if err != nil {
			// Use fallback if available
			if fallback, hasFallback := a.fallbacks[step.Name()]; hasFallback {
				results[step.Name()] = fallback
				ctx.Logger().Warn("Using fallback for failed step",
					zap.String("aggregation", a.Name()),
					zap.String("step", step.Name()),
					zap.Error(err))
			} else if a.required[step.Name()] {
				// Required step failed without fallback
				return nil, fmt.Errorf("required step %s failed: %w", step.Name(), err)
			} else if a.failFast {
				// Fail fast enabled
				return nil, fmt.Errorf("step %s failed (fail-fast enabled): %w", step.Name(), err)
			}
		} else {
			// Collect step results from context
			if stepResult, exists := stepCtx.Get(step.Name()); exists {
				results[step.Name()] = stepResult
			} else if stepResult, exists := stepCtx.Get("http_response"); exists {
				results[step.Name()] = stepResult
			}
		}
	}

	return results, nil
}

// Convenience constructors for common BFF aggregation patterns

// NewMobileDashboardAggregation creates aggregation for mobile dashboard
func NewMobileDashboardAggregation(baseURL string) *AggregationStep {
	agg := NewAggregationStep("mobile_dashboard").
		WithParallel(true).
		WithTimeout(15 * time.Second).
		WithFailFast(false)

	// Add required user profile
	userProfile := NewMobileUserProfileStep(baseURL)
	agg.AddRequiredStep(userProfile)

	// Add optional notifications with fallback
	notifications := NewMobileNotificationsStep(baseURL)
	agg.AddOptionalStep(notifications, map[string]interface{}{
		"items": []interface{}{},
		"total": 0,
	})

	// Add optional content with fallback
	content := NewMobileContentStep(baseURL)
	agg.AddOptionalStep(content, map[string]interface{}{
		"items":  []interface{}{},
		"total":  0,
		"status": "offline",
	})

	// Add mobile-optimized transformer
	agg.WithTransformer(transformers.NewTransformerChain("dashboard_chain",
		transformers.NewFieldTransformer("dashboard", []string{
			"user_profile", "notifications", "content",
		}),
		transformers.NewFlattenTransformer("mobile_flatten", "mobile").WithMaxDepth(2),
	))

	return agg
}

// NewMobileSearchAggregation creates aggregation for mobile search
func NewMobileSearchAggregation(baseURL string) *AggregationStep {
	agg := NewAggregationStep("mobile_search").
		WithParallel(true).
		WithTimeout(5 * time.Second). // Fast search timeout
		WithFailFast(false)

	// Add required search results
	search := NewMobileSearchStep(baseURL)
	agg.AddRequiredStep(search)

	// Add optional analytics (fire and forget)
	analytics := NewMobileAnalyticsStep(baseURL)
	agg.AddOptionalStep(analytics, map[string]interface{}{
		"status": "skipped",
	})

	return agg
}

// NewMobileProfileAggregation creates aggregation for mobile profile page
func NewMobileProfileAggregation(baseURL string) *AggregationStep {
	agg := NewAggregationStep("mobile_profile").
		WithParallel(true).
		WithTimeout(10 * time.Second).
		WithFailFast(false)

	// Add required user profile
	userProfile := NewMobileUserProfileStep(baseURL)
	agg.AddRequiredStep(userProfile)

	// Add optional user content
	userContent := NewMobileAPIStep("user_content", "GET", baseURL+"/api/v1/user/content",
		[]string{"id", "title", "type", "timestamp", "status"})
	agg.AddOptionalStep(userContent, map[string]interface{}{
		"items": []interface{}{},
		"total": 0,
	})

	// Add mobile profile transformer
	agg.WithTransformer(transformers.NewFieldTransformer("profile", []string{
		"user_profile", "user_content",
	}))

	return agg
}
