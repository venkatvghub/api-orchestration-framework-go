# Building a BFF Layer: A Complete Beginner's Guide

> Learn how to build Backend for Frontend (BFF) layers step by step using the API Orchestration Framework. This guide uses simple language and real examples from a mobile onboarding system.

## Table of Contents

1. [What is a BFF and Why Do You Need One?](#what-is-a-bff-and-why-do-you-need-one)
2. [Project Structure - How to Organize Your Code](#project-structure)
3. [Step 1: Setting Up Configuration](#step-1-configuration)
4. [Step 2: Dependency Injection (DI)](#step-2-dependency-injection)
5. [Step 3: Creating Data Models](#step-3-data-models)
6. [Step 4: Building Services](#step-4-services)
7. [Step 5: Adding Middleware](#step-5-middleware)
8. [Step 6: Creating Handlers](#step-6-handlers)
9. [Step 7: Main Application Setup](#step-7-main-application)
10. [Understanding Flows - The Heart of BFF](#understanding-flows)
11. [Testing Your BFF](#testing-your-bff)
12. [Common Patterns and Best Practices](#common-patterns)

## What is a BFF and Why Do You Need One?

Think of a BFF as a **smart middleman** between your mobile app and your backend services.

### Without BFF (The Problem)
```
Mobile App ──┐
             ├──► User Service (get user info)
             ├──► Content Service (get content)
             ├──► Analytics Service (track events)
             └──► Notification Service (get notifications)
```
**Problems:**
- Mobile app makes 4 separate API calls
- Slow loading (network requests are expensive on mobile)
- Complex error handling in the app
- Too much data transferred (mobile has limited bandwidth)

### With BFF (The Solution)
```
Mobile App ──► BFF ──┐
                     ├──► User Service
                     ├──► Content Service  
                     ├──► Analytics Service
                     └──► Notification Service
```
**Benefits:**
- Mobile app makes 1 API call
- BFF fetches data in parallel (faster)
- BFF handles errors and provides fallbacks
- BFF sends only the data mobile needs

## Project Structure

Here's how to organize your BFF project:

```
my-bff/
├── main.go                 # Application entry point
├── go.mod                  # Go dependencies
├── config/                 # Configuration management
│   └── config.go
├── di/                     # Dependency injection setup
│   └── di.go
├── models/                 # Data structures
│   ├── user.go
│   ├── content.go
│   └── response.go
├── services/               # Business logic
│   ├── user_service.go
│   ├── content_service.go
│   └── cache_service.go
├── middleware/             # HTTP middleware
│   ├── auth.go
│   ├── logging.go
│   └── flow_middleware.go
├── handlers/               # HTTP request handlers
│   ├── user_handler.go
│   └── content_handler.go
└── flows/                  # Flow definitions (optional)
    ├── user_flows.go
    └── content_flows.go
```

## Step 1: Configuration

Configuration manages all your app settings in one place.

### config/config.go
```go
package config

import "os"

// AppConfig holds all application configuration
type AppConfig struct {
    Port           string
    UserServiceURL string
    ContentServiceURL string
    CacheEnabled   bool
    LogLevel       string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *AppConfig {
    return &AppConfig{
        Port:              getEnv("PORT", "8080"),
        UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8081"),
        ContentServiceURL: getEnv("CONTENT_SERVICE_URL", "http://localhost:8082"),
        CacheEnabled:      getEnv("CACHE_ENABLED", "true") == "true",
        LogLevel:          getEnv("LOG_LEVEL", "info"),
    }
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

**Why this matters:** Configuration keeps your app flexible. You can change URLs or settings without changing code.

## Step 2: Dependency Injection (DI)

DI is like a **smart container** that creates and connects all your app components automatically.

### di/di.go
```go
package di

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
    "go.uber.org/zap"

    "your-app/config"
    "your-app/handlers"
    "your-app/services"
    frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

// Module provides all dependencies for the application
func Module() fx.Option {
    return fx.Options(
        fx.Provide(
            // Core dependencies
            provideConfig,
            provideLogger,
            
            // Services
            services.NewUserService,
            services.NewContentService,
            services.NewCacheService,
            
            // Handlers
            handlers.NewUserHandler,
            handlers.NewContentHandler,
            
            // Infrastructure
            provideRouter,
            provideServer,
        ),
    )
}

// provideConfig creates application config
func provideConfig() *config.AppConfig {
    return config.LoadConfig()
}

// provideLogger creates a logger
func provideLogger(cfg *config.AppConfig) (*zap.Logger, error) {
    if cfg.LogLevel == "debug" {
        return zap.NewDevelopment()
    }
    return zap.NewProduction()
}

// provideRouter creates HTTP router
func provideRouter() *gin.Engine {
    return gin.New()
}

// provideServer creates HTTP server
func provideServer(cfg *config.AppConfig, router *gin.Engine) *http.Server {
    return &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      router,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
}
```

**Think of DI like this:** Instead of manually creating every component and connecting them, DI does it automatically. It's like having a smart assistant that knows what each part needs.

## Step 3: Data Models

Models define the **shape of your data**. They're like blueprints for the information your app handles.

### models/user.go
```go
package models

import "time"

// User represents a user in the system
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Avatar    string    `json:"avatar"`
    CreatedAt time.Time `json:"created_at"`
}

// UserProfile is a mobile-optimized version of User
type UserProfile struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Avatar string `json:"avatar"`
}
```

### models/response.go
```go
package models

import "time"

// APIResponse is the standard response format
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an error response
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(data interface{}) APIResponse {
    return APIResponse{
        Success:   true,
        Data:      data,
        Timestamp: time.Now(),
    }
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message string) APIResponse {
    return APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
        },
        Timestamp: time.Now(),
    }
}
```

**Why models matter:** They ensure your data has a consistent structure and make your code easier to understand and maintain.

## Step 4: Services

Services contain your **business logic**. They're like specialized workers that know how to do specific tasks.

### services/user_service.go
```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "go.uber.org/zap"
    "your-app/config"
    "your-app/models"
)

// UserService handles user-related operations
type UserService struct {
    config     *config.AppConfig
    logger     *zap.Logger
    httpClient *http.Client
}

// NewUserService creates a new user service
func NewUserService(cfg *config.AppConfig, logger *zap.Logger) *UserService {
    return &UserService{
        config: cfg,
        logger: logger,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// GetUser fetches a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
    url := fmt.Sprintf("%s/users/%s", s.config.UserServiceURL, userID)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
    }
    
    var user models.User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("failed to decode user: %w", err)
    }
    
    return &user, nil
}

// GetUserProfile returns a mobile-optimized user profile
func (s *UserService) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
    user, err := s.GetUser(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Transform full user to mobile profile
    return &models.UserProfile{
        ID:     user.ID,
        Name:   user.Name,
        Avatar: user.Avatar,
    }, nil
}
```

### services/cache_service.go
```go
package services

import (
    "sync"
    "time"

    "go.uber.org/zap"
    "your-app/config"
)

// CacheEntry represents a cached item
type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}

// CacheService provides in-memory caching
type CacheService struct {
    config *config.AppConfig
    logger *zap.Logger
    cache  map[string]*CacheEntry
    mutex  sync.RWMutex
}

// NewCacheService creates a new cache service
func NewCacheService(cfg *config.AppConfig, logger *zap.Logger) *CacheService {
    service := &CacheService{
        config: cfg,
        logger: logger,
        cache:  make(map[string]*CacheEntry),
    }
    
    // Start cleanup routine
    go service.cleanup()
    
    return service
}

// Get retrieves an item from cache
func (s *CacheService) Get(key string) (interface{}, bool) {
    if !s.config.CacheEnabled {
        return nil, false
    }
    
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    entry, exists := s.cache[key]
    if !exists {
        return nil, false
    }
    
    if time.Now().After(entry.ExpiresAt) {
        delete(s.cache, key)
        return nil, false
    }
    
    return entry.Data, true
}

// Set stores an item in cache
func (s *CacheService) Set(key string, value interface{}, ttl time.Duration) {
    if !s.config.CacheEnabled {
        return
    }
    
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    s.cache[key] = &CacheEntry{
        Data:      value,
        ExpiresAt: time.Now().Add(ttl),
    }
}

// cleanup removes expired entries
func (s *CacheService) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        s.mutex.Lock()
        now := time.Now()
        for key, entry := range s.cache {
            if now.After(entry.ExpiresAt) {
                delete(s.cache, key)
            }
        }
        s.mutex.Unlock()
    }
}
```

**Services are like specialists:** Each service knows how to do one thing really well (user operations, caching, etc.).

## Step 5: Middleware

Middleware are like **security guards and helpers** that process every request before it reaches your handlers.

### middleware/flow_middleware.go
```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

// FlowContextMiddleware creates flow context for each request
func FlowContextMiddleware(cfg *frameworkConfig.FrameworkConfig, logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Create flow context from HTTP request
        flowCtx := flow.NewContextFromGinWithLogger(c, cfg, logger)
        
        // Extract common parameters
        if userID := c.Param("userId"); userID != "" {
            flowCtx.Set("user_id", userID)
        }
        if userID := c.Query("user_id"); userID != "" {
            flowCtx.Set("user_id", userID)
        }
        
        // Set device information
        flowCtx.Set("device_type", c.GetHeader("X-Device-Type"))
        flowCtx.Set("app_version", c.GetHeader("X-App-Version"))
        
        // Store flow context for handlers to use
        c.Set("flow_context", flowCtx)
        
        logger.Debug("Flow context created",
            zap.String("path", c.Request.URL.Path),
            zap.String("method", c.Request.Method))
        
        c.Next()
    }
}
```

### middleware/logging.go
```go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        logger.Info("HTTP Request",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", duration),
            zap.String("user_agent", c.Request.UserAgent()),
        )
    }
}
```

**Middleware is like a checkpoint:** Every request passes through middleware before reaching your business logic.

## Step 6: Handlers

Handlers are like **receptionists** - they receive HTTP requests and coordinate the response.

### handlers/user_handler.go
```go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "your-app/models"
    "your-app/services"
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
    frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
    config      *frameworkConfig.FrameworkConfig
    logger      *zap.Logger
    userService *services.UserService
    cacheService *services.CacheService
}

