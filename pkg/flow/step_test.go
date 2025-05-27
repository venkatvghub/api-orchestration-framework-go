package flow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// Mock step for testing that implements interfaces.Step
type mockStep struct {
	name        string
	description string
	runFunc     func(interfaces.ExecutionContext) error
}

func (m *mockStep) Run(ctx interfaces.ExecutionContext) error {
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

// Mock interfaces.Step for testing wrapper
type mockInterfaceStep struct {
	name        string
	description string
	runFunc     func(interfaces.ExecutionContext) error
}

func (m *mockInterfaceStep) Run(ctx interfaces.ExecutionContext) error {
	if m.runFunc != nil {
		return m.runFunc(ctx)
	}
	return nil
}

func (m *mockInterfaceStep) Name() string {
	return m.name
}

func (m *mockInterfaceStep) Description() string {
	return m.description
}

func TestStepWrapper(t *testing.T) {
	mockInterfaceStep := &mockInterfaceStep{
		name:        "test_step",
		description: "Test step description",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			ctx.Set("test_key", "test_value")
			return nil
		},
	}

	wrapper := NewStepWrapper(mockInterfaceStep)

	if wrapper.Name() != "test_step" {
		t.Errorf("Name() = %v, want 'test_step'", wrapper.Name())
	}

	if wrapper.Description() != "Test step description" {
		t.Errorf("Description() = %v, want 'Test step description'", wrapper.Description())
	}

	ctx := NewContext()
	err := wrapper.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	val, ok := ctx.Get("test_key")
	if !ok || val != "test_value" {
		t.Error("Wrapper should execute the wrapped step")
	}
}

func TestStepWrapper_Error(t *testing.T) {
	expectedError := errors.New("step error")
	mockInterfaceStep := &mockInterfaceStep{
		name:        "error_step",
		description: "Error step",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			return expectedError
		},
	}

	wrapper := NewStepWrapper(mockInterfaceStep)
	ctx := NewContext()
	err := wrapper.Run(ctx)

	if err != expectedError {
		t.Errorf("Run() error = %v, want %v", err, expectedError)
	}
}

func TestStepFunc(t *testing.T) {
	executed := false
	stepFunc := StepFunc(func(ctx interfaces.ExecutionContext) error {
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

	ctx := NewContext()
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
	stepFunc := StepFunc(func(ctx interfaces.ExecutionContext) error {
		return expectedError
	})

	ctx := NewContext()
	err := stepFunc.Run(ctx)

	if err != expectedError {
		t.Errorf("Run() error = %v, want %v", err, expectedError)
	}
}

func TestBaseStep(t *testing.T) {
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
}

func TestConditionalStep(t *testing.T) {
	executed := false
	mockStep := &mockStep{
		name:        "conditional_inner",
		description: "Inner step",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			executed = true
			return nil
		},
	}

	// Test condition that returns true
	condition := func(ctx interfaces.ExecutionContext) bool {
		return ctx.Has("execute")
	}

	conditionalStep := NewConditionalStep("conditional_test", condition, mockStep)

	if conditionalStep.Name() != "conditional_test" {
		t.Errorf("Name() = %v, want 'conditional_test'", conditionalStep.Name())
	}

	ctx := NewContext().WithLogger(zap.NewNop())
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
	ctx2 := NewContext().WithLogger(zap.NewNop())
	// Don't set "execute" key

	err = conditionalStep.Run(ctx2)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if executed {
		t.Error("Inner step should not have been executed when condition is false")
	}
}

func TestParallelStep(t *testing.T) {
	step1Executed := false
	step2Executed := false
	step3Executed := false

	step1 := &mockStep{
		name: "step1",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			step1Executed = true
			ctx.Set("step1", "done")
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			step2Executed = true
			ctx.Set("step2", "done")
			return nil
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			step3Executed = true
			ctx.Set("step3", "done")
			return nil
		},
	}

	parallelStep := NewParallelStep("parallel_test", step1, step2, step3)

	if parallelStep.Name() != "parallel_test" {
		t.Errorf("Name() = %v, want 'parallel_test'", parallelStep.Name())
	}

	ctx := NewContext().WithLogger(zap.NewNop())
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
	ctx := NewContext()

	err := parallelStep.Run(ctx)
	if err != nil {
		t.Errorf("Empty parallel step should not fail: %v", err)
	}
}

func TestParallelStep_Error(t *testing.T) {
	expectedError := errors.New("step error")

	step1 := &mockStep{
		name: "step1",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			return expectedError
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			return nil
		},
	}

	parallelStep := NewParallelStep("parallel_error", step1, step2, step3)
	ctx := NewContext().WithLogger(zap.NewNop())

	err := parallelStep.Run(ctx)
	if err != expectedError {
		t.Errorf("Run() error = %v, want %v", err, expectedError)
	}
}

