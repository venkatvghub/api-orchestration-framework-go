# Validators Package Documentation

## Overview

The `pkg/validators` package provides a comprehensive, generic validation system for data validation within the API Orchestration Framework. It offers a flexible, composable approach to validating data structures with support for field validation, type checking, custom validation logic, and chaining operations.

## Purpose

The validators package serves as the data validation engine that:
- Provides a unified interface for all data validation operations
- Enables composable validation pipelines through chaining
- Offers comprehensive field validation with nested support
- Supports type validation and custom validation logic
- Implements conditional validation based on data content
- Allows parallel validation execution for performance
- Provides detailed error reporting with context

## Core Architecture

### Validator Interface (`validator.go`)

The foundation of the validation system:
```go
type Validator interface {
    // Validate checks if the input data meets the validation criteria
    Validate(data map[string]interface{}) error
    
    // Name returns the validator name for logging and debugging
    Name() string
}
```

### BaseValidator
Common functionality for all validator implementations:
```go
type BaseValidator struct {
    name string
}

func NewBaseValidator(name string) *BaseValidator {
    return &BaseValidator{name: name}
}

func (bv *BaseValidator) Name() string {
    return bv.name
}
```

### Built-in Validator Types

#### FuncValidator
Wraps a function as a validator:
```go
funcValidator := validators.NewFuncValidator("customValidation", 
    func(data map[string]interface{}) error {
        // Custom validation logic
        if value, ok := data["custom_field"]; ok {
            if str, ok := value.(string); ok && len(str) > 0 {
                return nil
            }
        }
        return fmt.Errorf("custom_field is required and must be a non-empty string")
    })
```

#### NoOpValidator
Pass-through validator for testing or placeholder scenarios:
```go
noOpValidator := validators.NewNoOpValidator()
```

#### AlwaysFailValidator
Always fails validation (useful for testing):
```go
failValidator := validators.NewAlwaysFailValidator("This validator always fails")
```

### Validation Errors

#### ValidationError
Structured error with field context:
```go
type ValidationError struct {
    Field     string
    Value     interface{}
    Message   string
    Validator string
}

func (ve *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}
```

#### MultiValidationError
Aggregates multiple validation errors:
```go
type MultiValidationError struct {
    Errors []error
}

func (mve *MultiValidationError) Error() string {
    // Returns formatted string with all errors
}
```

## Required Fields Validation (`required.go`)

### RequiredFieldsValidator
Validates that specified fields are present and not empty:
```go
requiredValidator := validators.NewRequiredFieldsValidator("id", "email", "name").
    WithAllowEmpty(false).
    WithCustomNames(map[string]string{
        "id": "User ID",
        "email": "Email Address",
    })
```

#### Key Features:
- **Field Presence**: Checks if fields exist in data
- **Empty Value Handling**: Configurable empty value validation
- **Custom Field Names**: Custom names for error messages
- **Dynamic Field Management**: Add/remove fields at runtime

#### Usage Examples:
```go
// Basic required fields validation
basicValidator := validators.NewRequiredFieldsValidator("id", "name", "email")

// Allow empty values
allowEmptyValidator := validators.NewRequiredFieldsValidator("id", "name").
    WithAllowEmpty(true)

// Custom field names for better error messages
customNamesValidator := validators.NewRequiredFieldsValidator("user_id", "full_name").
    WithCustomNames(map[string]string{
        "user_id": "User ID",
        "full_name": "Full Name",
    })

// Dynamic field management
dynamicValidator := validators.NewRequiredFieldsValidator("id").
    AddField("email").
    AddField("name").
    RemoveField("id")
```

### RequiredStringFieldsValidator
Validates string fields with length constraints:
```go
stringValidator := validators.NewRequiredStringFieldsValidator("name", "email", "description").
    WithMinLength(3).
    WithMaxLength(100)
```

#### Features:
- **String Type Validation**: Ensures fields are strings
- **Length Constraints**: Configurable min/max length validation
- **Non-empty Validation**: Ensures strings are not empty

