package bff

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
)

// MockStep for testing
type MockStep struct {
	mock.Mock
	name        string
	description string
}

func NewMockStep(name string) *MockStep {
	return &MockStep{
		name:        name,
		description: "Mock step: " + name,
	}
}

func (m *MockStep) Name() string {
	return m.name
}

func (m *MockStep) Description() string {
	return m.description
}

func (m *MockStep) Run(ctx *flow.Context) error {
	args := m.Called(ctx)

	// Simulate setting some data in context
	ctx.Set(m.name, map[string]interface{}{
		"id":   123,
		"data": "test_data_" + m.name,
	})

	return args.Error(0)
}

func (m *MockStep) SetTimeout(timeout time.Duration) base.Step {
	return m
}

func (m *MockStep) GetTimeout() time.Duration {
	return 30 * time.Second
}

// MockTransformer for testing - fixed to match the correct interface
type MockTransformer struct {
	mock.Mock
}

func (m *MockTransformer) Name() string {
	return "mock_transformer"
}

func (m *MockTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(data)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestNewAggregationStep(t *testing.T) {
	step := NewAggregationStep("test_aggregation")

	assert.Equal(t, "test_aggregation", step.Name())
	assert.Contains(t, step.Description(), "BFF Aggregation: test_aggregation")
	assert.True(t, step.parallel)
	assert.Equal(t, 30*time.Second, step.timeout)
	assert.False(t, step.failFast)
	assert.NotNil(t, step.steps)
	assert.NotNil(t, step.required)
	assert.NotNil(t, step.fallbacks)
	assert.Len(t, step.steps, 0)
}

func TestAggregationStep_AddStep(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation")
	mockStep := NewMockStep("test_step")

	result := aggregation.AddStep(mockStep)

	assert.Equal(t, aggregation, result) // Fluent API
	assert.Len(t, aggregation.steps, 1)
	assert.Equal(t, mockStep, aggregation.steps[0])
	assert.False(t, aggregation.required["test_step"]) // Default is not required
}

func TestAggregationStep_AddRequiredStep(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation")
	mockStep := NewMockStep("required_step")

	result := aggregation.AddRequiredStep(mockStep)

	assert.Equal(t, aggregation, result) // Fluent API
	assert.Len(t, aggregation.steps, 1)
	assert.Equal(t, mockStep, aggregation.steps[0])
	assert.True(t, aggregation.required["required_step"])
}

func TestAggregationStep_AddOptionalStep(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation")
	mockStep := NewMockStep("optional_step")
	fallbackData := map[string]interface{}{
		"fallback": "data",
	}

	result := aggregation.AddOptionalStep(mockStep, fallbackData)

	assert.Equal(t, aggregation, result) // Fluent API
	assert.Len(t, aggregation.steps, 1)
	assert.Equal(t, mockStep, aggregation.steps[0])
	assert.False(t, aggregation.required["optional_step"])
	assert.Equal(t, fallbackData, aggregation.fallbacks["optional_step"])
}

func TestAggregationStep_Configuration(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation")
	mockTransformer := &MockTransformer{}

	// Test fluent configuration
	result := aggregation.
		WithTransformer(mockTransformer).
		WithParallel(false).
		WithTimeout(60 * time.Second).
		WithFailFast(true)

	assert.Equal(t, aggregation, result) // Fluent API
	assert.Equal(t, mockTransformer, aggregation.transformer)
	assert.False(t, aggregation.parallel)
	assert.Equal(t, 60*time.Second, aggregation.timeout)
	assert.True(t, aggregation.failFast)
}

func TestAggregationStep_RunParallel_Success(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation").
		WithParallel(true).
		WithTimeout(5 * time.Second)

	// Create mock steps
	step1 := NewMockStep("step1")
	step2 := NewMockStep("step2")
	step3 := NewMockStep("step3")

	step1.On("Run", mock.Anything).Return(nil)
	step2.On("Run", mock.Anything).Return(nil)
	step3.On("Run", mock.Anything).Return(nil)

	aggregation.AddStep(step1).AddStep(step2).AddStep(step3)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.NoError(t, err)

	// Check that aggregated results are stored
	bffResults, exists := ctx.Get("bff_aggregation")
	assert.True(t, exists)
	assert.NotNil(t, bffResults)

	namedResults, exists := ctx.Get("bff_test_aggregation")
	assert.True(t, exists)
	assert.NotNil(t, namedResults)

	// Verify all steps were called
	step1.AssertExpectations(t)
	step2.AssertExpectations(t)
	step3.AssertExpectations(t)
}

func TestAggregationStep_RunSequential_Success(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation").
		WithParallel(false).
		WithTimeout(5 * time.Second)

	// Create mock steps
	step1 := NewMockStep("step1")
	step2 := NewMockStep("step2")

	step1.On("Run", mock.Anything).Return(nil)
	step2.On("Run", mock.Anything).Return(nil)

	aggregation.AddStep(step1).AddStep(step2)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.NoError(t, err)

	// Check that aggregated results are stored
	bffResults, exists := ctx.Get("bff_aggregation")
	assert.True(t, exists)
	assert.NotNil(t, bffResults)

	// Verify all steps were called
	step1.AssertExpectations(t)
	step2.AssertExpectations(t)
}

func TestAggregationStep_RunWithTransformer(t *testing.T) {
	mockTransformer := &MockTransformer{}
	transformedData := map[string]interface{}{
		"transformed": "data",
		"count":       2,
	}
	mockTransformer.On("Transform", mock.Anything).Return(transformedData, nil)

	aggregation := NewAggregationStep("test_aggregation").
		WithTransformer(mockTransformer).
		WithParallel(true)

	step1 := NewMockStep("step1")
	step1.On("Run", mock.Anything).Return(nil)
	aggregation.AddStep(step1)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.NoError(t, err)

	// Check that transformed results are stored
	bffResults, exists := ctx.Get("bff_aggregation")
	assert.True(t, exists)
	assert.Equal(t, transformedData, bffResults)

	mockTransformer.AssertExpectations(t)
	step1.AssertExpectations(t)
}

func TestAggregationStep_RunWithTransformerError(t *testing.T) {
	mockTransformer := &MockTransformer{}
	mockTransformer.On("Transform", mock.Anything).Return(map[string]interface{}{}, fmt.Errorf("transformation failed"))

	aggregation := NewAggregationStep("test_aggregation").
		WithTransformer(mockTransformer)

	step1 := NewMockStep("step1")
	step1.On("Run", mock.Anything).Return(nil)
	aggregation.AddStep(step1)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "aggregation transformation failed")

	mockTransformer.AssertExpectations(t)
	step1.AssertExpectations(t)
}

