# Transformers Package Documentation

## Overview

The `pkg/transformers` package provides a comprehensive, generic transformation system for data manipulation within the API Orchestration Framework. It offers a flexible, composable approach to transforming data structures with specialized support for mobile optimization, field selection, flattening, and chaining operations.

## Purpose

The transformers package serves as the data processing engine that:
- Provides a unified interface for all data transformations
- Enables composable transformation pipelines through chaining
- Offers mobile-specific optimizations for BFF patterns
- Supports conditional transformations based on data content
- Implements efficient field selection and data flattening
- Allows parallel transformation execution for performance

## Core Architecture

### Transformer Interface (`transformer.go`)

The foundation of the transformation system:
```go
type Transformer interface {
    // Transform applies the transformation to the input data
    Transform(data map[string]interface{}) (map[string]interface{}, error)
    
    // Name returns the transformer name for logging and debugging
    Name() string
}
```

### BaseTransformer
Common functionality for all transformer implementations:
```go
type BaseTransformer struct {
    name string
}

func NewBaseTransformer(name string) *BaseTransformer {
    return &BaseTransformer{name: name}
}

func (bt *BaseTransformer) Name() string {
    return bt.name
}
```

### Built-in Transformer Types

#### FuncTransformer
Wraps a function as a transformer:
```go
funcTransformer := transformers.NewFuncTransformer("customTransform", 
    func(data map[string]interface{}) (map[string]interface{}, error) {
        // Custom transformation logic
        result := make(map[string]interface{})
        result["processed"] = true
        result["original"] = data
        return result, nil
    })
```

#### NoOpTransformer
Pass-through transformer for testing or placeholder scenarios:
```go
noOpTransformer := transformers.NewNoOpTransformer()
```

#### CopyTransformer
Creates deep copies of data structures:
```go
copyTransformer := transformers.NewCopyTransformer()
```

## Field Transformation (`field.go`)

### FieldTransformer
Powerful field selection and manipulation transformer:
```go
fieldTransformer := transformers.NewFieldTransformer("selectFields", 
    []string{"id", "name", "email", "profile.avatar"}).
    WithPrefix("user").
    WithMeta(true).
    WithFlatten(false).
    WithSeparator("_")
```

#### Key Features:
- **Field Selection**: Choose specific fields from complex data structures
- **Nested Field Access**: Support for dot notation (e.g., "user.profile.name")
- **Prefix Addition**: Add prefixes to selected fields
- **Meta Field Inclusion**: Include metadata fields like timestamps
- **Flattening Support**: Flatten nested structures with configurable separators

#### Usage Examples:
```go
// Basic field selection
includeFields := transformers.IncludeFieldsTransformer("id", "name", "email")

// Field exclusion
excludeFields := transformers.ExcludeFieldsTransformer("password", "internal_id")

// Field renaming
renameFields := transformers.RenameFieldsTransformer(map[string]string{
    "user_id": "id",
    "full_name": "name",
    "email_address": "email",
})

// Add new fields
addFields := transformers.AddFieldsTransformer(map[string]interface{}{
    "timestamp": time.Now().Unix(),
    "version": "1.0",
    "source": "api",
})
```

## Data Flattening (`flatten.go`)

### FlattenTransformer
Converts nested data structures into flat key-value pairs:
```go
flattenTransformer := transformers.NewFlattenTransformer("flattenUser", "user").
    WithSeparator("_").
    WithMaxDepth(3).
    WithMeta(true)
```

#### Configuration Options:
- **Prefix**: Add prefix to all flattened keys
- **Separator**: Configure key separator (default: ".")
- **Max Depth**: Limit flattening depth to prevent infinite recursion
- **Meta Inclusion**: Include metadata fields in flattening

#### Example:
```go
// Input data
input := map[string]interface{}{
    "user": map[string]interface{}{
        "id": 123,
        "profile": map[string]interface{}{
            "name": "John Doe",
            "settings": map[string]interface{}{
                "theme": "dark",
                "notifications": true,
            },
        },
    },
}

// Output after flattening with prefix "user" and separator "_"
output := map[string]interface{}{
    "user_id": 123,
    "user_profile_name": "John Doe",
    "user_profile_settings_theme": "dark",
    "user_profile_settings_notifications": true,
}
```

### UnflattenTransformer
Converts flat key-value pairs back to nested structures:
```go
unflattenTransformer := transformers.NewUnflattenTransformer("unflatten").
    WithSeparator("_")
```

