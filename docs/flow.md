# Flow Package Documentation

## Overview

The `pkg/flow` package is the core orchestration engine of the API Orchestration Framework. It provides the fundamental building blocks for creating, executing, and managing complex workflows with comprehensive observability, error handling, and performance optimization.

## Purpose

The flow package serves as the central coordinator that:
- Orchestrates step execution in sequential, parallel, and conditional patterns
- Manages execution context and data flow between steps
- Provides comprehensive observability through logging, metrics, and tracing
- Handles errors gracefully with retry mechanisms and circuit breakers
- Supports middleware for cross-cutting concerns
- Enables complex workflow patterns like branching and aggregation

## Core Components

### Flow (`flow.go`)

The `Flow` struct represents a complete orchestration workflow with a fluent DSL for building complex execution patterns.

#### Key Features:
- **Fluent Builder Pattern**: Chain methods to build workflows declaratively
- **Middleware Support**: Add cross-cutting concerns like logging, metrics, authentication
- **Timeout Management**: Configure timeouts at flow and step levels
- **Conditional Execution**: Support for `Choice`, `When`, `Otherwise` patterns
- **Parallel Execution**: Execute multiple steps concurrently with proper synchronization
- **Error Handling**: Built-in retry logic and graceful degradation

#### Basic Usage:
```go
flow := flow.NewFlow("UserProfileFlow").
    WithDescription("Fetches and transforms user profile data").
    WithTimeout(30 * time.Second).
    Use(middleware.LoggingMiddleware()).
    Step("validate", authStep).
    Step("fetch", httpStep).
    Step("transform", transformStep)

result, err := flow.Execute(ctx)
```

#### Advanced Patterns:
```go
// Conditional execution
flow.Choice("userType").
    When(func(ctx *flow.Context) bool {
        userType, _ := ctx.GetString("user.type")
        return userType == "premium"
    }).
        Step("premiumData", premiumStep).
    Otherwise().
        Step("basicData", basicStep).
EndChoice()

// Parallel execution
flow.Parallel("dataFetch").
    Step("profile", profileStep).
    Step("preferences", preferencesStep).
    Step("notifications", notificationsStep).
EndParallel()
```

### Context (`context.go`)

The `Context` provides thread-safe data storage and type-safe access patterns for sharing data between steps.

#### Key Features:
- **Thread-Safe Operations**: All context operations are protected by RWMutex
- **Type-Safe Access**: Strongly typed getters with error handling
- **Nested Value Support**: Access nested data using dot notation
- **Context Cloning**: Create isolated copies for parallel execution
- **Observability Integration**: Built-in logger, span, and metrics support
- **Configuration Access**: Centralized access to framework configuration

#### Usage Examples:
```go
// Basic operations
ctx.Set("userId", "123")
ctx.Set("user", map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
    "profile": map[string]interface{}{
        "age": 30,
        "active": true,
    },
})

// Type-safe access
userId, err := ctx.GetString("userId")
userAge, err := ctx.GetInt("user.profile.age")
isActive, err := ctx.GetBool("user.profile.active")
userMap, err := ctx.GetMap("user")

// Existence checks
if ctx.Has("user.profile") {
    // Handle profile data
}

// Context cloning for parallel operations
clonedCtx := ctx.Clone()
```

#### Advanced Context Operations:
```go
// Access observability tools
ctx.Logger().Info("Processing user", zap.String("userId", userId))
ctx.Span().SetAttributes(attribute.String("user.id", userId))

// Access configuration
httpConfig := ctx.Config().HTTP
cacheConfig := ctx.Config().Cache

// Metadata access
executionID := ctx.ExecutionID()
duration := ctx.Duration()
flowName := ctx.FlowName()
```

### Step Interface (`step.go`)

The step system provides a unified interface for all operations within a flow, with both legacy and modern implementations.

#### Step Interface:
```go
type Step interface {
    Run(ctx *Context) error
    Name() string
    Description() string
}
```

#### Built-in Step Types:

##### BaseStep
Provides common functionality for all step implementations:
```go
step := flow.NewBaseStep("myStep", "Description of what this step does").
    WithTimeout(10 * time.Second)
```

##### ConditionalStep
Executes a step only if a condition is met:
```go
conditionalStep := flow.NewConditionalStep("checkUser", 
    func(ctx *flow.Context) bool {
        status, _ := ctx.GetString("user.status")
        return status == "active"
    }, 
    actualStep)
```

##### ParallelStep
Executes multiple steps concurrently:
```go
parallelStep := flow.NewParallelStep("fetchData",
    profileStep,
    preferencesStep,
    notificationsStep)
```

