# Config Package Documentation

## Overview

The `pkg/config` package provides centralized configuration management for the API Orchestration Framework. It offers a comprehensive configuration system with environment variable support, default values, validation, and type-safe access patterns for all framework components.

## Purpose

The config package serves as the configuration backbone that:
- Centralizes all framework configuration in a single location
- Provides environment variable-based configuration with sensible defaults
- Offers type-safe configuration access across all components
- Supports configuration validation and error handling
- Enables easy configuration customization for different environments
- Provides configuration for HTTP clients, caching, logging, security, and mobile optimizations

## Core Architecture

### FrameworkConfig Structure

The main configuration structure that encompasses all framework settings:
```go
type FrameworkConfig struct {
    HTTP     HTTPConfig     `json:"http"`
    Cache    CacheConfig    `json:"cache"`
    Logging  LoggingConfig  `json:"logging"`
    Security SecurityConfig `json:"security"`
    Timeouts TimeoutConfig  `json:"timeouts"`
    Mobile   MobileConfig   `json:"mobile"`
}
```

## Configuration Categories

### HTTP Configuration (`HTTPConfig`)

Comprehensive HTTP client configuration for resilience and performance:
```go
type HTTPConfig struct {
    MaxIdleConns        int           `json:"max_idle_conns"`
    MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host"`
    IdleConnTimeout     time.Duration `json:"idle_conn_timeout"`
    MaxRetries          int           `json:"max_retries"`
    RetryDelay          time.Duration `json:"retry_delay"`
    MaxRetryDelay       time.Duration `json:"max_retry_delay"`
    FailureThreshold    uint          `json:"failure_threshold"`
    SuccessThreshold    uint          `json:"success_threshold"`
    CircuitBreakerDelay time.Duration `json:"circuit_breaker_delay"`
    RequestTimeout      time.Duration `json:"request_timeout"`
    EnableFallback      bool          `json:"enable_fallback"`
    UserAgent           string        `json:"user_agent"`
}
```

#### Environment Variables:
- `HTTP_MAX_IDLE_CONNS` (default: 100)
- `HTTP_MAX_IDLE_CONNS_PER_HOST` (default: 10)
- `HTTP_IDLE_CONN_TIMEOUT` (default: 90s)
- `HTTP_MAX_RETRIES` (default: 3)
- `HTTP_RETRY_DELAY` (default: 1s)
- `HTTP_MAX_RETRY_DELAY` (default: 10s)
- `HTTP_FAILURE_THRESHOLD` (default: 5)
- `HTTP_SUCCESS_THRESHOLD` (default: 3)
- `HTTP_CIRCUIT_BREAKER_DELAY` (default: 5s)
- `HTTP_REQUEST_TIMEOUT` (default: 15s)
- `HTTP_ENABLE_FALLBACK` (default: true)
- `HTTP_USER_AGENT` (default: "API-Orchestration-Framework/2.0")

#### Usage:
```go
config := config.DefaultConfig()
httpClient := http.NewResilientHTTPClient(&http.ClientConfig{
    MaxRetries:           config.HTTP.MaxRetries,
    InitialRetryDelay:    config.HTTP.RetryDelay,
    MaxRetryDelay:        config.HTTP.MaxRetryDelay,
    FailureThreshold:     config.HTTP.FailureThreshold,
    CircuitBreakerDelay:  config.HTTP.CircuitBreakerDelay,
    RequestTimeout:       config.HTTP.RequestTimeout,
    EnableFallback:       config.HTTP.EnableFallback,
})
```

### Cache Configuration (`CacheConfig`)

Configuration for caching behavior and performance:
```go
type CacheConfig struct {
    DefaultTTL    time.Duration `json:"default_ttl"`
    MaxSize       int           `json:"max_size"`
    CleanupPeriod time.Duration `json:"cleanup_period"`
    EnableMetrics bool          `json:"enable_metrics"`
}
```

#### Environment Variables:
- `CACHE_DEFAULT_TTL` (default: 5m)
- `CACHE_MAX_SIZE` (default: 1000)
- `CACHE_CLEANUP_PERIOD` (default: 10m)
- `CACHE_ENABLE_METRICS` (default: true)

#### Usage:
```go
cacheStep := core.NewCacheSetStep("cacheUser", "user_${userId}", "user_data", 
    config.Cache.DefaultTTL)