### PrefixTransformer
Adds prefixes to all keys:
```go
prefixTransformer := transformers.NewPrefixTransformer("addPrefix", "mobile").
    WithSeparator("_")
```

### RemovePrefixTransformer
Removes prefixes from keys:
```go
removePrefixTransformer := transformers.NewRemovePrefixTransformer("removePrefix", "api").
    WithSeparator("_")
```

## Mobile Optimizations (`mobile.go`)

### Mobile-Specific Transformers
Optimized transformers for mobile BFF patterns:

#### NewMobileTransformer
Basic mobile field selection:
```go
mobileTransformer := transformers.NewMobileTransformer([]string{
    "id", "name", "avatar", "status"
})
```

#### NewMobileFlattenTransformer
Mobile-optimized flattening:
```go
mobileFlattenTransformer := transformers.NewMobileFlattenTransformer("mobile")
```

#### NewMobileResponseTransformer
Complete mobile response transformation:
```go
mobileResponseTransformer := transformers.NewMobileResponseTransformer([]string{
    "id", "name", "avatar", "email"
})
```

#### NewMobileListTransformer
Optimizes list responses for mobile consumption:
```go
mobileListTransformer := transformers.NewMobileListTransformer([]string{
    "id", "title", "thumbnail", "status"
})
```

#### NewMobileErrorTransformer
Standardizes error responses for mobile clients:
```go
mobileErrorTransformer := transformers.NewMobileErrorTransformer()
```

#### NewMobilePaginationTransformer
Handles pagination metadata for mobile:
```go
mobilePaginationTransformer := transformers.NewMobilePaginationTransformer()
```

#### NewMobileImageTransformer
Optimizes image URLs for mobile devices:
```go
mobileImageTransformer := transformers.NewMobileImageTransformer()
```

### Pre-built Mobile Transformers
Domain-specific mobile transformers:
```go
// User profile transformer
userProfileTransformer := transformers.MobileUserProfileTransformer()

// Product transformer
productTransformer := transformers.MobileProductTransformer()

// Order transformer
orderTransformer := transformers.MobileOrderTransformer()
```

## Transformer Chaining (`chain.go`)

### TransformerChain
Sequential execution of multiple transformers:
```go
chain := transformers.NewTransformerChain("mobileProcessing",
    transformers.NewFieldTransformer("select", []string{"user", "profile", "settings"}),
    transformers.NewMobileTransformer([]string{"id", "name", "avatar"}),
    transformers.NewFlattenTransformer("flatten", "mobile"),
)

// Add transformers dynamically
chain.Add(transformers.NewPrefixTransformer("prefix", "api"))

// Insert at specific position
chain.Insert(1, transformers.NewCopyTransformer())

// Remove by name
chain.Remove("flatten")
```

### ParallelTransformerChain
Parallel execution of transformers with result merging:
```go
parallelChain := transformers.NewParallelTransformerChain("parallelProcessing",
    transformers.NewFieldTransformer("user", []string{"id", "name"}),
    transformers.NewFieldTransformer("profile", []string{"avatar", "bio"}),
    transformers.NewFieldTransformer("settings", []string{"theme", "language"}),
).WithMergeFunc(func(results []map[string]interface{}) map[string]interface{} {
    // Custom merge logic
    merged := make(map[string]interface{})
    for _, result := range results {
        for k, v := range result {
            merged[k] = v
        }
    }
    return merged
})
```

### ConditionalTransformerChain
Conditional transformer execution based on data content:
```go
conditionalChain := transformers.NewConditionalTransformerChain("conditional").
    When(func(data map[string]interface{}) bool {
        userType, ok := data["user_type"].(string)
        return ok && userType == "premium"
    }, transformers.NewFieldTransformer("premium", []string{
        "id", "name", "avatar", "premium_features", "subscription",
    })).
    When(func(data map[string]interface{}) bool {
        userType, ok := data["user_type"].(string)
        return ok && userType == "business"
    }, transformers.NewFieldTransformer("business", []string{
        "id", "name", "avatar", "company", "business_features",
    })).
    Otherwise(transformers.NewFieldTransformer("basic", []string{
        "id", "name", "avatar",
    }))
```

### Pre-built Chains
Common transformation patterns:
```go
// Field processing chain
fieldChain := transformers.FieldProcessingChain("fieldProcessing")

// Mobile optimization chain
mobileChain := transformers.MobileOptimizationChain()

// Data cleanup chain
cleanupChain := transformers.DataCleanupChain()
```