// NewUserHandler creates a new user handler
func NewUserHandler(
    cfg *frameworkConfig.FrameworkConfig,
    logger *zap.Logger,
    userService *services.UserService,
    cacheService *services.CacheService,
) *UserHandler {
    return &UserHandler{
        config:       cfg,
        logger:       logger,
        userService:  userService,
        cacheService: cacheService,
    }
}

// GetUserProfile handles GET /users/:userId/profile
func (h *UserHandler) GetUserProfile(c *gin.Context) {
    userID := c.Param("userId")
    
    h.logger.Info("Getting user profile", zap.String("user_id", userID))
    
    // Get flow context from middleware
    ctx, exists := flow.GetFlowContext(c)
    if !exists {
        c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
            "CONTEXT_ERROR", "Flow context not available"))
        return
    }
    
    // Create a flow to get user profile
    profileFlow := flow.NewFlow("get_user_profile").
        WithDescription("Get user profile with caching").
        WithTimeout(30 * time.Second)
    
    profileFlow.
        // Step 1: Check cache first
        Step("check_cache", h.createCacheCheckStep(userID)).
        
        // Step 2: If not in cache, fetch from service
        Choice("cache_status").
            When(func(ctx interfaces.ExecutionContext) bool {
                return !ctx.Has("cached_profile")
            }).
                Step("fetch_user", h.createFetchUserStep()).
                Step("cache_user", h.createCacheUserStep(userID)).
            Otherwise().
                Step("use_cached", h.createUseCachedStep()).
        EndChoice().
        
        // Step 3: Transform for mobile
        Transform("mobile_transform", h.createMobileTransformer())
    
    // Execute the flow
    _, err := profileFlow.Execute(ctx)
    if err != nil {
        h.logger.Error("Failed to get user profile", zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
            "PROFILE_ERROR", err.Error()))
        return
    }
    
    // Get the result and return it
    profile, _ := ctx.Get("mobile_profile")
    c.JSON(http.StatusOK, models.NewSuccessResponse(profile))
}