```

### Logging Configuration (`LoggingConfig`)

Structured logging configuration:
```go
type LoggingConfig struct {
    Level            string   `json:"level"`
    Format           string   `json:"format"` // json, console
    EnableStackTrace bool     `json:"enable_stack_trace"`
    SanitizeFields   []string `json:"sanitize_fields"`
}
```

#### Environment Variables:
- `LOG_LEVEL` (default: "info")
- `LOG_FORMAT` (default: "json")
- `LOG_ENABLE_STACK_TRACE` (default: false)

#### Default Sanitized Fields:
- password
- token
- secret
- key
- authorization

#### Usage:
```go
logger := zap.NewProduction()
if config.Logging.Format == "console" {
    logger = zap.NewDevelopment()
}
```

### Security Configuration (`SecurityConfig`)

Security-related settings including token validation and rate limiting:
```go
type SecurityConfig struct {
    TokenValidation TokenValidationConfig `json:"token_validation"`
    RateLimit       RateLimitConfig       `json:"rate_limit"`
}

type TokenValidationConfig struct {
    Algorithm    string        `json:"algorithm"`
    SecretKey    string        `json:"secret_key"`
    Issuer       string        `json:"issuer"`
    ExpiryBuffer time.Duration `json:"expiry_buffer"`
}

type RateLimitConfig struct {
    RequestsPerSecond int           `json:"requests_per_second"`
    BurstSize         int           `json:"burst_size"`
    WindowSize        time.Duration `json:"window_size"`
}
```

#### Environment Variables:
- `TOKEN_ALGORITHM` (default: "HS256")
- `TOKEN_SECRET_KEY` (default: "")
- `TOKEN_ISSUER` (default: "api-orchestration-framework")
- `TOKEN_EXPIRY_BUFFER` (default: 5m)
- `RATE_LIMIT_RPS` (default: 100)
- `RATE_LIMIT_BURST` (default: 200)
- `RATE_LIMIT_WINDOW` (default: 1m)

#### Usage:
```go
authStep := core.NewTokenValidationStep("auth", "Authorization").
    WithValidationFunc(func(token string) bool {
        // Use config.Security.TokenValidation settings
        return validateJWT(token, config.Security.TokenValidation.SecretKey)
    })
```

### Timeout Configuration (`TimeoutConfig`)

Comprehensive timeout settings for different operations:
```go
type TimeoutConfig struct {
    FlowExecution time.Duration `json:"flow_execution"`
    StepExecution time.Duration `json:"step_execution"`
    HTTPRequest   time.Duration `json:"http_request"`
    CacheOp       time.Duration `json:"cache_op"`
}
```

#### Environment Variables:
- `TIMEOUT_FLOW_EXECUTION` (default: 30s)
- `TIMEOUT_STEP_EXECUTION` (default: 10s)
- `TIMEOUT_HTTP_REQUEST` (default: 15s)
- `TIMEOUT_CACHE_OP` (default: 1s)

#### Usage:
```go
flow := flow.NewFlow("UserFlow").
    WithTimeout(config.Timeouts.FlowExecution)

httpStep := http.GET("/api/data").
    WithTimeout(config.Timeouts.HTTPRequest)
```

### Mobile Configuration (`MobileConfig`)

Mobile-specific optimizations and settings:
```go
type MobileConfig struct {
    MaxPayloadSize    int      `json:"max_payload_size"`
    CompressionLevel  int      `json:"compression_level"`
    DefaultFields     []string `json:"default_fields"`
    EnableCompression bool     `json:"enable_compression"`
    OptimizeImages    bool     `json:"optimize_images"`
}
```

#### Environment Variables:
- `MOBILE_MAX_PAYLOAD_SIZE` (default: 1MB)
- `MOBILE_COMPRESSION_LEVEL` (default: 6)
- `MOBILE_ENABLE_COMPRESSION` (default: true)
- `MOBILE_OPTIMIZE_IMAGES` (default: true)

#### Default Mobile Fields:
- id
- name
- status
- timestamp

#### Usage:
```go
mobileStep := bff.NewMobileAPIStep("profile", "GET", "/api/profile", 
    config.Mobile.DefaultFields).
    WithMobileHeaders("mobile", "1.0", "ios")
