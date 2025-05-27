# Dependency Injection Package Documentation

## Overview

The dependency injection (DI) approach in the API Orchestration Framework has been **decentralized and isolated** to provide maximum flexibility and extensibility. Instead of a single centralized DI package, each example application now contains its own independent DI module tailored to its specific needs.

This approach enables:
- **Application-Specific Configuration**: Each application can define its own DI providers optimized for its use case
- **Independent Evolution**: Applications can evolve their DI configuration without affecting others
- **Reduced Coupling**: No shared DI dependencies between different applications
- **Enhanced Testability**: Each application can mock and test its DI independently
- **Extensibility**: Easy to add new applications with custom DI configurations

## Architecture

### Isolated DI Structure

Each example application contains its own `di/` directory with:
```
examples/
├── mobile-onboarding-v2/
│   └── di/
│       └── di.go          # Onboarding-specific DI configuration
├── mobile-bff/
│   └── di/
│       └── di.go          # BFF-specific DI configuration
└── [other-examples]/
    └── di/
        └── di.go          # Application-specific DI configuration
```

### Core DI Pattern

Each DI module follows the same core pattern but with application-specific providers:

```go
package di

import (
    // Standard imports
    "go.uber.org/fx"
    "go.uber.org/zap"
    
    // Framework imports
    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    // ... other framework imports
)

// Module returns application-specific DI configuration
func Module() fx.Option {
    return fx.Options(
        fx.Provide(
            // Core dependencies (common across all apps)
            provideConfig,
            provideLogger,
            provideMetrics,
            
            // Application-specific providers
            // ...
        ),
    )
}
```

## Application-Specific DI Configurations

### Mobile Onboarding v2 DI

**Location**: `examples/mobile-onboarding-v2/di/di.go`

**Purpose**: Optimized for mobile onboarding flows with WireMock integration

**Key Features**:
- Single mobile transformer for onboarding screens
- Onboarding-specific validators (user_id, screen_id)
- Metrics namespace: "mobile_onboarding.v2"
- HTTP client optimized for mock API calls

```go
// Mobile Onboarding specific providers
func provideMobileTransformer() transformers.Transformer {
    return transformers.NewMobileTransformer([]string{
        "id", "title", "description", "type", "fields", "actions", "next_screen"
    })
}

func provideRequiredFieldsValidator() validators.Validator {
    return validators.NewRequiredFieldsValidator("user_id", "screen_id")
}

func provideMetrics() metrics.MetricsCollector {
    return metrics.NewPrometheusMetrics("mobile_onboarding", "v2")
}
```

### Mobile BFF DI

**Location**: `examples/mobile-bff/di/di.go`

**Purpose**: Comprehensive BFF layer with multiple transformation capabilities

**Key Features**:
- Multiple named transformers (mobile, field, flatten)
- BFF-specific validators (user_id, request_id)
- Metrics namespace: "mobile_bff.api"
- HTTP client optimized for backend API aggregation

```go
// Mobile BFF specific providers with named transformers
fx.Annotated{
    Name:   "mobile",
    Target: provideMobileTransformer,
},
fx.Annotated{
    Name:   "field", 
    Target: provideFieldTransformer,
},
fx.Annotated{
    Name:   "flatten",
    Target: provideFlattenTransformer,
},
```

## Quick Start

### Creating a New Application with DI

1. **Create DI Directory**
   ```bash
   mkdir examples/my-new-app/di
   ```

2. **Create DI Configuration**
   ```go
   // examples/my-new-app/di/di.go
   package di
   
   import (
       "go.uber.org/fx"
       // ... framework imports
   )
   
   func Module() fx.Option {
       return fx.Options(
           fx.Provide(
               // Core dependencies
               provideConfig,
               provideLogger,
               provideMetrics,
               
               // App-specific providers
               provideCustomTransformer,
               provideCustomValidator,
               
               // Infrastructure
               provideRouter,
               provideServer,
           ),
       )
   }
   
   // App-specific providers
   func provideCustomTransformer() transformers.Transformer {
       return transformers.NewFieldTransformer("my_app", []string{"custom", "fields"})
   }
   ```

3. **Use in Main Application**
   ```go
   // examples/my-new-app/main.go
   package main
   
   import (
       "go.uber.org/fx"
       "github.com/venkatvghub/api-orchestration-framework/examples/my-new-app/di"
   )
   
   func main() {
       app := fx.New(
           di.Module(),
           fx.Invoke(startServer),
       )
       app.Run()
   }
   ```

### Basic Application Setup