### RequiredNestedFieldsValidator
Validates nested field paths using dot notation:
```go
nestedValidator := validators.NewRequiredNestedFieldsValidator(
    "user.profile.name",
    "user.profile.email",
    "user.settings.theme",
)
```

#### Features:
- **Dot Notation Support**: Access nested fields with "parent.child.field"
- **Deep Validation**: Validates fields at any nesting level
- **Path Validation**: Ensures entire path exists

### ConditionalRequiredValidator
Validates fields based on conditions:
```go
conditionalValidator := validators.NewConditionalRequiredValidator().
    When(func(data map[string]interface{}) bool {
        userType, ok := data["user_type"].(string)
        return ok && userType == "business"
    }, []string{"company_name", "tax_id"}, "Business users must provide company information").
    When(func(data map[string]interface{}) bool {
        subscription, ok := data["subscription"].(string)
        return ok && subscription == "premium"
    }, []string{"payment_method", "billing_address"}, "Premium users must provide payment information")
```

### Specialized Required Validators

#### EmailRequiredValidator
Validates email fields with format checking:
```go
emailValidator := validators.EmailRequiredValidator("email", "backup_email")
```

#### IDRequiredValidator
Validates ID fields with format checking:
```go
idValidator := validators.IDRequiredValidator("user_id", "organization_id")
```

## Type Validation (`type.go`)

### TypeValidator
Validates field types with comprehensive type checking:
```go
typeValidator := validators.NewTypeValidator().
    RequireString("name", "email", "description").
    RequireInt("age", "count", "priority").
    RequireBool("active", "verified", "enabled").
    RequireFloat("price", "rating", "percentage").
    RequireArray("tags", "categories", "permissions").
    RequireObject("profile", "settings", "metadata")
```

#### Supported Types:
- **String**: Text values
- **Int**: Integer numbers
- **Float**: Floating-point numbers
- **Bool**: Boolean values
- **Array**: Slice/array values
- **Object**: Map/object values

#### Advanced Type Validation:
```go
advancedTypeValidator := validators.NewTypeValidator().
    RequireStringWithPattern("email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
    RequireIntInRange("age", 18, 120).
    RequireFloatInRange("rating", 0.0, 5.0).
    RequireArrayWithMinLength("tags", 1).
    RequireObjectWithRequiredFields("profile", []string{"name", "avatar"})
```

## Format Validation (`format.go`)

### EmailValidator
Validates email format:
```go
emailValidator := validators.NewEmailValidator("email", "backup_email").
    WithStrictMode(true).
    WithDomainWhitelist([]string{"company.com", "partner.org"})
```

### URLValidator
Validates URL format:
```go
urlValidator := validators.NewURLValidator("website", "profile_url").
    WithSchemes([]string{"https", "http"}).
    WithDomainValidation(true)
```

### PhoneValidator
Validates phone number format:
```go
phoneValidator := validators.NewPhoneValidator("phone", "mobile").
    WithCountryCode("US").
    WithInternationalFormat(true)
```

### DateValidator
Validates date format and ranges:
```go
dateValidator := validators.NewDateValidator("birth_date", "created_at").
    WithFormat("2006-01-02").
    WithMinDate(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)).
    WithMaxDate(time.Now())
```

## Range Validation (`range.go`)

### NumericRangeValidator
Validates numeric ranges:
```go
rangeValidator := validators.NewNumericRangeValidator().
    AddIntRange("age", 18, 120).
    AddFloatRange("rating", 0.0, 5.0).
    AddIntRange("priority", 1, 10)
```

### LengthValidator
Validates string and array lengths:
```go
lengthValidator := validators.NewLengthValidator().
    AddStringLength("name", 2, 50).
    AddStringLength("description", 10, 500).
    AddArrayLength("tags", 1, 10)
```

### ValueValidator
Validates values against allowed sets:
```go
valueValidator := validators.NewValueValidator().
    AddAllowedValues("status", []interface{}{"active", "inactive", "pending"}).
    AddAllowedValues("priority", []interface{}{1, 2, 3, 4, 5}).
    AddForbiddenValues("username", []interface{}{"admin", "root", "system"})
```