// Helper methods to create flow steps

func (h *UserHandler) createCacheCheckStep(userID string) interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        cacheKey := "user_profile_" + userID
        if profile, exists := h.cacheService.Get(cacheKey); exists {
            ctx.Set("cached_profile", profile)
        }
        return nil
    })
}

func (h *UserHandler) createFetchUserStep() interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        userID, _ := ctx.GetString("user_id")
        
        profile, err := h.userService.GetUserProfile(ctx.Context(), userID)
        if err != nil {
            return err
        }
        
        ctx.Set("user_profile", profile)
        return nil
    })
}

func (h *UserHandler) createCacheUserStep(userID string) interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        profile, _ := ctx.Get("user_profile")
        cacheKey := "user_profile_" + userID
        h.cacheService.Set(cacheKey, profile, 10*time.Minute)
        return nil
    })
}

func (h *UserHandler) createUseCachedStep() interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        cached, _ := ctx.Get("cached_profile")
        ctx.Set("user_profile", cached)
        return nil
    })
}

func (h *UserHandler) createMobileTransformer() func(interfaces.ExecutionContext) error {
    return func(ctx interfaces.ExecutionContext) error {
        profile, _ := ctx.Get("user_profile")
        
        // Transform to mobile-friendly format
        mobileProfile := map[string]interface{}{
            "id":     profile.(models.UserProfile).ID,
            "name":   profile.(models.UserProfile).Name,
            "avatar": profile.(models.UserProfile).Avatar,
        }
        
        ctx.Set("mobile_profile", mobileProfile)
        return nil
    }
}
```

**Handlers coordinate everything:** They receive requests, orchestrate the work through flows, and send responses.

## Step 7: Main Application

The main application ties everything together.

### main.go
```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
    "go.uber.org/zap"

    "your-app/config"
    "your-app/di"
    "your-app/handlers"
    "your-app/middleware"
    frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
)