func TestSequentialStep(t *testing.T) {
	executionOrder := []string{}

	step1 := &mockStep{
		name: "step1",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			executionOrder = append(executionOrder, "step1")
			ctx.Set("step1", "done")
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			executionOrder = append(executionOrder, "step2")
			ctx.Set("step2", "done")
			return nil
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			executionOrder = append(executionOrder, "step3")
			ctx.Set("step3", "done")
			return nil
		},
	}

	sequentialStep := NewSequentialStep("sequential_test", step1, step2, step3)

	if sequentialStep.Name() != "sequential_test" {
		t.Errorf("Name() = %v, want 'sequential_test'", sequentialStep.Name())
	}

	ctx := NewContext().WithLogger(zap.NewNop())
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
		runFunc: func(ctx interfaces.ExecutionContext) error {
			executionOrder = append(executionOrder, "step1")
			return nil
		},
	}

	step2 := &mockStep{
		name: "step2",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			executionOrder = append(executionOrder, "step2")
			return expectedError
		},
	}

	step3 := &mockStep{
		name: "step3",
		runFunc: func(ctx interfaces.ExecutionContext) error {
			executionOrder = append(executionOrder, "step3")
			return nil
		},
	}

	sequentialStep := NewSequentialStep("sequential_error", step1, step2, step3)
	ctx := NewContext().WithLogger(zap.NewNop())

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

func TestTransformStep(t *testing.T) {
	transformer := func(ctx interfaces.ExecutionContext) error {
		val, _ := ctx.GetString("input")
		ctx.Set("output", "transformed_"+val)
		return nil
	}

	transformStep := NewTransformStep("transform_test", transformer)

	if transformStep.Name() != "transform_test" {
		t.Errorf("Name() = %v, want 'transform_test'", transformStep.Name())
	}

	if transformStep.Description() != "Data transformation step" {
		t.Errorf("Description() = %v, want 'Data transformation step'", transformStep.Description())
	}

	ctx := NewContext().WithLogger(zap.NewNop())
	ctx.Set("input", "data")

	err := transformStep.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	output, err := ctx.GetString("output")
	if err != nil {
		t.Errorf("Failed to get output: %v", err)
	}

	if output != "transformed_data" {
		t.Errorf("Output = %v, want 'transformed_data'", output)
	}
}

func TestTransformStep_Error(t *testing.T) {
	expectedError := errors.New("transform error")
	transformer := func(ctx interfaces.ExecutionContext) error {
		return expectedError
	}

	transformStep := NewTransformStep("transform_error", transformer)
	ctx := NewContext().WithLogger(zap.NewNop())

	err := transformStep.Run(ctx)
	if err != expectedError {
		t.Errorf("Run() error = %v, want %v", err, expectedError)
	}
}

func TestRetryStep(t *testing.T) {
	attempts := 0
	mockStep := &mockStep{
		name: "retry_inner",
		runFunc: func(ctx interfaces.ExecutionContext) error {
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

	ctx := NewContext().WithLogger(zap.NewNop())
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
		runFunc: func(ctx interfaces.ExecutionContext) error {
			attempts++
			return errors.New("persistent error")
		},
	}

	retryStep := NewRetryStep("retry_fail", mockStep, 2, 1*time.Millisecond)
	ctx := NewContext().WithLogger(zap.NewNop())

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
		runFunc: func(ctx interfaces.ExecutionContext) error {
			attempts++
			return nonRetryableError
		},
	}

	retryStep := NewRetryStep("retry_conditional", mockStep, 3, 1*time.Millisecond)
	retryStep.WithRetryCondition(func(err error) bool {
		return err.Error() != "non-retryable error"
	})

	ctx := NewContext().WithLogger(zap.NewNop())
	err := retryStep.Run(ctx)

	if err == nil {
		t.Error("Run() should have failed")
	}

	// Check if the error contains the original error
	if !errors.Is(err, nonRetryableError) && err.Error() != nonRetryableError.Error() {
		// The error might be wrapped, so check if it contains the original message
		if err.Error() != "non-retryable error" && !strings.Contains(err.Error(), "non-retryable error") {
			t.Errorf("Run() error = %v, should contain 'non-retryable error'", err)
		}
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

	expectedDesc := "Delay for 50ms"
	if delayStep.Description() != expectedDesc {
		t.Errorf("Description() = %v, want %v", delayStep.Description(), expectedDesc)
	}

	ctx := NewContext().WithLogger(zap.NewNop())
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

	// Create a context that will be cancelled manually
	baseCtx, cancel := context.WithCancel(context.Background())
	ctx := &Context{
		ctx:         baseCtx,
		values:      make(map[string]interface{}),
		executionID: generateExecutionID(),
		startTime:   time.Now(),
		timeout:     0, // No timeout, we'll cancel manually
		config:      config.DefaultConfig(),
	}
	ctx.WithLogger(zap.NewNop())

	// Cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := delayStep.Run(ctx)
	duration := time.Since(start)

	if err == nil {
		t.Error("Run() should have failed due to context cancellation")
	}

	// Should have been cancelled before the full delay
	if duration >= delay {
		t.Errorf("Delay should have been cancelled, but took %v (expected less than %v)", duration, delay)
	}

	// Should be around 50ms (the cancellation delay)
	if duration < 40*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Duration %v should be around 50ms (cancellation time)", duration)
	}
}

func TestNamedStep(t *testing.T) {
	innerStep := &mockStep{
		name:        "inner",
		description: "Inner step",
	}

	namedStep := &namedStep{
		Step: innerStep,
		name: "custom_name",
	}

	if namedStep.Name() != "custom_name" {
		t.Errorf("Name() = %v, want 'custom_name'", namedStep.Name())
	}

	if namedStep.Description() != "Inner step" {
		t.Errorf("Description() = %v, want 'Inner step'", namedStep.Description())
	}

	ctx := NewContext()
	err := namedStep.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}
}
