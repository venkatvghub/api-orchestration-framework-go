
# Stateless Declarative API Orchestration Engine in Go

## 1. 🌍 Overview

This document outlines the architecture for a **stateless, declarative API orchestration engine** built in **Go**, inspired by Apache Camel but implemented purely in code using a fluent, type-safe DSL.

---

## 2. 🌟 Objectives

- Declarative composition of API flows entirely in Go
- Stateless execution with no persistence between invocations
- Type-safe context for inter-step data passing
- Sequential, parallel, and conditional execution support
- Observability: logging, metrics, tracing
- Pluggable and composable steps

---

## 3. 📊 High-Level Architecture

### Components:

| Component     | Responsibility |
|---------------|----------------|
| `Flow`        | Orchestration unit holding sequential steps |
| `Step`        | Interface for all operations within a flow |
| `Context`     | Shared, typed state between steps |
| `Runner`      | Execution engine for a `Flow` with error propagation |
| `Logger`      | Structured step-level logging |
| `Metrics`     | Prometheus-based step-level and flow-level metrics |

---

## 4. 🌐 Architecture Diagrams

### 4.1 Component Interaction Diagram
```
+---------+     Step1     +---------+     Step2     +----------+
| Runner  +-------------> | Step A  +-------------> | Step B   |
+----+----+              +----+----+              +-----+----+
     |                         |                         |
     |    Parallel Step        |     Conditional Branch  |
     v                         v                         v
+----+----+               +----+----+               +-----+-----+
| Logger  |               | Metric  |               | Context   |
+---------+               +---------+               +-----------+
```

### 4.2 Execution Flow
```
[Runner]
   |
   |-- executes --> [Flow]
                         |-- contains --> [Steps]
                         |-- uses --> [Context]
                         |-- emits --> [Logs, Metrics, Traces]
```

---

## 5. 🛠️ Low-Level Architecture

### 5.1 Step Interface
```go
type Step interface {
    Run(ctx *Context) error
}
```

### 5.2 Flow DSL and Execution
```go
type Flow struct {
    steps []Step
}

func NewFlow() *Flow {
    return &Flow{}
}

func (f *Flow) Step(name string, step Step) *Flow {
    f.steps = append(f.steps, step)
    return f
}

func (f *Flow) Run(ctx *Context) error {
    for _, step := range f.steps {
        if err := step.Run(ctx); err != nil {
            return err
        }
    }
    return nil
}
```

### 5.3 Context
```go
type Context struct {
    values map[string]interface{}
    mu     sync.RWMutex
}

func (c *Context) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.values[key] = value
}

func (c *Context) Get[T any](key string) (T, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.values[key]
    if !ok {
        var zero T
        return zero, fmt.Errorf("missing key: %s", key)
    }
    casted, ok := val.(T)
    if !ok {
        var zero T
        return zero, fmt.Errorf("type mismatch for key: %s", key)
    }
    return casted, nil
}
```

---

## 6. 📊 Observability

### Logging
- Step-level logs
- Use `zap` or `zerolog`

### Metrics
- Prometheus counters and histograms
- Instrumented per step and per flow

### Tracing
- OpenTelemetry-compatible
- Propagate trace IDs via context

---

## 7. ⚡ Performance Considerations

- `sync.WaitGroup` for parallel step execution
- Stateless context usage
- Step timeout via `context.WithTimeout`
- Shared HTTP clients

---

## 8. 🔐 Security

- Interpolation input validation
- Masking sensitive outputs in logs
- Header/token redaction

---

## 9. 🔮 Testing Strategy

- Unit tests for step and flow execution
- Integration tests with sample APIs
- Mockable interfaces for test control
- Benchmarks for concurrent execution

---

## 10. 🚀 Future Extensions

- `Retry()`, `Timeout()`, `OnErrorResume()`, `Delay()` DSL methods
- Middleware-based observability wrappers
- Flow-to-Graphviz export for UI visualization
- Persistence layer for rehydrated workflows
- CRON and Kafka-based triggering
- Validation and policy hook systems

---

## 11. 🎯 Sample DSL

```go
flow := NewFlow().
  Step("fetchUser", GET("https://api.example.com/users/123").SaveAs("user")).
  Choice("checkUserType").
    When(func(ctx *Context) bool {
      return ctx.Get("user.type") == "admin"
    }).
      Step("fetchAdmin", GET("https://api.example.com/admin/${user.id}/stats").SaveAs("adminStats")).
    Otherwise().
      Step("fetchUserStats", GET("https://api.example.com/users/${user.id}/stats").SaveAs("userStats")).
  EndChoice().
  Parallel("extras").
    Step("notifs", GET("https://api.example.com/notifs").SaveAs("notifs")).
    Step("tasks", GET("https://api.example.com/tasks").SaveAs("tasks")).
  EndParallel()
```

---

## 12. 🛒 API Model

### POST /execute-flow
Executes a named flow programmatically

**Request**:
```json
{
  "flow": "UserStatsFlow",
  "params": {
    "userId": "123",
    "authToken": "abc123"
  }
}
```

**Response**:
```json
{
  "status": "success",
  "context": {
    "user": { "id": 123, "type": "admin" },
    "adminStats": { ... },
    "notifs": [...],
    "tasks": [...]
  }
}
```

---

## 13. 🚀 Summary

This architecture provides a robust, extensible, and observable framework for declarative API orchestration written entirely in Go. It leverages Go’s type system, concurrency model, and functional programming primitives to deliver a lightweight alternative to XML/YAML-based orchestration engines like Apache Camel.

With future extensions, this can power use cases across microservices, backend workflows, integration hubs, and developer tools.