##### SequentialStep
Executes steps in sequence (useful for grouping):
```go
sequentialStep := flow.NewSequentialStep("userFlow",
    validateStep,
    fetchStep,
    transformStep)
```

##### RetryStep
Wraps another step with retry logic:
```go
retryStep := flow.NewRetryStep("retryAPI", apiStep, 3, 2*time.Second).
    WithRetryCondition(func(err error) bool {
        return errors.IsRetryable(err)
    })
```

##### TransformStep
Applies data transformations:
```go
transformStep := flow.NewTransformStep("mobileTransform", 
    func(ctx *flow.Context) error {
        // Custom transformation logic
        return nil
    })
```

##### DelayStep
Introduces controlled delays:
```go
delayStep := flow.NewDelayStep("rateLimitDelay", 1*time.Second)
```

## Integration with Other Packages

### Steps Package Integration
The flow package works seamlessly with the steps package through the `StepWrapper`:

```go
// Wrap new interface steps for use in flows
httpStep := flow.NewStepWrapper(
    http.NewMobileAPIStep("profile", "GET", "/api/profile", 
        []string{"id", "name", "avatar"}))

flow.Step("fetchProfile", httpStep)
```

### Transformers Integration
Transformers are integrated through transform steps:

```go
// Using transformers in flows
flow.Step("transform", flow.NewStepWrapper(
    core.NewTransformStep("mobileTransform",
        transformers.NewMobileTransformer([]string{"id", "name"}))))
```

### Validators Integration
Validators are integrated through validation steps:

```go
// Using validators in flows
flow.Step("validate", flow.NewStepWrapper(
    core.NewValidationStep("validateUser",
        validators.NewRequiredFieldsValidator("id", "email"))))
```

### Metrics Integration
Automatic metrics collection for all flow and step executions:

```go
// Metrics are automatically recorded for:
// - Flow execution duration
// - Step execution duration
// - Success/failure rates
// - Concurrent executions
```

### Error Handling Integration
Integration with the errors package for structured error handling:

```go
// Framework errors are automatically handled
if err := step.Run(ctx); err != nil {
    if errors.IsRetryable(err) {
        // Automatic retry logic
    }
    statusCode := errors.GetHTTPStatus(err)
    errorType := errors.GetErrorType(err)
}
```

## Extensibility

### Custom Steps
Create custom steps by implementing the Step interface:

```go
type CustomStep struct {
    *flow.BaseStep
    customField string
}

func NewCustomStep(name, customField string) *CustomStep {
    return &CustomStep{
        BaseStep:    flow.NewBaseStep(name, "Custom step description"),
        customField: customField,
    }
}

func (cs *CustomStep) Run(ctx *flow.Context) error {
    // Custom logic here
    ctx.Logger().Info("Executing custom step", 
        zap.String("custom_field", cs.customField))
    return nil
}
```

### Custom Middleware
Add cross-cutting concerns through middleware:

```go
func CustomMiddleware() flow.Middleware {
    return func(next func(*flow.Context) error) func(*flow.Context) error {
        return func(ctx *flow.Context) error {
            // Pre-execution logic
            start := time.Now()
            
            err := next(ctx)
            
            // Post-execution logic
            duration := time.Since(start)
            ctx.Logger().Info("Flow completed", 
                zap.Duration("duration", duration))
            
            return err
        }
    }
}

// Usage
flow.Use(CustomMiddleware())
```

### Flow Composition
Compose complex flows from simpler ones:

```go
func CreateUserFlow() *flow.Flow {
    return flow.NewFlow("UserFlow").
        Step("auth", authStep).
        Step("fetch", fetchStep)
}

func CreateComplexFlow() *flow.Flow {
    userFlow := CreateUserFlow()
    
    return flow.NewFlow("ComplexFlow").
        Step("userFlow", flow.NewSequentialStep("userSubflow", 
            userFlow.Steps()...)).
        Step("additional", additionalStep)
}
```

## Performance Considerations

### Context Cloning
Context cloning is optimized for parallel execution:
- Shallow copy of configuration and metadata
- Deep copy of data values only when necessary
- Efficient memory usage through copy-on-write patterns

### Parallel Execution
Parallel steps are optimized for performance:
- Goroutine pooling to prevent resource exhaustion
- Proper synchronization to avoid race conditions
- Error aggregation for comprehensive error reporting

### Memory Management
- Object pooling for frequently used objects
- Efficient string interpolation with minimal allocations
- Garbage collection friendly patterns

## Best Practices