## Integration with Steps

### HTTP Steps Integration
Transformers integrate seamlessly with HTTP steps:
```go
httpStep := http.GET("/api/users/${userId}").
    WithTransformer(transformers.NewTransformerChain("userProcessing",
        transformers.NewFieldTransformer("select", []string{"id", "name", "profile"}),
        transformers.NewMobileTransformer([]string{"id", "name", "avatar"}),
    )).
    SaveAs("user_data")
```

### Core Steps Integration
Use transformers in core transform steps:
```go
transformStep := core.NewTransformStep("mobileTransform",
    transformers.NewMobileTransformer([]string{"id", "name", "avatar"}))
```

### BFF Steps Integration
BFF steps can use complex transformer chains:
```go
bffStep := bff.NewMobileAPIStep("profile", "GET", "/api/profile", 
    []string{"id", "name", "avatar"}).
    WithTransformer(transformers.NewTransformerChain("bffProcessing",
        transformers.NewFieldTransformer("select", []string{"user", "preferences"}),
        transformers.NewMobileTransformer([]string{"id", "name", "avatar"}),
        transformers.NewFlattenTransformer("flatten", "mobile"),
    ))
```

## Performance Considerations

### Memory Efficiency
- Deep copy operations are optimized for minimal memory allocation
- Object pooling for frequently used transformation operations
- Efficient map operations with pre-allocated capacity

### Parallel Processing
- ParallelTransformerChain uses goroutines for concurrent execution
- Proper synchronization to avoid race conditions
- Configurable merge functions for result combination

### Field Selection Optimization
- Early field filtering to reduce data processing overhead
- Efficient nested field access using optimized path resolution
- Minimal string operations for key manipulation

## Extensibility

### Custom Transformers
Create custom transformers by implementing the Transformer interface:
```go
type CustomTransformer struct {
    *transformers.BaseTransformer
    customConfig string
}

func NewCustomTransformer(name, config string) *CustomTransformer {
    return &CustomTransformer{
        BaseTransformer: transformers.NewBaseTransformer(name),
        customConfig:    config,
    }
}

func (ct *CustomTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
    // Custom transformation logic
    result := make(map[string]interface{})
    
    // Apply custom transformation based on config
    for key, value := range data {
        if ct.shouldTransform(key) {
            result[ct.transformKey(key)] = ct.transformValue(value)
        } else {
            result[key] = value
        }
    }
    
    return result, nil
}

func (ct *CustomTransformer) shouldTransform(key string) bool {
    // Custom logic to determine if field should be transformed
    return strings.Contains(ct.customConfig, key)
}

func (ct *CustomTransformer) transformKey(key string) string {
    // Custom key transformation
    return "custom_" + key
}

func (ct *CustomTransformer) transformValue(value interface{}) interface{} {
    // Custom value transformation
    return value
}
```

### Transformer Composition
Build complex transformers by composing simpler ones:
```go
func NewAdvancedMobileTransformer(fields []string, prefix string) transformers.Transformer {
    return transformers.NewTransformerChain("advancedMobile",
        transformers.NewFieldTransformer("select", fields),
        transformers.NewMobileTransformer(fields),
        transformers.NewPrefixTransformer("prefix", prefix),
        transformers.NewMobileImageTransformer(),
    )
}
```

## Best Practices

### Transformer Design
1. **Single Responsibility**: Each transformer should have one clear purpose
2. **Immutability**: Don't modify input data; always return new data structures
3. **Error Handling**: Return meaningful errors with context
4. **Performance**: Consider memory and CPU implications of transformations
5. **Naming**: Use descriptive names that indicate the transformation purpose

### Chain Design
1. **Order Matters**: Consider the order of transformers in chains
2. **Early Filtering**: Apply field selection early to reduce processing overhead
3. **Conditional Logic**: Use conditional chains for data-dependent transformations
4. **Parallel Execution**: Use parallel chains for independent transformations
5. **Merge Strategy**: Design appropriate merge functions for parallel chains

### Mobile Optimization
1. **Field Selection**: Only include fields needed by mobile clients
2. **Data Flattening**: Flatten complex nested structures for easier consumption
3. **Image Optimization**: Transform image URLs for mobile-appropriate sizes
4. **Pagination**: Include pagination metadata for list responses
5. **Error Standardization**: Use consistent error response formats

## Examples

