package interfaces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockExecutionContext is a mock implementation of ExecutionContext for testing.
// We will add more fields and methods as needed for specific tests.
// For now, it includes all methods to satisfy the interface.
type MockExecutionContext struct {
	data map[string]interface{}
}

func NewMockExecutionContext() *MockExecutionContext {
	return &MockExecutionContext{
		data: make(map[string]interface{}),
	}
}

func (m *MockExecutionContext) Set(key string, value interface{})               { m.data[key] = value }
func (m *MockExecutionContext) Get(key string) (interface{}, bool)            { v, ok := m.data[key]; return v, ok }
func (m *MockExecutionContext) Has(key string) bool                           { _, ok := m.data[key]; return ok }
func (m *MockExecutionContext) Delete(key string)                             { delete(m.data, key) }
func (m *MockExecutionContext) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}
func (m *MockExecutionContext) GetString(key string) (string, error)          { return "", nil } // Simplified for mock
func (m *MockExecutionContext) GetInt(key string) (int, error)                { return 0, nil }    // Simplified for mock
func (m *MockExecutionContext) GetBool(key string) (bool, error)              { return false, nil } // Simplified for mock
func (m *MockExecutionContext) GetMap(key string) (map[string]interface{}, error) { return nil, nil }   // Simplified for mock
func (m *MockExecutionContext) Context() context.Context                      { return context.Background() }
func (m *MockExecutionContext) Clone() ExecutionContext                       { return NewMockExecutionContext() } // Simplified clone
func (m *MockExecutionContext) FlowName() string                            { return "test-flow" }
func (m *MockExecutionContext) ExecutionID() string                         { return "test-exec-id" }
func (m *MockExecutionContext) StartTime() time.Time                          { return time.Now() }
func (m *MockExecutionContext) Duration() time.Duration                       { return 0 }
func (m *MockExecutionContext) Logger() *zap.Logger                           { return zap.NewNop() }

// MockStep is a mock implementation of Step for testing.
type MockStep struct {
	MockName        string
	MockDescription string
	RunFunc         func(ctx ExecutionContext) error
}

func (ms *MockStep) Run(ctx ExecutionContext) error {
	if ms.RunFunc != nil {
		return ms.RunFunc(ctx)
	}
	return nil
}

func (ms *MockStep) Name() string {
	return ms.MockName
}

func (ms *MockStep) Description() string {
	return ms.MockDescription
}

func TestExecutionContextInterface(t *testing.T) {
	var ec ExecutionContext
	// Assign a mock implementation to the interface variable
	ec = NewMockExecutionContext()

	assert.NotNil(t, ec, "ExecutionContext should be assignable with a mock implementation")

	// Basic Set/Get test
	ec.Set("testKey", "testValue")
	val, ok := ec.Get("testKey")
	assert.True(t, ok, "Key should exist")
	assert.Equal(t, "testValue", val, "Value should match")
}

func TestStepInterface(t *testing.T) {
	var s Step
	// Assign a mock implementation to the interface variable
	s = &MockStep{
		MockName:        "Test Step",
		MockDescription: "This is a test step.",
	}

	assert.NotNil(t, s, "Step should be assignable with a mock implementation")
	assert.Equal(t, "Test Step", s.Name(), "Name should match")
	assert.Equal(t, "This is a test step.", s.Description(), "Description should match")

	// Test Run method
	mockCtx := NewMockExecutionContext()
	err := s.Run(mockCtx)
	assert.NoError(t, err, "Run should not return an error for basic mock")
}