func TestAggregationStep_RunWithRequiredStepFailure(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation").
		WithParallel(true).
		WithFailFast(false)

	// Create a required step that fails
	requiredStep := NewMockStep("required_step")
	requiredStep.On("Run", mock.Anything).Return(fmt.Errorf("required step failed"))

	// Create an optional step that succeeds
	optionalStep := NewMockStep("optional_step")
	optionalStep.On("Run", mock.Anything).Return(nil)

	aggregation.
		AddRequiredStep(requiredStep).
		AddOptionalStep(optionalStep, map[string]interface{}{"fallback": "data"})

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	// Should fail because required step failed without fallback
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required step required_step failed")

	requiredStep.AssertExpectations(t)
	optionalStep.AssertExpectations(t)
}

func TestAggregationStep_RunWithOptionalStepFailure(t *testing.T) {
	aggregation := NewAggregationStep("test_aggregation").
		WithParallel(true)

	// Create an optional step that fails
	optionalStep := NewMockStep("optional_step")
	optionalStep.On("Run", mock.Anything).Return(fmt.Errorf("optional step failed"))

	fallbackData := map[string]interface{}{
		"fallback": "data",
		"status":   "fallback_used",
	}

	aggregation.AddOptionalStep(optionalStep, fallbackData)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.NoError(t, err)

	// Check that aggregated results contain fallback data
	bffResults, exists := ctx.Get("bff_aggregation")
	assert.True(t, exists)

	resultsMap, ok := bffResults.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, fallbackData, resultsMap["optional_step"])

	optionalStep.AssertExpectations(t)
}

func TestAggregationStep_RunWithMixedSteps(t *testing.T) {
	aggregation := NewAggregationStep("mixed_aggregation").
		WithParallel(true)

	// Create steps with different configurations
	successStep := NewMockStep("success_step")
	successStep.On("Run", mock.Anything).Return(nil)

	failingOptionalStep := NewMockStep("failing_optional")
	failingOptionalStep.On("Run", mock.Anything).Return(fmt.Errorf("optional failure"))

	requiredStep := NewMockStep("required_step")
	requiredStep.On("Run", mock.Anything).Return(nil)

	fallbackData := map[string]interface{}{
		"fallback": "optional_data",
	}

	aggregation.
		AddStep(successStep).
		AddOptionalStep(failingOptionalStep, fallbackData).
		AddRequiredStep(requiredStep)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.NoError(t, err)

	// Check results
	bffResults, exists := ctx.Get("bff_aggregation")
	assert.True(t, exists)

	resultsMap, ok := bffResults.(map[string]interface{})
	assert.True(t, ok)

	// Should have data from success_step and required_step
	assert.Contains(t, resultsMap, "success_step")
	assert.Contains(t, resultsMap, "required_step")

	// Should have fallback data for failing_optional
	assert.Equal(t, fallbackData, resultsMap["failing_optional"])

	successStep.AssertExpectations(t)
	failingOptionalStep.AssertExpectations(t)
	requiredStep.AssertExpectations(t)
}

