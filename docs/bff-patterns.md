# BFF Patterns for Mobile Applications

This guide covers common Backend for Frontend (BFF) patterns implemented using the API Orchestration Framework, specifically optimized for mobile applications.

## Table of Contents

1. [Screen-Level Orchestration](#screen-level-orchestration)
2. [User-Specific Data Loading](#user-specific-data-loading)
3. [Progressive Data Loading](#progressive-data-loading)
4. [Caching Strategies](#caching-strategies)
5. [Error Handling Patterns](#error-handling-patterns)
6. [Authentication Patterns](#authentication-patterns)
7. [Data Transformation Patterns](#data-transformation-patterns)

## Screen-Level Orchestration

### Pattern: Single Screen Data Aggregation

Mobile screens often require data from multiple backend services. This pattern aggregates all required data in a single API call.

```go
func CreateHomeScreenFlow() *flow.Flow {
    return flow.NewFlow("HomeScreen").
        WithDescription("Aggregate data for mobile home screen").
        WithTimeout(10 * time.Second).
        
        // Authentication
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization").
                WithClaimsExtraction(true),
        )).
        
        // Fetch user context
        Step("user", flow.NewStepWrapper(
            http.GET("${baseURL}/users/${userId}").
                WithHeader("Authorization", "Bearer ${authToken}").
                SaveAs("user"),
        )).
        
        // Parallel data fetching for screen components
        Step("screenComponents", flow.NewStepWrapper(
            base.NewParallelStep("screenData",
                // Notifications
                http.NewMobileAPIStep("notifications", "GET", "${baseURL}/notifications",
                    []string{"id", "title", "type", "timestamp"}).
                    WithHeader("Authorization", "Bearer ${authToken}").
                    SaveAs("notifications"),
                
                // Feed
                http.NewMobileAPIStep("feed", "GET", "${baseURL}/feed?limit=10",
                    []string{"id", "content", "author", "timestamp", "likes"}).
                    WithHeader("Authorization", "Bearer ${authToken}").
                    SaveAs("feed"),
                
                // Weather
                http.GET("${weatherAPI}/current?lat=${user.location.lat}&lon=${user.location.lon}").
                    WithHeader("API-Key", "${weatherAPIKey}").
                    SaveAs("weather"),
            ),
        )).
        
        // Transform for mobile consumption
        Step("mobileResponse", flow.NewStepWrapper(
            core.NewTransformStep("homeScreenTransform", createHomeScreenTransformer()),
        ))
}

func createHomeScreenTransformer() transformers.Transformer {
    return transformers.NewFunctionTransformer("homeScreen", func(data map[string]interface{}) (map[string]interface{}, error) {
        user, _ := data["user"].(map[string]interface{})
        notifications, _ := data["notifications"].(map[string]interface{})
        feed, _ := data["feed"].(map[string]interface{})
        weather, _ := data["weather"].(map[string]interface{})
        
        response := map[string]interface{}{
            "user": map[string]interface{}{
                "name":   user["name"],
                "avatar": user["avatar_url"],
            },
            "summary": map[string]interface{}{
                "unread_notifications": len(notifications["items"].([]interface{})),
                "new_feed_items":      len(feed["items"].([]interface{})),
                "weather":             weather["current"],
            },
            "quick_actions": []map[string]interface{}{
                {"type": "notifications", "count": len(notifications["items"].([]interface{}))},
                {"type": "messages", "count": 0},
                {"type": "calendar", "count": 0},
            },
            "feed_preview": feed["items"],
            "timestamp":    time.Now().Unix(),
        }
        
        return map[string]interface{}{"response": response}, nil
    })
}
```

### Pattern: Tab-Based Data Loading

For tabbed interfaces, load data specific to each tab:

```go
func CreateTabDataFlow(tabName string) *flow.Flow {
    return flow.NewFlow(fmt.Sprintf("TabData_%s", tabName)).
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        Choice("tabType").
            When(func(ctx *flow.Context) bool { return tabName == "home" }).
                Step("homeData", flow.NewStepWrapper(createHomeTabStep())).
            When(func(ctx *flow.Context) bool { return tabName == "profile" }).
                Step("profileData", flow.NewStepWrapper(createProfileTabStep())).
            When(func(ctx *flow.Context) bool { return tabName == "settings" }).
                Step("settingsData", flow.NewStepWrapper(createSettingsTabStep())).
            Otherwise().
                Step("defaultData", flow.NewStepWrapper(
                    core.NewSetValueStep("error", "error", "Unknown tab"),
                )).
        EndChoice()
}

func createHomeTabStep() interfaces.Step {
    return bff.NewMobileDashboardAggregation("${baseURL}").
        WithParallel(true).
        WithTimeout(10 * time.Second)
}

func createProfileTabStep() interfaces.Step {
    return bff.NewMobileUserProfileStep("${baseURL}").
        WithCaching("user_profile", 10*time.Minute)
}

func createSettingsTabStep() interfaces.Step {
    return http.NewMobileAPIStep("settings", "GET", "${baseURL}/settings",
        []string{"theme", "notifications", "privacy"})
}
```

## User-Specific Data Loading

### Pattern: Role-Based Data Access

Load different data based on user roles and permissions:

```go
func CreateUserDashboardFlow() *flow.Flow {
    return flow.NewFlow("UserDashboard").
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization").
                WithClaimsExtraction(true),
        )).
        Step("user", flow.NewStepWrapper(
            http.GET("${baseURL}/users/${userId}").SaveAs("user"),
        )).
        
        Choice("userRole").
            When(func(ctx *flow.Context) bool {
                user, _ := ctx.GetMap("user")
                return user["role"] == "admin"
            }).
                Step("adminData", flow.NewStepWrapper(
                    base.NewParallelStep("adminData",
                        http.GET("${baseURL}/admin/users").SaveAs("users"),
                        http.GET("${baseURL}/admin/analytics").SaveAs("analytics"),
                        http.GET("${baseURL}/admin/reports").SaveAs("reports"),
                    ),
                )).
                
            When(func(ctx *flow.Context) bool {
                user, _ := ctx.GetMap("user")
                return user["role"] == "manager"
            }).
                Step("managerData", flow.NewStepWrapper(
                    base.NewParallelStep("managerData",
                        http.GET("${baseURL}/teams/${user.team_id}").SaveAs("team"),
                        http.GET("${baseURL}/projects?manager=${userId}").SaveAs("projects"),
                    ),
                )).
                
            Otherwise().
                Step("userData", flow.NewStepWrapper(
                    base.NewParallelStep("userData",
                        http.GET("${baseURL}/profile").SaveAs("profile"),
                        http.GET("${baseURL}/tasks?assignee=${userId}").SaveAs("tasks"),
                    ),
                )).
        EndChoice().
        
        Step("roleBasedResponse", flow.NewStepWrapper(
            core.NewTransformStep("roleTransform", createRoleBasedTransformer()),
        ))
}

func createRoleBasedTransformer() transformers.Transformer {
    return transformers.NewFunctionTransformer("roleBasedTransform", func(data map[string]interface{}) (map[string]interface{}, error) {
        user, _ := data["user"].(map[string]interface{})
        role := user["role"].(string)
        
        response := map[string]interface{}{
            "user": user,
            "role": role,
        }
        
        switch role {
        case "admin":
            response["admin_data"] = map[string]interface{}{
                "users":     data["users"],
                "analytics": data["analytics"],
                "reports":   data["reports"],
            }
        case "manager":
            response["manager_data"] = map[string]interface{}{
                "team":     data["team"],
                "projects": data["projects"],
            }
        default:
            response["user_data"] = map[string]interface{}{
                "profile": data["profile"],
                "tasks":   data["tasks"],
            }
        }
        
        return response, nil
    })
}
```

### Pattern: Subscription-Based Features

Load features based on user subscription level:

```go
func CreateSubscriptionBasedFlow() *flow.Flow {
    return flow.NewFlow("SubscriptionFeatures").
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        Step("subscription", flow.NewStepWrapper(
            http.GET("${baseURL}/subscriptions/${userId}").SaveAs("subscription"),
        )).
        
        Choice("subscriptionLevel").
            When(func(ctx *flow.Context) bool {
                sub, _ := ctx.GetMap("subscription")
                return sub["tier"] == "premium"
            }).
                Step("premiumFeatures", flow.NewStepWrapper(
                    base.NewParallelStep("premiumData",
                        http.GET("${baseURL}/premium/analytics").SaveAs("analytics"),
                        http.GET("${baseURL}/premium/reports").SaveAs("reports"),
                        http.GET("${baseURL}/premium/integrations").SaveAs("integrations"),
                    ),
                )).
                
            When(func(ctx *flow.Context) bool {
                sub, _ := ctx.GetMap("subscription")
                return sub["tier"] == "pro"
            }).
                Step("proFeatures", flow.NewStepWrapper(
                    http.GET("${baseURL}/pro/features").SaveAs("proFeatures"),
                )).
                
            Otherwise().
                Step("basicFeatures", flow.NewStepWrapper(
                    http.GET("${baseURL}/basic/features").SaveAs("basicFeatures"),
                )).
        EndChoice()
}
```

## Progressive Data Loading

### Pattern: Critical First, Optional Later

Load critical data first, then optional data:

```go
func CreateProgressiveLoadFlow() *flow.Flow {
    return flow.NewFlow("ProgressiveLoad").
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        // Critical data first
        Step("criticalData", flow.NewStepWrapper(
            base.NewParallelStep("critical",
                http.GET("${baseURL}/users/${userId}").SaveAs("user"),
                http.GET("${baseURL}/notifications?urgent=true").SaveAs("urgentNotifications"),
            ),
        )).
        
        // Send initial response
        Step("initialResponse", flow.NewStepWrapper(
            core.NewTransformStep("initialTransform", transformers.NewFunctionTransformer("initial", func(data map[string]interface{}) (map[string]interface{}, error) {
                user, _ := data["user"].(map[string]interface{})
                notifications, _ := data["urgentNotifications"].(map[string]interface{})
                
                initial := map[string]interface{}{
                    "user":                user,
                    "urgent_notifications": notifications,
                    "loading_additional":   true,
                }
                
                return map[string]interface{}{"initial_response": initial}, nil
            })),
        )).
        
        // Optional data (can be loaded asynchronously)
        Step("optionalData", flow.NewStepWrapper(
            base.NewParallelStep("optional",
                http.GET("${baseURL}/feed").SaveAs("feed"),
                http.GET("${baseURL}/recommendations").SaveAs("recommendations"),
                http.GET("${weatherAPI}/current").SaveAs("weather"),
            ),
        )).
        
        Step("completeResponse", flow.NewStepWrapper(
            core.NewTransformStep("completeTransform", transformers.NewFunctionTransformer("complete", func(data map[string]interface{}) (map[string]interface{}, error) {
                initial, _ := data["initial_response"].(map[string]interface{})
                feed, _ := data["feed"].(map[string]interface{})
                recommendations, _ := data["recommendations"].(map[string]interface{})
                weather, _ := data["weather"].(map[string]interface{})
                
                complete := initial
                complete["feed"] = feed
                complete["recommendations"] = recommendations
                complete["weather"] = weather
                complete["loading_additional"] = false
                
                return map[string]interface{}{"response": complete}, nil
            })),
        ))
}
```

## Caching Strategies

### Pattern: Multi-Level Caching

Implement caching at different levels:

```go
func CreateCachedDataFlow() *flow.Flow {
    return flow.NewFlow("CachedData").
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        // Check cache first
        Step("checkCache", flow.NewStepWrapper(
            core.NewCacheGetStep("checkUserCache", "user_data_${userId}", "cached_user_data"),
        )).
        
        Choice("cacheHit").
            When(func(ctx *flow.Context) bool {
                return ctx.Has("cached_user_data")
            }).
                Step("log", flow.NewStepWrapper(
                    core.NewLogStep("cacheHit", "info", "Cache hit for user data"),
                )).
            Otherwise().
                // Cache miss - fetch from API
                Step("fetchUser", flow.NewStepWrapper(
                    http.GET("${baseURL}/users/${userId}").SaveAs("user_data"),
                )).
                Step("cacheUser", flow.NewStepWrapper(
                    core.NewCacheSetStep("cacheUserData", "user_data_${userId}", "user_data", 5*time.Minute),
                )).
        EndChoice().
        
        Step("response", flow.NewStepWrapper(
            core.NewTransformStep("responseTransform", transformers.NewFunctionTransformer("response", func(data map[string]interface{}) (map[string]interface{}, error) {
                var userData interface{}
                if cachedData, ok := data["cached_user_data"]; ok {
                    userData = cachedData
                } else {
                    userData = data["user_data"]
                }
                return map[string]interface{}{"response": userData}, nil
            })),
        ))
}
```

## Error Handling Patterns

### Pattern: Graceful Degradation

Handle service failures gracefully:

```go
func CreateResilientFlow() *flow.Flow {
    return flow.NewFlow("ResilientData").
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        // Core data (must succeed)
        Step("coreData", flow.NewStepWrapper(
            base.NewRetryStep("coreDataRetry",
                http.GET("${baseURL}/core").SaveAs("core"),
                3, 1*time.Second),
        )).
        
        // Optional data (can fail gracefully)
        Step("optionalData", flow.NewStepWrapper(
            base.NewParallelStep("optionalData",
                createOptionalStep("recommendations", "${baseURL}/recommendations"),
                createOptionalStep("social", "${baseURL}/social"),
                createOptionalStep("analytics", "${baseURL}/analytics"),
            ),
        )).
        
        Step("resilientResponse", flow.NewStepWrapper(
            core.NewTransformStep("resilientTransform", transformers.NewFunctionTransformer("resilient", func(data map[string]interface{}) (map[string]interface{}, error) {
                core, _ := data["core"].(map[string]interface{})
                
                response := map[string]interface{}{
                    "core": core,
                    "optional": map[string]interface{}{},
                }
                
                // Add optional data if available
                if recommendations, ok := data["recommendations"]; ok {
                    response["optional"].(map[string]interface{})["recommendations"] = recommendations
                }
                if social, ok := data["social"]; ok {
                    response["optional"].(map[string]interface{})["social"] = social
                }
                if analytics, ok := data["analytics"]; ok {
                    response["optional"].(map[string]interface{})["analytics"] = analytics
                }
                
                return response, nil
            })),
        ))
}

func createOptionalStep(name, url string) interfaces.Step {
    return &OptionalStep{
        name: name,
        step: http.GET(url).SaveAs(name),
    }
}

type OptionalStep struct {
    name string
    step interfaces.Step
}

func (os *OptionalStep) Run(ctx interfaces.ExecutionContext) error {
    // Try to fetch data, but don't fail the flow if it fails
    if err := os.step.Run(ctx); err != nil {
        ctx.Logger().Warn("Optional step failed", 
            zap.String("step", os.name), 
            zap.Error(err))
        // Don't return error - continue flow
    }
    return nil
}

func (os *OptionalStep) Name() string {
    return os.name
}

func (os *OptionalStep) Description() string {
    return fmt.Sprintf("Optional step: %s", os.name)
}
```

## Authentication Patterns

### Pattern: Token Refresh

Handle token expiration automatically:

```go
func CreateTokenRefreshFlow() *flow.Flow {
    return flow.NewFlow("TokenRefresh").
        Step("validateToken", flow.NewStepWrapper(
            core.NewTokenValidationStep("validateToken", "Authorization").
                WithValidationFunc(func(token string) bool {
                    return !isTokenExpired(token)
                }),
        )).
        
        Choice("tokenStatus").
            When(func(ctx *flow.Context) bool {
                // Check if token validation failed
                return ctx.Has("token_error")
            }).
                Step("refreshToken", flow.NewStepWrapper(
                    http.POST("${authURL}/refresh").
                        WithJSONBody(map[string]interface{}{
                            "refresh_token": "${refreshToken}",
                        }).
                        SaveAs("newTokens"),
                )).
                Step("updateContext", flow.NewStepWrapper(
                    core.NewTransformStep("updateTokens", transformers.NewFunctionTransformer("updateTokens", func(data map[string]interface{}) (map[string]interface{}, error) {
                        tokens, _ := data["newTokens"].(map[string]interface{})
                        return map[string]interface{}{
                            "authToken": tokens["access_token"],
                        }, nil
                    })),
                )).
            Otherwise().
                Step("log", flow.NewStepWrapper(
                    core.NewLogStep("tokenValid", "info", "Token is valid"),
                )).
        EndChoice()
}

func isTokenExpired(token string) bool {
    // Custom token expiration logic
    return false // Placeholder
}
```

## Data Transformation Patterns

### Pattern: Mobile-Optimized Responses

Transform complex backend responses for mobile consumption:

```go
func CreateMobileOptimizedFlow() *flow.Flow {
    return flow.NewFlow("MobileOptimized").
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        
        Step("rawData", flow.NewStepWrapper(
            base.NewParallelStep("rawData",
                http.GET("${baseURL}/users/${userId}").SaveAs("rawUser"),
                http.GET("${baseURL}/posts?user=${userId}").SaveAs("rawPosts"),
                http.GET("${baseURL}/followers/${userId}").SaveAs("rawFollowers"),
            ),
        )).
        
        Step("mobileOptimization", flow.NewStepWrapper(
            core.NewTransformStep("mobileTransform", 
                transformers.NewTransformerChain("mobileChain",
                    transformers.NewFieldTransformer("selectFields", []string{"rawUser", "rawPosts", "rawFollowers"}),
                    transformers.NewMobileTransformer([]string{"id", "name", "avatar"}),
                    transformers.NewFunctionTransformer("mobileOptimize", func(data map[string]interface{}) (map[string]interface{}, error) {
                        rawUser, _ := data["rawUser"].(map[string]interface{})
                        rawPosts, _ := data["rawPosts"].(map[string]interface{})
                        rawFollowers, _ := data["rawFollowers"].(map[string]interface{})
                        
                        // Mobile-optimized response
                        mobile := map[string]interface{}{
                            "profile": map[string]interface{}{
                                "id":     rawUser["id"],
                                "name":   rawUser["full_name"],
                                "avatar": rawUser["profile_image_url"],
                                "bio":    truncateString(rawUser["biography"].(string), 100),
                            },
                            "stats": map[string]interface{}{
                                "posts":     len(rawPosts["items"].([]interface{})),
                                "followers": rawFollowers["count"],
                            },
                            "recent_posts": transformPosts(rawPosts["items"].([]interface{})),
                            "meta": map[string]interface{}{
                                "timestamp": time.Now().Unix(),
                                "version":   "mobile_v1",
                            },
                        }
                        
                        return map[string]interface{}{"response": mobile}, nil
                    }),
                ),
            ),
        ))
}

func transformPosts(posts []interface{}) []map[string]interface{} {
    var mobilePosts []map[string]interface{}
    
    for i, post := range posts {
        if i >= 5 { // Limit to 5 recent posts for mobile
            break
        }
        
        p := post.(map[string]interface{})
        mobilePost := map[string]interface{}{
            "id":        p["id"],
            "content":   truncateString(p["content"].(string), 200),
            "image":     p["featured_image_url"],
            "timestamp": p["created_at"],
            "likes":     p["like_count"],
        }
        mobilePosts = append(mobilePosts, mobilePost)
    }
    
    return mobilePosts
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

### Pattern: Field Selection and Filtering

Allow clients to specify which fields they need:

```go
func CreateFieldSelectionFlow(fields []string) *flow.Flow {
    return flow.NewFlow("FieldSelection").
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"),
        )).
        Step("data", flow.NewStepWrapper(
            http.GET("${baseURL}/data/${resourceId}").SaveAs("fullData"),
        )).
        
        Step("fieldSelection", flow.NewStepWrapper(
            core.NewTransformStep("fieldSelect", 
                transformers.NewFieldTransformer("selectFields", fields),
            ),
        ))
}
```

## Best Practices Summary

1. **Parallel Execution**: Use `base.NewParallelStep()` for independent data fetching
2. **Graceful Degradation**: Handle optional service failures gracefully with custom steps
3. **Mobile Optimization**: Use `transformers.NewMobileTransformer()` for mobile-specific data
4. **Caching**: Implement caching with `core.NewCacheGetStep()` and `core.NewCacheSetStep()`
5. **Error Handling**: Use `base.NewRetryStep()` and structured error handling
6. **Authentication**: Use `core.NewTokenValidationStep()` with automatic token refresh
7. **Field Selection**: Use `transformers.NewFieldTransformer()` for client-specific data
8. **Timeouts**: Set appropriate timeouts with `WithTimeout()` for different operations
9. **Logging**: Use `core.NewLogStep()` for important events and debugging
10. **Testing**: Write comprehensive tests for all flow patterns using the test utilities 