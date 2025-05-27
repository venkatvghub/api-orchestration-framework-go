package flow

import (
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()

	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}

	if ctx.ctx == nil {
		t.Error("Context should have a background context")
	}

	if ctx.values == nil {
		t.Error("Context should have initialized values map")
	}

	if ctx.executionID == "" {
		t.Error("Context should have an execution ID")
	}

	if ctx.startTime.IsZero() {
		t.Error("Context should have a start time")
	}

	if ctx.timeout != 30*time.Second {
		t.Errorf("Default timeout should be 30s, got %v", ctx.timeout)
	}

	if ctx.config == nil {
		t.Error("Context should have default config")
	}
}

func TestNewContextWithConfig(t *testing.T) {
	cfg := &config.FrameworkConfig{
		Timeouts: config.TimeoutConfig{
			FlowExecution: 60 * time.Second,
		},
	}

	ctx := NewContextWithConfig(cfg)

	if ctx == nil {
		t.Fatal("NewContextWithConfig() returned nil")
	}

	if ctx.timeout != 60*time.Second {
		t.Errorf("Timeout should be 60s, got %v", ctx.timeout)
	}

	if ctx.config != cfg {
		t.Error("Context should use provided config")
	}
}

func TestContext_SetAndGet(t *testing.T) {
	ctx := NewContext()

	// Test setting and getting a value
	ctx.Set("key1", "value1")
	val, ok := ctx.Get("key1")

	if !ok {
		t.Error("Get should return true for existing key")
	}

	if val != "value1" {
		t.Errorf("Get returned %v, want 'value1'", val)
	}

	// Test getting non-existent key
	_, ok = ctx.Get("nonexistent")
	if ok {
		t.Error("Get should return false for non-existent key")
	}
}

func TestContext_GetTyped(t *testing.T) {
	ctx := NewContext()

	// Test successful type assertion
	ctx.Set("string_key", "test_value")
	var result string
	err := ctx.GetTyped("string_key", &result)
	if err != nil {
		t.Errorf("GetTyped failed: %v", err)
	}

	if result != "test_value" {
		t.Errorf("GetTyped result = %v, want 'test_value'", result)
	}

	// Test non-existent key
	var nonExistent string
	err = ctx.GetTyped("nonexistent", &nonExistent)
	if err == nil {
		t.Error("GetTyped should return error for non-existent key")
	}

	// Test non-pointer target
	var notPointer string
	err = ctx.GetTyped("string_key", notPointer)
	if err == nil {
		t.Error("GetTyped should return error for non-pointer target")
	}

	// Test type mismatch
	ctx.Set("int_key", 42)
	var wrongType string
	err = ctx.GetTyped("int_key", &wrongType)
	if err == nil {
		t.Error("GetTyped should return error for type mismatch")
	}
}

func TestContext_GetString(t *testing.T) {
	ctx := NewContext()

	// Test successful string retrieval
	ctx.Set("string_key", "test_string")
	result, err := ctx.GetString("string_key")
	if err != nil {
		t.Errorf("GetString failed: %v", err)
	}

	if result != "test_string" {
		t.Errorf("GetString result = %v, want 'test_string'", result)
	}

	// Test non-existent key
	_, err = ctx.GetString("nonexistent")
	if err == nil {
		t.Error("GetString should return error for non-existent key")
	}

	// Test wrong type
	ctx.Set("int_key", 42)
	_, err = ctx.GetString("int_key")
	if err == nil {
		t.Error("GetString should return error for non-string value")
	}
}

func TestContext_GetInt(t *testing.T) {
	ctx := NewContext()

	// Test successful int retrieval
	ctx.Set("int_key", 42)
	result, err := ctx.GetInt("int_key")
	if err != nil {
		t.Errorf("GetInt failed: %v", err)
	}

	if result != 42 {
		t.Errorf("GetInt result = %v, want 42", result)
	}

	// Test non-existent key
	_, err = ctx.GetInt("nonexistent")
	if err == nil {
		t.Error("GetInt should return error for non-existent key")
	}

	// Test wrong type
	ctx.Set("string_key", "not_int")
	_, err = ctx.GetInt("string_key")
	if err == nil {
		t.Error("GetInt should return error for non-int value")
	}
}

func TestContext_GetBool(t *testing.T) {
	ctx := NewContext()

	// Test successful bool retrieval
	ctx.Set("bool_key", true)
	result, err := ctx.GetBool("bool_key")
	if err != nil {
		t.Errorf("GetBool failed: %v", err)
	}

	if result != true {
		t.Errorf("GetBool result = %v, want true", result)
	}

	// Test non-existent key
	_, err = ctx.GetBool("nonexistent")
	if err == nil {
		t.Error("GetBool should return error for non-existent key")
	}

	// Test wrong type
	ctx.Set("string_key", "not_bool")
	_, err = ctx.GetBool("string_key")
	if err == nil {
		t.Error("GetBool should return error for non-bool value")
	}
}

