# Dependency Injection Package Documentation

## Overview

The `pkg/di` package provides comprehensive dependency injection for the API Orchestration Framework using [Uber Fx](https://github.com/uber-go/fx). It centralizes dependency management across all framework modules, enabling clean architecture, testability, and modular design.

## Purpose

The dependency injection package serves as the foundation for:
- **Centralized Dependency Management**: All framework components use consistent DI patterns
- **Loose Coupling**: Components depend on interfaces, not concrete implementations
- **Testability**: Easy mocking and testing through dependency injection
- **Modularity**: Components can be easily swapped and extended
- **Configuration Management**: Centralized configuration injection
- **Lifecycle Management**: Automatic startup/shutdown coordination

## Core Architecture

### DI Module Structure

The main DI module provides all framework dependencies:
```go
func Module() fx.Option {
    return fx.Options(
        fx.Provide(
            // Core dependencies
            provideConfig,
            provideLogger,
            provideMetrics,
            
            // Framework components
            provideStepRegistry,
            provideFlowContext,
            provideHTTPClient,
            provideFlowFactory,
            
            // Transformers
            provideFieldTransformer,
            provideMobileTransformer,
            provideFlattenTransformer,
            
            // Validators
            provideRequiredFieldsValidator,
            provideEmailValidator,
            
            // Infrastructure
            provideRouter,
            provideServer,
        ),
    )
}
```

### Available Dependencies

#### Core Dependencies
- `*config.FrameworkConfig` - Framework configuration
- `*zap.Logger` - Structured logger
- `metrics.MetricsCollector` - Metrics collection

#### Framework Components
- `*registry.StepRegistry` - Step registry for dynamic step creation
- `interfaces.ExecutionContext` - Flow execution context
- `httpsteps.HTTPClient` - Resilient HTTP client
- `*di.FlowFactory` - Factory for creating flows and contexts

#### Transformers
- `transformers.Transformer` (Field) - Field selection transformer
- `transformers.Transformer` (Mobile) - Mobile optimization transformer
- `transformers.Transformer` (Flatten) - Data flattening transformer

#### Validators
- `validators.Validator` (Required Fields) - Required field validation
- `validators.Validator` (Email) - Email validation

#### Infrastructure
- `*gin.Engine` - HTTP router
- `*http.Server` - HTTP server

## Quick Start

### Basic Application Setup

```go
package main

import (
    "go.uber.org/fx"
    "github.com/venkatvghub/api-orchestration-framework/pkg/di"
)

func main() {
    app := fx.New(
        di.Module(),
        fx.Invoke(startApplication),
    )
    app.Run()
}

func startApplication(
    flowFactory *di.FlowFactory,
    logger *zap.Logger,
) {
    logger.Info("Application started with DI")
    
    // Create and execute flows
    flow := flowFactory.CreateFlow("example")
    ctx := flowFactory.CreateContext()
    
    result, err := flow.Execute(ctx)
    if err != nil {
        logger.Error("Flow execution failed", zap.Error(err))
    }
}
```

### Integration with Main Application

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
    "go.uber.org/zap"

    "github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/handlers"
    "github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/middleware"
    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/di"
    "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

func main() {
    app := fx.New(
        di.Module(),
        fx.Invoke(startServer),
    )
    app.Run()
}

func startServer(lc fx.Lifecycle, server *http.Server, logger *zap.Logger, router *gin.Engine, cfg *config.FrameworkConfig, metricsCollector metrics.MetricsCollector) {
    // Setup router with middleware and routes
    setupRouter(router, cfg, logger, metricsCollector)
    
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go func() {
                logger.Info("Starting Mobile BFF server",
                    zap.String("addr", server.Addr),
                    zap.String("version", "1.0.0"))

                if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                    logger.Fatal("Server failed to start", zap.Error(err))
                }
            }()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            logger.Info("Shutting down server")
            return server.Shutdown(ctx)
        },
    })
}

func setupRouter(router *gin.Engine, cfg *config.FrameworkConfig, logger *zap.Logger, metricsCollector metrics.MetricsCollector) {
    // Add middleware
    router.Use(middleware.LoggingMiddleware(logger))
    router.Use(middleware.MetricsMiddleware(metricsCollector))
    router.Use(middleware.CORSMiddleware())
    
    // Add routes
    api := router.Group("/api/v1")
    {
        api.GET("/health", handlers.HealthHandler())
        api.POST("/mobile/dashboard", handlers.MobileDashboardHandler())
        api.POST("/mobile/profile", handlers.MobileProfileHandler())
    }
    
    // Metrics endpoint
    router.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))
}
```

## Usage Patterns

### 1. Flow Orchestration with DI

```go
type UserService struct {
    flowFactory *di.FlowFactory
    logger      *zap.Logger
}

