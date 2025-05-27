package core

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
)

func TestNewLogStep(t *testing.T) {
	step := NewLogStep("test_log", "info", "Test message")

	assert.NotNil(t, step)
	assert.Equal(t, "test_log", step.Name())
	assert.Equal(t, "Logging", step.Description())
	assert.Equal(t, "info", step.level)
	assert.Equal(t, "Test message", step.message)
	assert.NotNil(t, step.fields)
	assert.True(t, step.sanitize)
	assert.False(t, step.includeCtx)
	assert.Empty(t, step.contextKeys)
}

func TestLogStep_WithFields(t *testing.T) {
	step := NewLogStep("test", "info", "message")
	fields := map[string]string{
		"user_id": "${user_id}",
		"action":  "login",
	}

	result := step.WithFields(fields)

	assert.Equal(t, step, result) // Fluent API
	assert.Equal(t, "${user_id}", step.fields["user_id"])
	assert.Equal(t, "login", step.fields["action"])
}

func TestLogStep_WithField(t *testing.T) {
	step := NewLogStep("test", "info", "message")

	result := step.WithField("key", "value")

	assert.Equal(t, step, result) // Fluent API
	assert.Equal(t, "value", step.fields["key"])
}

func TestLogStep_WithSanitization(t *testing.T) {
	step := NewLogStep("test", "info", "message")

	result := step.WithSanitization(false)

	assert.Equal(t, step, result) // Fluent API
	assert.False(t, step.sanitize)
}

func TestLogStep_WithContext(t *testing.T) {
	step := NewLogStep("test", "info", "message")

	result := step.WithContext(true, "user_id", "session_id")

	assert.Equal(t, step, result) // Fluent API
	assert.True(t, step.includeCtx)
	assert.Equal(t, []string{"user_id", "session_id"}, step.contextKeys)
}

func TestLogStep_Run_BasicLogging(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"warning level", "warning"},
		{"error level", "error"},
		{"unknown level", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := NewLogStep("test_log", tt.level, "Test message")
			ctx := flow.NewContext().WithFlowName("test_flow")

			err := step.Run(ctx)

			assert.NoError(t, err)
		})
	}
}

