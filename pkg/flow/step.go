package flow

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

// Step represents a single operation in a flow
// This is kept for backward compatibility, but new steps should implement interfaces.Step
type Step interface {
	// Run executes the step with the given context
	Run(ctx interfaces.ExecutionContext) error

	// Name returns the step name for logging and metrics
	Name() string

	// Description returns a human-readable description
	Description() string
}

// LegacyStepWrapper wraps the old flow.Step to work with new ExecutionContext
type LegacyStepWrapper struct {
	step Step
}

// NewLegacyStepWrapper creates a wrapper for old flow steps
func NewLegacyStepWrapper(step Step) *LegacyStepWrapper {
	return &LegacyStepWrapper{step: step}
}

func (w *LegacyStepWrapper) Run(ctx interfaces.ExecutionContext) error {
	start := time.Now()
	err := w.step.Run(ctx)
	duration := time.Since(start)

	// Record metrics
	metrics.RecordStepExecution(w.step.Name(), duration, err == nil)

	return err
}

func (w *LegacyStepWrapper) Name() string {
	return w.step.Name()
}

func (w *LegacyStepWrapper) Description() string {
	return w.step.Description()
}

// StepWrapper wraps interfaces.Step to work with *flow.Context
type StepWrapper struct {
	step interfaces.Step
}

// NewStepWrapper creates a wrapper for interfaces.Step to work with *flow.Context
func NewStepWrapper(step interfaces.Step) *StepWrapper {
	return &StepWrapper{step: step}
}

func (w *StepWrapper) Run(ctx *Context) error {
	start := time.Now()
	err := w.step.Run(ctx)
	duration := time.Since(start)

	// Record metrics
	metrics.RecordStepExecution(w.step.Name(), duration, err == nil)

	return err
}

func (w *StepWrapper) Name() string {
	return w.step.Name()
}

func (w *StepWrapper) Description() string {
	return w.step.Description()
}

// StepFunc is a function type that implements interfaces.Step
type StepFunc func(ctx interfaces.ExecutionContext) error

func (f StepFunc) Run(ctx interfaces.ExecutionContext) error {
	return f(ctx)
}

func (f StepFunc) Name() string {
	return "anonymous"
}

