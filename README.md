# Go API Orchestration Framework

> A powerful, declarative API orchestration engine built in Go, designed for high-performance Backend for Frontend (BFF) implementations with mobile-first use cases.

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen.svg)](docs/)

## 🚀 Quick Start (2 minutes)

```bash
go get github.com/venkatvghub/api-orchestration-framework
```

```go
package main

import (
    "log"
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
)

func main() {
    // Create a simple flow
    userFlow := flow.NewFlow("GetUser").
        Step("fetchUser", flow.NewStepWrapper(
            http.GET("https://jsonplaceholder.typicode.com/users/1").
                SaveAs("user")))

    // Execute and get results
    ctx := flow.NewContext()
    result, err := userFlow.Execute(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("User: %+v", result.GetResponse())
}
```

## 🎯 Why Use This Framework?

| Problem | Solution |
|---------|----------|
| **Multiple API calls from mobile** | Single BFF endpoint with parallel data fetching |
| **Complex data transformation** | Declarative transformation pipeline |
| **Poor error handling** | Built-in retry, circuit breakers, and fallbacks |
| **No observability** | Automatic Prometheus metrics and structured logging |
| **Slow mobile responses** | Field selection, caching, and compression |

## 🏗️ Core Concepts

### 1. Flows - Your Orchestration Pipeline
```go
// Sequential execution
flow.NewFlow("UserProfile").
    Step("auth", authStep).
    Step("fetch", fetchStep).
    Step("transform", transformStep)

// Parallel execution
flow.Parallel("userData").
    Step("profile", profileStep).
    Step("preferences", preferencesStep).
    Step("notifications", notificationsStep).
EndParallel()

// Conditional logic
flow.Choice("userType").
    When(isPremium).Step("premium", premiumStep).
    Otherwise().Step("basic", basicStep).
EndChoice()
```

### 2. Steps - Reusable Building Blocks
```go
// HTTP requests
http.GET("/api/users/${userId}").SaveAs("user")
http.POST("/api/orders").WithJSONBody(orderData)

// Authentication
core.NewTokenValidationStep("auth", "Authorization")

// Caching
core.NewCacheSetStep("cache", "user_${userId}", "userData", 5*time.Minute)

// Transformations
core.NewTransformStep("mobile", mobileTransformer)
```

### 3. Context - Type-Safe Data Flow
```go
ctx := flow.NewContext()
ctx.Set("userId", "123")
ctx.Set("user", userData)

// Type-safe access
userId, _ := ctx.GetString("userId")
userAge, _ := ctx.GetInt("user.profile.age")
isActive, _ := ctx.GetBool("user.active")
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Mobile App    │───▶│   BFF Layer     │───▶│  Microservices  │
│                 │    │  (This Framework)│    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  Orchestration  │
                       │     Engine      │
                       │                 │
                       │ • Flow Engine   │
                       │ • Step Library  │
                       │ • Transformers  │
                       │ • Validators    │
                       │ • Metrics       │
                       └─────────────────┘
```

### Package Structure
```
pkg/
├── flow/           # Core orchestration engine
├── steps/          # Pre-built step implementations
│   ├── base/       # Base step types
│   ├── core/       # Core utility steps
│   ├── http/       # HTTP client steps
│   └── bff/        # Mobile BFF steps
├── transformers/   # Data transformation
├── validators/     # Data validation
├── metrics/        # Prometheus metrics
├── config/         # Configuration management
├── errors/         # Structured error handling
└── interfaces/     # Core interfaces
```

## 📱 Mobile BFF Example

```go
func CreateMobileDashboard() *flow.Flow {
    return flow.NewFlow("MobileDashboard").
        WithTimeout(10 * time.Second).
        
        // 1. Authenticate user
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"))).
        
        // 2. Fetch data in parallel (fast!)
        Parallel("screenData").
            Step("profile", flow.NewStepWrapper(
                http.NewMobileAPIStep("profile", "GET", "/api/profile",
                    []string{"id", "name", "avatar"}). // Only get what you need
                    WithCaching("profile", 5*time.Minute))).
            Step("notifications", flow.NewStepWrapper(
                http.GET("/api/notifications").
                    WithFallback([]interface{}{}))).  // Graceful degradation
            Step("feed", flow.NewStepWrapper(
                http.GET("/api/feed?limit=10"))).
        EndParallel().
        
        // 3. Transform for mobile
        Step("optimize", flow.NewStepWrapper(
            core.NewTransformStep("mobile",
                transformers.NewMobileTransformer([]string{"id", "name", "avatar"}))))
}
```

## 🛠️ Installation & Setup

### 1. Install the Framework
```bash
go mod init your-project
go get github.com/venkatvghub/api-orchestration-framework
```

### 2. Basic Project Structure
```
your-project/
├── main.go
├── flows/
│   ├── user.go
│   └── mobile.go
├── steps/
│   └── custom.go
└── config/
    └── config.go
```

### 3. Add Monitoring (Optional)
```go
import "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"

// Setup Prometheus metrics
metrics.SetupDefaultPrometheusMetrics()
http.Handle("/metrics", metrics.MetricsHandler())
go http.ListenAndServe(":9090", nil)
```

## 🔧 Configuration

### Environment Variables
```bash
# HTTP settings
export HTTP_MAX_RETRIES=3
export HTTP_REQUEST_TIMEOUT=15s
export HTTP_ENABLE_FALLBACK=true

# Cache settings
export CACHE_DEFAULT_TTL=5m
export CACHE_MAX_SIZE=1000

# Logging
export LOG_LEVEL=info
export LOG_FORMAT=json
```

