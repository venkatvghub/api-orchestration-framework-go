# Metrics Package Documentation

## Overview

The `pkg/metrics` package provides comprehensive observability and monitoring capabilities for the API Orchestration Framework. It offers Prometheus-based metrics collection with support for counters, gauges, and histograms, along with automatic framework integration metrics.

## Purpose

The metrics package serves as the observability foundation that:
- Provides standardized Prometheus metrics collection across all framework components
- Offers real-time performance monitoring and alerting
- Enables custom business metrics and KPIs
- Integrates with Prometheus monitoring ecosystem
- Supports metrics-driven monitoring and alerting
- Provides automatic Go runtime and process metrics

## Core Architecture

### MetricsCollector Interface

The central interface for metrics collection:
```go
type MetricsCollector interface {
    // Counters
    IncrementCounter(name string, tags map[string]string)
    IncrementCounterBy(name string, value int64, tags map[string]string)

    // Gauges
    SetGauge(name string, value float64, tags map[string]string)
    IncrementGauge(name string, tags map[string]string)
    DecrementGauge(name string, tags map[string]string)

    // Histograms
    RecordDuration(name string, duration time.Duration, tags map[string]string)
    RecordValue(name string, value float64, tags map[string]string)

    // Framework metrics
    RecordFlowExecution(flowName string, duration time.Duration, success bool)
    RecordStepExecution(stepName string, duration time.Duration, success bool)
    RecordHTTPRequest(method, url string, statusCode int, duration time.Duration)
    RecordCacheOperation(operation string, hit bool, duration time.Duration)
    RecordTransformation(transformer string, duration time.Duration, success bool)
    RecordValidation(validator string, duration time.Duration, success bool)
}
```

### Metric Types

#### Counter
Monotonically increasing counter for counting events:
```go
// Usage examples
metrics.IncrementCounter("http_requests_total", map[string]string{
    "method": "GET",
    "status_code": "200",
})

metrics.IncrementCounterBy("bytes_processed_total", 1024, map[string]string{
    "service": "user-api",
})
```

#### Gauge
Current value metric that can go up and down:
```go
// Usage examples
metrics.SetGauge("active_connections", 150, map[string]string{
    "service": "user-api",
})

metrics.IncrementGauge("queue_size", map[string]string{
    "queue": "notifications",
})

metrics.DecrementGauge("active_workers", map[string]string{
    "worker_type": "background",
})
```

#### Histogram
Distribution of values with configurable buckets:
```go
// Usage examples
metrics.RecordDuration("http_request_duration_seconds", duration, map[string]string{
    "method": "POST",
    "endpoint": "/api/users",
})

metrics.RecordValue("request_size_bytes", 1024, map[string]string{
    "endpoint": "/api/upload",
})
```

## Framework Integration Metrics

### Flow Metrics
Automatic metrics collection for flow execution:
```go
// Automatically collected flow metrics:
// - api_orchestration_framework_flow_executions_total{flow_name, success}
// - api_orchestration_framework_flow_execution_duration_seconds{flow_name, success}

// Usage in framework (automatic)
metrics.RecordFlowExecution("UserRegistration", duration, true)
```

### Step Metrics
Comprehensive step execution metrics:
```go
// Automatically collected step metrics:
// - api_orchestration_framework_step_executions_total{step_name, success}
// - api_orchestration_framework_step_execution_duration_seconds{step_name, success}

// Usage in framework (automatic)
metrics.RecordStepExecution("validateUser", duration, true)
```

### HTTP Request Metrics
HTTP request/response metrics:
```go
// Automatically collected HTTP metrics:
// - api_orchestration_framework_http_requests_total{method, url, status_code}
// - api_orchestration_framework_http_request_duration_seconds{method, url, status_code}

// Usage in framework (automatic)
metrics.RecordHTTPRequest("GET", "/api/users/123", 200, duration)
```

### Cache Metrics
Cache performance and hit rate metrics:
```go
// Automatically collected cache metrics:
// - api_orchestration_framework_cache_operations_total{operation, hit}
// - api_orchestration_framework_cache_operation_duration_seconds{operation, hit}

// Usage in framework (automatic)
metrics.RecordCacheOperation("get", true, duration)
```

### Transformation Metrics
Data transformation metrics:
```go
// Automatically collected transformation metrics:
// - api_orchestration_framework_transformations_total{transformer, success}
// - api_orchestration_framework_transformation_duration_seconds{transformer, success}

// Usage in framework (automatic)
metrics.RecordTransformation("userDataTransformer", duration, true)
```

