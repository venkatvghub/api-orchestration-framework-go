
# Stateless Declarative API Orchestration Engine in Go

## 1. 🌍 Overview

This document outlines the architecture and rationale for a **stateless, declarative API orchestration engine built in Go**, designed to deliver high-performance, extensible Backend for Frontend (BFF) workflows. It replaces the need for Apache Camel-style implementations by leveraging Go's native capabilities in concurrency, type safety, and observability, while providing domain-specific flow orchestration primitives that are optimized for stateless execution.

---

## 2. 🎯 Project Requirements

- Must support **stateless orchestration** of multiple backend services
- Must allow **conditional, parallel, and retry-based flows**
- Must be **lightweight**, **embeddable**, and easy to deploy
- Must offer **high observability** (metrics, logging, tracing)
- Should support **mobile-friendly BFF transformations**
- Should avoid JVM overhead and runtime dependency complexity
- Should allow **custom extension points** (steps, validators, transformers)

---

## 3. ❓ Why Not Apache Camel (Directly or via Golang Ports)?

### ❌ Why Apache Camel (Java) Is Not Suitable
- Built for JVM: Requires Java runtime, increasing footprint
- Verbose: XML or Java DSL is cumbersome for simple flows
- Stateful by design: Routes often assume long-lived processes
- Overkill: Most use cases here require only stateless aggregation logic

### ⚠️ Why Camel-Style Go Frameworks Are Not Ideal
| Feature                     | Camel-Style Go                    | This Framework (Flow DSL)         |
|----------------------------|-----------------------------------|-----------------------------------|
| Stateful Routing Model     | Often mimics message brokers      | Fully stateless per HTTP request  |
| Configuration Model        | Tends to rely on central router   | Pure function-chaining DSL        |
| Route Definitions          | Route mapping with switches       | Declarative Step-based logic      |
| Observability              | Minimal by default                | Built-in tracing, logging, metrics|
| Concurrency Support        | Manual goroutines                 | Native parallel step support      |

Camel-like patterns work well for **integration platforms**, but for **API aggregation**, **data transformation**, and **BFF use cases**, they introduce unnecessary complexity.

---

## 4. ✅ Benefits of This Framework

- 🔧 **Pure Go, zero dependency runtime**
- 📦 **Declarative flow definition** via fluent builders
- 🔍 **Type-safe context passing**, nested data access
- ♻️ **Parallel, conditional, retry logic** built-in
- 🚦 **Prometheus metrics**, Zap logs, and OpenTelemetry tracing
- 📱 **Transformer chains** for mobile BFF responses
- 🔐 Built-in **validation and sanitization**
- 🧪 Easy unit testing of individual steps and flows

---

## 5. 📊 Comparison With Other Frameworks (Go & Non-Go)

### 5.1 Non-Go Frameworks

| Framework                | Language | Key Strengths                                              | Limitations Compared to This Framework                |
|--------------------------|----------|------------------------------------------------------------|--------------------------------------------------------|
| **Node.js BFF (e.g. BFF.js)** | JavaScript | Rapid prototyping, rich ecosystem                          | Weak type safety, harder to test, verbose observability |
| **Spring Cloud Gateway**| Java     | Enterprise routing, resilience                            | Heavyweight JVM, annotation config overhead             |
| **Temporal.io**         | Go       | Durable workflows, stateful retries                        | Complex setup for stateless API flows                  |
| **AWS Step Functions**  | Cloud    | Serverless orchestration                                   | Costly, externalized control, poor inline observability |

### 5.2 Go-Specific Frameworks

| Framework                   | Key Strengths                                                    | Limitations Compared to This Framework                            |
|-----------------------------|------------------------------------------------------------------|--------------------------------------------------------------------|
| **GoFlow**                 | Visual flow modeling, DSL for task coordination                  | Designed for general-purpose workflow, not optimized for APIs     |
| **Conductor-go (Netflix)** | Community port of Conductor for task coordination                | Heavy dependencies, not idiomatic Go DSL                          |
| **Koanf**                  | Configuration management and dynamic flow switching              | Not a flow engine, lacks orchestration primitives                 |
| **goflow (by Travis Jeffery)** | Simple DAG-like flow processing in Go                         | Lacks robust error handling, observability, BFF readiness         |
| **Cuelang**                | Declarative configuration & validation with scripting support    | Great for validation, but not ideal for dynamic runtime execution |

This framework fills the gap by combining:
- Stateless, declarative API orchestration
- Native support for BFF/mobile scenarios
- Fluent DSL with observability and transform layers
- Plug-and-play steps, validators, and transformers

...

## 11. 🚀 Final Summary

This framework offers a compelling alternative to traditional routing engines and Camel-based paradigms by:
- Embracing Go’s native strengths
- Enabling expressive, declarative orchestration
- Delivering first-class BFF/mobile API features
- Empowering developers to iterate, test, and deploy faster

It is best suited for high-throughput, stateless orchestration of APIs where you need control, visibility, and performance in a single package.
