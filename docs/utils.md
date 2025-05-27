# Utils Package Documentation

## Overview

The `pkg/utils` package provides essential utility functions and helpers for the API Orchestration Framework. It includes string interpolation, nested data manipulation, object pooling for performance optimization, and data sanitization utilities that are used throughout the framework.

## Purpose

The utils package serves as the foundation utility layer that:
- Provides string interpolation for dynamic value substitution
- Offers efficient nested data structure manipulation
- Implements object pooling for performance optimization
- Provides data sanitization for security and logging
- Supports common data transformation operations
- Enables efficient memory management across the framework

## Core Components

### String Interpolation (`interpolation.go`)

Advanced string interpolation system for dynamic value substitution in templates, URLs, and configuration.

#### InterpolationEngine
The core interpolation engine with support for multiple syntax patterns:
```go
type InterpolationEngine struct {
    // Configurable delimiters and options
    StartDelimiter string
    EndDelimiter   string
    CaseSensitive  bool
    AllowMissing   bool
    DefaultValue   string
}

// Create new interpolation engine
engine := utils.NewInterpolationEngine().
    WithDelimiters("${", "}").
    WithCaseSensitive(true).
    WithAllowMissing(false).
    WithDefaultValue("")
```

#### Basic Interpolation
Simple variable substitution:
```go
// Basic variable substitution
template := "Hello ${name}, welcome to ${app}!"
data := map[string]interface{}{
    "name": "John",
    "app":  "MyApp",
}
result := utils.Interpolate(template, data)
// Result: "Hello John, welcome to MyApp!"
```

#### Nested Field Access
Support for dot notation in variable names:
```go
// Nested field access
template := "User: ${user.profile.name} (${user.profile.email})"
data := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "name":  "John Doe",
            "email": "john@example.com",
        },
    },
}
result := utils.Interpolate(template, data)
// Result: "User: John Doe (john@example.com)"
```

#### Array Index Access
Access array elements by index:
```go
// Array index access
template := "First item: ${items[0]}, Second: ${items[1]}"
data := map[string]interface{}{
    "items": []interface{}{"apple", "banana", "cherry"},
}
result := utils.Interpolate(template, data)
// Result: "First item: apple, Second: banana"
```

#### Advanced Interpolation Features

##### Default Values
Provide default values for missing variables:
```go
// Default values
template := "Hello ${name:Guest}, your role is ${role:user}"
data := map[string]interface{}{
    "name": "John",
    // role is missing, will use default
}
result := utils.InterpolateWithDefaults(template, data)
// Result: "Hello John, your role is user"
```

##### Conditional Interpolation
Conditional value substitution:
```go
// Conditional interpolation
template := "Status: ${status?active:inactive}"
data := map[string]interface{}{
    "status": true,
}
result := utils.InterpolateConditional(template, data)
// Result: "Status: active"
```

##### Format Specifiers
Apply formatting to interpolated values:
```go
// Format specifiers
template := "Price: ${price|currency}, Date: ${date|iso}"
data := map[string]interface{}{
    "price": 29.99,
    "date":  time.Now(),
}
formatters := map[string]utils.Formatter{
    "currency": utils.CurrencyFormatter("$", 2),
    "iso":      utils.ISODateFormatter(),
}
result := utils.InterpolateWithFormatters(template, data, formatters)
// Result: "Price: $29.99, Date: 2023-12-01T10:30:00Z"
```

#### URL Interpolation
Specialized interpolation for URLs with proper encoding:
```go
// URL interpolation with encoding
urlTemplate := "/api/users/${userId}/posts/${postId}?filter=${filter}"
data := map[string]interface{}{
    "userId": 123,
    "postId": 456,
    "filter": "recent posts",
}
url := utils.InterpolateURL(urlTemplate, data)
// Result: "/api/users/123/posts/456?filter=recent%20posts"
```

