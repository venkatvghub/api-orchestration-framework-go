# Steps Package Documentation

## Overview

The `pkg/steps` package contains all step implementations organized into logical categories. It provides a comprehensive collection of pre-built steps for common operations like HTTP requests, authentication, caching, validation, and mobile-specific BFF patterns.

## Purpose

The steps package serves as the operational toolkit that:
- Provides ready-to-use step implementations for common operations
- Organizes steps by functional category for better discoverability
- Implements the new interfaces.Step interface for consistency
- Offers mobile-optimized steps for BFF patterns
- Supports extensibility through base step types

## Package Structure

```
pkg/steps/
├── base/           # Base step types and interfaces
├── core/           # Core utility steps (auth, cache, validation, etc.)
├── http/           # HTTP-related steps with resilience
└── bff/            # BFF-specific steps and mobile optimizations
```

## Base Package (`pkg/steps/base`)

### Purpose
Provides foundational step types and interfaces that other step packages build upon.

### Key Components

#### Step Interface
```go
type Step interface {
    Run(ctx *flow.Context) error
    Name() string
    Description() string
}
```

#### BaseStep
Common functionality for all step implementations:
```go
type BaseStep struct {
    name        string
    description string
    timeout     time.Duration
}

// Usage
step := base.NewBaseStep("myStep", "Description").
    WithTimeout(10 * time.Second)
```

#### Composite Step Types

##### ConditionalStep
Executes a step only when a condition is met:
```go
conditionalStep := base.NewConditionalStep("checkActive",
    func(ctx *flow.Context) bool {
        status, _ := ctx.GetString("user.status")
        return status == "active"
    },
    actualStep)
```

##### SequentialStep
Executes multiple steps in sequence:
```go
sequentialStep := base.NewSequentialStep("userWorkflow",
    authStep,
    fetchStep,
    transformStep)
```

##### ParallelStep
Executes multiple steps concurrently with proper synchronization:
```go
parallelStep := base.NewParallelStep("dataFetch",
    profileStep,
    preferencesStep,
    notificationsStep)
```

##### RetryStep
Wraps another step with configurable retry logic:
```go
retryStep := base.NewRetryStep("retryAPI", apiStep, 3, 2*time.Second).
    WithRetryCondition(func(err error) bool {
        return errors.IsRetryable(err)
    })
```

##### DelayStep
Introduces controlled delays for rate limiting or timing:
```go
delayStep := base.NewDelayStep("rateLimitDelay", 1*time.Second)
```

## Core Package (`pkg/steps/core`)

### Purpose
Provides essential utility steps for authentication, caching, validation, logging, and data manipulation.

### Authentication Steps (`auth.go`)

#### TokenValidationStep
Validates authentication tokens with configurable validation logic:
```go
authStep := core.NewTokenValidationStep("auth", "Authorization").
    WithTokenPrefix("Bearer ").
    WithClaimsExtraction(true).
    WithValidationFunc(func(token string) bool {
        // Custom validation logic
        return isValidToken(token)
    })
```

Features:
- Thread-safe token storage with sync.Map
- Configurable token prefix handling
- Claims extraction for JWT tokens
- Custom validation functions
- Header extraction capabilities

#### HeaderExtractionStep
Extracts and validates HTTP headers:
```go
headerStep := core.NewHeaderExtractionStep("extractHeaders", 
    "Authorization", "X-User-ID", "X-Request-ID").
    WithRequired("Authorization").
    WithSanitization(true)
```

### Caching Steps (`cache.go`)

#### CacheStep
Thread-safe caching with TTL support:
```go
// Cache operations
cacheGet := core.NewCacheGetStep("getUser", "user_${userId}", "cached_user")
cacheSet := core.NewCacheSetStep("setUser", "user_${userId}", "user_data", 10*time.Minute)
cacheDelete := core.NewCacheDeleteStep("deleteUser", "user_${userId}")
cacheClear := core.NewCacheClearStep("clearAll")
```

Features:
- Thread-safe operations with sync.Map
- TTL-based expiration
- Automatic cleanup of expired entries
- Context-aware key interpolation
- Cache statistics and monitoring

### Conditional Logic (`condition.go`)

#### ConditionStep
Evaluates conditions on context data:
```go
// Field-based conditions
conditionStep := core.NewConditionStep("checkStatus", "user.status", "equals", "active")

// Custom condition functions
customCondition := core.NewConditionStep("customCheck", "", "", nil).
    WithCustomCondition(func(ctx *flow.Context) bool {
        user, _ := ctx.GetMap("user")
        return user["role"] == "admin" && user["active"] == true
    })

// Convenience constructors
existsCondition := core.NewExistsCondition("userExists", "user.id")
equalsCondition := core.NewEqualsCondition("isActive", "user.status", "active")
containsCondition := core.NewContainsCondition("hasEmail", "user.email", "@")
inCondition := core.NewInCondition("validRole", "user.role", []interface{}{"admin", "user", "guest"})
```

