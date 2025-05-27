# Getting Started with API Orchestration Framework

A beginner-friendly guide to building Backend for Frontend (BFF) APIs with the Go API Orchestration Framework.

## What is this Framework?

This framework helps you build APIs that combine data from multiple services into a single response - perfect for mobile apps that need data from different sources. Think of it as a smart middleman that fetches, combines, and optimizes data for your frontend.

## Installation

```bash
go get github.com/venkatvghub/api-orchestration-framework
```

## Your First Flow (5 minutes)

Let's start with the simplest possible example - fetching user data:

```go
package main

import (
    "log"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
)

func main() {
    // Create a simple flow that fetches user data
    userFlow := flow.NewFlow("GetUser").
        Step("fetchUser", flow.NewStepWrapper(
            http.GET("https://jsonplaceholder.typicode.com/users/1").
                SaveAs("user"),
        ))

    // Execute the flow
    ctx := flow.NewContext()
    result, err := userFlow.Execute(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("User data: %+v", result.GetResponse())
}
```

**What's happening here?**
1. We create a "flow" (a sequence of steps)
2. We add one step that makes an HTTP GET request
3. We save the response as "user"
4. We execute the flow and print the result

## Adding Authentication (10 minutes)

Most real APIs need authentication. Let's add token validation:

```go
package main

import (
    "log"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
)

func main() {
    // Create a flow with authentication
    userFlow := flow.NewFlow("GetUserWithAuth").
        // Step 1: Validate the token
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        // Step 2: Fetch user data
        Step("fetchUser", flow.NewStepWrapper(
            http.GET("https://api.example.com/users/${userId}").
                WithHeader("Authorization", "Bearer ${token}").
                SaveAs("user"),
        ))

    // Execute with authentication data
    ctx := flow.NewContext()
    ctx.Set("userId", "123")
    ctx.Set("token", "your-auth-token")
    
    result, err := userFlow.Execute(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("User data: %+v", result.GetResponse())
}
```

**What's new?**
- We added authentication validation before fetching data
- We use `${userId}` and `${token}` to insert values from our context
- Steps run in order: auth first, then fetch

## Fetching Multiple Things at Once (15 minutes)

Mobile apps often need data from multiple sources. Let's fetch user profile, notifications, and settings simultaneously:

```go
package main

import (
    "log"
    "time"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
)

func main() {
    // Create a flow that fetches multiple things in parallel
    dashboardFlow := flow.NewFlow("MobileDashboard").
        WithTimeout(10 * time.Second).
        
        // Step 1: Authenticate
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        // Step 2: Fetch multiple things at the same time
        Step("fetchData", flow.NewStepWrapper(
            base.NewParallelStep("dashboardData",
                // These all run at the same time
                http.GET("https://api.example.com/users/${userId}").SaveAs("profile"),
                http.GET("https://api.example.com/notifications").SaveAs("notifications"),
                http.GET("https://api.example.com/settings").SaveAs("settings"),
            ),
        ))

    // Execute the flow
    ctx := flow.NewContext()
    ctx.Set("userId", "123")
    ctx.Set("Authorization", "Bearer your-token")
    
    result, err := dashboardFlow.Execute(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Now you have profile, notifications, and settings all in one response
    log.Printf("Dashboard data: %+v", result.GetResponse())
}
```

**What's new?**
- `base.NewParallelStep()` runs multiple HTTP requests at the same time
- This is much faster than doing them one by one
- All the data gets combined into a single response

## Making it Mobile-Friendly (20 minutes)

Mobile apps need smaller, optimized responses. Let's add field selection and caching:

```go
package main

import (
    "log"
    "time"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
    "github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
    "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

func main() {
    // Setup monitoring (optional but recommended)
    metrics.SetupDefaultPrometheusMetrics()
    
    // Create a mobile-optimized flow
    mobileFlow := flow.NewFlow("MobileUserProfile").
        WithDescription("Optimized user profile for mobile").
        WithTimeout(15 * time.Second).
        
        // Step 1: Authenticate
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        // Step 2: Fetch user data with mobile optimization
        Step("fetchUser", flow.NewStepWrapper(
            http.NewMobileAPIStep("profile", "GET", "https://api.example.com/users/${userId}",
                []string{"id", "name", "avatar", "email"}). // Only get these fields
                WithCaching("user_profile", 5*time.Minute).  // Cache for 5 minutes
                SaveAs("user"),
        )).
        
        // Step 3: Transform data for mobile
        Step("optimize", flow.NewStepWrapper(
            core.NewTransformStep("mobileTransform",
                transformers.NewMobileTransformer([]string{"id", "name", "avatar"})),
        ))

    // Execute the flow
    ctx := flow.NewContext()
    ctx.Set("userId", "123")
    ctx.Set("Authorization", "Bearer your-token")
    
    result, err := mobileFlow.Execute(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Mobile-optimized response: %+v", result.GetResponse())
}
```

**What's new?**
- `http.NewMobileAPIStep()` only fetches the fields you specify
- `WithCaching()` caches responses to make subsequent requests faster
- `transformers.NewMobileTransformer()` optimizes data for mobile apps
- Smaller responses = faster mobile apps

## Handling Different User Types (25 minutes)

Let's show different data based on user type (free vs premium):