func main() {
    app := fx.New(
        di.Module(),
        fx.Invoke(startServer),
    )
    app.Run()
}

// startServer sets up and starts the HTTP server
func startServer(
    lc fx.Lifecycle,
    server *http.Server,
    router *gin.Engine,
    cfg *frameworkConfig.FrameworkConfig,
    logger *zap.Logger,
    userHandler *handlers.UserHandler,
) {
    setupRouter(router, cfg, logger, userHandler)

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go func() {
                logger.Info("Starting BFF server", zap.String("addr", server.Addr))
                if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                    logger.Fatal("Server failed", zap.Error(err))
                }
            }()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            logger.Info("Shutting down server...")
            shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
            defer cancel()
            return server.Shutdown(shutdownCtx)
        },
    })
}

// setupRouter configures all routes and middleware
func setupRouter(
    router *gin.Engine,
    cfg *frameworkConfig.FrameworkConfig,
    logger *zap.Logger,
    userHandler *handlers.UserHandler,
) {
    // Add framework middleware
    router.Use(flow.RequestIDMiddleware())
    router.Use(flow.TimingMiddleware())
    router.Use(flow.ErrorHandlingMiddleware(logger))
    
    // Add application middleware
    router.Use(middleware.FlowContextMiddleware(cfg, logger))
    router.Use(middleware.LoggingMiddleware(logger))
    router.Use(gin.Recovery())

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status":    "healthy",
            "timestamp": time.Now(),
        })
    })

    // API routes
    api := router.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.GET("/:userId/profile", userHandler.GetUserProfile)
        }
    }
}
```

## Understanding Flows - The Heart of BFF

Flows are the **magic** of the API Orchestration Framework. They let you define complex operations as a series of simple steps.

### What is a Flow?

Think of a flow like a **recipe**:
1. Check if we have ingredients (cache check)
2. If not, go shopping (fetch from API)
3. Cook the meal (transform data)
4. Serve it (return response)

### Types of Flow Steps

#### 1. Sequential Steps (One After Another)
```go
flow.NewFlow("sequential_example").
    Step("step1", doFirstThing).
    Step("step2", doSecondThing).
    Step("step3", doThirdThing)