func (f StepFunc) Description() string {
	return "Anonymous step function"
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

// ConditionalStep represents a step that executes based on a condition
type ConditionalStep struct {
	*BaseStep
	condition func(interfaces.ExecutionContext) bool
	step      interfaces.Step
}

// NewConditionalStep creates a conditional step
func NewConditionalStep(name string, condition func(interfaces.ExecutionContext) bool, step interfaces.Step) *ConditionalStep {
	return &ConditionalStep{
		BaseStep:  NewBaseStep(name, fmt.Sprintf("Conditional: %s", step.Description())),
		condition: condition,
		step:      step,
	}
}

func (s *ConditionalStep) Run(ctx interfaces.ExecutionContext) error {
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

// ParallelStep executes multiple steps concurrently
type ParallelStep struct {
	*BaseStep
	steps []interfaces.Step
}

// NewParallelStep creates a parallel step
func NewParallelStep(name string, steps ...interfaces.Step) *ParallelStep {
	return &ParallelStep{
		BaseStep: NewBaseStep(name, fmt.Sprintf("Parallel execution of %d steps", len(steps))),
		steps:    steps,
	}
}

func (s *ParallelStep) Run(ctx interfaces.ExecutionContext) error {
	if len(s.steps) == 0 {
		return nil
	}

	// Create error channel
	errChan := make(chan error, len(s.steps))

	// Check if context is already cancelled
	select {
	case <-ctx.Context().Done():
		return ctx.Context().Err()
	default:
	}

	// Execute steps concurrently
	for _, step := range s.steps {
		go func(step interfaces.Step) {
			// Clone context for each parallel step to avoid race conditions
			clonedCtx := ctx.Clone()

			clonedCtx.Logger().Info("Starting parallel step",
				zap.String("parent_step", s.Name()),
				zap.String("step", step.Name()))

			// Check for cancellation before starting step
			select {
			case <-clonedCtx.Context().Done():
				errChan <- clonedCtx.Context().Err()
				return
			default:
			}

			err := step.Run(clonedCtx)
			if err != nil {
				clonedCtx.Logger().Error("Parallel step failed",
					zap.String("parent_step", s.Name()),
					zap.String("step", step.Name()),
					zap.Error(err))
			} else {
				// Merge successful results back to main context
				// Use a mutex to avoid race conditions when writing to main context
				for _, key := range clonedCtx.Keys() {
					if val, ok := clonedCtx.Get(key); ok {
						ctx.Set(key, val)
					}
				}
			}

			errChan <- err
		}(step)
	}

	// Wait for all steps to complete or context cancellation
	var firstError error
	for i := 0; i < len(s.steps); i++ {
		select {
		case err := <-errChan:
			if err != nil && firstError == nil {
				firstError = err
			}
		case <-ctx.Context().Done():
			// Context was cancelled, return immediately
			return ctx.Context().Err()
		}
	}

	return firstError
}

// SequentialStep executes steps in sequence
type SequentialStep struct {
	*BaseStep
	steps []interfaces.Step
}

// NewSequentialStep creates a sequential step
func NewSequentialStep(name string, steps ...interfaces.Step) *SequentialStep {
	return &SequentialStep{
		BaseStep: NewBaseStep(name, fmt.Sprintf("Sequential execution of %d steps", len(steps))),
		steps:    steps,
	}
}

func (s *SequentialStep) Run(ctx interfaces.ExecutionContext) error {
	for _, step := range s.steps {
		ctx.Logger().Info("Executing sequential step",
			zap.String("parent_step", s.Name()),
			zap.String("step", step.Name()))

		if err := step.Run(ctx); err != nil {
			ctx.Logger().Error("Sequential step failed",
				zap.String("parent_step", s.Name()),
				zap.String("step", step.Name()),
				zap.Error(err))
			return err
		}
	}
	return nil
}

// TransformStep applies a transformation function to context data
type TransformStep struct {
	*BaseStep
	transformer func(interfaces.ExecutionContext) error
}

// NewTransformStep creates a transform step
func NewTransformStep(name string, transformer func(interfaces.ExecutionContext) error) *TransformStep {
	return &TransformStep{
		BaseStep:    NewBaseStep(name, "Data transformation step"),
		transformer: transformer,
	}
}

func (s *TransformStep) Run(ctx interfaces.ExecutionContext) error {
	ctx.Logger().Info("Executing transform step", zap.String("step", s.Name()))
	return s.transformer(ctx)
}

// RetryStep wraps another step with retry logic
type RetryStep struct {
	*BaseStep
	step        interfaces.Step
	maxRetries  int
	retryDelay  time.Duration
	shouldRetry func(error) bool
}

// NewRetryStep creates a retry step
func NewRetryStep(name string, step interfaces.Step, maxRetries int, retryDelay time.Duration) *RetryStep {
	return &RetryStep{
		BaseStep:    NewBaseStep(name, fmt.Sprintf("Retry wrapper for: %s", step.Description())),
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

func (s *RetryStep) Run(ctx interfaces.ExecutionContext) error {
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

	return fmt.Errorf("step failed after %d retries: %w", s.maxRetries, lastErr)
}

// DelayStep introduces a delay in the flow
type DelayStep struct {
	*BaseStep
	delay time.Duration
}

// NewDelayStep creates a delay step
func NewDelayStep(name string, delay time.Duration) *DelayStep {
	return &DelayStep{
		BaseStep: NewBaseStep(name, fmt.Sprintf("Delay for %v", delay)),
		delay:    delay,
	}
}

func (s *DelayStep) Run(ctx interfaces.ExecutionContext) error {
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