### Flow Design
1. **Keep flows focused**: Each flow should have a single responsibility
2. **Use descriptive names**: Flow and step names should be self-documenting
3. **Handle errors gracefully**: Use retry logic and fallback mechanisms
4. **Set appropriate timeouts**: Different operations need different timeout values

### Context Usage
1. **Use type-safe getters**: Always use typed getters with error handling
2. **Minimize context size**: Only store necessary data in context
3. **Clone for parallel operations**: Always clone context for parallel steps
4. **Use nested keys**: Organize data hierarchically for better structure

### Step Implementation
1. **Implement proper logging**: Use structured logging with relevant fields
2. **Handle timeouts**: Respect context cancellation and timeouts
3. **Return meaningful errors**: Use the errors package for structured errors
4. **Keep steps stateless**: Steps should not maintain internal state

## Examples

### Complete Mobile BFF Flow
```go
func CreateMobileBFFFlow() *flow.Flow {
    return flow.NewFlow("MobileBFF").
        WithDescription("Mobile Backend for Frontend aggregation").
        WithTimeout(15 * time.Second).
        Use(middleware.LoggingMiddleware()).
        Use(middleware.MetricsMiddleware()).
        
        // Authentication
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization").
                WithClaimsExtraction(true))).
        
        // Parallel data fetching
        Parallel("dataAggregation").
            Step("profile", flow.NewStepWrapper(
                http.NewMobileAPIStep("profile", "GET", "/api/profile",
                    []string{"id", "name", "avatar"}).
                    WithCaching("profile", 10*time.Minute))).
            Step("notifications", flow.NewStepWrapper(
                http.NewMobileAPIStep("notifications", "GET", "/api/notifications",
                    []string{"id", "title", "timestamp"}).
                    WithFallback(map[string]interface{}{"items": []interface{}{}}))).
            Step("preferences", flow.NewStepWrapper(
                http.NewMobileAPIStep("preferences", "GET", "/api/preferences",
                    []string{"theme", "language", "notifications"}).
                    WithCaching("preferences", 30*time.Minute))).
        EndParallel().
        
        // Data transformation
        Step("transform", flow.NewStepWrapper(
            core.NewTransformStep("mobileTransform",
                transformers.NewTransformerChain("mobile",
                    transformers.NewFieldTransformer("select", 
                        []string{"profile", "notifications", "preferences"}),
                    transformers.NewMobileTransformer([]string{"id", "name", "avatar"}))))).
        
        // Validation
        Step("validate", flow.NewStepWrapper(
            core.NewValidationStep("validateResponse",
                validators.NewRequiredFieldsValidator("profile", "notifications"))))
}
```

### Error Handling Flow
```go
func CreateResilientFlow() *flow.Flow {
    return flow.NewFlow("ResilientFlow").
        WithTimeout(30 * time.Second).
        
        // Primary data source with retry
        Step("primary", flow.NewRetryStep("primaryAPI",
            flow.NewStepWrapper(
                http.GET("/api/primary").SaveAs("primaryData")),
            3, 2*time.Second).
            WithRetryCondition(func(err error) bool {
                return errors.IsRetryable(err)
            })).
        
        // Fallback to secondary source if primary fails
        Choice("dataSource").
            When(func(ctx *flow.Context) bool {
                return ctx.Has("primaryData")
            }).
                Step("log", flow.NewStepWrapper(
                    core.NewLogStep("success", "info", "Primary data source succeeded"))).
            Otherwise().
                Step("fallback", flow.NewStepWrapper(
                    http.GET("/api/fallback").SaveAs("fallbackData"))).
                Step("logFallback", flow.NewStepWrapper(
                    core.NewLogStep("fallback", "warn", "Using fallback data source"))).
        EndChoice()
}
```

## Troubleshooting

### Common Issues

1. **Context Race Conditions**: Always clone context for parallel operations
2. **Memory Leaks**: Ensure proper cleanup of goroutines in parallel steps
3. **Timeout Issues**: Set appropriate timeouts for different operation types
4. **Error Propagation**: Use structured errors for better error handling

### Debugging

1. **Enable Debug Logging**: Set log level to debug for detailed execution traces
2. **Use Metrics**: Monitor flow execution metrics for performance issues
3. **Trace Execution**: Use distributed tracing for complex flow debugging
4. **Context Inspection**: Log context state at key points for debugging

### Performance Tuning

1. **Optimize Parallel Execution**: Balance parallelism with resource usage
2. **Cache Frequently Used Data**: Use caching steps for expensive operations
3. **Minimize Context Size**: Only store necessary data in context
4. **Use Connection Pooling**: Configure HTTP clients with appropriate pooling 