# Interfaces Package Documentation

## Overview

The `pkg/interfaces` package provides the foundational interfaces and contracts for the API Orchestration Framework. It defines the core abstractions that enable loose coupling, extensibility, and consistent behavior across all framework components.

## Purpose

The interfaces package serves as the contract foundation that:
- Defines core interfaces for all framework components
- Enables loose coupling between components
- Provides extensibility points for custom implementations
- Ensures consistent behavior across different implementations
- Supports dependency injection and testing
- Facilitates plugin architecture and component composition

## Core Interfaces

### Execution Context Interface

The fundamental interface for execution context management:
```go
type ExecutionContext interface {
    // Data access
    Get(key string) (interface{}, bool)
    Set(key string, value interface{})
    Delete(key string)
    Has(key string) bool
    GetData() map[string]interface{}
    
    // Type-safe getters
    GetString(key string) (string, error)
    GetInt(key string) (int, error)
    GetFloat(key string) (float64, error)
    GetBool(key string) (bool, error)
    GetMap(key string) (map[string]interface{}, error)
    GetSlice(key string) ([]interface{}, error)
    
    // Nested access
    GetNested(path string) (interface{}, error)
    SetNested(path string, value interface{}) error
    HasNested(path string) bool
    
    // Context management
    Clone() ExecutionContext
    Merge(other ExecutionContext) error
    
    // Observability
    Logger() Logger
    Metrics() MetricsCollector
    Tracer() Tracer
    
    // Configuration
    Config() Configuration
    
    // Lifecycle
    ExecutionID() string
    StartTime() time.Time
    Duration() time.Duration
    
    // Cancellation
    Context() context.Context
    WithTimeout(timeout time.Duration) ExecutionContext
    WithCancel() (ExecutionContext, context.CancelFunc)
}
```

### Step Interface

The core interface for all step implementations:
```go
type Step interface {
    // Execution
    Run(ctx ExecutionContext) error
    
    // Metadata
    Name() string
    Description() string
    
    // Configuration
    Configure(config map[string]interface{}) error
    Validate() error
    
    // Dependencies
    Dependencies() []string
    
    // Lifecycle hooks
    BeforeRun(ctx ExecutionContext) error
    AfterRun(ctx ExecutionContext, err error) error
    
    // Timeout and cancellation
    Timeout() time.Duration
    SupportsCancel() bool
    
    // Retry support
    IsRetryable(err error) bool
    MaxRetries() int
    RetryDelay() time.Duration
}
```

### Transformer Interface

Interface for data transformation operations:
```go
type Transformer interface {
    // Core transformation
    Transform(data map[string]interface{}) (map[string]interface{}, error)
    
    // Metadata
    Name() string
    Description() string
    Version() string
    
    // Configuration
    Configure(config map[string]interface{}) error
    Validate() error
    
    // Capabilities
    SupportedInputTypes() []string
    SupportedOutputTypes() []string
    
    // Performance
    IsThreadSafe() bool
    EstimatedComplexity() ComplexityLevel
    
    // Chaining support
    CanChainWith(other Transformer) bool
    ChainPriority() int
}
```

### Validator Interface

Interface for data validation operations:
```go
type Validator interface {
    // Core validation
    Validate(data map[string]interface{}) error
    
    // Metadata
    Name() string
    Description() string
    
    // Configuration
    Configure(config map[string]interface{}) error
    
    // Validation capabilities
    SupportedDataTypes() []string
    ValidationRules() []ValidationRule
    
    // Error handling
    ContinueOnError() bool
    ErrorSeverity() SeverityLevel
    
    // Performance
    IsThreadSafe() bool
    ValidationComplexity() ComplexityLevel
}
```

### Cache Interface