func NewUserService(flowFactory *di.FlowFactory, logger *zap.Logger) *UserService {
    return &UserService{
        flowFactory: flowFactory,
        logger:      logger,
    }
}

func (s *UserService) ProcessUser(userData map[string]interface{}) error {
    flow := s.flowFactory.CreateFlow("user_processing").
        StepFunc("validate", func(ctx interfaces.ExecutionContext) error {
            s.logger.Info("Validating user data")
            // Validation logic
            return nil
        }).
        StepFunc("transform", func(ctx interfaces.ExecutionContext) error {
            s.logger.Info("Transforming user data")
            // Transformation logic
            return nil
        }).
        StepFunc("save", func(ctx interfaces.ExecutionContext) error {
            s.logger.Info("Saving user data")
            // Save logic
            return nil
        })
    
    ctx := s.flowFactory.CreateContext()
    ctx.Set("user_data", userData)
    
    result, err := flow.Execute(ctx)
    if err != nil {
        s.logger.Error("User processing failed", zap.Error(err))
        return err
    }
    
    s.logger.Info("User processed successfully", 
        zap.Duration("duration", result.Duration))
    return nil
}
```

### 2. HTTP Service with DI

```go
type APIService struct {
    httpClient httpsteps.HTTPClient
    logger     *zap.Logger
}

func NewAPIService(httpClient httpsteps.HTTPClient, logger *zap.Logger) *APIService {
    return &APIService{
        httpClient: httpClient,
        logger:     logger,
    }
}

func (s *APIService) FetchUserData(userID string) (map[string]interface{}, error) {
    url := fmt.Sprintf("/api/users/%s", userID)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := s.httpClient.Do(req)
    if err != nil {
        s.logger.Error("API call failed", zap.Error(err))
        return nil, err
    }
    defer resp.Body.Close()
    
    // Process response...
    return userData, nil
}
```

### 3. Data Processing with Transformers and Validators

```go
type DataProcessor struct {
    transformer transformers.Transformer
    validator   validators.Validator
    logger      *zap.Logger
}

func NewDataProcessor(
    transformer transformers.Transformer,
    validator validators.Validator,
    logger *zap.Logger,
) *DataProcessor {
    return &DataProcessor{
        transformer: transformer,
        validator:   validator,
        logger:      logger,
    }
}

func (p *DataProcessor) ProcessData(data map[string]interface{}) (map[string]interface{}, error) {
    // Validate input
    if err := p.validator.Validate(data); err != nil {
        p.logger.Error("Validation failed", zap.Error(err))
        return nil, err
    }
    
    // Transform data
    result, err := p.transformer.Transform(data)
    if err != nil {
        p.logger.Error("Transformation failed", zap.Error(err))
        return nil, err
    }
    
    p.logger.Info("Data processed successfully")
    return result, nil
}
```

## Creating Custom Modules

### Service-Specific Module

```go
func UserModule() fx.Option {
    return fx.Options(
        fx.Provide(
            NewUserService,
            NewUserRepository,
            NewUserValidator,
        ),
        fx.Invoke(func(service *UserService, logger *zap.Logger) {
            logger.Info("User module initialized")
        }),
    )
}