```go
package main

import (
    "log"
    "time"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
)

func main() {
    // Create a flow with conditional logic
    userFlow := flow.NewFlow("ConditionalUserData").
        WithTimeout(15 * time.Second).
        
        // Step 1: Authenticate
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        // Step 2: Get user info
        Step("user", flow.NewStepWrapper(
            http.GET("https://api.example.com/users/${userId}").SaveAs("user"),
        )).
        
        // Step 3: Conditional data based on user type
        Choice("userType").
            When(func(ctx *flow.Context) bool {
                user, _ := ctx.GetMap("user")
                return user["subscription"] == "premium"
            }).
                // Premium users get extra data
                Step("premiumData", flow.NewStepWrapper(
                    http.GET("https://api.example.com/premium/features").SaveAs("features"),
                )).
            Otherwise().
                // Free users get basic data
                Step("freeData", flow.NewStepWrapper(
                    http.GET("https://api.example.com/basic/features").SaveAs("features"),
                )).
        EndChoice()

    // Execute the flow
    ctx := flow.NewContext()
    ctx.Set("userId", "123")
    ctx.Set("Authorization", "Bearer your-token")
    
    result, err := userFlow.Execute(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("User-specific data: %+v", result.GetResponse())
}
```

**What's new?**
- `Choice()` lets you run different steps based on conditions
- `When()` checks if something is true
- `Otherwise()` runs if the condition is false
- Perfect for showing different features to different user types

## Adding Error Handling (30 minutes)

Real apps need to handle failures gracefully:

```go
package main

import (
    "log"
    "time"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
)

func main() {
    // Create a resilient flow that handles failures
    resilientFlow := flow.NewFlow("ResilientUserData").
        WithTimeout(20 * time.Second).
        
        // Step 1: Authenticate
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        // Step 2: Try to get user data with retries
        Step("userData", flow.NewStepWrapper(
            base.NewRetryStep("retryUserFetch",
                http.GET("https://api.example.com/users/${userId}").SaveAs("user"),
                3,                    // Retry 3 times
                2*time.Second),       // Wait 2 seconds between retries
        )).
        
        // Step 3: Try to get optional data (won't fail the whole flow)
        Step("optionalData", flow.NewStepWrapper(
            http.GET("https://api.example.com/optional-data").
                WithFallback(map[string]interface{}{
                    "message": "Optional data not available",
                }).
                SaveAs("optional"),
        ))

    // Execute the flow
    ctx := flow.NewContext()
    ctx.Set("userId", "123")
    ctx.Set("Authorization", "Bearer your-token")
    
    result, err := userFlow.Execute(ctx)
    if err != nil {
        log.Printf("Flow failed: %v", err)
        return
    }

    log.Printf("Resilient response: %+v", result.GetResponse())
}
```

**What's new?**
- `base.NewRetryStep()` automatically retries failed requests
- `WithFallback()` provides default data if a request fails
- Your app keeps working even when some services are down

## Configuration and Monitoring

### Environment Variables
Set these in your environment for production:

```bash
# How many times to retry failed requests
export HTTP_MAX_RETRIES=3

# How long to wait for responses
export HTTP_REQUEST_TIMEOUT=15s

# Enable fallback responses
export HTTP_ENABLE_FALLBACK=true

# How long to cache data
export CACHE_DEFAULT_TTL=5m

# Log level (debug, info, warn, error)
export LOG_LEVEL=info
```

### Monitoring Your Flows
Add this to see how your flows are performing:

```go
import (
    "net/http"
    "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

func main() {
    // Setup metrics collection
    metrics.SetupDefaultPrometheusMetrics()
    
    // Expose metrics at /metrics endpoint
    http.Handle("/metrics", metrics.MetricsHandler())
    go http.ListenAndServe(":9090", nil)
    
    // Your flows will now automatically collect metrics
    // Visit http://localhost:9090/metrics to see them
}
```

## Common Patterns

### 1. Mobile Screen Data
```go
// Get all data needed for a mobile screen
screenFlow := flow.NewFlow("HomeScreen").
    Step("auth", authStep).
    Step("data", parallelDataStep).
    Step("optimize", mobileTransformStep)
```

### 2. User-Specific Features
```go
// Show different features based on user type
featureFlow := flow.NewFlow("UserFeatures").
    Step("user", getUserStep).
    Choice("userType").
        When(isPremium).Step("premium", premiumStep).
        Otherwise().Step("basic", basicStep).
    EndChoice()
```

### 3. Resilient Data Fetching
```go
// Handle failures gracefully
resilientFlow := flow.NewFlow("ResilientData").
    Step("required", retryStep).
    Step("optional", fallbackStep)
```

## Next Steps

Now that you understand the basics, explore these advanced topics:

1. **[BFF Patterns](bff-patterns.md)** - Learn mobile-specific patterns
2. **[Steps Reference](steps.md)** - All available step types
3. **[Transformers Guide](transformers.md)** - Data transformation
4. **[Error Handling](errors.md)** - Advanced error handling
5. **[Configuration Guide](config.md)** - Production configuration

## Common Questions

**Q: When should I use this framework?**
A: When you need to combine data from multiple APIs into a single response, especially for mobile apps.

**Q: How is this different from a regular API?**
A: Instead of making multiple API calls from your mobile app, you make one call to your BFF, which handles all the complexity.

**Q: Can I use this in production?**
A: Yes! The framework includes monitoring, error handling, caching, and other production features.

**Q: How do I test my flows?**
A: Create test contexts and mock your external API calls. See the examples folder for testing patterns.

**Q: What if I need custom logic?**
A: You can create custom steps, transformers, and validators. The framework is designed to be extensible.

## Getting Help

- Check the [examples](../examples/) folder for complete working examples
- Read the detailed guides in the [docs](.) folder
- Look at the test files to see how components work
- Create an issue if you find bugs or need features

Happy coding! 🚀 