Interface for caching operations:
```go
type Cache interface {
    // Basic operations
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Exists(key string) bool
    
    // Batch operations
    GetMulti(keys []string) (map[string]interface{}, error)
    SetMulti(items map[string]interface{}, ttl time.Duration) error
    DeleteMulti(keys []string) error
    
    // Advanced operations
    Increment(key string, delta int64) (int64, error)
    Decrement(key string, delta int64) (int64, error)
    Touch(key string, ttl time.Duration) error
    
    // Management
    Clear() error
    Size() int64
    Keys(pattern string) ([]string, error)
    
    // Statistics
    Stats() CacheStats
    HitRate() float64
    
    // Lifecycle
    Close() error
}
```

### HTTP Client Interface

Interface for HTTP operations:
```go
type HTTPClient interface {
    // Basic HTTP methods
    Get(url string, headers map[string]string) (*HTTPResponse, error)
    Post(url string, body interface{}, headers map[string]string) (*HTTPResponse, error)
    Put(url string, body interface{}, headers map[string]string) (*HTTPResponse, error)
    Delete(url string, headers map[string]string) (*HTTPResponse, error)
    Patch(url string, body interface{}, headers map[string]string) (*HTTPResponse, error)
    
    // Generic request
    Do(request *HTTPRequest) (*HTTPResponse, error)
    
    // Configuration
    SetTimeout(timeout time.Duration)
    SetRetryPolicy(policy RetryPolicy)
    SetCircuitBreaker(cb CircuitBreaker)
    
    // Middleware support
    AddMiddleware(middleware HTTPMiddleware)
    
    // Connection management
    SetMaxIdleConns(n int)
    SetMaxIdleConnsPerHost(n int)
    SetIdleConnTimeout(timeout time.Duration)
    
    // Lifecycle
    Close() error
}
```

### Logger Interface

Interface for logging operations:
```go
type Logger interface {
    // Log levels
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    
    // Structured logging
    With(fields ...Field) Logger
    WithError(err error) Logger
    WithContext(ctx context.Context) Logger
    
    // Level management
    SetLevel(level LogLevel)
    GetLevel() LogLevel
    IsLevelEnabled(level LogLevel) bool
    
    // Output management
    SetOutput(output io.Writer)
    AddHook(hook LogHook)
    
    // Formatting
    SetFormatter(formatter LogFormatter)
    
    // Sampling
    SetSampling(config SamplingConfig)
}
```

### Metrics Collector Interface

Interface for metrics collection:
```go
type MetricsCollector interface {
    // Counter metrics
    IncrementCounter(name string, tags map[string]string)
    AddToCounter(name string, value int64, tags map[string]string)
    
    // Gauge metrics
    SetGauge(name string, value float64, tags map[string]string)
    AddToGauge(name string, value float64, tags map[string]string)
    
    // Histogram metrics
    RecordHistogram(name string, value float64, tags map[string]string)
    
    // Timer metrics
    RecordTimer(name string, duration time.Duration, tags map[string]string)
    StartTimer(name string, tags map[string]string) TimerHandle
    
    // Custom metrics
    RecordCustom(metric CustomMetric)
    
    // Batch operations
    RecordBatch(metrics []Metric)
    
    // Configuration
    SetDefaultTags(tags map[string]string)
    SetSampling(rate float64)
    
    // Lifecycle
    Flush() error
    Close() error
}
```

### Configuration Interface

Interface for configuration management:
```go
type Configuration interface {
    // Basic access
    Get(key string) interface{}
    GetString(key string) string
    GetInt(key string) int
    GetFloat(key string) float64
    GetBool(key string) bool
    GetDuration(key string) time.Duration
    
    // Default values
    GetWithDefault(key string, defaultValue interface{}) interface{}
    GetStringWithDefault(key string, defaultValue string) string
    GetIntWithDefault(key string, defaultValue int) int
    
    // Nested access
    GetNested(path string) interface{}
    
    // Existence checks
    Has(key string) bool
    HasNested(path string) bool
    
    // Sections
    GetSection(section string) Configuration
    GetAllSections() map[string]Configuration
    
    // Environment
    GetEnvironment() string
    IsProduction() bool
    IsDevelopment() bool
    
    // Validation
    Validate() error
    ValidateSection(section string) error
    
    // Change notification
    Watch(key string, callback ConfigChangeCallback) error
    Unwatch(key string) error
}
```