func TestAggregationStep_RunSequentialWithFailure(t *testing.T) {
	aggregation := NewAggregationStep("sequential_aggregation").
		WithParallel(false).
		WithFailFast(false)

	step1 := NewMockStep("step1")
	step1.On("Run", mock.Anything).Return(nil)

	step2 := NewMockStep("step2")
	step2.On("Run", mock.Anything).Return(fmt.Errorf("step2 failed"))

	step3 := NewMockStep("step3")
	step3.On("Run", mock.Anything).Return(nil)

	aggregation.
		AddStep(step1).
		AddOptionalStep(step2, map[string]interface{}{"fallback": "step2_fallback"}).
		AddStep(step3)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.NoError(t, err)

	step1.AssertExpectations(t)
	step2.AssertExpectations(t)
	step3.AssertExpectations(t)
}

func TestAggregationStep_EmptySteps(t *testing.T) {
	aggregation := NewAggregationStep("empty_aggregation")

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	assert.NoError(t, err)

	// Should still create empty results
	bffResults, exists := ctx.Get("bff_aggregation")
	assert.True(t, exists)

	resultsMap, ok := bffResults.(map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, resultsMap, 0)
}

func TestNewMobileDashboardAggregation(t *testing.T) {
	baseURL := "https://api.example.com"
	aggregation := NewMobileDashboardAggregation(baseURL)

	assert.NotNil(t, aggregation)
	assert.Equal(t, "mobile_dashboard", aggregation.Name())
	assert.Contains(t, aggregation.Description(), "BFF Aggregation: mobile_dashboard")
	assert.True(t, aggregation.parallel)
	assert.Len(t, aggregation.steps, 3)
}

func TestNewMobileSearchAggregation(t *testing.T) {
	baseURL := "https://api.example.com"
	aggregation := NewMobileSearchAggregation(baseURL)

	assert.NotNil(t, aggregation)
	assert.Equal(t, "mobile_search", aggregation.Name())
	assert.Contains(t, aggregation.Description(), "BFF Aggregation: mobile_search")
	assert.True(t, aggregation.parallel)
	assert.Len(t, aggregation.steps, 2)
}

func TestNewMobileProfileAggregation(t *testing.T) {
	baseURL := "https://api.example.com"
	aggregation := NewMobileProfileAggregation(baseURL)

	assert.NotNil(t, aggregation)
	assert.Equal(t, "mobile_profile", aggregation.Name())
	assert.Contains(t, aggregation.Description(), "BFF Aggregation: mobile_profile")
	assert.True(t, aggregation.parallel)
	assert.Len(t, aggregation.steps, 2)
}

func TestAggregationStep_Timeout(t *testing.T) {
	aggregation := NewAggregationStep("timeout_test").
		WithTimeout(100 * time.Millisecond).
		WithParallel(true)

	// Create a step that takes longer than the timeout
	slowStep := NewMockStep("slow_step")
	slowStep.On("Run", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(200 * time.Millisecond)
	}).Return(nil)

	aggregation.AddStep(slowStep)

	ctx := flow.NewContext().WithFlowName("test_flow")
	err := aggregation.Run(ctx)

	// Should complete (timeout is handled at context level)
	assert.NoError(t, err)
}

func TestAggregationStep_FluentAPI(t *testing.T) {
	mockTransformer := &MockTransformer{}
	step1 := NewMockStep("step1")
	step2 := NewMockStep("step2")

	// Test complete fluent API chain
	aggregation := NewAggregationStep("fluent_test").
		AddStep(step1).
		AddRequiredStep(step2).
		AddOptionalStep(NewMockStep("optional"), map[string]interface{}{"fallback": true}).
		WithTransformer(mockTransformer).
		WithParallel(false).
		WithTimeout(45 * time.Second).
		WithFailFast(true)

	assert.Equal(t, "fluent_test", aggregation.Name())
	assert.Equal(t, mockTransformer, aggregation.transformer)
	assert.False(t, aggregation.parallel)
	assert.Equal(t, 45*time.Second, aggregation.timeout)
	assert.True(t, aggregation.failFast)
	assert.Len(t, aggregation.steps, 3)
	assert.True(t, aggregation.required["step2"])
	assert.False(t, aggregation.required["optional"])
	assert.Contains(t, aggregation.fallbacks, "optional")
}
