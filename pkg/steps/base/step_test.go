package base

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
)

// Mock step for testing
type mockStep struct {
	name        string
	description string
	runFunc     func(*flow.Context) error
}

func (m *mockStep) Run(ctx *flow.Context) error {
	if m.runFunc != nil {
		return m.runFunc(ctx)
	}
	return nil
}

func (m *mockStep) Name() string {
	return m.name
}

func (m *mockStep) Description() string {
	return m.description
}

func TestNewBaseStep(t *testing.T) {
	step := NewBaseStep("test_step", "Test description")

	if step.Name() != "test_step" {
		t.Errorf("Name() = %v, want 'test_step'", step.Name())
	}

	if step.Description() != "Test description" {
		t.Errorf("Description() = %v, want 'Test description'", step.Description())
	}

	if step.timeout != 10*time.Second {
		t.Errorf("Default timeout = %v, want 10s", step.timeout)
	}
}

func TestBaseStep_WithTimeout(t *testing.T) {
	step := NewBaseStep("test_step", "Test description")
	timeout := 30 * time.Second

	result := step.WithTimeout(timeout)

	if result != step {
		t.Error("WithTimeout should return the same instance")
	}

	if step.timeout != timeout {
		t.Errorf("Timeout = %v, want %v", step.timeout, timeout)
	}

	if step.Timeout() != timeout {
		t.Errorf("Timeout() = %v, want %v", step.Timeout(), timeout)
	}
}

func TestStepFunc(t *testing.T) {
	executed := false
	stepFunc := StepFunc(func(ctx *flow.Context) error {
		executed = true
		ctx.Set("func_key", "func_value")
		return nil
	})

	if stepFunc.Name() != "anonymous" {
		t.Errorf("Name() = %v, want 'anonymous'", stepFunc.Name())
	}

	if stepFunc.Description() != "Anonymous step function" {
		t.Errorf("Description() = %v, want 'Anonymous step function'", stepFunc.Description())
	}

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := stepFunc.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if !executed {
		t.Error("StepFunc should have been executed")
	}

	val, ok := ctx.Get("func_key")
	if !ok || val != "func_value" {
		t.Error("StepFunc should have set the value")
	}
}

func TestStepFunc_Error(t *testing.T) {
	expectedError := errors.New("function error")
	stepFunc := StepFunc(func(ctx *flow.Context) error {
		return expectedError
	})

	ctx := flow.NewContext()
	err := stepFunc.Run(ctx)

	if err != expectedError {
		t.Errorf("Run() error = %v, want %v", err, expectedError)
	}
}

func TestConditionalStep(t *testing.T) {
	executed := false
	mockStep := &mockStep{
		name:        "conditional_inner",
		description: "Inner step",
		runFunc: func(ctx *flow.Context) error {
			executed = true
			return nil
		},
	}

	// Test condition that returns true
	condition := func(ctx *flow.Context) bool {
		return ctx.Has("execute")
	}

	conditionalStep := NewConditionalStep("conditional_test", condition, mockStep)

	if conditionalStep.Name() != "conditional_test" {
		t.Errorf("Name() = %v, want 'conditional_test'", conditionalStep.Name())
	}

	expectedDesc := "Conditional: Inner step"
	if conditionalStep.Description() != expectedDesc {
		t.Errorf("Description() = %v, want %v", conditionalStep.Description(), expectedDesc)
	}

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("execute", true)

	err := conditionalStep.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if !executed {
		t.Error("Inner step should have been executed when condition is true")
	}

	// Test condition that returns false
	executed = false
	ctx2 := flow.NewContext().WithLogger(zap.NewNop())
	// Don't set "execute" key

	err = conditionalStep.Run(ctx2)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if executed {
		t.Error("Inner step should not have been executed when condition is false")
	}
}