#### SQL Interpolation
Safe SQL interpolation with parameter binding:
```go
// SQL interpolation (safe parameter binding)
sqlTemplate := "SELECT * FROM users WHERE id = ${userId} AND status = ${status}"
data := map[string]interface{}{
    "userId": 123,
    "status": "active",
}
query, params := utils.InterpolateSQL(sqlTemplate, data)
// Result: query = "SELECT * FROM users WHERE id = ? AND status = ?"
//         params = []interface{}{123, "active"}
```

### Nested Data Manipulation (`nested.go`)

Comprehensive utilities for working with nested data structures, maps, and complex object hierarchies.

#### GetNestedValue
Safely retrieve values from nested structures:
```go
// Get nested value with dot notation
data := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "settings": map[string]interface{}{
                "theme": "dark",
                "notifications": true,
            },
        },
    },
}

// Get nested value
theme := utils.GetNestedValue(data, "user.profile.settings.theme")
// Result: "dark"

// Get with default value
language := utils.GetNestedValueWithDefault(data, "user.profile.settings.language", "en")
// Result: "en" (default value since path doesn't exist)
```

#### SetNestedValue
Set values in nested structures, creating intermediate maps as needed:
```go
// Set nested value, creating intermediate structures
data := make(map[string]interface{})
utils.SetNestedValue(data, "user.profile.settings.theme", "light")

// Result: data = {
//     "user": {
//         "profile": {
//             "settings": {
//                 "theme": "light"
//             }
//         }
//     }
// }
```

#### DeleteNestedValue
Remove values from nested structures:
```go
// Delete nested value
utils.DeleteNestedValue(data, "user.profile.settings.theme")
// Removes the theme setting while preserving the rest of the structure
```

#### HasNestedValue
Check if a nested path exists:
```go
// Check if nested path exists
exists := utils.HasNestedValue(data, "user.profile.settings.theme")
// Result: true or false
```

#### FlattenMap
Convert nested maps to flat key-value pairs:
```go
// Flatten nested map
nested := map[string]interface{}{
    "user": map[string]interface{}{
        "id":   123,
        "name": "John",
        "profile": map[string]interface{}{
            "email": "john@example.com",
            "settings": map[string]interface{}{
                "theme":         "dark",
                "notifications": true,
            },
        },
    },
}

flattened := utils.FlattenMap(nested, ".")
// Result: {
//     "user.id": 123,
//     "user.name": "John",
//     "user.profile.email": "john@example.com",
//     "user.profile.settings.theme": "dark",
//     "user.profile.settings.notifications": true,
// }
```

#### UnflattenMap
Convert flat key-value pairs back to nested structures:
```go
// Unflatten map
flat := map[string]interface{}{
    "user.id":                           123,
    "user.name":                         "John",
    "user.profile.email":                "john@example.com",
    "user.profile.settings.theme":       "dark",
    "user.profile.settings.notifications": true,
}

nested := utils.UnflattenMap(flat, ".")
// Result: Original nested structure
```

#### MergeNestedMaps
Deep merge multiple nested maps:
```go
// Deep merge maps
map1 := map[string]interface{}{
    "user": map[string]interface{}{
        "id":   123,
        "name": "John",
        "profile": map[string]interface{}{
            "email": "john@example.com",
        },
    },
}

map2 := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "phone": "+1234567890",
            "settings": map[string]interface{}{
                "theme": "dark",
            },
        },
    },
}

merged := utils.MergeNestedMaps(map1, map2)
// Result: Deep merged structure with all fields
```

#### CloneNestedMap
Create deep copies of nested structures:
```go
// Deep clone nested map
original := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "settings": map[string]interface{}{
                "theme": "dark",
            },
        },
    },
}

cloned := utils.CloneNestedMap(original)
// Result: Complete deep copy, safe to modify independently
```

### Object Pooling (`pools.go`)

High-performance object pooling system for memory optimization and garbage collection reduction.