## Supporting Types and Enums

### Complexity Levels
```go
type ComplexityLevel int

const (
    ComplexityLow ComplexityLevel = iota
    ComplexityMedium
    ComplexityHigh
    ComplexityVeryHigh
)
```

### Severity Levels
```go
type SeverityLevel int

const (
    SeverityInfo SeverityLevel = iota
    SeverityWarning
    SeverityError
    SeverityCritical
)
```

### Log Levels
```go
type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
    LogLevelFatal
)
```

### HTTP Request/Response Types
```go
type HTTPRequest struct {
    Method  string
    URL     string
    Headers map[string]string
    Body    interface{}
    Timeout time.Duration
    Context context.Context
}

type HTTPResponse struct {
    StatusCode int
    Headers    map[string]string
    Body       []byte
    Duration   time.Duration
    Request    *HTTPRequest
}
```

### Cache Statistics
```go
type CacheStats struct {
    Hits        int64
    Misses      int64
    Sets        int64
    Deletes     int64
    Size        int64
    Memory      int64
    Evictions   int64
    Errors      int64
    LastAccess  time.Time
}
```

## Interface Implementations

### Base Implementations

The package provides base implementations that can be embedded:

#### BaseStep
```go
type BaseStep struct {
    name        string
    description string
    timeout     time.Duration
    maxRetries  int
    retryDelay  time.Duration
    config      map[string]interface{}
}

func (bs *BaseStep) Name() string { return bs.name }
func (bs *BaseStep) Description() string { return bs.description }
func (bs *BaseStep) Timeout() time.Duration { return bs.timeout }
func (bs *BaseStep) MaxRetries() int { return bs.maxRetries }
func (bs *BaseStep) RetryDelay() time.Duration { return bs.retryDelay }
func (bs *BaseStep) SupportsCancel() bool { return true }
func (bs *BaseStep) Dependencies() []string { return []string{} }

func (bs *BaseStep) Configure(config map[string]interface{}) error {
    bs.config = config
    return nil
}

func (bs *BaseStep) Validate() error {
    if bs.name == "" {
        return fmt.Errorf("step name is required")
    }
    return nil
}

func (bs *BaseStep) BeforeRun(ctx ExecutionContext) error { return nil }
func (bs *BaseStep) AfterRun(ctx ExecutionContext, err error) error { return nil }
func (bs *BaseStep) IsRetryable(err error) bool { return false }
```

#### BaseTransformer
```go
type BaseTransformer struct {
    name        string
    description string
    version     string
    config      map[string]interface{}
}

func (bt *BaseTransformer) Name() string { return bt.name }
func (bt *BaseTransformer) Description() string { return bt.description }
func (bt *BaseTransformer) Version() string { return bt.version }
func (bt *BaseTransformer) IsThreadSafe() bool { return true }
func (bt *BaseTransformer) EstimatedComplexity() ComplexityLevel { return ComplexityMedium }
func (bt *BaseTransformer) ChainPriority() int { return 0 }

func (bt *BaseTransformer) Configure(config map[string]interface{}) error {
    bt.config = config
    return nil
}

func (bt *BaseTransformer) Validate() error {
    if bt.name == "" {
        return fmt.Errorf("transformer name is required")
    }
    return nil
}

func (bt *BaseTransformer) SupportedInputTypes() []string {
    return []string{"map[string]interface{}"}
}

func (bt *BaseTransformer) SupportedOutputTypes() []string {
    return []string{"map[string]interface{}"}
}

func (bt *BaseTransformer) CanChainWith(other Transformer) bool {
    return true
}
```

### Adapter Patterns

The package provides adapters for integrating external libraries:

#### Function Adapters
```go
// Adapt functions to Step interface
type FunctionStep struct {
    *BaseStep
    fn func(ExecutionContext) error
}

func NewFunctionStep(name string, fn func(ExecutionContext) error) *FunctionStep {
    return &FunctionStep{
        BaseStep: &BaseStep{name: name},
        fn:       fn,
    }
}

func (fs *FunctionStep) Run(ctx ExecutionContext) error {
    return fs.fn(ctx)
}

// Adapt functions to Transformer interface
type FunctionTransformer struct {
    *BaseTransformer
    fn func(map[string]interface{}) (map[string]interface{}, error)
}

func NewFunctionTransformer(name string, fn func(map[string]interface{}) (map[string]interface{}, error)) *FunctionTransformer {
    return &FunctionTransformer{
        BaseTransformer: &BaseTransformer{name: name},
        fn:              fn,
    }
}

func (ft *FunctionTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
    return ft.fn(data)
}
```

## Interface Composition

### Composite Interfaces

Complex interfaces built from simpler ones:

#### Executable
```go
type Executable interface {
    Step
    Transformer
    Validator
}
```

#### Observable
```go
type Observable interface {
    Logger
    MetricsCollector
    Tracer
}
```

#### Configurable
```go
type Configurable interface {
    Configure(config map[string]interface{}) error
    Validate() error
    GetConfig() map[string]interface{}
}
```

#### Lifecycle
```go
type Lifecycle interface {
    Initialize() error
    Start() error
    Stop() error
    Shutdown() error
    IsRunning() bool
    Health() HealthStatus
}
```

### Plugin Interface

Interface for plugin architecture:
```go
type Plugin interface {
    // Plugin metadata
    Name() string
    Version() string
    Description() string
    Author() string
    
    // Plugin lifecycle
    Initialize(config map[string]interface{}) error
    Start() error
    Stop() error
    
    // Plugin capabilities
    Provides() []string
    Requires() []string
    
    // Component registration
    RegisterComponents(registry ComponentRegistry) error
    
    // Health and status
    Health() HealthStatus
    Status() PluginStatus
}
```

## Interface Extensions

### Middleware Interfaces

Support for middleware patterns:

#### StepMiddleware
```go
type StepMiddleware interface {
    BeforeStep(ctx ExecutionContext, step Step) error
    AfterStep(ctx ExecutionContext, step Step, result interface{}, err error) error
    OnError(ctx ExecutionContext, step Step, err error) error
}
```

#### TransformerMiddleware
```go
type TransformerMiddleware interface {
    BeforeTransform(data map[string]interface{}, transformer Transformer) (map[string]interface{}, error)
    AfterTransform(input, output map[string]interface{}, transformer Transformer) (map[string]interface{}, error)
    OnTransformError(data map[string]interface{}, transformer Transformer, err error) error
}
```

### Event Interfaces

Support for event-driven patterns:

#### EventEmitter
```go
type EventEmitter interface {
    Emit(event Event) error
    EmitAsync(event Event) error
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
}
```

#### EventHandler
```go
type EventHandler interface {
    Handle(event Event) error
    CanHandle(eventType string) bool
    Priority() int
}
```

## Testing Interfaces

### Mock Interfaces

Interfaces designed for testing:

#### MockStep
```go
type MockStep struct {
    *BaseStep
    RunFunc        func(ExecutionContext) error
    ConfigureFunc  func(map[string]interface{}) error
    ValidateFunc   func() error
    CallCount      int
    LastContext    ExecutionContext
}

func (ms *MockStep) Run(ctx ExecutionContext) error {
    ms.CallCount++
    ms.LastContext = ctx
    if ms.RunFunc != nil {
        return ms.RunFunc(ctx)
    }
    return nil
}
```