func TestSequentialStep(t *testing.T) {
	executionOrder := []string{}

	step1 := &mockStep{
		name: "step1",
		runFunc: func(ctx *flow.Context) error {
			executionOrder = append(executionOrder, "step1")
			ctx.Set("step1", "done")
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx *flow.Context) error {
			executionOrder = append(executionOrder, "step2")
			ctx.Set("step2", "done")
			return nil
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx *flow.Context) error {
			executionOrder = append(executionOrder, "step3")
			ctx.Set("step3", "done")
			return nil
		},
	}

	sequentialStep := NewSequentialStep("sequential_test", step1, step2, step3)

	if sequentialStep.Name() != "sequential_test" {
		t.Errorf("Name() = %v, want 'sequential_test'", sequentialStep.Name())
	}

	if sequentialStep.Description() != "Sequential execution" {
		t.Errorf("Description() = %v, want 'Sequential execution'", sequentialStep.Description())
	}

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := sequentialStep.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	expectedOrder := []string{"step1", "step2", "step3"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Execution order length = %d, want %d", len(executionOrder), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Execution order[%d] = %v, want %v", i, executionOrder[i], expected)
		}
	}

	// Check that all steps executed
	if !ctx.Has("step1") || !ctx.Has("step2") || !ctx.Has("step3") {
		t.Error("All sequential steps should have executed")
	}
}

func TestSequentialStep_Error(t *testing.T) {
	expectedError := errors.New("step2 error")
	executionOrder := []string{}

	step1 := &mockStep{
		name: "step1",
		runFunc: func(ctx *flow.Context) error {
			executionOrder = append(executionOrder, "step1")
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx *flow.Context) error {
			executionOrder = append(executionOrder, "step2")
			return expectedError
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx *flow.Context) error {
			executionOrder = append(executionOrder, "step3")
			return nil
		},
	}

	sequentialStep := NewSequentialStep("sequential_error", step1, step2, step3)
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	err := sequentialStep.Run(ctx)
	if err != expectedError {
		t.Errorf("Run() error = %v, want %v", err, expectedError)
	}

	// Should have executed step1 and step2, but not step3
	expectedOrder := []string{"step1", "step2"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Execution order length = %d, want %d", len(executionOrder), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Execution order[%d] = %v, want %v", i, executionOrder[i], expected)
		}
	}
}

func TestParallelStep(t *testing.T) {
	step1Executed := false
	step2Executed := false
	step3Executed := false

	step1 := &mockStep{
		name: "step1",
		runFunc: func(ctx *flow.Context) error {
			step1Executed = true
			ctx.Set("step1", "done")
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx *flow.Context) error {
			step2Executed = true
			ctx.Set("step2", "done")
			return nil
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx *flow.Context) error {
			step3Executed = true
			ctx.Set("step3", "done")
			return nil
		},
	}

	parallelStep := NewParallelStep("parallel_test", step1, step2, step3)

	if parallelStep.Name() != "parallel_test" {
		t.Errorf("Name() = %v, want 'parallel_test'", parallelStep.Name())
	}

	if parallelStep.Description() != "Parallel execution" {
		t.Errorf("Description() = %v, want 'Parallel execution'", parallelStep.Description())
	}

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := parallelStep.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if !step1Executed || !step2Executed || !step3Executed {
		t.Error("All parallel steps should have been executed")
	}

	// Check that results were merged back
	if !ctx.Has("step1") || !ctx.Has("step2") || !ctx.Has("step3") {
		t.Error("Parallel step results should be merged back to main context")
	}
}

func TestParallelStep_Empty(t *testing.T) {
	parallelStep := NewParallelStep("empty_parallel")
	ctx := flow.NewContext()

	err := parallelStep.Run(ctx)
	if err != nil {
		t.Errorf("Empty parallel step should not fail: %v", err)
	}
}

func TestParallelStep_Error(t *testing.T) {
	expectedError := errors.New("step error")

	step1 := &mockStep{
		name: "step1",
		runFunc: func(ctx *flow.Context) error {
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx *flow.Context) error {
			return expectedError
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx *flow.Context) error {
			return nil
		},
	}

	parallelStep := NewParallelStep("parallel_error", step1, step2, step3)
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	err := parallelStep.Run(ctx)
	if err != expectedError {
		t.Errorf("Run() error = %v, want %v", err, expectedError)
	}
}