Supported operators:
- `equals`, `not_equals`
- `greater_than`, `less_than`, `greater_equal`, `less_equal`
- `contains`, `starts_with`, `ends_with`
- `in`, `not_in`
- `exists`, `not_exists`
- `empty`, `not_empty`

### Logging Steps (`log.go`)

#### LogStep
Structured logging with field interpolation:
```go
logStep := core.NewLogStep("logUser", "info", "Processing user ${userId}").
    WithFields(map[string]string{
        "operation": "user_fetch",
        "source": "api",
    }).
    WithContext(true, "user", "request_id").
    WithSanitization(true)
```

#### MetricsLogStep
Logging with metrics collection:
```go
metricsStep := core.NewMetricsLogStep("userMetrics", "user_processing", "counter").
    WithValue("1").
    WithTags(map[string]string{
        "operation": "fetch",
        "source": "database",
    }).
    WithStats(true)
```

### Validation Steps (`validation.go`)

#### ValidationStep
Single validator execution:
```go
validationStep := core.NewValidationStep("validateUser", 
    validators.NewRequiredFieldsValidator("id", "email")).
    WithDataField("user").
    WithContinueOnError(false).
    WithResultStorage(true, "validation_result")
```

#### ValidationChainStep
Multiple validator execution with configurable behavior:
```go
chainStep := core.NewValidationChainStep("validateUserChain",
    validators.NewRequiredFieldsValidator("id", "email"),
    validators.NewTypeValidator().RequireString("email"),
    validators.NewCustomValidator("emailFormat", emailValidationFunc)).
    WithStopOnFirst(false).
    WithContinueOnError(true).
    WithResultStorage(true, "validation_results")
```

### Value Manipulation (`value.go`)

#### ValueStep
Comprehensive value operations:
```go
// Set values
setValue := core.NewSetValueStep("setTimestamp", "timestamp", time.Now().Unix())

// Copy values
copyValue := core.NewCopyValueStep("copyUser", "source_user", "target_user")

// Move values
moveValue := core.NewMoveValueStep("moveData", "temp_data", "final_data")

// Delete values
deleteValue := core.NewDeleteValueStep("cleanup", "temp_data")

// Transform values
transformValue := core.NewTransformValueStep("processUser", "raw_user", 
    func(value interface{}) interface{} {
        // Custom transformation logic
        return processedValue
    })
```

## HTTP Package (`pkg/steps/http`)

### Purpose
Provides HTTP client functionality with resilience patterns, mobile optimizations, and comprehensive error handling.

### HTTP Client (`client.go`)

#### ResilientHTTPClient
Advanced HTTP client with resilience patterns:
```go
clientConfig := &http.ClientConfig{
    // Retry configuration
    MaxRetries:           3,
    InitialRetryDelay:    1 * time.Second,
    MaxRetryDelay:        10 * time.Second,
    RetryableStatusCodes: []int{429, 502, 503, 504},
    
    // Circuit breaker configuration
    FailureThreshold:    5,
    SuccessThreshold:    3,
    CircuitBreakerDelay: 5 * time.Second,
    
    // Timeout configuration
    RequestTimeout:    15 * time.Second,
    ConnectionTimeout: 5 * time.Second,
    
    // Fallback configuration
    EnableFallback:     true,
    FallbackStatusCode: 200,
    FallbackBody:       []byte(`{"status": "fallback"}`),
}

client := http.NewResilientHTTPClient(clientConfig)
```

### HTTP Steps (`step.go`)

#### HTTPStep
Comprehensive HTTP request step with full configuration:
```go
httpStep := http.NewHTTPStep("GET", "https://api.example.com/users/${userId}").
    WithHeader("Authorization", "Bearer ${token}").
    WithHeader("Content-Type", "application/json").
    WithQueryParam("include", "profile,preferences").
    WithTimeout(10 * time.Second).
    WithExpectedStatus(200, 201).
    WithTransformer(transformers.NewMobileTransformer([]string{"id", "name"})).
    WithValidator(validators.NewRequiredFieldsValidator("id", "status")).
    SaveAs("user_data")
```

#### Convenience Constructors
```go
// Basic HTTP methods
getStep := http.GET("https://api.example.com/users/${userId}")
postStep := http.POST("https://api.example.com/users")
putStep := http.PUT("https://api.example.com/users/${userId}")
deleteStep := http.DELETE("https://api.example.com/users/${userId}")
patchStep := http.PATCH("https://api.example.com/users/${userId}")

// Specialized steps
jsonAPIStep := http.NewJSONAPIStep("GET", "https://api.example.com/data")
mobileAPIStep := http.NewMobileAPIStep("GET", "https://api.example.com/profile", 
    []string{"id", "name", "avatar"})
authenticatedStep := http.NewAuthenticatedStep("GET", "https://api.example.com/secure", "auth_token")
```

