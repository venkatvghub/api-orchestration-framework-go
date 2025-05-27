package flow

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewFlow(t *testing.T) {
	flow := NewFlow("test_flow")

	if flow == nil {
		t.Fatal("NewFlow() returned nil")
	}

	if flow.name != "test_flow" {
		t.Errorf("Flow name = %v, want 'test_flow'", flow.name)
	}

	if flow.description != "Flow: test_flow" {
		t.Errorf("Flow description = %v, want 'Flow: test_flow'", flow.description)
	}

	if len(flow.steps) != 0 {
		t.Errorf("New flow should have 0 steps, got %d", len(flow.steps))
	}

	if len(flow.middleware) != 0 {
		t.Errorf("New flow should have 0 middleware, got %d", len(flow.middleware))
	}

	if flow.timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want 30s", flow.timeout)
	}
}

func TestFlow_WithDescription(t *testing.T) {
	flow := NewFlow("test_flow")
	description := "Custom description"

	result := flow.WithDescription(description)

	if result != flow {
		t.Error("WithDescription should return the same flow instance")
	}

	if flow.description != description {
		t.Errorf("Description = %v, want %v", flow.description, description)
	}
}

func TestFlow_WithTimeout(t *testing.T) {
	flow := NewFlow("test_flow")
	timeout := 60 * time.Second

	result := flow.WithTimeout(timeout)

	if result != flow {
		t.Error("WithTimeout should return the same flow instance")
	}

	if flow.timeout != timeout {
		t.Errorf("Timeout = %v, want %v", flow.timeout, timeout)
	}
}