#### Generic Pool
Type-safe generic pool implementation:
```go
// Create a pool for map[string]interface{}
mapPool := utils.NewPool(func() map[string]interface{} {
    return make(map[string]interface{})
}, func(m map[string]interface{}) {
    // Reset function to clean the map
    for k := range m {
        delete(m, k)
    }
})

// Get object from pool
data := mapPool.Get()

// Use the object
data["key"] = "value"

// Return to pool when done
mapPool.Put(data)
```

#### Specialized Pools

##### String Builder Pool
Optimized pool for string building operations:
```go
// String builder pool
builderPool := utils.NewStringBuilderPool()

// Get builder from pool
builder := builderPool.Get()

// Use builder
builder.WriteString("Hello ")
builder.WriteString("World")
result := builder.String()

// Return to pool
builderPool.Put(builder)
```

##### Byte Buffer Pool
Pool for byte buffer operations:
```go
// Byte buffer pool
bufferPool := utils.NewByteBufferPool()

// Get buffer from pool
buffer := bufferPool.Get()

// Use buffer
buffer.Write([]byte("data"))
data := buffer.Bytes()

// Return to pool
bufferPool.Put(buffer)
```

##### HTTP Request Pool
Pool for HTTP request objects:
```go
// HTTP request pool
requestPool := utils.NewHTTPRequestPool()

// Get request from pool
req := requestPool.Get()

// Configure request
req.Method = "GET"
req.URL = url
req.Header.Set("Content-Type", "application/json")

// Use request
resp, err := client.Do(req)

// Return to pool
requestPool.Put(req)
```

#### Pool Monitoring
Monitor pool performance and usage:
```go
// Pool with monitoring
monitoredPool := utils.NewMonitoredPool(
    func() interface{} { return make(map[string]interface{}) },
    func(obj interface{}) { /* reset */ },
).WithMetrics(true).WithMaxSize(1000)

// Get pool statistics
stats := monitoredPool.Stats()
log.Info("Pool statistics",
    zap.Int("gets", stats.Gets),
    zap.Int("puts", stats.Puts),
    zap.Int("hits", stats.Hits),
    zap.Int("misses", stats.Misses),
    zap.Int("size", stats.Size))
```

### Data Sanitization (`sanitize.go`)

Comprehensive data sanitization for security, logging, and data protection.

#### Field Sanitization
Sanitize sensitive fields in data structures:
```go
// Sanitize sensitive fields
data := map[string]interface{}{
    "username": "john_doe",
    "password": "secret123",
    "email":    "john@example.com",
    "token":    "abc123xyz",
    "profile": map[string]interface{}{
        "name":        "John Doe",
        "credit_card": "1234-5678-9012-3456",
    },
}

// Default sensitive fields
sensitiveFields := []string{"password", "token", "secret", "key", "credit_card"}
sanitized := utils.SanitizeFields(data, sensitiveFields)

// Result: {
//     "username": "john_doe",
//     "password": "[REDACTED]",
//     "email":    "john@example.com",
//     "token":    "[REDACTED]",
//     "profile": {
//         "name":        "John Doe",
//         "credit_card": "[REDACTED]",
//     },
// }
```

#### Custom Sanitization Rules
Define custom sanitization patterns:
```go
// Custom sanitization rules
rules := []utils.SanitizationRule{
    {
        Pattern:     regexp.MustCompile(`\b\d{4}-\d{4}-\d{4}-\d{4}\b`),
        Replacement: "[CREDIT_CARD]",
        Description: "Credit card numbers",
    },
    {
        Pattern:     regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
        Replacement: "[EMAIL]",
        Description: "Email addresses",
    },
    {
        Pattern:     regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
        Replacement: "[SSN]",
        Description: "Social Security Numbers",
    },
}

sanitizer := utils.NewSanitizer(rules)
sanitized := sanitizer.SanitizeString("Contact: john@example.com, Card: 1234-5678-9012-3456")
// Result: "Contact: [EMAIL], Card: [CREDIT_CARD]"
```