```

## Configuration Access Patterns

### Default Configuration
Get configuration with environment variable overrides:
```go
config := config.DefaultConfig()
```

### Environment-based Configuration
Create configuration entirely from environment variables:
```go
config := config.ConfigFromEnv()
```

### Custom Configuration
Create and customize configuration programmatically:
```go
config := &config.FrameworkConfig{
    HTTP: config.HTTPConfig{
        MaxRetries:      5,
        RequestTimeout:  20 * time.Second,
        EnableFallback:  true,
    },
    Cache: config.CacheConfig{
        DefaultTTL: 10 * time.Minute,
        MaxSize:    2000,
    },
    Timeouts: config.TimeoutConfig{
        FlowExecution: 45 * time.Second,
        HTTPRequest:   20 * time.Second,
    },
}
```

### Configuration Validation
Validate configuration before use:
```go
config := config.DefaultConfig()
if err := config.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

## Integration with Framework Components

### Flow Context Integration
Configuration is automatically available in flow contexts:
```go
ctx := flow.NewContextWithConfig(config)

// Access configuration in steps
httpConfig := ctx.Config().HTTP
cacheConfig := ctx.Config().Cache
timeoutConfig := ctx.Config().Timeouts
```

### HTTP Steps Integration
HTTP steps automatically use configuration:
```go
// Configuration is applied automatically
httpStep := http.NewJSONAPIStep("GET", "/api/data")

// Or override with custom configuration
customConfig := &http.ClientConfig{
    MaxRetries:      config.HTTP.MaxRetries,
    RequestTimeout:  config.HTTP.RequestTimeout,
    EnableFallback:  config.HTTP.EnableFallback,
}
httpStep.WithClientConfig(customConfig)
```

### Cache Steps Integration
Cache steps use configuration for TTL and behavior:
```go
// Uses default TTL from configuration
cacheStep := core.NewCacheSetStep("cache", "key", "data", config.Cache.DefaultTTL)

// Cache configuration is automatically applied
cacheStep.WithConfig(&config.Cache)
```

### BFF Steps Integration
BFF steps use mobile configuration:
```go
mobileStep := bff.NewMobileAPIStep("profile", "GET", "/api/profile", 
    config.Mobile.DefaultFields)

// Mobile configuration is automatically applied for:
// - Payload size limits
// - Compression settings
// - Image optimization
// - Default field selection
```

## Environment Variable Helpers

The package provides helper functions for environment variable processing:

### String Values
```go
value := getEnvString("CONFIG_KEY", "default_value")
```

### Integer Values
```go
value := getEnvInt("CONFIG_KEY", 100)
```

### Boolean Values
```go
value := getEnvBool("CONFIG_KEY", true)
```

### Duration Values
```go
value := getEnvDuration("CONFIG_KEY", 30*time.Second)
```

## Configuration Best Practices

### Environment-Specific Configuration

#### Development Environment
```bash
# Development settings
export LOG_LEVEL=debug
export LOG_FORMAT=console
export HTTP_MAX_RETRIES=1
export CACHE_DEFAULT_TTL=1m
export TIMEOUT_FLOW_EXECUTION=60s
```

#### Production Environment
```bash
# Production settings
export LOG_LEVEL=info
export LOG_FORMAT=json
export HTTP_MAX_RETRIES=3
export CACHE_DEFAULT_TTL=10m
export TIMEOUT_FLOW_EXECUTION=30s
export HTTP_ENABLE_FALLBACK=true
export CACHE_ENABLE_METRICS=true
```

#### Testing Environment
```bash
# Testing settings
export LOG_LEVEL=warn
export HTTP_MAX_RETRIES=0
export CACHE_DEFAULT_TTL=1s
export TIMEOUT_FLOW_EXECUTION=5s
export HTTP_ENABLE_FALLBACK=false
```

### Configuration Validation

Implement custom validation for specific requirements:
```go
func ValidateProductionConfig(config *config.FrameworkConfig) error {
    if config.Security.TokenValidation.SecretKey == "" {
        return fmt.Errorf("token secret key is required in production")
    }
    
    if config.HTTP.RequestTimeout > 30*time.Second {
        return fmt.Errorf("HTTP timeout too high for production")
    }
    
    if config.Logging.Level == "debug" {
        return fmt.Errorf("debug logging not allowed in production")
    }
    
    return nil
}
```