## Validator Chaining (`chain.go`)

### ValidatorChain
Sequential execution of multiple validators:
```go
chain := validators.NewValidatorChain("userValidation",
    validators.NewRequiredFieldsValidator("id", "email", "name"),
    validators.NewTypeValidator().RequireString("email").RequireInt("age"),
    validators.NewEmailValidator("email"),
    validators.NewNumericRangeValidator().AddIntRange("age", 18, 120),
)

// Configure chain behavior
chain.WithStopOnFirst(false).        // Continue validation after first error
    WithContinueOnError(true).       // Don't fail entire chain on single error
    WithResultStorage(true, "validation_results")  // Store detailed results
```

#### Chain Configuration:
- **StopOnFirst**: Stop validation on first error
- **ContinueOnError**: Continue chain execution even if validators fail
- **ResultStorage**: Store detailed validation results
- **Parallel Execution**: Run validators concurrently

### ParallelValidatorChain
Parallel execution of validators:
```go
parallelChain := validators.NewParallelValidatorChain("parallelValidation",
    validators.NewRequiredFieldsValidator("id", "name"),
    validators.NewTypeValidator().RequireString("email"),
    validators.NewEmailValidator("email"),
).WithErrorAggregation(true)  // Aggregate all errors
```

### ConditionalValidatorChain
Conditional validator execution:
```go
conditionalChain := validators.NewConditionalValidatorChain("conditionalValidation").
    When(func(data map[string]interface{}) bool {
        userType, ok := data["user_type"].(string)
        return ok && userType == "business"
    }, validators.NewValidatorChain("businessValidation",
        validators.NewRequiredFieldsValidator("company_name", "tax_id"),
        validators.NewTypeValidator().RequireString("company_name"),
    )).
    When(func(data map[string]interface{}) bool {
        userType, ok := data["user_type"].(string)
        return ok && userType == "individual"
    }, validators.NewValidatorChain("individualValidation",
        validators.NewRequiredFieldsValidator("first_name", "last_name"),
        validators.NewTypeValidator().RequireString("first_name", "last_name"),
    )).
    Otherwise(validators.NewRequiredFieldsValidator("name"))
```

## Integration with Steps

### Core Steps Integration
Validators integrate with validation steps:
```go
validationStep := core.NewValidationStep("validateUser",
    validators.NewValidatorChain("userValidation",
        validators.NewRequiredFieldsValidator("id", "email"),
        validators.NewTypeValidator().RequireString("email"),
        validators.NewEmailValidator("email"),
    )).
    WithDataField("user").
    WithContinueOnError(false).
    WithResultStorage(true, "validation_result")
```

### HTTP Steps Integration
Validators can validate HTTP responses:
```go
httpStep := http.GET("/api/users/${userId}").
    WithValidator(validators.NewValidatorChain("responseValidation",
        validators.NewRequiredFieldsValidator("id", "name", "status"),
        validators.NewTypeValidator().RequireString("name").RequireString("status"),
        validators.NewValueValidator().AddAllowedValues("status", 
            []interface{}{"active", "inactive", "pending"}),
    )).
    SaveAs("user_data")
```

### BFF Steps Integration
BFF steps can use complex validation chains:
```go
bffStep := bff.NewMobileAPIStep("profile", "GET", "/api/profile", 
    []string{"id", "name", "avatar"}).
    WithValidator(validators.NewValidatorChain("mobileValidation",
        validators.NewRequiredFieldsValidator("id", "name"),
        validators.NewTypeValidator().RequireString("name"),
        validators.NewLengthValidator().AddStringLength("name", 2, 50),
    ))
```

## Performance Considerations

### Validation Efficiency
- Early termination on critical validation failures
- Parallel validation for independent checks
- Efficient field access with minimal reflection
- Optimized error aggregation

### Memory Management
- Reusable validator instances
- Efficient error message formatting
- Minimal memory allocation for validation results
- Object pooling for frequently used validators

### Caching
- Compiled regex patterns for format validation
- Cached validation results for expensive operations
- Memoized validation functions

## Extensibility