func TestFlow_Use(t *testing.T) {
	flow := NewFlow("test_flow")
	middlewareCalled := false

	middleware := func(next func(*Context) error) func(*Context) error {
		return func(ctx *Context) error {
			middlewareCalled = true
			return next(ctx)
		}
	}

	result := flow.Use(middleware)

	if result != flow {
		t.Error("Use should return the same flow instance")
	}

	if len(flow.middleware) != 1 {
		t.Errorf("Flow should have 1 middleware, got %d", len(flow.middleware))
	}

	// Test middleware execution
	ctx := NewContext().WithLogger(zap.NewNop())
	_, err := flow.Execute(ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if !middlewareCalled {
		t.Error("Middleware should have been called")
	}
}

func TestFlow_Step(t *testing.T) {
	flow := NewFlow("test_flow")
	step := &mockStep{name: "test_step", description: "Test step"}

	result := flow.Step("test_step", step)

	if result != flow {
		t.Error("Step should return the same flow instance")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should have 1 step, got %d", len(flow.steps))
	}

	if flow.steps[0].Name() != "test_step" {
		t.Errorf("Step name = %v, want 'test_step'", flow.steps[0].Name())
	}
}

func TestFlow_Step_AnonymousStep(t *testing.T) {
	flow := NewFlow("test_flow")
	step := StepFunc(func(ctx *Context) error { return nil })

	result := flow.Step("custom_name", step)

	if result != flow {
		t.Error("Step should return the same flow instance")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should have 1 step, got %d", len(flow.steps))
	}

	// Should wrap anonymous step with custom name
	if flow.steps[0].Name() != "custom_name" {
		t.Errorf("Step name = %v, want 'custom_name'", flow.steps[0].Name())
	}
}

func TestFlow_StepFunc(t *testing.T) {
	flow := NewFlow("test_flow")
	executed := false

	stepFunc := func(ctx *Context) error {
		executed = true
		ctx.Set("step_executed", true)
		return nil
	}

	result := flow.StepFunc("func_step", stepFunc)

	if result != flow {
		t.Error("StepFunc should return the same flow instance")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should have 1 step, got %d", len(flow.steps))
	}

	if flow.steps[0].Name() != "func_step" {
		t.Errorf("Step name = %v, want 'func_step'", flow.steps[0].Name())
	}

	// Test execution
	ctx := NewContext()
	err := flow.steps[0].Run(ctx)
	if err != nil {
		t.Errorf("Step execution failed: %v", err)
	}

	if !executed {
		t.Error("Step function should have been executed")
	}

	val, _ := ctx.GetBool("step_executed")
	if !val {
		t.Error("Step should have set the value")
	}
}

func TestFlow_Transform(t *testing.T) {
	flow := NewFlow("test_flow")

	transformer := func(ctx *Context) error {
		val, _ := ctx.GetString("input")
		ctx.Set("output", "transformed_"+val)
		return nil
	}

	result := flow.Transform("transform_step", transformer)

	if result != flow {
		t.Error("Transform should return the same flow instance")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should have 1 step, got %d", len(flow.steps))
	}

	// Test execution
	ctx := NewContext()
	ctx.Set("input", "data")

	err := flow.steps[0].Run(ctx)
	if err != nil {
		t.Errorf("Transform step failed: %v", err)
	}

	output, _ := ctx.GetString("output")
	if output != "transformed_data" {
		t.Errorf("Output = %v, want 'transformed_data'", output)
	}
}

func TestFlow_Delay(t *testing.T) {
	flow := NewFlow("test_flow")
	delay := 10 * time.Millisecond

	result := flow.Delay("delay_step", delay)

	if result != flow {
		t.Error("Delay should return the same flow instance")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should have 1 step, got %d", len(flow.steps))
	}

	// Test execution
	ctx := NewContext()
	start := time.Now()

	err := flow.steps[0].Run(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Delay step failed: %v", err)
	}

	if duration < delay {
		t.Errorf("Delay duration = %v, should be at least %v", duration, delay)
	}
}

func TestFlow_Retry(t *testing.T) {
	flow := NewFlow("test_flow")

	// Add a step first
	step := &mockStep{name: "original_step"}
	flow.Step("original", step)

	result := flow.Retry(3, 1*time.Millisecond)

	if result != flow {
		t.Error("Retry should return the same flow instance")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should still have 1 step, got %d", len(flow.steps))
	}

	// The step should now be wrapped with retry logic
	stepName := flow.steps[0].Name()
	if stepName != "retry_original_step" {
		t.Errorf("Step name = %v, want 'retry_original_step'", stepName)
	}
}

func TestFlow_Retry_NoSteps(t *testing.T) {
	flow := NewFlow("test_flow")

	result := flow.Retry(3, 1*time.Millisecond)

	if result != flow {
		t.Error("Retry should return the same flow instance")
	}

	if len(flow.steps) != 0 {
		t.Errorf("Flow should have 0 steps, got %d", len(flow.steps))
	}
}

func TestFlow_Execute(t *testing.T) {
	flow := NewFlow("test_flow")
	step1Executed := false
	step2Executed := false

	flow.StepFunc("step1", func(ctx *Context) error {
		step1Executed = true
		ctx.Set("step1_result", "done")
		return nil
	})

	flow.StepFunc("step2", func(ctx *Context) error {
		step2Executed = true
		ctx.Set("step2_result", "done")
		return nil
	})

	ctx := NewContext().WithLogger(zap.NewNop())
	result, err := flow.Execute(ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Execute should return a result")
	}

	if result.FlowName != "test_flow" {
		t.Errorf("Result flow name = %v, want 'test_flow'", result.FlowName)
	}

	if !result.Success {
		t.Error("Result should indicate success")
	}

	if result.Error != nil {
		t.Errorf("Result should not have error, got %v", result.Error)
	}

	if !step1Executed || !step2Executed {
		t.Error("All steps should have been executed")
	}

	// Check context values
	val1, _ := ctx.GetString("step1_result")
	val2, _ := ctx.GetString("step2_result")

	if val1 != "done" || val2 != "done" {
		t.Error("Steps should have set their values")
	}
}

func TestFlow_Execute_WithError(t *testing.T) {
	flow := NewFlow("test_flow")
	expectedError := errors.New("step error")

	flow.StepFunc("step1", func(ctx *Context) error {
		return nil
	})

	flow.StepFunc("step2", func(ctx *Context) error {
		return expectedError
	})

	flow.StepFunc("step3", func(ctx *Context) error {
		t.Error("Step3 should not be executed after step2 fails")
		return nil
	})

	ctx := NewContext().WithLogger(zap.NewNop())
	result, err := flow.Execute(ctx)

	if err == nil {
		t.Error("Execute should have failed")
	}

	if result == nil {
		t.Fatal("Execute should return a result even on failure")
	}

	if result.Success {
		t.Error("Result should indicate failure")
	}

	if result.Error == nil {
		t.Error("Result should have error")
	}
}

func TestFlow_Name(t *testing.T) {
	flow := NewFlow("test_flow")
	if flow.Name() != "test_flow" {
		t.Errorf("Name() = %v, want 'test_flow'", flow.Name())
	}
}

func TestFlow_Description(t *testing.T) {
	flow := NewFlow("test_flow").WithDescription("Custom description")
	if flow.Description() != "Custom description" {
		t.Errorf("Description() = %v, want 'Custom description'", flow.Description())
	}
}

func TestFlow_Steps(t *testing.T) {
	flow := NewFlow("test_flow")
	step1 := &mockStep{name: "step1"}
	step2 := &mockStep{name: "step2"}

	flow.Step("step1", step1).Step("step2", step2)

	steps := flow.Steps()
	if len(steps) != 2 {
		t.Errorf("Steps() length = %d, want 2", len(steps))
	}
}

func TestChoiceBuilder(t *testing.T) {
	flow := NewFlow("test_flow")
	executed := ""

	choice := flow.Choice("test_choice").
		When(func(ctx *Context) bool {
			return ctx.Has("condition1")
		}).
		StepFunc("branch1", func(ctx *Context) error {
			executed = "branch1"
			return nil
		}).
		When(func(ctx *Context) bool {
			return ctx.Has("condition2")
		}).
		StepFunc("branch2", func(ctx *Context) error {
			executed = "branch2"
			return nil
		}).
		Otherwise().
		StepFunc("default", func(ctx *Context) error {
			executed = "default"
			return nil
		}).
		EndChoice()

	if choice != flow {
		t.Error("EndChoice should return the original flow")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should have 1 choice step, got %d", len(flow.steps))
	}

	// Test first condition
	ctx1 := NewContext().WithLogger(zap.NewNop())
	ctx1.Set("condition1", true)

	_, err := flow.Execute(ctx1)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if executed != "branch1" {
		t.Errorf("Executed = %v, want 'branch1'", executed)
	}

	// Test second condition
	executed = ""
	ctx2 := NewContext().WithLogger(zap.NewNop())
	ctx2.Set("condition2", true)

	_, err = flow.Execute(ctx2)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if executed != "branch2" {
		t.Errorf("Executed = %v, want 'branch2'", executed)
	}

	// Test otherwise branch
	executed = ""
	ctx3 := NewContext().WithLogger(zap.NewNop())

	_, err = flow.Execute(ctx3)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if executed != "default" {
		t.Errorf("Executed = %v, want 'default'", executed)
	}
}

func TestChoiceBuilder_NoOtherwise(t *testing.T) {
	flow := NewFlow("test_flow")
	executed := false

	flow.Choice("test_choice").
		When(func(ctx *Context) bool {
			return ctx.Has("condition1")
		}).
		StepFunc("branch1", func(ctx *Context) error {
			executed = true
			return nil
		}).
		EndChoice()

	// Test with no matching condition and no otherwise
	ctx := NewContext().WithLogger(zap.NewNop())

	_, err := flow.Execute(ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if executed {
		t.Error("No branch should have executed")
	}
}

func TestParallelBuilder(t *testing.T) {
	flow := NewFlow("test_flow")
	step1Executed := false
	step2Executed := false
	step3Executed := false

	parallel := flow.Parallel("test_parallel").
		StepFunc("step1", func(ctx *Context) error {
			step1Executed = true
			ctx.Set("step1", "done")
			return nil
		}).
		StepFunc("step2", func(ctx *Context) error {
			step2Executed = true
			ctx.Set("step2", "done")
			return nil
		}).
		StepFunc("step3", func(ctx *Context) error {
			step3Executed = true
			ctx.Set("step3", "done")
			return nil
		}).
		EndParallel()

	if parallel != flow {
		t.Error("EndParallel should return the original flow")
	}

	if len(flow.steps) != 1 {
		t.Errorf("Flow should have 1 parallel step, got %d", len(flow.steps))
	}

	ctx := NewContext().WithLogger(zap.NewNop())
	_, err := flow.Execute(ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if !step1Executed || !step2Executed || !step3Executed {
		t.Error("All parallel steps should have been executed")
	}

	// Check that results were merged
	if !ctx.Has("step1") || !ctx.Has("step2") || !ctx.Has("step3") {
		t.Error("Parallel step results should be available in context")
	}
}

func TestExecutionResult_GetResponse(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)

	result := &ExecutionResult{
		FlowName:  "test_flow",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Duration:  100 * time.Millisecond,
		Success:   true,
		Error:     nil,
		Context:   ctx,
	}

	response := result.GetResponse()

	if response["flow_name"] != "test_flow" {
		t.Errorf("Response flow_name = %v, want 'test_flow'", response["flow_name"])
	}

	if response["success"] != true {
		t.Errorf("Response success = %v, want true", response["success"])
	}

	if response["duration_ms"] != int64(100) {
		t.Errorf("Response duration_ms = %v, want 100", response["duration_ms"])
	}

	if response["execution_id"] != ctx.ExecutionID() {
		t.Error("Response should include execution ID")
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should include data map")
	}

	if data["key1"] != "value1" || data["key2"] != 42 {
		t.Error("Response data should include context values")
	}

	// Test with error
	result.Success = false
	result.Error = errors.New("test error")

	response = result.GetResponse()
	if response["error"] != "test error" {
		t.Errorf("Response error = %v, want 'test error'", response["error"])
	}
}

func TestMiddleware(t *testing.T) {
	flow := NewFlow("test_flow")
	middlewareOrder := []string{}

	middleware1 := func(next func(*Context) error) func(*Context) error {
		return func(ctx *Context) error {
			middlewareOrder = append(middlewareOrder, "middleware1_before")
			err := next(ctx)
			middlewareOrder = append(middlewareOrder, "middleware1_after")
			return err
		}
	}

	middleware2 := func(next func(*Context) error) func(*Context) error {
		return func(ctx *Context) error {
			middlewareOrder = append(middlewareOrder, "middleware2_before")
			err := next(ctx)
			middlewareOrder = append(middlewareOrder, "middleware2_after")
			return err
		}
	}

	flow.Use(middleware1).Use(middleware2)

	flow.StepFunc("test_step", func(ctx *Context) error {
		middlewareOrder = append(middlewareOrder, "step_executed")
		return nil
	})

	ctx := NewContext().WithLogger(zap.NewNop())
	_, err := flow.Execute(ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	expectedOrder := []string{
		"middleware1_before", // First middleware added executes first
		"middleware2_before", // Second middleware added executes second
		"step_executed",
		"middleware2_after", // Second middleware completes first
		"middleware1_after", // First middleware completes last
	}

	if len(middlewareOrder) != len(expectedOrder) {
		t.Errorf("Middleware order length = %d, want %d", len(middlewareOrder), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(middlewareOrder) || middlewareOrder[i] != expected {
			t.Errorf("Middleware order[%d] = %v, want %v", i, middlewareOrder[i], expected)
		}
	}
}

func TestChoiceStep_Run(t *testing.T) {
	branches := []conditionalBranch{
		{
			condition: func(ctx *Context) bool { return ctx.Has("condition1") },
			steps: []Step{
				StepFunc(func(ctx *Context) error {
					ctx.Set("executed", "branch1")
					return nil
				}),
			},
		},
		{
			condition: func(ctx *Context) bool { return ctx.Has("condition2") },
			steps: []Step{
				StepFunc(func(ctx *Context) error {
					ctx.Set("executed", "branch2")
					return nil
				}),
			},
		},
	}

	otherwiseStep := StepFunc(func(ctx *Context) error {
		ctx.Set("executed", "otherwise")
		return nil
	})

	choiceStep := &choiceStep{
		BaseStep:  NewBaseStep("test_choice", "Test choice"),
		branches:  branches,
		otherwise: otherwiseStep,
	}

	// Test first condition
	ctx1 := NewContext().WithLogger(zap.NewNop())
	ctx1.Set("condition1", true)

	err := choiceStep.Run(ctx1)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	executed, _ := ctx1.GetString("executed")
	if executed != "branch1" {
		t.Errorf("Executed = %v, want 'branch1'", executed)
	}

	// Test otherwise
	ctx2 := NewContext().WithLogger(zap.NewNop())

	err = choiceStep.Run(ctx2)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	executed, _ = ctx2.GetString("executed")
	if executed != "otherwise" {
		t.Errorf("Executed = %v, want 'otherwise'", executed)
	}
}