#### Logging Sanitization
Sanitize data for safe logging:
```go
// Sanitize for logging
logData := utils.SanitizeForLogging(data, utils.DefaultSensitiveFields())

// Log safely
log.Info("User data processed", zap.Any("data", logData))
```

#### URL Sanitization
Sanitize URLs by removing sensitive query parameters:
```go
// Sanitize URLs
originalURL := "https://api.example.com/users?token=secret123&api_key=xyz789&user_id=123"
sensitiveParams := []string{"token", "api_key", "password", "secret"}
sanitizedURL := utils.SanitizeURL(originalURL, sensitiveParams)
// Result: "https://api.example.com/users?token=[REDACTED]&api_key=[REDACTED]&user_id=123"
```

#### Header Sanitization
Sanitize HTTP headers for logging:
```go
// Sanitize HTTP headers
headers := http.Header{
    "Authorization": []string{"Bearer secret_token"},
    "X-API-Key":     []string{"api_key_123"},
    "Content-Type":  []string{"application/json"},
    "User-Agent":    []string{"MyApp/1.0"},
}

sanitizedHeaders := utils.SanitizeHeaders(headers, []string{"authorization", "x-api-key"})
// Result: Authorization and X-API-Key values are redacted
```

## Integration with Framework Components

### Flow Context Integration
Utils are integrated throughout the flow execution:
```go
// String interpolation in flow context
ctx.SetValue("user_id", 123)
url := utils.Interpolate("/api/users/${user_id}/profile", ctx.GetData())

// Nested data access in context
userEmail := utils.GetNestedValue(ctx.GetData(), "user.profile.email")

// Sanitize context data for logging
sanitizedData := utils.SanitizeForLogging(ctx.GetData(), utils.DefaultSensitiveFields())
```

### HTTP Steps Integration
Utils support HTTP step operations:
```go
// URL interpolation in HTTP steps
httpStep := http.GET("/api/users/${userId}/posts/${postId}").
    WithURLInterpolation(true)

// Header sanitization for logging
httpStep.WithHeaderSanitization([]string{"authorization", "x-api-key"})

// Response data manipulation
httpStep.WithResponseProcessor(func(data map[string]interface{}) map[string]interface{} {
    // Use nested utils to process response
    utils.SetNestedValue(data, "metadata.processed_at", time.Now())
    return data
})
```

### Cache Steps Integration
Utils support cache key generation and data processing:
```go
// Cache key interpolation
cacheKey := utils.Interpolate("user_${userId}_profile_${version}", ctx.GetData())
cacheStep := core.NewCacheGetStep("getCache", cacheKey, "cached_data")

// Sanitize cached data
cacheStep.WithDataProcessor(func(data interface{}) interface{} {
    if mapData, ok := data.(map[string]interface{}); ok {
        return utils.SanitizeFields(mapData, utils.DefaultSensitiveFields())
    }
    return data
})
```

### Transformer Integration
Utils support transformer operations:
```go
// Custom transformer using utils
customTransformer := transformers.NewFuncTransformer("processUser", 
    func(data map[string]interface{}) (map[string]interface{}, error) {
        // Use nested utils for data manipulation
        result := utils.CloneNestedMap(data)
        
        // Set computed fields
        fullName := fmt.Sprintf("%s %s", 
            utils.GetNestedValue(result, "profile.first_name"),
            utils.GetNestedValue(result, "profile.last_name"))
        utils.SetNestedValue(result, "profile.full_name", fullName)
        
        // Sanitize sensitive data
        result = utils.SanitizeFields(result, []string{"password", "token"})
        
        return result, nil
    })
```

## Performance Optimizations