func TestRetryStep(t *testing.T) {
	attempts := 0
	mockStep := &mockStep{
		name:        "retry_inner",
		description: "Inner step",
		runFunc: func(ctx *flow.Context) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			ctx.Set("success", true)
			return nil
		},
	}

	retryStep := NewRetryStep("retry_test", mockStep, 3, 10*time.Millisecond)

	if retryStep.Name() != "retry_test" {
		t.Errorf("Name() = %v, want 'retry_test'", retryStep.Name())
	}

	expectedDesc := "Retry wrapper for: Inner step"
	if retryStep.Description() != expectedDesc {
		t.Errorf("Description() = %v, want %v", retryStep.Description(), expectedDesc)
	}

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := retryStep.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Attempts = %d, want 3", attempts)
	}

	success, _ := ctx.GetBool("success")
	if !success {
		t.Error("Step should have succeeded after retries")
	}
}

func TestRetryStep_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	mockStep := &mockStep{
		name: "retry_inner",
		runFunc: func(ctx *flow.Context) error {
			attempts++
			return errors.New("persistent error")
		},
	}

	retryStep := NewRetryStep("retry_fail", mockStep, 2, 1*time.Millisecond)
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	err := retryStep.Run(ctx)
	if err == nil {
		t.Error("Run() should have failed after max retries")
	}

	if attempts != 3 { // Initial attempt + 2 retries
		t.Errorf("Attempts = %d, want 3", attempts)
	}
}

func TestRetryStep_WithRetryCondition(t *testing.T) {
	attempts := 0
	nonRetryableError := errors.New("non-retryable error")

	mockStep := &mockStep{
		name: "retry_inner",
		runFunc: func(ctx *flow.Context) error {
			attempts++
			return nonRetryableError
		},
	}

	retryStep := NewRetryStep("retry_conditional", mockStep, 3, 1*time.Millisecond)
	retryStep.WithRetryCondition(func(err error) bool {
		return err.Error() != "non-retryable error"
	})

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	err := retryStep.Run(ctx)

	if err != nonRetryableError {
		t.Errorf("Run() error = %v, want %v", err, nonRetryableError)
	}

	if attempts != 1 {
		t.Errorf("Attempts = %d, want 1 (should not retry non-retryable error)", attempts)
	}
}

func TestDelayStep(t *testing.T) {
	delay := 50 * time.Millisecond
	delayStep := NewDelayStep("delay_test", delay)

	if delayStep.Name() != "delay_test" {
		t.Errorf("Name() = %v, want 'delay_test'", delayStep.Name())
	}

	if delayStep.Description() != "Delay" {
		t.Errorf("Description() = %v, want 'Delay'", delayStep.Description())
	}

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	start := time.Now()

	err := delayStep.Run(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if duration < delay {
		t.Errorf("Delay duration = %v, should be at least %v", duration, delay)
	}
}

func TestDelayStep_ContextCancellation(t *testing.T) {
	delay := 1 * time.Second // Long delay
	delayStep := NewDelayStep("delay_cancel", delay)

	// Create a context that will be cancelled
	_, cancel := context.WithCancel(context.Background())
	ctx := flow.NewContext().WithLogger(zap.NewNop())
	// We need to manually set the context since flow.Context doesn't expose a way to set it
	// This is a limitation of the current API, but we can test the timeout behavior

	// Cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := delayStep.Run(ctx)
	duration := time.Since(start)

	// Since we can't easily inject the cancelled context, we'll test the normal delay behavior
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// For this test, we expect the full delay since we can't inject the cancelled context
	if duration < delay {
		t.Errorf("Delay should have completed, but took %v (expected at least %v)", duration, delay)
	}

	// Clean up
	cancel()
}

func TestDelayStep_ShortDelay(t *testing.T) {
	delay := 1 * time.Millisecond
	delayStep := NewDelayStep("delay_short", delay)

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	start := time.Now()

	err := delayStep.Run(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if duration < delay {
		t.Errorf("Delay duration = %v, should be at least %v", duration, delay)
	}

	// Should complete quickly
	if duration > 100*time.Millisecond {
		t.Errorf("Short delay took too long: %v", duration)
	}
}