func TestContext_GetMap(t *testing.T) {
	ctx := NewContext()

	// Test successful map retrieval
	testMap := map[string]interface{}{"key": "value"}
	ctx.Set("map_key", testMap)
	result, err := ctx.GetMap("map_key")
	if err != nil {
		t.Errorf("GetMap failed: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("GetMap result incorrect, got %v", result)
	}

	// Test non-existent key
	_, err = ctx.GetMap("nonexistent")
	if err == nil {
		t.Error("GetMap should return error for non-existent key")
	}

	// Test wrong type
	ctx.Set("string_key", "not_map")
	_, err = ctx.GetMap("string_key")
	if err == nil {
		t.Error("GetMap should return error for non-map value")
	}
}

func TestContext_Has(t *testing.T) {
	ctx := NewContext()

	// Test non-existent key
	if ctx.Has("nonexistent") {
		t.Error("Has should return false for non-existent key")
	}

	// Test existing key
	ctx.Set("existing", "value")
	if !ctx.Has("existing") {
		t.Error("Has should return true for existing key")
	}
}

func TestContext_Delete(t *testing.T) {
	ctx := NewContext()

	// Set a value and verify it exists
	ctx.Set("to_delete", "value")
	if !ctx.Has("to_delete") {
		t.Error("Key should exist before deletion")
	}

	// Delete the value
	ctx.Delete("to_delete")
	if ctx.Has("to_delete") {
		t.Error("Key should not exist after deletion")
	}
}

func TestContext_Keys(t *testing.T) {
	ctx := NewContext()

	// Test empty context
	keys := ctx.Keys()
	if len(keys) != 0 {
		t.Errorf("Empty context should have 0 keys, got %d", len(keys))
	}

	// Add some keys
	ctx.Set("key1", "value1")
	ctx.Set("key2", "value2")
	ctx.Set("key3", "value3")

	keys = ctx.Keys()
	if len(keys) != 3 {
		t.Errorf("Context should have 3 keys, got %d", len(keys))
	}

	// Verify all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	expectedKeys := []string{"key1", "key2", "key3"}
	for _, expectedKey := range expectedKeys {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key %s not found in keys", expectedKey)
		}
	}
}

func TestContext_Clone(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)
	ctx.WithFlowName("test_flow")

	cloned := ctx.Clone()
	clonedCtx, ok := cloned.(*Context)
	if !ok {
		t.Fatal("Clone should return *Context")
	}

	// Verify cloned context has same values
	val, ok := clonedCtx.Get("key1")
	if !ok || val != "value1" {
		t.Error("Cloned context should have same values")
	}

	val, ok = clonedCtx.Get("key2")
	if !ok || val != 42 {
		t.Error("Cloned context should have same values")
	}

	// Verify flow name is copied
	if clonedCtx.FlowName() != "test_flow" {
		t.Error("Cloned context should have same flow name")
	}

	// Verify execution IDs are different
	if clonedCtx.ExecutionID() == ctx.ExecutionID() {
		t.Error("Cloned context should have different execution ID")
	}

	// Verify modifying clone doesn't affect original
	clonedCtx.Set("new_key", "new_value")
	if ctx.Has("new_key") {
		t.Error("Modifying clone should not affect original")
	}
}

func TestContext_WithTimeout(t *testing.T) {
	ctx := NewContext()
	timeout := 45 * time.Second

	result := ctx.WithTimeout(timeout)

	if result != ctx {
		t.Error("WithTimeout should return the same context instance")
	}

	if ctx.timeout != timeout {
		t.Errorf("Timeout should be %v, got %v", timeout, ctx.timeout)
	}
}

func TestContext_WithLogger(t *testing.T) {
	ctx := NewContext()
	logger := zap.NewNop()

	result := ctx.WithLogger(logger)

	if result != ctx {
		t.Error("WithLogger should return the same context instance")
	}

	if ctx.logger != logger {
		t.Error("Logger should be set")
	}
}

func TestContext_WithSpan(t *testing.T) {
	ctx := NewContext()
	// Create a mock span (using nil for simplicity in test)
	var span trace.Span

	result := ctx.WithSpan(span)

	if result != ctx {
		t.Error("WithSpan should return the same context instance")
	}

	if ctx.span != span {
		t.Error("Span should be set")
	}
}

func TestContext_WithFlowName(t *testing.T) {
	ctx := NewContext()
	flowName := "test_flow"

	result := ctx.WithFlowName(flowName)

	if result != ctx {
		t.Error("WithFlowName should return the same context instance")
	}

	if ctx.flowName != flowName {
		t.Errorf("Flow name should be %s, got %s", flowName, ctx.flowName)
	}
}