```go
package main

import (
    "go.uber.org/fx"
    "github.com/venkatvghub/api-orchestration-framework/examples/mobile-onboarding-v2/di"
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
    logger.Info("Application started with isolated DI")
    
    // Create and execute flows
    flow := flowFactory.CreateFlow("example")
    ctx := flowFactory.CreateContext()
    
    result, err := flow.Execute(ctx)
    if err != nil {
        logger.Error("Flow execution failed", zap.Error(err))
    }
}
```

## Usage Patterns

### 1. Application-Specific Flow Orchestration

```go
// In your application's service layer
type OnboardingService struct {
    flowFactory *di.FlowFactory
    logger      *zap.Logger
}

func NewOnboardingService(flowFactory *di.FlowFactory, logger *zap.Logger) *OnboardingService {
    return &OnboardingService{
        flowFactory: flowFactory,
        logger:      logger,
    }
}

func (s *OnboardingService) ProcessOnboarding(userData map[string]interface{}) error {
    flow := s.flowFactory.CreateFlow("onboarding_flow").
        StepFunc("validate", func(ctx interfaces.ExecutionContext) error {
            s.logger.Info("Validating onboarding data")
            return nil
        }).
        StepFunc("transform", func(ctx interfaces.ExecutionContext) error {
            s.logger.Info("Transforming for mobile")
            return nil
        })
    
    ctx := s.flowFactory.CreateContext()
    ctx.Set("user_data", userData)
    
    result, err := flow.Execute(ctx)
    if err != nil {
        s.logger.Error("Onboarding failed", zap.Error(err))
        return err
    }
    
    s.logger.Info("Onboarding completed", 
        zap.Duration("duration", result.Duration))
    return nil
}
```

### 2. Named Transformer Usage (Mobile BFF)

```go
type BFFService struct {
    mobileTransformer  transformers.Transformer `name:"mobile"`
    fieldTransformer   transformers.Transformer `name:"field"`
    flattenTransformer transformers.Transformer `name:"flatten"`
    logger             *zap.Logger
}

func NewBFFService(
    mobileTransformer transformers.Transformer,
    fieldTransformer transformers.Transformer,
    flattenTransformer transformers.Transformer,
    logger *zap.Logger,
) *BFFService {
    return &BFFService{
        mobileTransformer:  mobileTransformer,
        fieldTransformer:   fieldTransformer,
        flattenTransformer: flattenTransformer,
        logger:             logger,
    }
}
```

### 3. Custom Application DI Module

```go
// Custom application with specific needs
func CustomAppModule() fx.Option {
    return fx.Options(
        fx.Provide(
            // Core framework dependencies
            provideConfig,
            provideLogger,
            provideMetrics,
            
            // Custom application providers
            provideCustomHTTPClient,
            provideCustomTransformer,
            provideCustomValidator,
            
            // Application-specific services
            NewCustomService,
            NewCustomHandler,
        ),
        fx.Invoke(func(
            service *CustomService,
            logger *zap.Logger,
        ) {
            logger.Info("Custom application initialized")
        }),
    )
}

func provideCustomHTTPClient(cfg *config.FrameworkConfig) httpsteps.HTTPClient {
    // Custom HTTP client configuration for this app
    clientConfig := &httpsteps.ClientConfig{
        RequestTimeout: 10 * time.Second,
        MaxRetries:     5,
        // ... custom settings
    }
    return httpsteps.NewResilientHTTPClient(clientConfig)
}
```

## Testing with Isolated DI

### Unit Testing

```go
func TestOnboardingService(t *testing.T) {
    // Create test-specific DI configuration
    cfg := &config.FrameworkConfig{
        HTTP: config.HTTPConfig{
            RequestTimeout: 5 * time.Second,
        },
    }
    logger := zap.NewNop()
    registry := registry.NewStepRegistry()
    
    flowFactory := di.NewFlowFactory(cfg, logger, registry)
    service := NewOnboardingService(flowFactory, logger)
    
    // Test the service
    testData := map[string]interface{}{
        "user_id":   "123",
        "screen_id": "welcome",
        "data":      map[string]interface{}{"accepted_terms": true},
    }
    
    err := service.ProcessOnboarding(testData)
    assert.NoError(t, err)
}
```

### Integration Testing

```go
func TestIntegration(t *testing.T) {
    var onboardingService *OnboardingService
    var bffService *BFFService
    
    app := fx.New(
        di.Module(), // Use the application's DI module
        fx.Provide(
            NewOnboardingService,
            NewBFFService,
        ),
        fx.Populate(&onboardingService, &bffService),
    )
    
    err := app.Start(context.Background())
    require.NoError(t, err)
    defer app.Stop(context.Background())
    
    // Test integrated services
    result, err := bffService.ProcessRequest("123")
    require.NoError(t, err)
    
    err = onboardingService.ProcessOnboarding(result)
    require.NoError(t, err)
}
```

### Mock Dependencies

