package flow

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

// Context represents the execution context for a flow
// It provides type-safe data passing between steps and includes observability
// Now implements interfaces.ExecutionContext
type Context struct {
	// Core context
	ctx context.Context

	// Data storage with thread safety
	values map[string]interface{}
	mu     sync.RWMutex

	// Metadata
	flowName    string
	executionID string
	startTime   time.Time

	// Observability
	logger *zap.Logger
	span   trace.Span

	// Configuration
	timeout time.Duration
	config  *config.FrameworkConfig
}

// NewContext creates a new execution context
func NewContext() *Context {
	return &Context{
		ctx:         context.Background(),
		values:      make(map[string]interface{}),
		executionID: generateExecutionID(),
		startTime:   time.Now(),
		timeout:     30 * time.Second,
		config:      config.DefaultConfig(),
	}
}

// NewContextWithConfig creates a new context with custom configuration
func NewContextWithConfig(cfg *config.FrameworkConfig) *Context {
	return &Context{
		ctx:         context.Background(),
		values:      make(map[string]interface{}),
		executionID: generateExecutionID(),
		startTime:   time.Now(),
		timeout:     cfg.Timeouts.FlowExecution,
		config:      cfg,
	}
}

// Set stores a value in the context with thread safety
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// Get retrieves a typed value from the context
func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.values[key]
	return val, ok
}

// GetTyped retrieves a value with type assertion
func (c *Context) GetTyped(key string, target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.values[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	// Use reflection to set the target value
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	sourceValue := reflect.ValueOf(val)
	targetType := targetValue.Elem().Type()

	if !sourceValue.Type().AssignableTo(targetType) {
		return fmt.Errorf("type assertion failed for key %s: expected %v, got %v", key, targetType, sourceValue.Type())
	}

	targetValue.Elem().Set(sourceValue)
	return nil
}

// GetString is a convenience method for string values
func (c *Context) GetString(key string) (string, error) {
	val, ok := c.Get(key)
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}

	if str, ok := val.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("value is not a string: %T", val)
}

// GetInt is a convenience method for int values
func (c *Context) GetInt(key string) (int, error) {
	val, ok := c.Get(key)
	if !ok {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	if i, ok := val.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("value is not an int: %T", val)
}

// GetBool is a convenience method for bool values
func (c *Context) GetBool(key string) (bool, error) {
	val, ok := c.Get(key)
	if !ok {
		return false, fmt.Errorf("key not found: %s", key)
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("value is not a bool: %T", val)
}

// GetMap is a convenience method for map values
func (c *Context) GetMap(key string) (map[string]interface{}, error) {
	val, ok := c.Get(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if m, ok := val.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, fmt.Errorf("value is not a map[string]interface{}: %T", val)
}

// Has checks if a key exists in the context
func (c *Context) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.values[key]
	return ok
}

// Delete removes a key from the context
func (c *Context) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}

// Keys returns all keys in the context
func (c *Context) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	return keys
}

// Clone creates a deep copy of the context
func (c *Context) Clone() interfaces.ExecutionContext {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newCtx := &Context{
		ctx:         c.ctx,
		values:      make(map[string]interface{}),
		flowName:    c.flowName,
		executionID: generateExecutionID(), // New execution ID for clone
		startTime:   time.Now(),
		logger:      c.logger,
		span:        c.span,
		timeout:     c.timeout,
		config:      c.config,
	}

	// Deep copy values
	for k, v := range c.values {
		newCtx.values[k] = v
	}

	return newCtx
}

// WithTimeout sets the timeout for the context
func (c *Context) WithTimeout(timeout time.Duration) *Context {
	c.timeout = timeout
	return c
}

// WithLogger sets the logger for the context
func (c *Context) WithLogger(logger *zap.Logger) *Context {
	c.logger = logger
	return c
}

// WithSpan sets the tracing span for the context
func (c *Context) WithSpan(span trace.Span) *Context {
	c.span = span
	return c
}

// WithFlowName sets the flow name for the context
func (c *Context) WithFlowName(name string) *Context {
	c.flowName = name
	return c
}

// Context returns the underlying Go context
func (c *Context) Context() context.Context {
	if c.timeout > 0 {
		ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
		// Store cancel function to be called when context is done
		go func() {
			<-ctx.Done()
			cancel()
		}()
		return ctx
	}
	return c.ctx
}

// Logger returns the logger instance
func (c *Context) Logger() *zap.Logger {
	if c.logger == nil {
		c.logger = zap.NewNop()
	}
	return c.logger
}

// Span returns the tracing span
func (c *Context) Span() trace.Span {
	return c.span
}

// FlowName returns the flow name
func (c *Context) FlowName() string {
	return c.flowName
}

// ExecutionID returns the unique execution ID
func (c *Context) ExecutionID() string {
	return c.executionID
}

// StartTime returns the execution start time
func (c *Context) StartTime() time.Time {
	return c.startTime
}

// Duration returns the elapsed time since start
func (c *Context) Duration() time.Duration {
	return time.Since(c.startTime)
}

// ToMap returns all values as a map (for serialization)
func (c *Context) ToMap() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.values {
		result[k] = v
	}
	return result
}

// Config returns the framework configuration
func (c *Context) Config() *config.FrameworkConfig {
	return c.config
}

// RecordMetrics records execution metrics
func (c *Context) RecordMetrics(stepName string, duration time.Duration, success bool) {
	metrics.RecordStepExecution(stepName, duration, success)
}

// generateExecutionID creates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}