```

#### 2. Parallel Steps (All at Once)
```go
flow.NewFlow("parallel_example").
    Parallel("fetch_data").
        Step("user", fetchUser).
        Step("posts", fetchPosts).
        Step("notifications", fetchNotifications).
    EndParallel()
```

#### 3. Conditional Steps (If/Else)
```go
flow.NewFlow("conditional_example").
    Choice("check_cache").
        When(func(ctx interfaces.ExecutionContext) bool {
            return ctx.Has("cached_data")
        }).
            Step("use_cache", useCachedData).
        Otherwise().
            Step("fetch_fresh", fetchFreshData).
    EndChoice()
```

#### 4. Transform Steps (Change Data)
```go
flow.NewFlow("transform_example").
    Step("fetch", fetchRawData).
    Transform("mobile_optimize", func(ctx interfaces.ExecutionContext) error {
        rawData, _ := ctx.Get("raw_data")
        mobileData := optimizeForMobile(rawData)
        ctx.Set("mobile_data", mobileData)
        return nil
    })
```

### Context Propagation

Context is like a **shared notebook** that all steps can read and write to.

```go
// Setting data in context
ctx.Set("user_id", "123")
ctx.Set("user_data", userData)

// Getting data from context
userID, _ := ctx.GetString("user_id")
userData, _ := ctx.Get("user_data")

// Checking if data exists
if ctx.Has("cached_data") {
    // Use cached data
}
```

### Real Example: User Dashboard Flow

```go
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
    ctx, _ := flow.GetFlowContext(c)
    
    dashboardFlow := flow.NewFlow("user_dashboard").
        WithDescription("Get user dashboard with all data").
        WithTimeout(30 * time.Second)
    
    dashboardFlow.
        // Step 1: Validate user
        Step("validate", h.validateUser()).
        
        // Step 2: Fetch data in parallel (FAST!)
        Parallel("fetch_data").
            Step("user", h.fetchUser()).
            Step("posts", h.fetchRecentPosts()).
            Step("notifications", h.fetchNotifications()).
            Step("analytics", h.fetchAnalytics()).
        EndParallel().
        
        // Step 3: Check if we have everything
        Choice("data_complete").
            When(func(ctx interfaces.ExecutionContext) bool {
                return ctx.Has("user") && ctx.Has("posts")
            }).
                // We have essential data, continue
                Step("combine", h.combineData()).
            Otherwise().
                // Missing essential data, return error
                Step("error", h.returnError()).
        EndChoice().
        
        // Step 4: Optimize for mobile
        Transform("mobile_optimize", h.optimizeForMobile())
    
    // Execute the flow
    _, err := dashboardFlow.Execute(ctx)
    if err != nil {
        c.JSON(500, models.NewErrorResponse("DASHBOARD_ERROR", err.Error()))
        return
    }
    
    dashboard, _ := ctx.Get("mobile_dashboard")
    c.JSON(200, models.NewSuccessResponse(dashboard))
}
```

### Why Flows Are Powerful

1. **Parallel Execution**: Fetch multiple things at once (faster)
2. **Error Handling**: If one step fails, handle it gracefully
3. **Caching**: Check cache first, fetch if needed
4. **Transformation**: Convert data to mobile-friendly format
5. **Conditional Logic**: Different paths based on conditions
6. **Timeout Management**: Don't wait forever for slow services

## Testing Your BFF

### Unit Testing a Handler

```go
func TestUserHandler_GetUserProfile(t *testing.T) {
    // Setup
    cfg := &frameworkConfig.FrameworkConfig{}
    logger := zap.NewNop()
    
    // Mock services
    userService := &MockUserService{}
    cacheService := &MockCacheService{}
    
    handler := handlers.NewUserHandler(cfg, logger, userService, cacheService)
    
    // Create test request
    router := gin.New()
    router.GET("/users/:userId/profile", handler.GetUserProfile)
    
    req := httptest.NewRequest("GET", "/users/123/profile", nil)
    w := httptest.NewRecorder()
    
    // Execute
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response models.APIResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.True(t, response.Success)
}
```

### Integration Testing

```go
func TestUserProfileFlow(t *testing.T) {
    // Start test server
    server := startTestServer()
    defer server.Close()
    
    // Test the complete flow
    resp, err := http.Get(server.URL + "/api/v1/users/123/profile")
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var response models.APIResponse
    json.NewDecoder(resp.Body).Decode(&response)
    assert.True(t, response.Success)
    assert.NotNil(t, response.Data)
}
```

## Common Patterns and Best Practices

### 1. Cache-First Pattern

```go
flow.NewFlow("cache_first").
    Step("check_cache", checkCache).
    Choice("cache_hit").
        When(func(ctx interfaces.ExecutionContext) bool {
            return ctx.Has("cached_data")
        }).
            Step("use_cache", useCachedData).
        Otherwise().
            Step("fetch_fresh", fetchFreshData).
            Step("store_cache", storeInCache).
    EndChoice()