// Use in main application
func main() {
    app := fx.New(
        di.Module(),    // Core framework dependencies
        UserModule(),   // User-specific module
        fx.Invoke(startApp),
    )
    app.Run()
}
```

### Feature Module

```go
func NotificationModule() fx.Option {
    return fx.Options(
        fx.Provide(
            NewNotificationService,
            NewEmailProvider,
            NewSMSProvider,
            NewPushProvider,
        ),
        fx.Invoke(func(
            service *NotificationService,
            logger *zap.Logger,
        ) {
            logger.Info("Notification module initialized")
        }),
    )
}
```

## Testing with DI

### Unit Testing

```go
func TestUserService(t *testing.T) {
    // Create test dependencies
    cfg := &config.FrameworkConfig{
        HTTP: config.HTTPConfig{
            RequestTimeout: 5 * time.Second,
        },
    }
    logger := zap.NewNop()
    registry := registry.NewStepRegistry()
    
    flowFactory := di.NewFlowFactory(cfg, logger, registry)
    service := NewUserService(flowFactory, logger)
    
    // Test the service
    testData := map[string]interface{}{
        "id":    "123",
        "name":  "John Doe",
        "email": "john@example.com",
    }
    
    err := service.ProcessUser(testData)
    assert.NoError(t, err)
}
```

### Integration Testing

```go
func TestIntegration(t *testing.T) {
    var userService *UserService
    var apiService *APIService
    
    app := fx.New(
        di.Module(),
        fx.Provide(
            NewUserService,
            NewAPIService,
        ),
        fx.Populate(&userService, &apiService),
    )
    
    err := app.Start(context.Background())
    require.NoError(t, err)
    defer app.Stop(context.Background())
    
    // Test integrated services
    result, err := apiService.FetchUserData("123")
    require.NoError(t, err)
    
    err = userService.ProcessUser(result)
    require.NoError(t, err)
}
```

### Mock Dependencies

```go
func TestWithMocks(t *testing.T) {
    // Create mock dependencies
    mockHTTPClient := &MockHTTPClient{}
    mockLogger := zap.NewNop()
    
    service := NewAPIService(mockHTTPClient, mockLogger)
    
    // Setup mock expectations
    mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{
        StatusCode: 200,
        Body:       ioutil.NopCloser(strings.NewReader(`{"id": "123"}`)),
    }, nil)
    
    // Test with mocks
    result, err := service.FetchUserData("123")
    assert.NoError(t, err)
    assert.Equal(t, "123", result["id"])
    
    mockHTTPClient.AssertExpectations(t)
}
```

## Configuration with DI

### Environment-Specific Configuration

```go
func DevelopmentModule() fx.Option {
    return fx.Options(
        fx.Decorate(func(cfg *config.FrameworkConfig) *config.FrameworkConfig {
            // Override for development
            cfg.Logging.Level = "debug"
            cfg.HTTP.MaxRetries = 1
            cfg.Cache.DefaultTTL = 1 * time.Minute
            return cfg
        }),
    )
}

func ProductionModule() fx.Option {
    return fx.Options(
        fx.Decorate(func(cfg *config.FrameworkConfig) *config.FrameworkConfig {
            // Override for production
            cfg.Logging.Level = "info"
            cfg.HTTP.MaxRetries = 3
            cfg.Cache.DefaultTTL = 10 * time.Minute
            return cfg
        }),
    )
}