### Memory Management
Efficient memory usage patterns:
```go
// Use object pools for frequently allocated objects
var mapPool = utils.NewPool(
    func() map[string]interface{} { return make(map[string]interface{}) },
    func(m map[string]interface{}) {
        for k := range m {
            delete(m, k)
        }
    },
)

// Reuse maps instead of creating new ones
func processData(input map[string]interface{}) map[string]interface{} {
    result := mapPool.Get()
    defer mapPool.Put(result)
    
    // Process data using pooled map
    for k, v := range input {
        result[k] = v
    }
    
    return utils.CloneNestedMap(result) // Return copy, recycle original
}
```

### String Operations
Optimized string building and manipulation:
```go
// Use string builder pool for efficient string concatenation
func buildLargeString(parts []string) string {
    builder := utils.GetStringBuilder()
    defer utils.PutStringBuilder(builder)
    
    for _, part := range parts {
        builder.WriteString(part)
    }
    
    return builder.String()
}
```

### Interpolation Caching
Cache compiled interpolation patterns:
```go
// Cache interpolation engines for reuse
var interpolationCache = make(map[string]*utils.InterpolationEngine)

func getInterpolationEngine(pattern string) *utils.InterpolationEngine {
    if engine, exists := interpolationCache[pattern]; exists {
        return engine
    }
    
    engine := utils.NewInterpolationEngine().WithPattern(pattern)
    interpolationCache[pattern] = engine
    return engine
}
```

## Best Practices

### String Interpolation
1. **Use Appropriate Delimiters**: Choose delimiters that don't conflict with your data
2. **Validate Templates**: Validate interpolation templates before use
3. **Handle Missing Values**: Decide how to handle missing interpolation values
4. **Cache Engines**: Cache interpolation engines for repeated use
5. **Escape Special Characters**: Properly escape special characters in templates

### Nested Data Manipulation
1. **Validate Paths**: Validate nested paths before accessing
2. **Use Safe Access**: Use safe access methods that handle missing paths
3. **Clone Before Modify**: Clone data structures before modification when needed
4. **Efficient Merging**: Use efficient merging strategies for large data structures
5. **Memory Management**: Be mindful of memory usage with deep cloning

### Object Pooling
1. **Pool Appropriate Objects**: Pool objects that are frequently allocated/deallocated
2. **Proper Reset**: Ensure objects are properly reset before returning to pool
3. **Monitor Pool Size**: Monitor pool size to prevent memory leaks
4. **Thread Safety**: Ensure thread safety for concurrent pool access
5. **Pool Lifecycle**: Manage pool lifecycle appropriately

### Data Sanitization
1. **Comprehensive Rules**: Define comprehensive sanitization rules
2. **Performance Impact**: Consider performance impact of sanitization
3. **Consistent Application**: Apply sanitization consistently across the application
4. **Regular Updates**: Regularly update sanitization patterns
5. **Audit Sanitization**: Audit sanitization effectiveness regularly

## Examples

### Complete Data Processing Pipeline
```go
func ProcessUserData(rawData map[string]interface{}) (map[string]interface{}, error) {
    // Get pooled map for processing
    processedData := mapPool.Get()
    defer mapPool.Put(processedData)
    
    // Clone input data to avoid modifying original
    workingData := utils.CloneNestedMap(rawData)
    
    // Interpolate dynamic values
    if template, exists := utils.GetNestedValue(workingData, "template"); exists {
        if templateStr, ok := template.(string); ok {
            interpolated := utils.Interpolate(templateStr, workingData)
            utils.SetNestedValue(workingData, "interpolated_value", interpolated)
        }
    }
    
    // Process nested data
    if userProfile := utils.GetNestedValue(workingData, "user.profile"); userProfile != nil {
        // Flatten profile for easier processing
        if profileMap, ok := userProfile.(map[string]interface{}); ok {
            flattened := utils.FlattenMap(profileMap, "_")
            utils.SetNestedValue(workingData, "user.profile_flat", flattened)
        }
    }
    
    // Sanitize sensitive data
    sanitized := utils.SanitizeFields(workingData, []string{
        "password", "token", "secret", "credit_card", "ssn",
    })
    
    // Return processed data
    return utils.CloneNestedMap(sanitized), nil
}
```