```

### 2. Parallel Fetch with Fallback

```go
flow.NewFlow("parallel_with_fallback").
    Parallel("fetch_all").
        Step("primary", fetchFromPrimaryService).
        Step("secondary", fetchFromSecondaryService).
    EndParallel().
    Choice("data_available").
        When(func(ctx interfaces.ExecutionContext) bool {
            return ctx.Has("primary_data")
        }).
            Step("use_primary", usePrimaryData).
        When(func(ctx interfaces.ExecutionContext) bool {
            return ctx.Has("secondary_data")
        }).
            Step("use_secondary", useSecondaryData).
        Otherwise().
            Step("use_default", useDefaultData).
    EndChoice()
```

### 3. Mobile Optimization Pattern

```go
Transform("mobile_optimize", func(ctx interfaces.ExecutionContext) error {
    rawData, _ := ctx.GetMap("raw_data")
    
    // Only include fields mobile needs
    mobileData := map[string]interface{}{
        "id":     rawData["id"],
        "name":   rawData["full_name"], // Rename field
        "avatar": rawData["profile_image_url"],
        // Skip heavy fields like "full_bio", "detailed_stats"
    }
    
    ctx.Set("mobile_data", mobileData)
    return nil
})
```

### 4. Error Handling Pattern

```go
flow.NewFlow("with_error_handling").
    Step("risky_operation", riskyOperation).
    Choice("operation_success").
        When(func(ctx interfaces.ExecutionContext) bool {
            return !ctx.Has("error")
        }).
            Step("continue", continueNormally).
        Otherwise().
            Step("handle_error", handleError).
            Step("fallback", provideFallbackData).
    EndChoice()
```

### Best Practices Summary

1. **Keep flows simple**: Each flow should do one main thing
2. **Use parallel steps**: Fetch independent data simultaneously
3. **Always have fallbacks**: Don't let one service failure break everything
4. **Cache wisely**: Cache expensive operations, not real-time data
5. **Optimize for mobile**: Send only what the mobile app needs
6. **Handle errors gracefully**: Provide meaningful error messages
7. **Use timeouts**: Don't wait forever for slow services
8. **Log everything**: You'll need logs when things go wrong

### Common Mistakes to Avoid

1. **Sequential when you could parallel**: Don't fetch data one by one
2. **No error handling**: Always plan for failures
3. **Over-caching**: Don't cache data that changes frequently
4. **Under-caching**: Do cache expensive computations
5. **Too much data**: Don't send everything to mobile
6. **No timeouts**: Always set reasonable timeouts
7. **Complex flows**: Keep flows simple and focused

This guide gives you everything you need to build a robust BFF layer. Start simple, add complexity gradually, and always think about what your mobile app actually needs! 