### Custom Validators
Create custom validators by implementing the Validator interface:
```go
type CustomBusinessValidator struct {
    *validators.BaseValidator
    businessRules map[string]interface{}
}

func NewCustomBusinessValidator(name string, rules map[string]interface{}) *CustomBusinessValidator {
    return &CustomBusinessValidator{
        BaseValidator: validators.NewBaseValidator(name),
        businessRules: rules,
    }
}

func (cbv *CustomBusinessValidator) Validate(data map[string]interface{}) error {
    // Custom business logic validation
    for field, rule := range cbv.businessRules {
        if err := cbv.validateBusinessRule(data, field, rule); err != nil {
            return validators.NewValidationError(field, data[field], err.Error(), cbv.Name())
        }
    }
    return nil
}

func (cbv *CustomBusinessValidator) validateBusinessRule(data map[string]interface{}, field string, rule interface{}) error {
    // Implement custom business rule validation
    value, exists := data[field]
    if !exists {
        return fmt.Errorf("required business field missing")
    }
    
    // Apply business rule logic
    switch r := rule.(type) {
    case string:
        return cbv.validateStringRule(value, r)
    case int:
        return cbv.validateNumericRule(value, r)
    default:
        return fmt.Errorf("unsupported business rule type")
    }
}
```

### Validator Composition
Build complex validators by composing simpler ones:
```go
func NewUserRegistrationValidator() validators.Validator {
    return validators.NewValidatorChain("userRegistration",
        // Basic field validation
        validators.NewRequiredFieldsValidator("email", "password", "name"),
        
        // Type validation
        validators.NewTypeValidator().
            RequireString("email", "password", "name").
            RequireInt("age"),
        
        // Format validation
        validators.NewEmailValidator("email"),
        
        // Business rules
        validators.NewLengthValidator().
            AddStringLength("password", 8, 128).
            AddStringLength("name", 2, 50),
        
        // Range validation
        validators.NewNumericRangeValidator().
            AddIntRange("age", 13, 120),
        
        // Custom business validation
        NewCustomBusinessValidator("businessRules", map[string]interface{}{
            "email": "unique_email_rule",
            "username": "unique_username_rule",
        }),
    )
}
```

## Best Practices

### Validator Design
1. **Single Responsibility**: Each validator should validate one specific aspect
2. **Clear Error Messages**: Provide descriptive, actionable error messages
3. **Performance**: Consider validation performance for large datasets
4. **Reusability**: Design validators to be reusable across different contexts
5. **Composability**: Build complex validators from simpler components

### Chain Design
1. **Logical Order**: Order validators from basic to complex
2. **Early Termination**: Use stop-on-first for critical validations
3. **Error Aggregation**: Collect all errors for better user experience
4. **Conditional Logic**: Use conditional chains for context-dependent validation
5. **Parallel Execution**: Use parallel chains for independent validations

### Error Handling
1. **Structured Errors**: Use ValidationError for consistent error format
2. **Field Context**: Include field names and values in error messages
3. **Error Aggregation**: Collect multiple errors when appropriate
4. **User-Friendly Messages**: Provide clear, actionable error messages
5. **Localization**: Support localized error messages

## Examples

### Complete User Validation
```go
func CreateUserValidator() validators.Validator {
    return validators.NewValidatorChain("completeUserValidation",
        // Required fields
        validators.NewRequiredFieldsValidator("id", "email", "name", "profile").
            WithCustomNames(map[string]string{
                "id": "User ID",
                "email": "Email Address",
                "name": "Full Name",
            }),
        
        // Type validation
        validators.NewTypeValidator().
            RequireString("email", "name").
            RequireInt("id").
            RequireObject("profile"),
        
        // Nested field validation
        validators.NewRequiredNestedFieldsValidator(
            "profile.first_name",
            "profile.last_name",
            "profile.birth_date",
        ),
        
        // Format validation
        validators.NewEmailValidator("email").WithStrictMode(true),
        validators.NewDateValidator("profile.birth_date").
            WithFormat("2006-01-02").
            WithMaxDate(time.Now().AddDate(-13, 0, 0)), // Minimum age 13
        
        // Length validation
        validators.NewLengthValidator().
            AddStringLength("name", 2, 100).
            AddStringLength("profile.first_name", 1, 50).
            AddStringLength("profile.last_name", 1, 50),
        
        // Custom business validation
        validators.NewFuncValidator("businessRules", func(data map[string]interface{}) error {
            // Custom business logic
            email := data["email"].(string)
            if strings.Contains(email, "+") {
                return fmt.Errorf("email addresses with '+' are not allowed")
            }
            return nil
        }),
    )
}
```