// Use based on environment
func main() {
    modules := []fx.Option{di.Module()}
    
    if os.Getenv("ENVIRONMENT") == "production" {
        modules = append(modules, ProductionModule())
    } else {
        modules = append(modules, DevelopmentModule())
    }
    
    app := fx.New(modules...)
    app.Run()
}
```

### Custom Providers

```go
func CustomModule() fx.Option {
    return fx.Options(
        fx.Provide(
            // Custom HTTP client with specific configuration
            func(cfg *config.FrameworkConfig) httpsteps.HTTPClient {
                clientConfig := &httpsteps.ClientConfig{
                    RequestTimeout: cfg.HTTP.RequestTimeout,
                    MaxRetries:     cfg.HTTP.MaxRetries,
                    EnableFallback: cfg.HTTP.EnableFallback,
                }
                return httpsteps.NewResilientHTTPClient(clientConfig)
            },
            
            // Custom transformer with business logic
            func() transformers.Transformer {
                return transformers.NewFunctionTransformer("business", func(data map[string]interface{}) (map[string]interface{}, error) {
                    // Custom business logic
                    return data, nil
                })
            },
        ),
    )
}
```

## Advanced Patterns

### Conditional Dependencies

```go
func ConditionalModule() fx.Option {
    return fx.Options(
        fx.Provide(
            fx.Annotated{
                Name: "cache",
                Target: func(cfg *config.FrameworkConfig) interface{} {
                    if cfg.Cache.EnableRedis {
                        return NewRedisCache(cfg.Cache.RedisURL)
                    }
                    return NewInMemoryCache(cfg.Cache.MaxSize)
                },
            },
        ),
    )
}
```

### Lifecycle Hooks

```go
func LifecycleModule() fx.Option {
    return fx.Options(
        fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) {
            lc.Append(fx.Hook{
                OnStart: func(ctx context.Context) error {
                    logger.Info("Application starting")
                    // Initialize resources
                    return nil
                },
                OnStop: func(ctx context.Context) error {
                    logger.Info("Application stopping")
                    // Cleanup resources
                    return nil
                },
            })
        }),
    )
}
```

### Decorator Pattern

```go
func EnhancedModule() fx.Option {
    return fx.Options(
        fx.Decorate(func(logger *zap.Logger) *zap.Logger {
            // Add additional fields to logger
            return logger.With(
                zap.String("service", "api-orchestration"),
                zap.String("version", "2.0.0"),
            )
        }),
        
        fx.Decorate(func(httpClient httpsteps.HTTPClient) httpsteps.HTTPClient {
            // Wrap HTTP client with additional functionality
            return NewInstrumentedHTTPClient(httpClient)
        }),
    )
}
```

## Best Practices

### Dependency Design
1. **Use Interfaces**: Define interfaces for dependencies to enable easy testing and mocking
2. **Constructor Injection**: Use constructor functions that accept dependencies as parameters
3. **Avoid Globals**: Use DI instead of global variables
4. **Single Responsibility**: Each provider should have a single responsibility
5. **Error Handling**: Providers can return errors for validation

### Module Organization
1. **Group Related Dependencies**: Organize related providers into modules
2. **Layer Separation**: Separate infrastructure, domain, and application layers
3. **Feature Modules**: Create modules for specific features or domains
4. **Environment Modules**: Use different modules for different environments
5. **Testing Modules**: Create specific modules for testing scenarios

### Performance Considerations
1. **Lazy Initialization**: Use fx.Lazy for expensive dependencies
2. **Singleton Pattern**: Most dependencies should be singletons
3. **Avoid Circular Dependencies**: Design dependencies to avoid cycles
4. **Resource Management**: Use lifecycle hooks for resource cleanup
5. **Memory Usage**: Monitor memory usage with large dependency graphs

## Integration with Framework Components

### Flow Integration
```go
// Flows automatically use injected dependencies
func CreateUserFlow(
    flowFactory *di.FlowFactory,
    httpClient httpsteps.HTTPClient,
    transformer transformers.Transformer,
) *flow.Flow {
    return flowFactory.CreateFlow("user_flow").
        Step("fetch", httpsteps.GET("/api/users/{{userId}}")).
        Step("transform", core.NewTransformStep("mobile", transformer))
}
```

### Step Integration
```go
// Steps can be created with injected dependencies
func NewCustomStep(
    httpClient httpsteps.HTTPClient,
    logger *zap.Logger,
) interfaces.Step {
    return &CustomStep{
        httpClient: httpClient,
        logger:     logger,
    }
}
```

### Middleware Integration
```go
// Middleware can use injected dependencies
func NewAuthMiddleware(
    validator validators.Validator,
    logger *zap.Logger,
) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Use injected validator and logger
        token := c.GetHeader("Authorization")
        if err := validator.Validate(map[string]interface{}{"token": token}); err != nil {
            logger.Error("Authentication failed", zap.Error(err))
            c.AbortWithStatus(401)
            return
        }
        c.Next()
    }
}
```

## Troubleshooting

### Common Issues

1. **Circular Dependencies**
   ```go
   // Wrong: A depends on B, B depends on A
   func ProvideA(b B) A { return NewA(b) }
   func ProvideB(a A) B { return NewB(a) }
   
   // Correct: Use interfaces or refactor dependencies
   func ProvideA(b BInterface) A { return NewA(b) }
   func ProvideB() B { return NewB() }
   ```

2. **Missing Dependencies**
   ```go
   // Ensure all required dependencies are provided
   fx.New(
       di.Module(),
       fx.Provide(
           NewMyService, // Make sure all dependencies of MyService are available
       ),
   )
   ```

3. **Type Conflicts**
   ```go
   // Use fx.Annotated for multiple instances of same type
   fx.Provide(
       fx.Annotated{
           Name:   "primary",
           Target: NewPrimaryDatabase,
       },
       fx.Annotated{
           Name:   "secondary", 
           Target: NewSecondaryDatabase,
       },
   )
   ```

### Debugging

1. **Enable Debug Logging**: Use fx.WithLogger for detailed DI logs
2. **Dependency Visualization**: Use fx.Visualize to see dependency graph
3. **Startup Errors**: Check fx.New errors for missing dependencies
4. **Lifecycle Issues**: Monitor fx.Hook execution for startup/shutdown problems

## Examples

See the complete examples in `pkg/di/examples.go` for:
- Flow orchestration with DI
- Data transformation services
- Validation services  
- HTTP client services
- Custom step registration
- Complete demo application

## Related Documentation

- [Configuration Guide](config.md) - Framework configuration management
- [Flow Documentation](flow.md) - Flow orchestration patterns
- [Steps Reference](steps.md) - Available step types
- [Transformers Guide](transformers.md) - Data transformation
- [Validators Guide](validators.md) - Data validation
- [Error Handling](errors.md) - Structured error handling
- [Metrics & Monitoring](metrics.md) - Observability
- [Getting Started](getting-started.md) - Quick start guide 