### Complete Mobile User Profile Transformation
```go
func CreateMobileUserProfileTransformer() transformers.Transformer {
    return transformers.NewTransformerChain("mobileUserProfile",
        // Select relevant user fields
        transformers.NewFieldTransformer("userFields", []string{
            "id", "username", "email", "profile", "preferences", "subscription",
        }),
        
        // Apply mobile-specific transformations
        transformers.NewMobileTransformer([]string{
            "id", "username", "email", "profile.avatar", "profile.display_name",
            "preferences.theme", "preferences.language", "subscription.tier",
        }),
        
        // Optimize images for mobile
        transformers.NewMobileImageTransformer(),
        
        // Flatten for easier mobile consumption
        transformers.NewFlattenTransformer("mobileFlat", "user").
            WithMaxDepth(2).
            WithSeparator("_"),
        
        // Add mobile-specific metadata
        transformers.AddFieldsTransformer(map[string]interface{}{
            "mobile_optimized": true,
            "api_version": "mobile_v1",
            "timestamp": time.Now().Unix(),
        }),
    )
}
```

### Conditional Business Logic Transformation
```go
func CreateBusinessLogicTransformer() transformers.Transformer {
    return transformers.NewConditionalTransformerChain("businessLogic").
        When(func(data map[string]interface{}) bool {
            // Premium users get full data
            subscription, ok := data["subscription"].(map[string]interface{})
            if !ok {
                return false
            }
            tier, ok := subscription["tier"].(string)
            return ok && tier == "premium"
        }, transformers.NewTransformerChain("premiumTransform",
            transformers.NewFieldTransformer("premium", []string{
                "id", "name", "email", "profile", "premium_features", 
                "analytics", "advanced_settings",
            }),
            transformers.NewMobileTransformer([]string{
                "id", "name", "profile.avatar", "premium_features",
            }),
        )).
        When(func(data map[string]interface{}) bool {
            // Business users get business-specific data
            userType, ok := data["user_type"].(string)
            return ok && userType == "business"
        }, transformers.NewTransformerChain("businessTransform",
            transformers.NewFieldTransformer("business", []string{
                "id", "name", "email", "company", "business_features",
            }),
            transformers.NewMobileTransformer([]string{
                "id", "name", "company.name", "business_features",
            }),
        )).
        Otherwise(transformers.NewTransformerChain("basicTransform",
            transformers.NewFieldTransformer("basic", []string{
                "id", "name", "email", "profile.avatar",
            }),
            transformers.NewMobileTransformer([]string{
                "id", "name", "profile.avatar",
            }),
        ))
}
```

### Parallel Data Processing
```go
func CreateParallelDataProcessor() transformers.Transformer {
    return transformers.NewParallelTransformerChain("parallelProcessing",
        // Process user data
        transformers.NewTransformerChain("userData",
            transformers.NewFieldTransformer("user", []string{"id", "name", "email"}),
            transformers.NewPrefixTransformer("userPrefix", "user"),
        ),
        
        // Process profile data
        transformers.NewTransformerChain("profileData",
            transformers.NewFieldTransformer("profile", []string{"avatar", "bio", "location"}),
            transformers.NewMobileImageTransformer(),
            transformers.NewPrefixTransformer("profilePrefix", "profile"),
        ),
        
        // Process preferences data
        transformers.NewTransformerChain("preferencesData",
            transformers.NewFieldTransformer("preferences", []string{"theme", "language", "notifications"}),
            transformers.NewPrefixTransformer("prefsPrefix", "prefs"),
        ),
    ).WithMergeFunc(func(results []map[string]interface{}) map[string]interface{} {
        // Custom merge logic that combines all results
        merged := make(map[string]interface{})
        
        // Add metadata
        merged["processed_at"] = time.Now().Unix()
        merged["processing_type"] = "parallel"
        
        // Merge all results
        for _, result := range results {
            for k, v := range result {
                merged[k] = v
            }
        }
        
        return merged
    })
}
```

## Troubleshooting

### Common Issues
1. **Field Not Found**: Check field paths and nested structure access
2. **Type Assertions**: Ensure proper type checking before transformations
3. **Memory Usage**: Monitor memory usage with large data sets
4. **Chain Order**: Verify transformer order in chains
5. **Parallel Conflicts**: Check for data conflicts in parallel transformations

### Debugging
1. **Enable Logging**: Use transformer names for debugging
2. **Intermediate Results**: Log intermediate transformation results
3. **Performance Monitoring**: Track transformation execution times
4. **Data Validation**: Validate input and output data structures
5. **Error Context**: Include transformation context in error messages 