### Advanced Interpolation with Custom Formatters
```go
func CreateAdvancedInterpolationEngine() *utils.InterpolationEngine {
    formatters := map[string]utils.Formatter{
        "upper":    utils.UpperCaseFormatter(),
        "lower":    utils.LowerCaseFormatter(),
        "currency": utils.CurrencyFormatter("$", 2),
        "date":     utils.DateFormatter("2006-01-02"),
        "time":     utils.TimeFormatter("15:04:05"),
        "url":      utils.URLEncodeFormatter(),
        "base64":   utils.Base64EncodeFormatter(),
    }
    
    return utils.NewInterpolationEngine().
        WithDelimiters("{{", "}}").
        WithFormatters(formatters).
        WithCaseSensitive(false).
        WithAllowMissing(true).
        WithDefaultValue("[MISSING]")
}

func ProcessTemplate(template string, data map[string]interface{}) string {
    engine := CreateAdvancedInterpolationEngine()
    
    // Example template: "Hello {{name|upper}}, your balance is {{balance|currency}}"
    // With data: {"name": "john", "balance": 1234.56}
    // Result: "Hello JOHN, your balance is $1234.56"
    
    return engine.Interpolate(template, data)
}
```

### High-Performance Data Sanitization
```go
func CreateProductionSanitizer() *utils.Sanitizer {
    rules := []utils.SanitizationRule{
        // Credit card numbers
        {
            Pattern:     regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
            Replacement: "[CREDIT_CARD]",
            Description: "Credit card numbers",
        },
        // Social Security Numbers
        {
            Pattern:     regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
            Replacement: "[SSN]",
            Description: "Social Security Numbers",
        },
        // Email addresses
        {
            Pattern:     regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
            Replacement: "[EMAIL]",
            Description: "Email addresses",
        },
        // Phone numbers
        {
            Pattern:     regexp.MustCompile(`\b(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b`),
            Replacement: "[PHONE]",
            Description: "Phone numbers",
        },
        // API keys and tokens
        {
            Pattern:     regexp.MustCompile(`\b[A-Za-z0-9]{32,}\b`),
            Replacement: "[TOKEN]",
            Description: "API keys and tokens",
        },
    }
    
    return utils.NewSanitizer(rules).
        WithCaching(true).
        WithPerformanceMonitoring(true)
}
```

## Troubleshooting

### Common Issues

1. **Interpolation Failures**
   ```go
   // Check for missing variables
   template := "Hello ${name}"
   data := map[string]interface{}{} // missing "name"
   
   // Use safe interpolation
   result := utils.InterpolateWithDefault(template, data, "[UNKNOWN]")
   ```

2. **Nested Path Errors**
   ```go
   // Check if path exists before accessing
   if utils.HasNestedValue(data, "user.profile.email") {
       email := utils.GetNestedValue(data, "user.profile.email")
       // Process email
   }
   ```

3. **Pool Memory Leaks**
   ```go
   // Ensure proper pool usage
   obj := pool.Get()
   defer pool.Put(obj) // Always return to pool
   
   // Reset object state before returning
   pool.PutWithReset(obj, func(o interface{}) {
       // Reset object state
   })
   ```

### Performance Debugging

Monitor utils performance:
```go
func MonitorUtilsPerformance() {
    // Monitor interpolation performance
    start := time.Now()
    result := utils.Interpolate(template, data)
    duration := time.Since(start)
    
    if duration > 10*time.Millisecond {
        log.Warn("Slow interpolation detected",
            zap.Duration("duration", duration),
            zap.String("template", template))
    }
    
    // Monitor pool efficiency
    stats := pool.Stats()
    hitRate := float64(stats.Hits) / float64(stats.Gets)
    if hitRate < 0.8 {
        log.Warn("Low pool hit rate",
            zap.Float64("hit_rate", hitRate),
            zap.Int("size", stats.Size))
    }
}
``` 