### Validation Metrics
Validation performance and error metrics:
```go
// Automatically collected validation metrics:
// - api_orchestration_framework_validations_total{validator, success}
// - api_orchestration_framework_validation_duration_seconds{validator, success}

// Usage in framework (automatic)
metrics.RecordValidation("emailValidator", duration, true)
```

## Prometheus Configuration

### Basic Setup
```go
import "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"

// Use default Prometheus metrics (recommended)
metrics.SetupDefaultPrometheusMetrics()

// Or create with custom configuration
config := &metrics.PrometheusConfig{
    Namespace: "my_app",
    Subsystem: "api",
    Registry:  prometheus.NewRegistry(), // Optional custom registry
}
metrics.SetupPrometheusMetricsWithConfig(config)
```

### HTTP Metrics Endpoint
```go
import (
    "net/http"
    "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

// Expose metrics endpoint
http.Handle("/metrics", metrics.MetricsHandler())

// Or with custom registry
http.Handle("/metrics", metrics.MetricsHandlerWithRegistry(customRegistry))

// Start metrics server
go http.ListenAndServe(":9090", nil)
```

### Custom Metrics Implementation
```go
// Create custom Prometheus metrics
prometheusMetrics := metrics.NewPrometheusMetrics("my_app", "service")

// Set as global metrics collector
metrics.SetGlobalMetrics(prometheusMetrics)

// Use global convenience functions
metrics.IncrementCounter("custom_events_total", map[string]string{
    "event_type": "user_signup",
})

metrics.RecordDuration("operation_duration_seconds", duration, map[string]string{
    "operation": "data_processing",
})
```

## Custom Business Metrics

### Business KPIs
Track business-specific metrics:
```go
// User registration metrics
metrics.IncrementCounter("user_registrations_total", map[string]string{
    "source": "web",
    "plan":   "premium",
})

// Revenue metrics
metrics.SetGauge("revenue_total", 50000.0, map[string]string{
    "currency": "USD",
    "period":   "monthly",
})

// Conversion metrics
metrics.SetGauge("conversion_rate", 0.15, map[string]string{
    "funnel_step": "checkout",
    "variant":     "A",
})
```

### Feature Usage Metrics
Track feature adoption and usage:
```go
// Feature usage tracking
metrics.IncrementCounter("feature_usage_total", map[string]string{
    "feature": "mobile_app",
    "action":  "profile_view",
})

// A/B testing metrics
metrics.IncrementCounter("experiment_events_total", map[string]string{
    "experiment": "checkout_flow_v2",
    "variant":    "treatment",
    "event":      "conversion",
})
```

## Metric Names and Labels

### Standard Metric Names
The package provides constants for consistent metric naming:
```go
const (
    // Flow metrics
    FlowExecutionsTotal   = "flow_executions_total"
    FlowExecutionDuration = "flow_execution_duration_seconds"
    StepExecutionsTotal   = "step_executions_total"
    StepExecutionDuration = "step_execution_duration_seconds"

    // HTTP metrics
    HTTPRequestsTotal   = "http_requests_total"
    HTTPRequestDuration = "http_request_duration_seconds"

    // Cache metrics
    CacheOperationsTotal   = "cache_operations_total"
    CacheOperationDuration = "cache_operation_duration_seconds"

    // Transformation metrics
    TransformationsTotal   = "transformations_total"
    TransformationDuration = "transformation_duration_seconds"

    // Validation metrics
    ValidationsTotal   = "validations_total"
    ValidationDuration = "validation_duration_seconds"
)
```

### Standard Label Names
```go
const (
    LabelFlowName    = "flow_name"
    LabelStepName    = "step_name"
    LabelSuccess     = "success"
    LabelMethod      = "method"
    LabelURL         = "url"
    LabelStatusCode  = "status_code"
    LabelOperation   = "operation"
    LabelHit         = "hit"
    LabelTransformer = "transformer"
    LabelValidator   = "validator"
)
```

## Alternative Implementations

### In-Memory Metrics (Testing)
For testing or development environments:
```go
// Create in-memory metrics collector
inMemoryMetrics := metrics.NewInMemoryMetrics()
metrics.SetGlobalMetrics(inMemoryMetrics)

// Access collected metrics
counters := inMemoryMetrics.GetCounters()
gauges := inMemoryMetrics.GetGauges()
histograms := inMemoryMetrics.GetHistograms()

// Reset metrics
inMemoryMetrics.Reset()
```

### No-Op Metrics (Disabled)
For environments where metrics are disabled:
```go
// Create no-op metrics collector
noOpMetrics := metrics.NewNoOpMetrics()
metrics.SetGlobalMetrics(noOpMetrics)
```

## Best Practices