### Conditional Business Validation
```go
func CreateBusinessUserValidator() validators.Validator {
    return validators.NewConditionalValidatorChain("businessUserValidation").
        When(func(data map[string]interface{}) bool {
            userType, ok := data["user_type"].(string)
            return ok && userType == "business"
        }, validators.NewValidatorChain("businessValidation",
            validators.NewRequiredFieldsValidator(
                "company_name", "tax_id", "business_address", "contact_person",
            ),
            validators.NewTypeValidator().
                RequireString("company_name", "tax_id", "contact_person").
                RequireObject("business_address"),
            validators.NewLengthValidator().
                AddStringLength("company_name", 2, 200).
                AddStringLength("tax_id", 5, 50),
            validators.NewRequiredNestedFieldsValidator(
                "business_address.street",
                "business_address.city",
                "business_address.country",
            ),
        )).
        When(func(data map[string]interface{}) bool {
            userType, ok := data["user_type"].(string)
            return ok && userType == "individual"
        }, validators.NewValidatorChain("individualValidation",
            validators.NewRequiredFieldsValidator("first_name", "last_name", "birth_date"),
            validators.NewTypeValidator().RequireString("first_name", "last_name"),
            validators.NewDateValidator("birth_date").WithFormat("2006-01-02"),
        )).
        Otherwise(validators.NewValidatorChain("defaultValidation",
            validators.NewRequiredFieldsValidator("name"),
            validators.NewTypeValidator().RequireString("name"),
        ))
}
```

### API Response Validation
```go
func CreateAPIResponseValidator() validators.Validator {
    return validators.NewParallelValidatorChain("apiResponseValidation",
        // Validate response structure
        validators.NewValidatorChain("structureValidation",
            validators.NewRequiredFieldsValidator("status", "data", "metadata"),
            validators.NewTypeValidator().
                RequireString("status").
                RequireObject("data", "metadata"),
            validators.NewValueValidator().
                AddAllowedValues("status", []interface{}{"success", "error", "partial"}),
        ),
        
        // Validate data content
        validators.NewValidatorChain("dataValidation",
            validators.NewRequiredNestedFieldsValidator("data.id", "data.name"),
            validators.NewTypeValidator().
                RequireString("data.name").
                RequireInt("data.id"),
        ),
        
        // Validate metadata
        validators.NewValidatorChain("metadataValidation",
            validators.NewRequiredNestedFieldsValidator("metadata.timestamp", "metadata.version"),
            validators.NewTypeValidator().
                RequireInt("metadata.timestamp").
                RequireString("metadata.version"),
        ),
    ).WithErrorAggregation(true)
}
```

## Troubleshooting

### Common Issues
1. **Field Not Found**: Check field paths and nested structure access
2. **Type Mismatches**: Ensure proper type checking before validation
3. **Performance Issues**: Monitor validation performance with large datasets
4. **Chain Order**: Verify validator order in chains
5. **Error Aggregation**: Check error collection and reporting

### Debugging
1. **Enable Logging**: Use validator names for debugging
2. **Validation Results**: Log detailed validation results
3. **Performance Monitoring**: Track validation execution times
4. **Data Inspection**: Log input data for validation debugging
5. **Error Context**: Include validation context in error messages

### Performance Tuning
1. **Early Termination**: Use stop-on-first for critical validations
2. **Parallel Validation**: Use parallel chains for independent checks
3. **Caching**: Cache expensive validation results
4. **Field Access**: Optimize nested field access patterns
5. **Memory Usage**: Monitor memory usage with large validation chains 