```go
func TestWithMocks(t *testing.T) {
    // Create mock dependencies
    mockHTTPClient := &MockHTTPClient{}
    mockLogger := zap.NewNop()
    
    // Create service with mocks instead of DI
    service := NewOnboardingService(mockHTTPClient, mockLogger)
    
    // Setup mock expectations
    mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{
        StatusCode: 200,
        Body:       ioutil.NopCloser(strings.NewReader(`{"success": true}`)),
    }, nil)
    
    // Test with mocks
    result, err := service.ProcessOnboarding(testData)
    assert.NoError(t, err)
    
    mockHTTPClient.AssertExpectations(t)
}
```

## Configuration Patterns

### Environment-Specific DI

```go
func DevelopmentModule() fx.Option {
    return fx.Options(
        fx.Decorate(func(cfg *config.FrameworkConfig) *config.FrameworkConfig {
            // Override for development
            cfg.Logging.Level = "debug"
            cfg.HTTP.MaxRetries = 1
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

### Feature-Specific Modules

```go
func OnboardingModule() fx.Option {
    return fx.Options(
        fx.Provide(
            NewOnboardingService,
            NewOnboardingHandler,
            NewOnboardingValidator,
        ),
    )
}

func AnalyticsModule() fx.Option {
    return fx.Options(
        fx.Provide(
            NewAnalyticsService,
            NewAnalyticsHandler,
            NewEventTracker,
        ),
    )
}

// Combine modules
func main() {
    app := fx.New(
        di.Module(),           // Core DI
        OnboardingModule(),    // Onboarding features
        AnalyticsModule(),     // Analytics features
        fx.Invoke(startApp),
    )
    app.Run()
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

### Lifecycle Management

```go
func LifecycleModule() fx.Option {
    return fx.Options(
        fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) {
            lc.Append(fx.Hook{
                OnStart: func(ctx context.Context) error {
                    logger.Info("Application starting")
                    return nil
                },
                OnStop: func(ctx context.Context) error {
                    logger.Info("Application stopping")
                    return nil
                },
            })
        }),
    )
}
```

## Best Practices

### DI Design Principles
1. **Application Isolation**: Each application should have its own DI configuration
2. **Single Responsibility**: Each provider should have a single, clear purpose
3. **Interface Usage**: Use interfaces for dependencies to enable easy testing
4. **Named Providers**: Use `fx.Annotated` when providing multiple instances of the same type
5. **Environment Awareness**: Configure providers based on environment needs

### Module Organization
1. **Core Dependencies First**: Always provide config, logger, metrics first
2. **Framework Components**: Provide framework components (registry, context, HTTP client)
3. **Application Logic**: Provide transformers, validators, and business logic
4. **Infrastructure Last**: Provide router and server last

### Performance Considerations
1. **Lazy Initialization**: Use `fx.Lazy` for expensive dependencies
2. **Singleton Pattern**: Most dependencies should be singletons
3. **Resource Management**: Use lifecycle hooks for cleanup
4. **Memory Efficiency**: Monitor memory usage with large dependency graphs

## Migration Guide

### From Centralized to Isolated DI

If you have an existing application using the old centralized DI:

1. **Create Local DI Directory**
   ```bash
   mkdir examples/your-app/di
   ```

2. **Copy and Customize DI Configuration**
   ```go
   // Start with mobile-onboarding-v2 or mobile-bff as template
   cp examples/mobile-onboarding-v2/di/di.go examples/your-app/di/
   ```

3. **Update Imports**
   ```go
   // Change from:
   import "github.com/venkatvghub/api-orchestration-framework/pkg/di"
   
   // To:
   import "github.com/venkatvghub/api-orchestration-framework/examples/your-app/di"
   ```

4. **Customize Providers**
   - Update metrics namespace
   - Customize transformers for your use case
   - Add application-specific validators
   - Configure HTTP client for your backend APIs

5. **Test Thoroughly**
   - Ensure all dependencies resolve correctly
   - Test application startup and shutdown
   - Verify functionality with new DI configuration

## Examples

### Complete Mobile Onboarding DI
See: `examples/mobile-onboarding-v2/di/di.go`

### Complete Mobile BFF DI  
See: `examples/mobile-bff/di/di.go`

### Creating Custom Application DI
See the Quick Start section above for step-by-step instructions.

## Related Documentation

- [Configuration Guide](config.md) - Framework configuration management
- [Flow Documentation](flow.md) - Flow orchestration patterns
- [Steps Reference](steps.md) - Available step types
- [Transformers Guide](transformers.md) - Data transformation
- [Validators Guide](validators.md) - Data validation
- [Error Handling](errors.md) - Structured error handling
- [Metrics & Monitoring](metrics.md) - Observability
- [Getting Started](getting-started.md) - Quick start guide 