func TestContext_Context(t *testing.T) {
	ctx := NewContext()

	// Test without timeout
	ctx.timeout = 0
	goCtx := ctx.Context()
	if goCtx != ctx.ctx {
		t.Error("Context should return the underlying context when no timeout")
	}

	// Test with timeout
	ctx.timeout = 100 * time.Millisecond
	goCtx = ctx.Context()
	if goCtx == ctx.ctx {
		t.Error("Context should return a new context with timeout")
	}

	// Verify timeout works
	select {
	case <-goCtx.Done():
		// Expected after timeout
	case <-time.After(200 * time.Millisecond):
		t.Error("Context should have timed out")
	}
}

func TestContext_Logger(t *testing.T) {
	ctx := NewContext()

	// Test default logger (should be nop)
	logger := ctx.Logger()
	if logger == nil {
		t.Error("Logger should not be nil")
	}

	// Test custom logger
	customLogger := zap.NewNop()
	ctx.WithLogger(customLogger)
	logger = ctx.Logger()
	if logger != customLogger {
		t.Error("Should return custom logger")
	}
}

func TestContext_Span(t *testing.T) {
	ctx := NewContext()
	var span trace.Span

	ctx.WithSpan(span)
	result := ctx.Span()
	if result != span {
		t.Error("Should return the set span")
	}
}

func TestContext_FlowName(t *testing.T) {
	ctx := NewContext()
	flowName := "test_flow"

	ctx.WithFlowName(flowName)
	result := ctx.FlowName()
	if result != flowName {
		t.Errorf("FlowName should be %s, got %s", flowName, result)
	}
}

func TestContext_ExecutionID(t *testing.T) {
	ctx := NewContext()
	executionID := ctx.ExecutionID()

	if executionID == "" {
		t.Error("ExecutionID should not be empty")
	}

	// Verify different contexts have different IDs
	ctx2 := NewContext()
	if ctx2.ExecutionID() == executionID {
		t.Error("Different contexts should have different execution IDs")
	}
}

func TestContext_StartTime(t *testing.T) {
	before := time.Now()
	ctx := NewContext()
	after := time.Now()

	startTime := ctx.StartTime()
	if startTime.Before(before) || startTime.After(after) {
		t.Error("StartTime should be between before and after creation")
	}
}

func TestContext_Duration(t *testing.T) {
	ctx := NewContext()
	time.Sleep(10 * time.Millisecond)

	duration := ctx.Duration()
	if duration < 10*time.Millisecond {
		t.Error("Duration should be at least 10ms")
	}
}

func TestContext_ToMap(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)
	ctx.Set("key3", true)

	result := ctx.ToMap()

	if len(result) != 3 {
		t.Errorf("ToMap should return 3 items, got %d", len(result))
	}

	if result["key1"] != "value1" {
		t.Error("ToMap should include all values")
	}

	if result["key2"] != 42 {
		t.Error("ToMap should include all values")
	}

	if result["key3"] != true {
		t.Error("ToMap should include all values")
	}
}

func TestContext_Config(t *testing.T) {
	cfg := &config.FrameworkConfig{
		Timeouts: config.TimeoutConfig{
			FlowExecution: 60 * time.Second,
		},
	}

	ctx := NewContextWithConfig(cfg)
	result := ctx.Config()

	if result != cfg {
		t.Error("Config should return the set configuration")
	}
}

func TestContext_RecordMetrics(t *testing.T) {
	ctx := NewContext()

	// This test just ensures the method doesn't panic
	// In a real implementation, you might want to mock the metrics package
	ctx.RecordMetrics("test_step", 100*time.Millisecond, true)
	ctx.RecordMetrics("test_step", 200*time.Millisecond, false)
}

func TestGenerateExecutionID(t *testing.T) {
	id1 := generateExecutionID()
	time.Sleep(1 * time.Millisecond) // Small delay to ensure different timestamps
	id2 := generateExecutionID()

	if id1 == "" {
		t.Error("generateExecutionID should not return empty string")
	}

	if id2 == "" {
		t.Error("generateExecutionID should not return empty string")
	}

	if id1 == id2 {
		t.Error("generateExecutionID should return unique IDs")
	}

	// Verify format
	if len(id1) < 5 || id1[:5] != "exec_" {
		t.Error("generateExecutionID should start with 'exec_'")
	}
}

func TestContext_ThreadSafety(t *testing.T) {
	ctx := NewContext()
	done := make(chan bool)

	// Test concurrent access
	go func() {
		for i := 0; i < 100; i++ {
			ctx.Set("key1", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			ctx.Get("key1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			ctx.Has("key1")
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we reach here without panic, thread safety test passed
}