#### TestExecutionContext
```go
type TestExecutionContext struct {
    data    map[string]interface{}
    logger  Logger
    metrics MetricsCollector
    config  Configuration
}

func NewTestExecutionContext() *TestExecutionContext {
    return &TestExecutionContext{
        data:    make(map[string]interface{}),
        logger:  NewTestLogger(),
        metrics: NewTestMetricsCollector(),
        config:  NewTestConfiguration(),
    }
}
```

## Interface Validation

### Contract Testing

Utilities for testing interface implementations:

```go
// Test that an implementation satisfies the interface contract
func TestStepContract(t *testing.T, step Step) {
    // Test basic interface compliance
    assert.NotEmpty(t, step.Name())
    assert.NotNil(t, step.Description())
    assert.True(t, step.Timeout() > 0)
    
    // Test configuration
    config := map[string]interface{}{"test": "value"}
    err := step.Configure(config)
    assert.NoError(t, err)
    
    // Test validation
    err = step.Validate()
    assert.NoError(t, err)
    
    // Test execution
    ctx := NewTestExecutionContext()
    err = step.Run(ctx)
    assert.NoError(t, err)
}

// Test transformer contract
func TestTransformerContract(t *testing.T, transformer Transformer) {
    // Test metadata
    assert.NotEmpty(t, transformer.Name())
    assert.NotEmpty(t, transformer.Version())
    
    // Test capabilities
    inputTypes := transformer.SupportedInputTypes()
    assert.NotEmpty(t, inputTypes)
    
    outputTypes := transformer.SupportedOutputTypes()
    assert.NotEmpty(t, outputTypes)
    
    // Test transformation
    input := map[string]interface{}{"test": "data"}
    output, err := transformer.Transform(input)
    assert.NoError(t, err)
    assert.NotNil(t, output)
}
```

## Best Practices

### Interface Design
1. **Keep Interfaces Small**: Follow the Interface Segregation Principle
2. **Use Composition**: Build complex interfaces from simpler ones
3. **Avoid Breaking Changes**: Design interfaces for extensibility
4. **Document Contracts**: Clearly document expected behavior
5. **Provide Base Implementations**: Offer default implementations where appropriate

### Implementation Guidelines
1. **Implement Completely**: Implement all interface methods meaningfully
2. **Handle Errors Gracefully**: Return appropriate errors with context
3. **Support Cancellation**: Respect context cancellation where applicable
4. **Be Thread-Safe**: Ensure implementations are thread-safe when needed
5. **Validate Inputs**: Validate inputs and return meaningful errors

### Testing Strategies
1. **Contract Testing**: Test that implementations satisfy interface contracts
2. **Mock Implementations**: Provide mock implementations for testing
3. **Integration Testing**: Test interface interactions
4. **Performance Testing**: Test interface performance characteristics
5. **Error Handling**: Test error conditions and edge cases

## Examples

### Custom Step Implementation
```go
type DatabaseStep struct {
    *BaseStep
    db    *sql.DB
    query string
}

func NewDatabaseStep(name, query string, db *sql.DB) *DatabaseStep {
    return &DatabaseStep{
        BaseStep: &BaseStep{
            name:        name,
            description: "Execute database query",
            timeout:     30 * time.Second,
            maxRetries:  3,
            retryDelay:  1 * time.Second,
        },
        db:    db,
        query: query,
    }
}

func (ds *DatabaseStep) Run(ctx ExecutionContext) error {
    // Get query parameters from context
    params := make([]interface{}, 0)
    if paramData, ok := ctx.Get("query_params"); ok {
        if paramSlice, ok := paramData.([]interface{}); ok {
            params = paramSlice
        }
    }
    
    // Execute query with timeout
    queryCtx, cancel := context.WithTimeout(ctx.Context(), ds.Timeout())
    defer cancel()
    
    rows, err := ds.db.QueryContext(queryCtx, ds.query, params...)
    if err != nil {
        return fmt.Errorf("database query failed: %w", err)
    }
    defer rows.Close()
    
    // Process results
    results := make([]map[string]interface{}, 0)
    columns, err := rows.Columns()
    if err != nil {
        return fmt.Errorf("failed to get columns: %w", err)
    }
    
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return fmt.Errorf("failed to scan row: %w", err)
        }
        
        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        results = append(results, row)
    }
    
    // Store results in context
    ctx.Set("query_results", results)
    
    return nil
}

func (ds *DatabaseStep) IsRetryable(err error) bool {
    // Retry on connection errors, not on syntax errors
    return strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "timeout")
}

func (ds *DatabaseStep) Dependencies() []string {
    return []string{"database"}
}
```