func TestLogStep_Run_WithInterpolation(t *testing.T) {
	step := NewLogStep("test_log", "info", "User ${user_id} performed ${action}")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("action", "login")

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestLogStep_Run_WithFields(t *testing.T) {
	step := NewLogStep("test_log", "info", "Test message").
		WithField("user_id", "${user_id}").
		WithField("static_field", "static_value")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestLogStep_Run_WithContextData(t *testing.T) {
	step := NewLogStep("test_log", "info", "Test message").
		WithContext(true, "user_id", "session_id")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("session_id", "abc456")
	ctx.Set("other_field", "should_not_appear")

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestLogStep_Run_WithAllContextData(t *testing.T) {
	step := NewLogStep("test_log", "info", "Test message").
		WithContext(true) // No specific keys = include all

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("session_id", "abc456")
	ctx.Set("password", "secret123") // Should be sanitized

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestLogStep_Run_WithSanitization(t *testing.T) {
	step := NewLogStep("test_log", "info", "Password is ${password}").
		WithSanitization(true).
		WithField("token", "${auth_token}")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("password", "secret123")
	ctx.Set("auth_token", "bearer_token_123")

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestLogStep_Run_WithoutSanitization(t *testing.T) {
	step := NewLogStep("test_log", "info", "Password is ${password}").
		WithSanitization(false)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("password", "secret123")

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestLogStep_IsSensitiveKey(t *testing.T) {
	step := &LogStep{}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"password", "password", true},
		{"user_password", "user_password", true},
		{"PASSWORD", "PASSWORD", true},
		{"token", "token", true},
		{"auth_token", "auth_token", true},
		{"secret", "secret", true},
		{"api_key", "api_key", true},
		{"authorization", "authorization", true},
		{"cookie", "cookie", true},
		{"session", "session", true},
		{"credential", "credential", true},
		{"username", "username", false},
		{"user_id", "user_id", false},
		{"email", "email", false},
		{"name", "name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := step.isSensitiveKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogStep_GetContextData_SpecificKeys(t *testing.T) {
	step := &LogStep{
		contextKeys: []string{"user_id", "session_id"},
		sanitize:    true,
	}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("session_id", "abc456")
	ctx.Set("password", "secret")
	ctx.Set("other_field", "value")

	result := step.getContextData(ctx)

	assert.Len(t, result, 2)
	assert.Equal(t, "123", result["user_id"])
	assert.Equal(t, "***", result["session_id"]) // session_id is sanitized because it contains "session"
	assert.NotContains(t, result, "password")
	assert.NotContains(t, result, "other_field")
}

func TestLogStep_GetContextData_AllKeys(t *testing.T) {
	step := &LogStep{
		contextKeys: []string{}, // Empty means all keys
		sanitize:    true,
	}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("password", "secret")

	result := step.getContextData(ctx)

	assert.Contains(t, result, "user_id")
	assert.Contains(t, result, "password")
	assert.Equal(t, "123", result["user_id"])
	assert.Equal(t, "***", result["password"]) // Should be sanitized
}

func TestLogStep_GetContextData_NoSanitization(t *testing.T) {
	step := &LogStep{
		contextKeys: []string{},
		sanitize:    false,
	}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("password", "secret")

	result := step.getContextData(ctx)

	assert.Equal(t, "secret", result["password"]) // Should not be sanitized
}

// MetricsLogStep tests

func TestNewMetricsLogStep(t *testing.T) {
	step := NewMetricsLogStep("test_metrics", "api_requests", "counter")

	assert.NotNil(t, step)
	assert.Equal(t, "test_metrics", step.Name())
	assert.Equal(t, "Metrics logging", step.Description())
	assert.Equal(t, "api_requests", step.metricName)
	assert.Equal(t, "counter", step.metricType)
	assert.Empty(t, step.metricValue)
	assert.NotNil(t, step.tags)
	assert.True(t, step.includeStats)
}

func TestMetricsLogStep_WithValue(t *testing.T) {
	step := NewMetricsLogStep("test", "metric", "gauge")

	result := step.WithValue("${response_time}")

	assert.Equal(t, step, result) // Fluent API
	assert.Equal(t, "${response_time}", step.metricValue)
}

func TestMetricsLogStep_WithTags(t *testing.T) {
	step := NewMetricsLogStep("test", "metric", "counter")
	tags := map[string]string{
		"service": "api",
		"version": "${app_version}",
	}

	result := step.WithTags(tags)

	assert.Equal(t, step, result) // Fluent API
	assert.Equal(t, "api", step.tags["service"])
	assert.Equal(t, "${app_version}", step.tags["version"])
}

func TestMetricsLogStep_WithStats(t *testing.T) {
	step := NewMetricsLogStep("test", "metric", "counter")

	result := step.WithStats(false)

	assert.Equal(t, step, result) // Fluent API
	assert.False(t, step.includeStats)
}

func TestMetricsLogStep_Run_Counter(t *testing.T) {
	step := NewMetricsLogStep("test_counter", "api_requests", "counter")

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	assert.NoError(t, err)

	// Check that metric was stored in context
	metric, exists := ctx.Get("metric_api_requests")
	assert.True(t, exists)

	metricMap, ok := metric.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "api_requests", metricMap["name"])
	assert.Equal(t, "counter", metricMap["type"])
	assert.Equal(t, 1, metricMap["value"]) // Default counter value
}

func TestMetricsLogStep_Run_GaugeWithValue(t *testing.T) {
	step := NewMetricsLogStep("test_gauge", "response_time", "gauge").
		WithValue("${duration}")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("duration", "150ms")

	err := step.Run(ctx)

	assert.NoError(t, err)

	metric, exists := ctx.Get("metric_response_time")
	assert.True(t, exists)

	metricMap, ok := metric.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "response_time", metricMap["name"])
	assert.Equal(t, "gauge", metricMap["type"])
	assert.Equal(t, "150ms", metricMap["value"])
}

func TestMetricsLogStep_Run_WithTags(t *testing.T) {
	step := NewMetricsLogStep("test_tagged", "requests", "counter").
		WithTags(map[string]string{
			"service": "api",
			"version": "${app_version}",
			"static":  "value",
		})

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("app_version", "1.2.3")

	err := step.Run(ctx)

	assert.NoError(t, err)

	metric, exists := ctx.Get("metric_requests")
	assert.True(t, exists)

	metricMap, ok := metric.(map[string]interface{})
	assert.True(t, ok)

	tags, ok := metricMap["tags"].(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "api", tags["service"])
	assert.Equal(t, "1.2.3", tags["version"])
	assert.Equal(t, "value", tags["static"])
}

func TestMetricsLogStep_Run_WithStats(t *testing.T) {
	step := NewMetricsLogStep("test_stats", "metric", "counter").
		WithStats(true)

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("start_time", "2023-01-01T00:00:00Z")
	ctx.Set("other_field", "value")

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestMetricsLogStep_Run_WithoutStats(t *testing.T) {
	step := NewMetricsLogStep("test_no_stats", "metric", "counter").
		WithStats(false)

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestMetricsLogStep_GatherStats(t *testing.T) {
	step := &MetricsLogStep{}

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("start_time", "2023-01-01T00:00:00Z")
	ctx.Set("field1", "value1")
	ctx.Set("field2", "value2")

	stats := step.gatherStats(ctx)

	assert.Contains(t, stats, "context_keys")
	assert.Contains(t, stats, "start_time")
	assert.Equal(t, 3, stats["context_keys"]) // 3 fields set
	assert.Equal(t, "2023-01-01T00:00:00Z", stats["start_time"])
}

// Test helper functions

func TestNewInfoLogStep(t *testing.T) {
	step := NewInfoLogStep("info_test", "Info message")

	assert.Equal(t, "info_test", step.Name())
	assert.Equal(t, "info", step.level)
	assert.Equal(t, "Info message", step.message)
}

func TestNewErrorLogStep(t *testing.T) {
	step := NewErrorLogStep("error_test", "Error message")

	assert.Equal(t, "error_test", step.Name())
	assert.Equal(t, "error", step.level)
	assert.Equal(t, "Error message", step.message)
}

func TestNewDebugLogStep(t *testing.T) {
	step := NewDebugLogStep("debug_test", "Debug message")

	assert.Equal(t, "debug_test", step.Name())
	assert.Equal(t, "debug", step.level)
	assert.Equal(t, "Debug message", step.message)
}

func TestNewWarnLogStep(t *testing.T) {
	step := NewWarnLogStep("warn_test", "Warning message")

	assert.Equal(t, "warn_test", step.Name())
	assert.Equal(t, "warn", step.level)
	assert.Equal(t, "Warning message", step.message)
}

func TestNewContextLogStep(t *testing.T) {
	step := NewContextLogStep("context_test", "Context message", "user_id", "session_id")

	assert.Equal(t, "context_test", step.Name())
	assert.Equal(t, "info", step.level)
	assert.Equal(t, "Context message", step.message)
	assert.True(t, step.includeCtx)
	assert.Equal(t, []string{"user_id", "session_id"}, step.contextKeys)
}

func TestNewCounterMetricStep(t *testing.T) {
	step := NewCounterMetricStep("counter_test", "api_calls")

	assert.Equal(t, "counter_test", step.Name())
	assert.Equal(t, "api_calls", step.metricName)
	assert.Equal(t, "counter", step.metricType)
}

func TestNewGaugeMetricStep(t *testing.T) {
	step := NewGaugeMetricStep("gauge_test", "memory_usage", "${memory}")

	assert.Equal(t, "gauge_test", step.Name())
	assert.Equal(t, "memory_usage", step.metricName)
	assert.Equal(t, "gauge", step.metricType)
	assert.Equal(t, "${memory}", step.metricValue)
}

func TestNewTimerMetricStep(t *testing.T) {
	step := NewTimerMetricStep("timer_test", "request_duration", "elapsed_time")

	assert.Equal(t, "timer_test", step.Name())
	assert.Equal(t, "request_duration", step.metricName)
	assert.Equal(t, "timer", step.metricType)
	assert.Equal(t, "${elapsed_time}", step.metricValue)
}

func TestLogStep_PrepareLogFields(t *testing.T) {
	step := NewLogStep("test_step", "info", "test message").
		WithField("user_id", "${user_id}").
		WithField("action", "login").
		WithContext(true, "session_id")

	ctx := flow.NewContext().WithFlowName("test_flow")
	ctx.Set("user_id", "123")
	ctx.Set("session_id", "abc456")

	fields := step.prepareLogFields(ctx)

	// Should have at least step name, custom fields, and context
	assert.GreaterOrEqual(t, len(fields), 3)
}

func TestLogStep_Run_InterpolationError(t *testing.T) {
	step := NewLogStep("test_log", "info", "Message with ${unclosed")

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	// Should not fail even with interpolation error
	assert.NoError(t, err)
}

func TestMetricsLogStep_Run_InterpolationError(t *testing.T) {
	step := NewMetricsLogStep("test_metrics", "metric", "gauge").
		WithValue("${unclosed").
		WithTags(map[string]string{
			"tag": "${unclosed_tag",
		})

	ctx := flow.NewContext().WithFlowName("test_flow")

	err := step.Run(ctx)

	// Should not fail even with interpolation errors
	assert.NoError(t, err)
}