### Configuration Monitoring

Monitor configuration changes and performance impact:
```go
func MonitorConfiguration(config *config.FrameworkConfig) {
    // Log configuration on startup
    log.Info("Framework configuration loaded",
        zap.Int("http_max_retries", config.HTTP.MaxRetries),
        zap.Duration("http_timeout", config.HTTP.RequestTimeout),
        zap.Duration("cache_ttl", config.Cache.DefaultTTL),
        zap.String("log_level", config.Logging.Level))
    
    // Monitor configuration-dependent metrics
    metrics.RecordConfigurationMetrics(config)
}
```

## Advanced Configuration Patterns

### Configuration Composition
Compose configurations from multiple sources:
```go
func CreateCompositeConfig() *config.FrameworkConfig {
    // Start with defaults
    cfg := config.DefaultConfig()
    
    // Override with environment-specific settings
    if env := os.Getenv("ENVIRONMENT"); env == "production" {
        cfg.HTTP.MaxRetries = 5
        cfg.HTTP.EnableFallback = true
        cfg.Logging.Level = "info"
        cfg.Cache.EnableMetrics = true
    } else if env == "development" {
        cfg.HTTP.MaxRetries = 1
        cfg.Logging.Level = "debug"
        cfg.Logging.Format = "console"
    }
    
    // Override with service-specific settings
    if service := os.Getenv("SERVICE_NAME"); service == "mobile-api" {
        cfg.Mobile.EnableCompression = true
        cfg.Mobile.OptimizeImages = true
        cfg.Mobile.MaxPayloadSize = 512 * 1024 // 512KB for mobile
    }
    
    return cfg
}
```

### Dynamic Configuration Updates
Support runtime configuration updates:
```go
type ConfigManager struct {
    config *config.FrameworkConfig
    mu     sync.RWMutex
}

func (cm *ConfigManager) UpdateConfig(newConfig *config.FrameworkConfig) error {
    if err := newConfig.Validate(); err != nil {
        return err
    }
    
    cm.mu.Lock()
    defer cm.mu.Unlock()
    cm.config = newConfig
    
    // Notify components of configuration change
    cm.notifyConfigChange()
    return nil
}

func (cm *ConfigManager) GetConfig() *config.FrameworkConfig {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return cm.config
}
```

### Configuration Profiles
Support different configuration profiles:
```go
func LoadConfigProfile(profile string) (*config.FrameworkConfig, error) {
    switch profile {
    case "high-performance":
        return &config.FrameworkConfig{
            HTTP: config.HTTPConfig{
                MaxIdleConns:        200,
                MaxIdleConnsPerHost: 20,
                MaxRetries:          5,
                RequestTimeout:      10 * time.Second,
            },
            Cache: config.CacheConfig{
                MaxSize:    5000,
                DefaultTTL: 30 * time.Minute,
            },
        }, nil
        
    case "low-latency":
        return &config.FrameworkConfig{
            HTTP: config.HTTPConfig{
                MaxRetries:     1,
                RequestTimeout: 5 * time.Second,
            },
            Timeouts: config.TimeoutConfig{
                FlowExecution: 10 * time.Second,
                StepExecution: 3 * time.Second,
            },
        }, nil
        
    case "mobile-optimized":
        return &config.FrameworkConfig{
            Mobile: config.MobileConfig{
                MaxPayloadSize:    256 * 1024, // 256KB
                EnableCompression: true,
                OptimizeImages:    true,
                CompressionLevel:  9,
            },
            HTTP: config.HTTPConfig{
                RequestTimeout: 8 * time.Second,
                MaxRetries:     2,
            },
        }, nil
        
    default:
        return config.DefaultConfig(), nil
    }
}
```

## Configuration Examples