### Custom Transformer Implementation
```go
type JSONTransformer struct {
    *BaseTransformer
    indent bool
}

func NewJSONTransformer(name string, indent bool) *JSONTransformer {
    return &JSONTransformer{
        BaseTransformer: &BaseTransformer{
            name:        name,
            description: "Transform data to/from JSON",
            version:     "1.0.0",
        },
        indent: indent,
    }
}

func (jt *JSONTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
    result := make(map[string]interface{})
    
    for key, value := range data {
        switch v := value.(type) {
        case string:
            // Try to parse as JSON
            var parsed interface{}
            if err := json.Unmarshal([]byte(v), &parsed); err == nil {
                result[key] = parsed
            } else {
                result[key] = v
            }
        case map[string]interface{}, []interface{}:
            // Convert to JSON string
            var jsonBytes []byte
            var err error
            if jt.indent {
                jsonBytes, err = json.MarshalIndent(v, "", "  ")
            } else {
                jsonBytes, err = json.Marshal(v)
            }
            if err != nil {
                return nil, fmt.Errorf("failed to marshal %s to JSON: %w", key, err)
            }
            result[key+"_json"] = string(jsonBytes)
            result[key] = v
        default:
            result[key] = v
        }
    }
    
    return result, nil
}

func (jt *JSONTransformer) SupportedInputTypes() []string {
    return []string{"string", "map[string]interface{}", "[]interface{}"}
}

func (jt *JSONTransformer) SupportedOutputTypes() []string {
    return []string{"string", "map[string]interface{}", "[]interface{}"}
}

func (jt *JSONTransformer) EstimatedComplexity() ComplexityLevel {
    return ComplexityLow
}
```

## Troubleshooting

### Common Interface Issues

1. **Interface Not Implemented Completely**
   ```go
   // Compile-time check to ensure interface is implemented
   var _ Step = (*MyStep)(nil)
   var _ Transformer = (*MyTransformer)(nil)
   ```

2. **Nil Interface Values**
   ```go
   // Check for nil interfaces
   func SafeRun(step Step, ctx ExecutionContext) error {
       if step == nil {
           return fmt.Errorf("step is nil")
       }
       return step.Run(ctx)
   }
   ```

3. **Interface Type Assertions**
   ```go
   // Safe type assertion
   if configurable, ok := step.(Configurable); ok {
       err := configurable.Configure(config)
       if err != nil {
           return fmt.Errorf("configuration failed: %w", err)
       }
   }
   ```

### Performance Considerations

1. **Interface Method Calls**: Interface method calls have slight overhead
2. **Type Assertions**: Minimize type assertions in hot paths
3. **Interface Composition**: Be mindful of interface composition complexity
4. **Memory Allocation**: Interfaces can cause heap allocations

### Testing Interface Implementations

```go
func TestInterfaceImplementation(t *testing.T) {
    // Test that implementation satisfies interface
    var step Step = NewMyStep("test")
    
    // Test interface methods
    assert.Equal(t, "test", step.Name())
    assert.NotEmpty(t, step.Description())
    
    // Test execution
    ctx := NewTestExecutionContext()
    err := step.Run(ctx)
    assert.NoError(t, err)
    
    // Test configuration if supported
    if configurable, ok := step.(Configurable); ok {
        config := map[string]interface{}{"key": "value"}
        err := configurable.Configure(config)
        assert.NoError(t, err)
    }
}
``` 