#### Request Configuration
```go
step := http.POST("https://api.example.com/orders").
    // Body configuration
    WithJSONBody(map[string]interface{}{
        "user_id": "${userId}",
        "items": []string{"item1", "item2"},
    }).
    WithFormBody(map[string]string{
        "username": "${username}",
        "password": "${password}",
    }).
    WithRawBody([]byte("custom data"), "application/octet-stream").
    
    // Authentication
    WithBasicAuth("${username}", "${password}").
    WithBearerToken("${access_token}").
    
    // Headers and cookies
    WithHeaders(map[string]string{
        "X-Custom-Header": "value",
        "X-Request-ID": "${request_id}",
    }).
    WithCookie(&http.Cookie{Name: "session", Value: "${session_id}"}).
    
    // Client configuration
    WithClientConfig(clientConfig).
    WithUserAgent("MyApp/1.0")
```

## BFF Package (`pkg/steps/bff`)

### Purpose
Provides Backend for Frontend (BFF) specific steps optimized for mobile applications, including aggregation patterns and mobile-specific transformations.

### Mobile API Steps (`mobile.go`)

#### MobileAPIStep
HTTP step optimized for mobile consumption:
```go
mobileStep := bff.NewMobileAPIStep("userProfile", "GET", 
    "https://api.example.com/users/${userId}",
    []string{"id", "name", "avatar", "email"}).
    WithCaching("user_profile", 10*time.Minute).
    WithFallback(map[string]interface{}{
        "id": "unknown",
        "name": "Guest User",
        "status": "offline",
    }).
    WithMobileHeaders("mobile", "1.0", "ios").
    WithAuth("access_token").
    WithRetry(3, 2*time.Second)
```

Features:
- Automatic field selection for mobile optimization
- Built-in caching with TTL
- Fallback data for offline scenarios
- Mobile-specific headers
- Integrated retry logic
- Performance metrics collection

#### Convenience Constructors
```go
// Pre-configured mobile steps
userProfile := bff.NewMobileUserProfileStep("https://api.example.com")
notifications := bff.NewMobileNotificationsStep("https://api.example.com")
content := bff.NewMobileContentStep("https://api.example.com")
search := bff.NewMobileSearchStep("https://api.example.com")
analytics := bff.NewMobileAnalyticsStep("https://api.example.com")
```

### Aggregation Steps (`aggregation.go`)

#### AggregationStep
Combines multiple API responses for complex mobile screens:
```go
aggregation := bff.NewAggregationStep("dashboardData").
    AddRequiredStep(userProfileStep).
    AddOptionalStep(notificationsStep, map[string]interface{}{
        "items": []interface{}{},
        "total": 0,
    }).
    AddOptionalStep(contentStep, map[string]interface{}{
        "items": []interface{}{},
        "status": "offline",
    }).
    WithTransformer(transformers.NewMobileTransformer([]string{"id", "name", "avatar"})).
    WithParallel(true).
    WithTimeout(15 * time.Second).
    WithFailFast(false)
```

Features:
- Required vs optional step classification
- Fallback data for failed optional steps
- Parallel or sequential execution
- Configurable failure handling
- Integrated transformation
- Comprehensive error handling

#### Pre-built Aggregation Patterns
```go
// Mobile dashboard aggregation
dashboard := bff.NewMobileDashboardAggregation("https://api.example.com")

// Mobile search aggregation
search := bff.NewMobileSearchAggregation("https://api.example.com")

// Mobile profile aggregation
profile := bff.NewMobileProfileAggregation("https://api.example.com")
```

## Integration Patterns

### With Flow Package
Steps integrate seamlessly with the flow package:
```go
flow := flow.NewFlow("UserFlow").
    Step("auth", flow.NewStepWrapper(
        core.NewTokenValidationStep("auth", "Authorization"))).
    Step("fetch", flow.NewStepWrapper(
        http.NewMobileAPIStep("profile", "GET", "/api/profile", 
            []string{"id", "name", "avatar"}))).
    Step("validate", flow.NewStepWrapper(
        core.NewValidationStep("validate", 
            validators.NewRequiredFieldsValidator("id", "name"))))
```

### With Transformers Package
Steps can use transformers for data processing:
```go
httpStep := http.GET("/api/data").
    WithTransformer(transformers.NewTransformerChain("mobile",
        transformers.NewFieldTransformer("select", []string{"id", "name"}),
        transformers.NewMobileTransformer([]string{"id", "name"})))
```

### With Validators Package
Steps can use validators for data validation:
```go
httpStep := http.GET("/api/data").
    WithValidator(validators.NewValidatorChain("validation",
        validators.NewRequiredFieldsValidator("id", "status"),
        validators.NewTypeValidator().RequireString("name")))
```