### Complete Production Configuration
```go
func ProductionConfig() *config.FrameworkConfig {
    return &config.FrameworkConfig{
        HTTP: config.HTTPConfig{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
            MaxRetries:          3,
            RetryDelay:          2 * time.Second,
            MaxRetryDelay:       10 * time.Second,
            FailureThreshold:    5,
            SuccessThreshold:    3,
            CircuitBreakerDelay: 5 * time.Second,
            RequestTimeout:      15 * time.Second,
            EnableFallback:      true,
            UserAgent:           "MyApp-Production/1.0",
        },
        Cache: config.CacheConfig{
            DefaultTTL:    10 * time.Minute,
            MaxSize:       2000,
            CleanupPeriod: 5 * time.Minute,
            EnableMetrics: true,
        },
        Logging: config.LoggingConfig{
            Level:            "info",
            Format:           "json",
            EnableStackTrace: false,
            SanitizeFields:   []string{"password", "token", "secret", "key"},
        },
        Security: config.SecurityConfig{
            TokenValidation: config.TokenValidationConfig{
                Algorithm:    "HS256",
                SecretKey:    os.Getenv("JWT_SECRET_KEY"),
                Issuer:       "myapp-production",
                ExpiryBuffer: 5 * time.Minute,
            },
            RateLimit: config.RateLimitConfig{
                RequestsPerSecond: 100,
                BurstSize:         200,
                WindowSize:        1 * time.Minute,
            },
        },
        Timeouts: config.TimeoutConfig{
            FlowExecution: 30 * time.Second,
            StepExecution: 10 * time.Second,
            HTTPRequest:   15 * time.Second,
            CacheOp:       1 * time.Second,
        },
        Mobile: config.MobileConfig{
            MaxPayloadSize:    1024 * 1024, // 1MB
            CompressionLevel:  6,
            DefaultFields:     []string{"id", "name", "status", "timestamp"},
            EnableCompression: true,
            OptimizeImages:    true,
        },
    }
}
```

### Development Configuration
```go
func DevelopmentConfig() *config.FrameworkConfig {
    return &config.FrameworkConfig{
        HTTP: config.HTTPConfig{
            MaxRetries:     1,
            RequestTimeout: 30 * time.Second,
            EnableFallback: false, // Fail fast in development
        },
        Cache: config.CacheConfig{
            DefaultTTL: 1 * time.Minute, // Short TTL for development
            MaxSize:    100,
        },
        Logging: config.LoggingConfig{
            Level:            "debug",
            Format:           "console",
            EnableStackTrace: true,
        },
        Timeouts: config.TimeoutConfig{
            FlowExecution: 60 * time.Second, // Longer for debugging
            StepExecution: 30 * time.Second,
        },
    }
}
```

## Troubleshooting

### Common Configuration Issues

1. **Environment Variable Type Mismatch**
   ```go
   // Wrong: Setting duration as string without unit
   export TIMEOUT_FLOW_EXECUTION=30
   
   // Correct: Include time unit
   export TIMEOUT_FLOW_EXECUTION=30s
   ```

2. **Missing Required Configuration**
   ```go
   // Check for required configuration
   if config.Security.TokenValidation.SecretKey == "" {
       log.Fatal("JWT_SECRET_KEY environment variable is required")
   }
   ```

3. **Invalid Configuration Values**
   ```go
   // Validate configuration ranges
   if config.HTTP.MaxRetries < 0 || config.HTTP.MaxRetries > 10 {
       log.Fatal("HTTP_MAX_RETRIES must be between 0 and 10")
   }
   ```

### Configuration Debugging

Enable configuration debugging:
```go
func DebugConfiguration(config *config.FrameworkConfig) {
    log.Debug("Configuration loaded",
        zap.Any("http", config.HTTP),
        zap.Any("cache", config.Cache),
        zap.Any("logging", config.Logging),
        zap.Any("security", config.Security),
        zap.Any("timeouts", config.Timeouts),
        zap.Any("mobile", config.Mobile))
}
```

### Performance Impact Monitoring

Monitor configuration impact on performance:
```go
func MonitorConfigurationPerformance(config *config.FrameworkConfig) {
    // Monitor HTTP configuration impact
    if config.HTTP.MaxRetries > 3 {
        log.Warn("High retry count may impact performance",
            zap.Int("max_retries", config.HTTP.MaxRetries))
    }
    
    // Monitor cache configuration impact
    if config.Cache.DefaultTTL < 1*time.Minute {
        log.Warn("Low cache TTL may increase backend load",
            zap.Duration("ttl", config.Cache.DefaultTTL))
    }
    
    // Monitor timeout configuration
    if config.Timeouts.FlowExecution > 60*time.Second {
        log.Warn("High flow timeout may impact user experience",
            zap.Duration("timeout", config.Timeouts.FlowExecution))
    }
}
``` 