### Programmatic Configuration
```go
config := &config.FrameworkConfig{
    HTTP: config.HTTPConfig{
        MaxRetries:     3,
        RequestTimeout: 15 * time.Second,
        EnableFallback: true,
    },
    Cache: config.CacheConfig{
        DefaultTTL: 10 * time.Minute,
        MaxSize:    2000,
    },
}

ctx := flow.NewContextWithConfig(config)
```

## 📊 Built-in Features

### ✅ Observability
- **Prometheus Metrics**: Automatic collection of flow, step, and HTTP metrics
- **Structured Logging**: JSON logging with correlation IDs
- **Distributed Tracing**: OpenTelemetry integration ready

### ✅ Resilience
- **Retries**: Configurable retry logic with exponential backoff
- **Circuit Breakers**: Automatic failure detection and recovery
- **Timeouts**: Granular timeout control at flow and step levels
- **Fallbacks**: Graceful degradation with default responses

### ✅ Performance
- **Parallel Execution**: Concurrent step execution
- **Connection Pooling**: Efficient HTTP client management
- **Caching**: Built-in TTL-based caching
- **Mobile Optimization**: Field selection and response compression

### ✅ Developer Experience
- **Type Safety**: Strongly typed context operations
- **Fluent API**: Readable, declarative flow definitions
- **Hot Reload**: Configuration changes without restarts
- **Rich Errors**: Structured error handling with context

## 📚 Documentation

### 🚀 Getting Started
- [**5-Minute Tutorial**](docs/getting-started.md) - Get up and running quickly
- [**BFF Patterns**](docs/bff-patterns.md) - Mobile-specific implementation patterns
- [**Configuration**](docs/config.md) - Environment and programmatic configuration

### 🔧 Core Components
- [**Flow Engine**](docs/flow.md) - Orchestration patterns and advanced features
- [**Steps Library**](docs/steps.md) - All available steps and custom step creation
- [**Transformers**](docs/transformers.md) - Data transformation and mobile optimization
- [**Validators**](docs/validators.md) - Input validation and error handling

### 🏗️ Architecture & Extensibility
- [**Interfaces**](docs/interfaces.md) - Core interfaces and plugin development
- [**Error Handling**](docs/errors.md) - Structured errors and recovery patterns
- [**Metrics & Monitoring**](docs/metrics.md) - Observability and alerting
- [**Service Registry**](docs/registry.md) - Service discovery and load balancing

### 📖 Reference
- [**Examples**](examples/) - Complete working examples
- [**Utilities**](docs/utils.md) - Helper functions and common patterns
- [**API Reference**](https://pkg.go.dev/github.com/venkatvghub/api-orchestration-framework) - Complete API documentation

## 🎨 Common Patterns

<details>
<summary><strong>🔐 Authentication Flow</strong></summary>

```go
authFlow := flow.NewFlow("Authentication").
    Step("validate", flow.NewStepWrapper(
        core.NewTokenValidationStep("auth", "Authorization"))).
    Choice("tokenStatus").
        When(func(ctx *flow.Context) bool {
            return ctx.Has("token_error")
        }).
            Step("refresh", flow.NewStepWrapper(
                http.POST("/auth/refresh").
                    WithJSONBody(map[string]interface{}{
                        "refresh_token": "${refreshToken}",
                    }).SaveAs("newTokens"))).
        Otherwise().
            Step("proceed", flow.NewStepWrapper(
                core.NewLogStep("valid", "info", "Token is valid"))).
    EndChoice()
```
</details>

<details>
<summary><strong>📱 Mobile Screen Aggregation</strong></summary>

```go
screenFlow := flow.NewFlow("HomeScreen").
    Step("auth", authStep).
    Parallel("screenData").
        Step("user", userStep).
        Step("notifications", notificationsStep).
        Step("feed", feedStep).
    EndParallel().
    Step("transform", mobileTransformStep)
```
</details>

<details>
<summary><strong>🔄 Retry with Fallback</strong></summary>

```go
resilientFlow := flow.NewFlow("ResilientAPI").
    Step("primary", flow.NewRetryStep("primaryAPI",
        http.GET("/api/primary").SaveAs("data"),
        3, 2*time.Second)).
    Choice("dataAvailable").
        When(func(ctx *flow.Context) bool {
            return ctx.Has("data")
        }).
            Step("success", successStep).
        Otherwise().
            Step("fallback", flow.NewStepWrapper(
                http.GET("/api/fallback").SaveAs("data"))).
    EndChoice()
```
</details>

## 🚦 Performance Benchmarks

| Operation | Latency (p95) | Throughput |
|-----------|---------------|------------|
| Simple Flow (3 steps) | 2ms | 50,000 req/s |
| Parallel Flow (5 steps) | 5ms | 30,000 req/s |
| HTTP Step | 1ms overhead | 45,000 req/s |
| Transform Step | 0.5ms | 100,000 req/s |

*Benchmarks run on MacBook Pro M1, Go 1.21*

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes and add tests
4. **Run** tests: `go test ./...`
5. **Commit** your changes: `git commit -m 'Add amazing feature'`
6. **Push** to the branch: `git push origin feature/amazing-feature`
7. **Open** a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**[⭐ Star this repo](https://github.com/venkatvghub/api-orchestration-framework)** if you find it useful!

[Documentation](docs/) • [Examples](examples/) • [Issues](https://github.com/venkatvghub/api-orchestration-framework/issues) • [Discussions](https://github.com/venkatvghub/api-orchestration-framework/discussions)

</div> 