## Extensibility

### Creating Custom Steps
Implement the interfaces.Step interface:
```go
type CustomStep struct {
    name        string
    description string
    customField string
}

func NewCustomStep(name, customField string) *CustomStep {
    return &CustomStep{
        name:        name,
        description: "Custom step implementation",
        customField: customField,
    }
}

func (cs *CustomStep) Run(ctx interfaces.ExecutionContext) error {
    // Custom logic here
    ctx.Logger().Info("Executing custom step", 
        zap.String("custom_field", cs.customField))
    return nil
}

func (cs *CustomStep) Name() string {
    return cs.name
}

func (cs *CustomStep) Description() string {
    return cs.description
}
```

### Extending Existing Steps
Build upon existing step types:
```go
type EnhancedHTTPStep struct {
    *http.HTTPStep
    customBehavior string
}

func NewEnhancedHTTPStep(method, url string) *EnhancedHTTPStep {
    return &EnhancedHTTPStep{
        HTTPStep:       http.NewHTTPStep(method, url),
        customBehavior: "enhanced",
    }
}

func (ehs *EnhancedHTTPStep) Run(ctx interfaces.ExecutionContext) error {
    // Pre-processing
    ctx.Logger().Info("Enhanced HTTP step starting")
    
    // Call parent implementation
    err := ehs.HTTPStep.Run(ctx)
    
    // Post-processing
    if err == nil {
        ctx.Logger().Info("Enhanced HTTP step completed successfully")
    }
    
    return err
}
```

## Performance Considerations

### HTTP Steps
- Connection pooling for efficient resource usage
- Request/response body streaming for large payloads
- Automatic compression support
- Circuit breaker patterns for resilience

### Caching Steps
- Thread-safe operations with minimal locking
- Efficient TTL-based expiration
- Memory-efficient storage patterns
- Automatic cleanup of expired entries

### BFF Steps
- Parallel execution for independent operations
- Field selection to minimize payload size
- Caching strategies for frequently accessed data
- Fallback mechanisms for offline scenarios

## Best Practices

### Step Design
1. **Single Responsibility**: Each step should have one clear purpose
2. **Stateless Design**: Steps should not maintain internal state
3. **Error Handling**: Use structured errors with proper context
4. **Logging**: Include relevant context in log messages
5. **Timeouts**: Set appropriate timeouts for operations

### HTTP Steps
1. **Use Resilient Clients**: Configure retry and circuit breaker patterns
2. **Set Timeouts**: Configure appropriate request timeouts
3. **Handle Status Codes**: Validate expected status codes
4. **Use Transformers**: Transform responses for specific use cases
5. **Cache Responses**: Cache frequently accessed data

### BFF Steps
1. **Optimize for Mobile**: Use field selection and compression
2. **Provide Fallbacks**: Handle offline scenarios gracefully
3. **Use Aggregation**: Combine multiple API calls efficiently
4. **Monitor Performance**: Track metrics for optimization
5. **Handle Failures**: Implement graceful degradation

## Examples

### Complete User Profile Flow
```go
func CreateUserProfileFlow() *flow.Flow {
    return flow.NewFlow("UserProfile").
        WithTimeout(15 * time.Second).
        
        // Authentication
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization").
                WithClaimsExtraction(true))).
        
        // Fetch user data
        Step("fetchUser", flow.NewStepWrapper(
            bff.NewMobileUserProfileStep("https://api.example.com").
                WithCaching("user_profile", 10*time.Minute).
                WithFallback(map[string]interface{}{
                    "id": "unknown",
                    "name": "Guest User",
                }))).
        
        // Validate response
        Step("validate", flow.NewStepWrapper(
            core.NewValidationStep("validateUser",
                validators.NewRequiredFieldsValidator("id", "name")))).
        
        // Log completion
        Step("log", flow.NewStepWrapper(
            core.NewLogStep("completion", "info", "User profile fetched for ${userId}").
                WithFields(map[string]string{"operation": "profile_fetch"})))
}
```

### Mobile Dashboard Aggregation
```go
func CreateMobileDashboard() *flow.Flow {
    return flow.NewFlow("MobileDashboard").
        WithTimeout(20 * time.Second).
        
        // Authentication
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"))).
        
        // Aggregate dashboard data
        Step("aggregate", flow.NewStepWrapper(
            bff.NewMobileDashboardAggregation("https://api.example.com").
                WithParallel(true).
                WithTimeout(15 * time.Second).
                WithFailFast(false))).
        
        // Cache aggregated result
        Step("cache", flow.NewStepWrapper(
            core.NewCacheSetStep("cacheDashboard", "dashboard_${userId}", 
                "bff_aggregation", 5*time.Minute)))
} 