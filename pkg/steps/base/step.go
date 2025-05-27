package base

import (
	"time"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"go.uber.org/zap"
)

// Step represents a single operation in a flow
type Step interface {
	// Run executes the step with the given context
	Run(ctx *flow.Context) error

	// Name returns the step name for logging and metrics
	Name() string

	// Description returns a human-readable description
	Description() string
}

// BaseStep provides common functionality for steps
type BaseStep struct {
	name        string
	description string
	timeout     time.Duration
}

// NewBaseStep creates a new base step
func NewBaseStep(name, description string) *BaseStep {
	return &BaseStep{
		name:        name,
		description: description,
		timeout:     10 * time.Second,
	}
}

func (s *BaseStep) Name() string {
	return s.name
}

func (s *BaseStep) Description() string {
	return s.description
}

func (s *BaseStep) WithTimeout(timeout time.Duration) *BaseStep {
	s.timeout = timeout
	return s
}

func (s *BaseStep) Timeout() time.Duration {
	return s.timeout
}

// StepFunc is a function type that implements Step
type StepFunc func(ctx *flow.Context) error

func (f StepFunc) Run(ctx *flow.Context) error {
	return f(ctx)
}

func (f StepFunc) Name() string {
	return "anonymous"
}

func (f StepFunc) Description() string {
	return "Anonymous step function"
}

// ConditionalStep represents a step that executes based on a condition
type ConditionalStep struct {
	*BaseStep
	condition func(*flow.Context) bool
	step      Step
}

// NewConditionalStep creates a conditional step
func NewConditionalStep(name string, condition func(*flow.Context) bool, step Step) *ConditionalStep {
	return &ConditionalStep{
		BaseStep:  NewBaseStep(name, "Conditional: "+step.Description()),
		condition: condition,
		step:      step,
	}
}

func (s *ConditionalStep) Run(ctx *flow.Context) error {
	if s.condition(ctx) {
		ctx.Logger().Info("Condition met, executing step",
			zap.String("step", s.Name()),
			zap.String("conditional_step", s.step.Name()))
		return s.step.Run(ctx)
	}

	ctx.Logger().Info("Condition not met, skipping step",
		zap.String("step", s.Name()),
		zap.String("conditional_step", s.step.Name()))
	return nil
}

// SequentialStep executes steps in sequence
type SequentialStep struct {
	*BaseStep
	steps []Step
}

// NewSequentialStep creates a sequential step
func NewSequentialStep(name string, steps ...Step) *SequentialStep {
	return &SequentialStep{
		BaseStep: NewBaseStep(name, "Sequential execution"),
		steps:    steps,
	}
}

func (s *SequentialStep) Run(ctx *flow.Context) error {
	for i, step := range s.steps {
		ctx.Logger().Info("Executing sequential step",
			zap.String("parent_step", s.Name()),
			zap.String("step", step.Name()),
			zap.Int("index", i))

		if err := step.Run(ctx); err != nil {
			ctx.Logger().Error("Sequential step failed",
				zap.String("parent_step", s.Name()),
				zap.String("step", step.Name()),
				zap.Int("index", i),
				zap.Error(err))
			return err
		}
	}
	return nil
}

// ParallelStep executes multiple steps concurrently
type ParallelStep struct {
	*BaseStep
	steps []Step
}

// NewParallelStep creates a parallel step
func NewParallelStep(name string, steps ...Step) *ParallelStep {
	return &ParallelStep{
		BaseStep: NewBaseStep(name, "Parallel execution"),
		steps:    steps,
	}
}

func (s *ParallelStep) Run(ctx *flow.Context) error {
	if len(s.steps) == 0 {
		return nil
	}

	// Create error channel
	errChan := make(chan error, len(s.steps))

	// Execute steps concurrently
	for _, step := range s.steps {
		go func(step Step) {
			// Clone context for each parallel step to avoid race conditions
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

			stepCtx.Logger().Info("Starting parallel step",
				zap.String("parent_step", s.Name()),
				zap.String("step", step.Name()))

			err := step.Run(stepCtx)
			if err != nil {
				stepCtx.Logger().Error("Parallel step failed",
					zap.String("parent_step", s.Name()),
					zap.String("step", step.Name()),
					zap.Error(err))
			} else {
				// Merge successful results back to main context
				for _, key := range stepCtx.Keys() {
					if val, ok := stepCtx.Get(key); ok {
						ctx.Set(key, val)
					}
				}
			}

			errChan <- err
		}(step)
	}

	// Wait for all steps to complete
	var firstError error
	for i := 0; i < len(s.steps); i++ {
		if err := <-errChan; err != nil && firstError == nil {
			firstError = err
		}
	}

	return firstError
}

// RetryStep wraps another step with retry logic
type RetryStep struct {
	*BaseStep
	step        Step
	maxRetries  int
	retryDelay  time.Duration
	shouldRetry func(error) bool
}

// NewRetryStep creates a retry step
func NewRetryStep(name string, step Step, maxRetries int, retryDelay time.Duration) *RetryStep {
	return &RetryStep{
		BaseStep:    NewBaseStep(name, "Retry wrapper for: "+step.Description()),
		step:        step,
		maxRetries:  maxRetries,
		retryDelay:  retryDelay,
		shouldRetry: func(error) bool { return true }, // Retry all errors by default
	}
}

func (s *RetryStep) WithRetryCondition(shouldRetry func(error) bool) *RetryStep {
	s.shouldRetry = shouldRetry
	return s
}

func (s *RetryStep) Run(ctx *flow.Context) error {
	var lastErr error

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			ctx.Logger().Info("Retrying step",
				zap.String("step", s.Name()),
				zap.String("wrapped_step", s.step.Name()),
				zap.Int("attempt", attempt),
				zap.Int("max_retries", s.maxRetries))

			time.Sleep(s.retryDelay)
		}

		err := s.step.Run(ctx)
		if err == nil {
			if attempt > 0 {
				ctx.Logger().Info("Step succeeded after retry",
					zap.String("step", s.Name()),
					zap.String("wrapped_step", s.step.Name()),
					zap.Int("attempt", attempt))
			}
			return nil
		}

		lastErr = err

		if !s.shouldRetry(err) {
			ctx.Logger().Info("Error not retryable, stopping",
				zap.String("step", s.Name()),
				zap.String("wrapped_step", s.step.Name()),
				zap.Error(err))
			break
		}

		ctx.Logger().Warn("Step failed, will retry",
			zap.String("step", s.Name()),
			zap.String("wrapped_step", s.step.Name()),
			zap.Int("attempt", attempt),
			zap.Error(err))
	}

	return lastErr
}

// DelayStep introduces a delay in the flow
type DelayStep struct {
	*BaseStep
	delay time.Duration
}

// NewDelayStep creates a delay step
func NewDelayStep(name string, delay time.Duration) *DelayStep {
	return &DelayStep{
		BaseStep: NewBaseStep(name, "Delay"),
		delay:    delay,
	}
}

func (s *DelayStep) Run(ctx *flow.Context) error {
	ctx.Logger().Info("Executing delay step",
		zap.String("step", s.Name()),
		zap.Duration("delay", s.delay))

	select {
	case <-time.After(s.delay):
		return nil
	case <-ctx.Context().Done():
		return ctx.Context().Err()
	}
}