### Metric Naming
1. **Use Consistent Naming**: Follow Prometheus naming conventions
2. **Include Units**: Include units in metric names (e.g., `_seconds`, `_bytes`, `_total`)
3. **Use Descriptive Names**: Make metric names self-explanatory
4. **Avoid High Cardinality**: Be careful with label values that can have many unique values
5. **Namespace Metrics**: Use namespace and subsystem for organization

### Performance Considerations
1. **Label Cardinality**: Keep the number of unique label combinations reasonable
2. **Metric Creation**: Metrics are created dynamically but cached for reuse
3. **Memory Usage**: Monitor memory usage with high-cardinality metrics
4. **Collection Frequency**: Balance collection frequency with performance impact

### Monitoring Strategy
1. **Golden Signals**: Focus on latency, traffic, errors, and saturation
2. **Business Metrics**: Include business-relevant metrics alongside technical ones
3. **Alerting**: Use Prometheus AlertManager for alerting rules
4. **Dashboards**: Create Grafana dashboards for visualization
5. **Historical Analysis**: Leverage Prometheus's time-series capabilities

## Examples

### Complete Setup
```go
package main

import (
    "net/http"
    "time"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

func main() {
    // Setup Prometheus metrics
    metrics.SetupDefaultPrometheusMetrics()
    
    // Expose metrics endpoint
    http.Handle("/metrics", metrics.MetricsHandler())
    
    // Start metrics server
    go http.ListenAndServe(":9090", nil)
    
    // Example usage
    for {
        // Record some metrics
        metrics.IncrementCounter("app_requests_total", map[string]string{
            "method": "GET",
            "status": "200",
        })
        
        start := time.Now()
        // Simulate work
        time.Sleep(100 * time.Millisecond)
        
        metrics.RecordDuration("app_request_duration_seconds", time.Since(start), map[string]string{
            "endpoint": "/api/health",
        })
        
        time.Sleep(1 * time.Second)
    }
}
```

### Custom Metrics Wrapper
```go
type AppMetrics struct {
    collector metrics.MetricsCollector
}

func NewAppMetrics() *AppMetrics {
    return &AppMetrics{
        collector: metrics.GetGlobalMetrics(),
    }
}

func (am *AppMetrics) RecordUserAction(action string, success bool, duration time.Duration) {
    tags := map[string]string{
        "action":  action,
        "success": boolToString(success),
    }
    
    am.collector.IncrementCounter("user_actions_total", tags)
    am.collector.RecordDuration("user_action_duration_seconds", duration, tags)
}

func (am *AppMetrics) UpdateActiveUsers(count int) {
    am.collector.SetGauge("active_users", float64(count), nil)
}

func boolToString(b bool) string {
    if b {
        return "true"
    }
    return "false"
}
```

## Troubleshooting

### Common Issues

1. **High Cardinality Metrics**
   ```go
   // Wrong: Using user ID as label (high cardinality)
   metrics.IncrementCounter("requests_total", map[string]string{
       "user_id": userID, // This creates too many unique metrics
   })
   
   // Correct: Use user type or segment instead
   metrics.IncrementCounter("requests_total", map[string]string{
       "user_type": getUserType(userID),
   })
   ```

2. **Memory Usage Monitoring**
   ```go
   // Monitor metrics memory usage
   go func() {
       ticker := time.NewTicker(1 * time.Minute)
       for range ticker.C {
           var m runtime.MemStats
           runtime.ReadMemStats(&m)
           
           metrics.SetGauge("metrics_memory_usage_bytes", float64(m.Alloc), nil)
           
           if m.Alloc > 100*1024*1024 { // 100MB threshold
               log.Warn("High metrics memory usage", zap.Uint64("bytes", m.Alloc))
           }
       }
   }()
   ```

3. **Metrics Endpoint Health**
   ```go
   // Check if metrics endpoint is responding
   func checkMetricsHealth() error {
       resp, err := http.Get("http://localhost:9090/metrics")
       if err != nil {
           return err
       }
       defer resp.Body.Close()
       
       if resp.StatusCode != http.StatusOK {
           return fmt.Errorf("metrics endpoint returned status %d", resp.StatusCode)
       }
       
       return nil
   }
   ```

### Performance Debugging

Monitor metrics system performance:
```go
func monitorMetricsPerformance() {
    // Track metrics collection overhead
    start := time.Now()
    metrics.IncrementCounter("test_metric", nil)
    overhead := time.Since(start)
    
    metrics.RecordDuration("metrics_collection_overhead_seconds", overhead, nil)
    
    // Monitor metric creation rate
    metrics.IncrementCounter("metrics_created_total", map[string]string{
        "type": "counter",
    })
}
``` 