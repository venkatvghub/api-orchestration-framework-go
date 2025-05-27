# Building a BFF Layer: Complete Developer Guide

> A comprehensive guide to building Backend for Frontend (BFF) layers using the API Orchestration Framework, with real-world examples from mobile onboarding.

## Table of Contents

1. [What is a BFF Layer?](#what-is-a-bff-layer)
2. [Setting Up Your BFF Project](#setting-up-your-bff-project)
3. [Basic BFF Patterns](#basic-bff-patterns)
4. [Mobile Onboarding Example](#mobile-onboarding-example)
5. [Advanced Patterns](#advanced-patterns)
6. [Testing Your BFF](#testing-your-bff)
7. [Deployment and Monitoring](#deployment-and-monitoring)
8. [Common Pitfalls and Solutions](#common-pitfalls-and-solutions)

## What is a BFF Layer?

A Backend for Frontend (BFF) is a service layer that sits between your frontend applications (mobile apps, web apps) and your backend microservices. It aggregates, transforms, and optimizes data specifically for each frontend's needs.

### Why Use a BFF?

| Problem | BFF Solution |
|---------|--------------|
| **Multiple API calls from mobile** | Single endpoint with parallel data fetching |
| **Over-fetching data** | Field selection and data filtering |
| **Complex data transformation** | Server-side data shaping |
| **Poor mobile performance** | Caching, compression, and optimization |
| **Inconsistent error handling** | Centralized error management |

### BFF Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Mobile App    │───▶│   BFF Layer     │───▶│  Microservices  │
│                 │    │                 │    │                 │
│ • iOS           │    │ • Aggregation   │    │ • User Service  │
│ • Android       │    │ • Transformation│    │ • Content API   │
│ • React Native  │    │ • Caching       │    │ • Analytics     │
└─────────────────┘    │ • Error Handling│    │ • Notifications │
                       └─────────────────┘    └─────────────────┘
```

## Setting Up Your BFF Project

### 1. Project Structure

Create a well-organized project structure:

```
mobile-bff/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── handlers/               # HTTP request handlers
│   ├── onboarding.go      # Onboarding endpoints
│   ├── profile.go         # User profile endpoints
│   └── dashboard.go       # Dashboard endpoints
├── services/              # Business logic services
│   ├── user_service.go    # User-related operations
│   ├── content_service.go # Content management
│   └── analytics_service.go # Analytics operations
├── models/                # Data models
│   ├── user.go           # User models
│   ├── onboarding.go     # Onboarding models
│   └── response.go       # API response models
├── middleware/            # HTTP middleware
│   ├── auth.go           # Authentication
│   ├── logging.go        # Request logging
│   └── metrics.go        # Metrics collection
├── flows/                 # Flow definitions
│   ├── onboarding_flows.go # Onboarding orchestration
│   └── user_flows.go      # User data flows
└── config/               # Configuration
    └── config.go         # App configuration
```

### 2. Basic Setup

**main.go** - Application entry point with dependency injection:

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
    "go.uber.org/zap"

    "your-project/handlers"
    "your-project/middleware"
    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/di"
    "github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

func main() {
    app := fx.New(
        di.Module(),                    // Framework DI
        fx.Invoke(startServer),         // Start HTTP server
    )
    app.Run()
}

func startServer(lc fx.Lifecycle, server *http.Server, router *gin.Engine, cfg *config.FrameworkConfig, logger *zap.Logger) {
    setupRouter(router, cfg, logger)

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
            return server.Shutdown(ctx)
        },
    })
}

func setupRouter(router *gin.Engine, cfg *config.FrameworkConfig, logger *zap.Logger) {
    // Middleware
    router.Use(middleware.LoggingMiddleware(logger))
    router.Use(middleware.MetricsMiddleware())
    router.Use(middleware.CORSMiddleware())
    router.Use(gin.Recovery())

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })

    // API routes
    api := router.Group("/api/v1")
    {
        // Onboarding endpoints
        onboarding := api.Group("/onboarding")
        {
            handler := handlers.NewOnboardingHandler(cfg, logger)
            onboarding.GET("/screens/:screenId", handler.GetScreen)
            onboarding.POST("/screens/:screenId/submit", handler.SubmitScreen)
            onboarding.GET("/flow/:userId", handler.GetOnboardingFlow)
        }
    }
}
```

### 3. Environment Configuration

Create a `.env` file for configuration:

```bash
# Server Configuration
PORT=8080
GIN_MODE=release

# Backend Services
USER_SERVICE_URL=http://localhost:8081
CONTENT_SERVICE_URL=http://localhost:8082
ANALYTICS_SERVICE_URL=http://localhost:8083

# Authentication
JWT_SECRET=your-jwt-secret
API_TOKEN=your-api-token

# Cache Configuration
CACHE_DEFAULT_TTL=5m
CACHE_MAX_SIZE=1000

# HTTP Configuration
HTTP_REQUEST_TIMEOUT=15s
HTTP_MAX_RETRIES=3

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## Basic BFF Patterns

### 1. Simple Data Aggregation

**Problem**: Mobile app needs user profile + notifications + recent activity in one call.

**Solution**: Create an aggregation flow that fetches data in parallel.

```go
// handlers/profile.go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
    httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
    "github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
)

type ProfileHandler struct {
    config *config.FrameworkConfig
    logger *zap.Logger
}

func NewProfileHandler(cfg *config.FrameworkConfig, logger *zap.Logger) *ProfileHandler {
    return &ProfileHandler{config: cfg, logger: logger}
}

// GetUserDashboard aggregates user dashboard data
func (h *ProfileHandler) GetUserDashboard(c *gin.Context) {
    userID := c.Param("userId")
    
    // Create flow context
    ctx := flow.NewContextWithConfig(h.config)
    ctx.WithLogger(h.logger)
    ctx.Set("user_id", userID)
    
    // Create aggregation flow
    dashboardFlow := flow.NewFlow("user_dashboard").
        WithDescription("Aggregate user dashboard data").
        WithTimeout(10 * time.Second)
    
    dashboardFlow.
        // Step 1: Validate authentication
        Step("auth", flow.NewStepWrapper(
            core.NewTokenValidationStep("auth", "Authorization"))).
        
        // Step 2: Fetch data in parallel (this is the key BFF pattern!)
        Parallel("dashboard_data").
            Step("profile", flow.NewStepWrapper(
                httpsteps.GET("${user_service_url}/users/${user_id}").
                    WithHeader("Authorization", "Bearer ${api_token}").
                    SaveAs("user_profile"))).
            Step("notifications", flow.NewStepWrapper(
                httpsteps.GET("${notification_service_url}/notifications/${user_id}?limit=5").
                    WithHeader("Authorization", "Bearer ${api_token}").
                    SaveAs("notifications"))).
            Step("activity", flow.NewStepWrapper(
                httpsteps.GET("${activity_service_url}/activity/${user_id}?limit=10").
                    WithHeader("Authorization", "Bearer ${api_token}").
                    SaveAs("recent_activity"))).
        EndParallel().
        
        // Step 3: Transform for mobile consumption
        Transform("mobile_transform", h.createDashboardTransformer())
    
    // Execute flow
    _, err := dashboardFlow.Execute(ctx)
    if err != nil {
        h.logger.Error("Dashboard flow failed", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dashboard"})
        return
    }
    
    // Return aggregated data
    dashboardData, _ := ctx.Get("dashboard_response")
    c.JSON(http.StatusOK, dashboardData)
}

// Transform raw backend data into mobile-friendly format
func (h *ProfileHandler) createDashboardTransformer() func(interfaces.ExecutionContext) error {
    return func(ctx interfaces.ExecutionContext) error {
        // Get raw data from context
        profile, _ := ctx.GetMap("user_profile")
        notifications, _ := ctx.GetMap("notifications")
        activity, _ := ctx.GetMap("recent_activity")
        
        // Transform into mobile-optimized response
        dashboard := map[string]interface{}{
            "user": map[string]interface{}{
                "id":     profile["id"],
                "name":   profile["full_name"],
                "avatar": profile["avatar_url"],
                "email":  profile["email"],
            },
            "summary": map[string]interface{}{
                "unread_notifications": len(notifications["items"].([]interface{})),
                "recent_activities":    len(activity["items"].([]interface{})),
            },
            "quick_access": map[string]interface{}{
                "notifications": h.transformNotifications(notifications),
                "activities":    h.transformActivities(activity),
            },
            "timestamp": time.Now().Unix(),
        }
        
        ctx.Set("dashboard_response", map[string]interface{}{
            "success": true,
            "data":    dashboard,
        })
        
        return nil
    }
}

func (h *ProfileHandler) transformNotifications(notifications map[string]interface{}) []map[string]interface{} {
    items, _ := notifications["items"].([]interface{})
    var transformed []map[string]interface{}
    
    for _, item := range items {
        notif := item.(map[string]interface{})
        transformed = append(transformed, map[string]interface{}{
            "id":      notif["id"],
            "title":   notif["title"],
            "message": notif["message"],
            "type":    notif["type"],
            "read":    notif["read"],
        })
    }
    
    return transformed
}

func (h *ProfileHandler) transformActivities(activity map[string]interface{}) []map[string]interface{} {
    items, _ := activity["items"].([]interface{})
    var transformed []map[string]interface{}
    
    for _, item := range items {
        act := item.(map[string]interface{})
        transformed = append(transformed, map[string]interface{}{
            "id":          act["id"],
            "action":      act["action_type"],
            "description": act["description"],
            "timestamp":   act["created_at"],
        })
    }
    
    return transformed
}
```

### 2. Field Selection Pattern

**Problem**: Mobile app only needs specific fields, but backend returns everything.

**Solution**: Use field transformers to select only needed data.

```go
// handlers/user.go
package handlers

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
    "github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
)

func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("userId")
    fields := c.Query("fields") // e.g., "id,name,email,avatar"
    
    ctx := flow.NewContextWithConfig(h.config)
    ctx.Set("user_id", userID)
    
    // Parse requested fields
    var selectedFields []string
    if fields != "" {
        selectedFields = strings.Split(fields, ",")
    } else {
        // Default fields for mobile
        selectedFields = []string{"id", "name", "email", "avatar", "created_at"}
    }
    
    userFlow := flow.NewFlow("get_user_with_fields").
        Step("fetch", flow.NewStepWrapper(
            httpsteps.GET("${user_service_url}/users/${user_id}").
                SaveAs("full_user_data"))).
        
        // Apply field selection
        Transform("field_selection", func(ctx interfaces.ExecutionContext) error {
            transformer := transformers.NewFieldTransformer("select_fields", selectedFields)
            
            fullData, _ := ctx.GetMap("full_user_data")
            selectedData, err := transformer.Transform(fullData)
            if err != nil {
                return err
            }
            
            ctx.Set("user_response", map[string]interface{}{
                "success": true,
                "data":    selectedData,
            })
            
            return nil
        })
    
    _, err := userFlow.Execute(ctx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    response, _ := ctx.Get("user_response")
    c.JSON(http.StatusOK, response)
}
```

### 3. Caching Pattern

**Problem**: Expensive API calls that don't change frequently.

**Solution**: Implement multi-level caching with TTL.

```go
func (h *ContentHandler) GetContent(c *gin.Context) {
    contentID := c.Param("contentId")
    
    ctx := flow.NewContextWithConfig(h.config)
    ctx.Set("content_id", contentID)
    
    contentFlow := flow.NewFlow("cached_content").
        // Step 1: Check cache first
        Step("check_cache", flow.NewStepWrapper(
            core.NewCacheGetStep("content_cache", "content_${content_id}", "cached_content"))).
        
        // Step 2: If cache miss, fetch from API
        Choice("cache_status").
            When(func(ctx interfaces.ExecutionContext) bool {
                return !ctx.Has("cached_content")
            }).
                Step("fetch_content", flow.NewStepWrapper(
                    httpsteps.GET("${content_service_url}/content/${content_id}").
                        SaveAs("fresh_content"))).
                Step("cache_content", flow.NewStepWrapper(
                    core.NewCacheSetStep("set_cache", "content_${content_id}", "fresh_content", 10*time.Minute))).
                Step("set_response", flow.NewStepWrapper(
                    core.NewSetValueStep("response", "content_data", "${fresh_content}"))).
            Otherwise().
                Step("use_cached", flow.NewStepWrapper(
                    core.NewSetValueStep("response", "content_data", "${cached_content}"))).
        EndChoice()
    
    _, err := contentFlow.Execute(ctx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    contentData, _ := ctx.Get("content_data")
    c.JSON(http.StatusOK, map[string]interface{}{
        "success": true,
        "data":    contentData,
        "cached":  ctx.Has("cached_content"),
    })
}
```

## Mobile Onboarding Example

Let's build a complete mobile onboarding BFF based on the real implementation:

### 1. Onboarding Models

```go
// models/onboarding.go
package models

import "time"

type OnboardingScreen struct {
    ID         string                 `json:"id"`
    Title      string                 `json:"title"`
    Subtitle   string                 `json:"subtitle,omitempty"`
    Type       string                 `json:"type"`
    Fields     []ScreenField          `json:"fields"`
    Actions    []ScreenAction         `json:"actions"`
    Validation map[string]interface{} `json:"validation"`
    NextScreen string                 `json:"next_screen,omitempty"`
}

type ScreenField struct {
    ID          string      `json:"id"`
    Type        string      `json:"type"`
    Label       string      `json:"label"`
    Placeholder string      `json:"placeholder,omitempty"`
    Required    bool        `json:"required"`
    Options     []FieldOption `json:"options,omitempty"`
}

type FieldOption struct {
    Value string `json:"value"`
    Label string `json:"label"`
}

type ScreenAction struct {
    ID    string `json:"id"`
    Type  string `json:"type"`
    Label string `json:"label"`
    Style string `json:"style,omitempty"`
}

type ScreenSubmission struct {
    ScreenID   string                 `json:"screen_id"`
    UserID     string                 `json:"user_id"`
    Data       map[string]interface{} `json:"data"`
    Timestamp  time.Time              `json:"timestamp"`
    DeviceInfo DeviceInfo             `json:"device_info"`
}

type DeviceInfo struct {
    Type       string `json:"type"`
    Platform   string `json:"platform"`
    AppVersion string `json:"app_version"`
    UserAgent  string `json:"user_agent,omitempty"`
}

type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### 2. Onboarding Handler

```go
// handlers/onboarding.go
package handlers

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "your-project/models"
    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
    "github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
    httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
    "github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
    "github.com/venkatvghub/api-orchestration-framework/pkg/validators"
)

type OnboardingHandler struct {
    config *config.FrameworkConfig
    logger *zap.Logger
}

func NewOnboardingHandler(cfg *config.FrameworkConfig, logger *zap.Logger) *OnboardingHandler {
    return &OnboardingHandler{
        config: cfg,
        logger: logger,
    }
}

// GetScreen retrieves an onboarding screen configuration
func (h *OnboardingHandler) GetScreen(c *gin.Context) {
    screenID := c.Param("screenId")
    userID := c.Query("user_id")
    version := c.GetHeader("X-API-Version")
    deviceType := c.GetHeader("X-Device-Type")

    if version == "" {
        version = "v1"
    }

    h.logger.Info("Getting onboarding screen",
        zap.String("screen_id", screenID),
        zap.String("user_id", userID),
        zap.String("version", version))

    // Create flow context
    ctx := flow.NewContextWithConfig(h.config)
    ctx.WithLogger(h.logger)
    ctx.Set("screen_id", screenID)
    ctx.Set("user_id", userID)
    ctx.Set("version", version)
    ctx.Set("device_type", deviceType)

    // Create screen retrieval flow
    screenFlow := flow.NewFlow("get_onboarding_screen").
        WithDescription("Retrieve onboarding screen with caching and optimization").
        WithTimeout(h.config.Timeouts.FlowExecution)

    screenFlow.
        // Step 1: Validate request
        Step("validate", h.createValidationStep()).
        
        // Step 2: Check cache
        Step("check_cache", flow.NewStepWrapper(
            core.NewCacheGetStep("screen_cache", 
                fmt.Sprintf("screen:%s:%s", screenID, version), 
                "cached_screen"))).
        
        // Step 3: Fetch if not cached
        Choice("cache_status").
            When(func(ctx interfaces.ExecutionContext) bool {
                return !ctx.Has("cached_screen")
            }).
                // Fetch screen configuration
                Step("fetch_screen", flow.NewStepWrapper(
                    httpsteps.GET("${content_service_url}/api/screens/${screen_id}").
                        WithQueryParam("version", "${version}").
                        WithHeader("Authorization", "Bearer ${api_token}").
                        SaveAs("screen_config"))).
                
                // Fetch user progress
                Step("fetch_progress", flow.NewStepWrapper(
                    httpsteps.GET("${user_service_url}/api/users/${user_id}/onboarding/progress").
                        WithHeader("Authorization", "Bearer ${api_token}").
                        SaveAs("user_progress"))).
                
                // Apply personalization
                Step("personalize", h.createPersonalizationStep()).
                
                // Cache the result
                Step("cache_screen", flow.NewStepWrapper(
                    core.NewCacheSetStep("cache_result", 
                        fmt.Sprintf("screen:%s:%s", screenID, version), 
                        "personalized_screen", 
                        5*time.Minute))).
            Otherwise().
                Step("use_cached", flow.NewStepWrapper(
                    core.NewSetValueStep("final", "personalized_screen", "${cached_screen}"))).
        EndChoice().
        
        // Step 4: Transform for mobile
        Transform("mobile_transform", h.createMobileTransformer())

    // Execute flow
    _, err := screenFlow.Execute(ctx)
    if err != nil {
        h.logger.Error("Screen flow failed", zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success:   false,
            Error:     &models.APIError{Code: "SCREEN_FETCH_ERROR", Message: err.Error()},
            Timestamp: time.Now(),
        })
        return
    }

    screenData, _ := ctx.Get("mobile_screen")
    c.JSON(http.StatusOK, models.APIResponse{
        Success:   true,
        Data:      screenData,
        Timestamp: time.Now(),
    })
}

// SubmitScreen handles screen form submission
func (h *OnboardingHandler) SubmitScreen(c *gin.Context) {
    screenID := c.Param("screenId")

    var submission models.ScreenSubmission
    if err := c.ShouldBindJSON(&submission); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Success:   false,
            Error:     &models.APIError{Code: "INVALID_REQUEST", Message: err.Error()},
            Timestamp: time.Now(),
        })
        return
    }

    submission.ScreenID = screenID
    submission.Timestamp = time.Now()
    submission.DeviceInfo = models.DeviceInfo{
        Type:       c.GetHeader("X-Device-Type"),
        Platform:   c.GetHeader("X-Platform"),
        AppVersion: c.GetHeader("X-App-Version"),
        UserAgent:  c.GetHeader("User-Agent"),
    }

    h.logger.Info("Submitting screen", 
        zap.String("screen_id", screenID),
        zap.String("user_id", submission.UserID))

    ctx := flow.NewContextWithConfig(h.config)
    ctx.WithLogger(h.logger)
    ctx.Set("submission", submission)

    // Create submission flow
    submissionFlow := flow.NewFlow("submit_screen").
        WithDescription("Process screen submission with validation and progress tracking").
        WithTimeout(h.config.Timeouts.FlowExecution)

    submissionFlow.
        // Step 1: Validate submission
        Step("validate_submission", h.createSubmissionValidationStep(screenID)).
        
        // Step 2: Process in parallel
        Parallel("process_submission").
            // Save user data
            Step("save_data", flow.NewStepWrapper(
                httpsteps.POST("${user_service_url}/api/users/${submission.user_id}/data").
                    WithJSONBody("${submission.data}").
                    WithHeader("Authorization", "Bearer ${api_token}").
                    SaveAs("save_result"))).
            
            // Update progress
            Step("update_progress", flow.NewStepWrapper(
                httpsteps.POST("${user_service_url}/api/users/${submission.user_id}/onboarding/progress").
                    WithJSONBody(map[string]interface{}{
                        "completed_screen": "${submission.screen_id}",
                        "timestamp":        time.Now(),
                    }).
                    WithHeader("Authorization", "Bearer ${api_token}").
                    SaveAs("progress_result"))).
            
            // Log analytics
            Step("log_analytics", flow.NewStepWrapper(
                httpsteps.POST("${content_service_url}/api/analytics/events").
                    WithJSONBody(map[string]interface{}{
                        "event_type": "screen_submit",
                        "user_id":    "${submission.user_id}",
                        "screen_id":  "${submission.screen_id}",
                        "data":       "${submission.data}",
                        "timestamp":  time.Now(),
                    }).
                    WithHeader("Authorization", "Bearer ${api_token}").
                    SaveAs("analytics_result"))).
        EndParallel().
        
        // Step 3: Determine next screen
        Step("next_screen", h.createNextScreenStep()).
        
        // Step 4: Format response
        Transform("format_response", h.createSubmissionResponseTransformer())

    _, err := submissionFlow.Execute(ctx)
    if err != nil {
        h.logger.Error("Submission flow failed", zap.Error(err))
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Success:   false,
            Error:     &models.APIError{Code: "SUBMISSION_ERROR", Message: err.Error()},
            Timestamp: time.Now(),
        })
        return
    }

    responseData, _ := ctx.Get("submission_response")
    c.JSON(http.StatusOK, models.APIResponse{
        Success:   true,
        Data:      responseData,
        Timestamp: time.Now(),
    })
}

// Helper methods for creating flow steps

func (h *OnboardingHandler) createValidationStep() interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        // Validate required fields
        validator := validators.NewRequiredFieldsValidator("screen_id", "user_id")
        
        data := map[string]interface{}{
            "screen_id": ctx.MustGet("screen_id"),
            "user_id":   ctx.MustGet("user_id"),
        }
        
        return validator.Validate(data)
    })
}

func (h *OnboardingHandler) createPersonalizationStep() interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        screenConfig, _ := ctx.GetMap("screen_config")
        userProgress, _ := ctx.GetMap("user_progress")
        
        // Apply personalization based on user progress
        personalized := make(map[string]interface{})
        for k, v := range screenConfig {
            personalized[k] = v
        }
        
        // Add progress information
        if progress, ok := userProgress["data"].(map[string]interface{}); ok {
            personalized["user_progress"] = progress
        }
        
        ctx.Set("personalized_screen", personalized)
        return nil
    })
}

func (h *OnboardingHandler) createMobileTransformer() func(interfaces.ExecutionContext) error {
    return func(ctx interfaces.ExecutionContext) error {
        transformer := transformers.NewMobileTransformer([]string{
            "id", "title", "subtitle", "type", "fields", "actions", "next_screen",
        })
        
        screenData, _ := ctx.GetMap("personalized_screen")
        if screenData == nil {
            screenData, _ = ctx.GetMap("cached_screen")
        }
        
        transformed, err := transformer.Transform(screenData)
        if err != nil {
            return err
        }
        
        ctx.Set("mobile_screen", transformed)
        return nil
    }
}

func (h *OnboardingHandler) createSubmissionValidationStep(screenID string) interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        submission, _ := ctx.Get("submission").(models.ScreenSubmission)
        
        // Basic validation
        if submission.UserID == "" {
            return fmt.Errorf("user_id is required")
        }
        
        if submission.Data == nil {
            return fmt.Errorf("submission data is required")
        }
        
        // Screen-specific validation
        switch screenID {
        case "personal_info":
            emailValidator := validators.EmailRequiredValidator("email")
            return emailValidator.Validate(submission.Data)
        case "preferences":
            requiredValidator := validators.NewRequiredFieldsValidator("theme", "language")
            return requiredValidator.Validate(submission.Data)
        }
        
        return nil
    })
}

func (h *OnboardingHandler) createNextScreenStep() interfaces.Step {
    return flow.StepWrapper(
        httpsteps.GET("${content_service_url}/api/flows/onboarding").
            WithQueryParam("current_screen", "${submission.screen_id}").
            WithQueryParam("user_id", "${submission.user_id}").
            WithHeader("Authorization", "Bearer ${api_token}").
            SaveAs("next_screen_info"))
}

func (h *OnboardingHandler) createSubmissionResponseTransformer() func(interfaces.ExecutionContext) error {
    return func(ctx interfaces.ExecutionContext) error {
        nextScreenInfo, _ := ctx.GetMap("next_screen_info")
        
        response := map[string]interface{}{
            "success":     true,
            "next_screen": nextScreenInfo,
            "timestamp":   time.Now(),
        }
        
        ctx.Set("submission_response", response)
        return nil
    }
}
```

### 3. Services Layer

```go
// services/onboarding_service.go
package services

import (
    "go.uber.org/zap"
    
    "your-project/models"
    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/validators"
)

type OnboardingService struct {
    config *config.FrameworkConfig
    logger *zap.Logger
}

func NewOnboardingService(cfg *config.FrameworkConfig, logger *zap.Logger) *OnboardingService {
    return &OnboardingService{
        config: cfg,
        logger: logger,
    }
}

func (s *OnboardingService) GetScreenValidator(screenID string) validators.Validator {
    switch screenID {
    case "welcome":
        return validators.NewRequiredFieldsValidator("user_id")
    case "personal_info":
        return validators.NewFuncValidator("personal_info", func(data map[string]interface{}) error {
            // Validate required fields
            if err := validators.NewRequiredFieldsValidator("first_name", "last_name", "email").Validate(data); err != nil {
                return err
            }
            // Validate email format
            return validators.EmailRequiredValidator("email").Validate(data)
        })
    case "preferences":
        return validators.NewRequiredFieldsValidator("theme", "language")
    default:
        return validators.NewRequiredFieldsValidator("user_id")
    }
}

func (s *OnboardingService) PersonalizeScreen(screen *models.OnboardingScreen, userProgress map[string]interface{}) *models.OnboardingScreen {
    // Clone the screen
    personalized := *screen
    
    // Apply personalization based on user progress
    if progress, ok := userProgress["progress"].(float64); ok {
        if progress > 0.5 {
            // User is halfway through, show encouragement
            personalized.Subtitle = "You're doing great! Just a few more steps."
        }
    }
    
    return &personalized
}
```

## Advanced Patterns

### 1. A/B Testing Integration

```go
func (h *OnboardingHandler) GetScreenWithABTesting(c *gin.Context) {
    userID := c.Query("user_id")
    screenID := c.Param("screenId")
    
    ctx := flow.NewContextWithConfig(h.config)
    ctx.Set("user_id", userID)
    ctx.Set("screen_id", screenID)
    
    abTestFlow := flow.NewFlow("ab_test_screen").
        // Step 1: Determine A/B test variant
        Step("ab_variant", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
            variant := h.getABTestVariant(userID, "onboarding_flow")
            ctx.Set("ab_variant", variant)
            return nil
        })).
        
        // Step 2: Fetch screen based on variant
        Step("fetch_screen", flow.NewStepWrapper(
            httpsteps.GET("${content_service_url}/api/screens/${screen_id}").
                WithQueryParam("variant", "${ab_variant}").
                SaveAs("screen_config"))).
        
        // Step 3: Log A/B test exposure
        Step("log_exposure", flow.NewStepWrapper(
            httpsteps.POST("${analytics_service_url}/ab-tests/exposure").
                WithJSONBody(map[string]interface{}{
                    "user_id":   "${user_id}",
                    "test_name": "onboarding_flow",
                    "variant":   "${ab_variant}",
                    "timestamp": time.Now(),
                }).
                SaveAs("exposure_logged")))
    
    _, err := abTestFlow.Execute(ctx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    screenData, _ := ctx.Get("screen_config")
    c.JSON(http.StatusOK, screenData)
}

func (h *OnboardingHandler) getABTestVariant(userID, testName string) string {
    // Simple hash-based A/B testing
    hash := fnv.New32a()
    hash.Write([]byte(userID + testName))
    hashValue := hash.Sum32()
    
    if hashValue%100 < 50 {
        return "variant_a"
    }
    return "variant_b"
}
```

### 2. Progressive Data Loading

```go
func (h *DashboardHandler) GetProgressiveDashboard(c *gin.Context) {
    userID := c.Param("userId")
    
    ctx := flow.NewContextWithConfig(h.config)
    ctx.Set("user_id", userID)
    
    progressiveFlow := flow.NewFlow("progressive_dashboard").
        // Step 1: Load critical data first
        Step("critical_data", flow.NewStepWrapper(
            base.NewParallelStep("critical",
                httpsteps.GET("${user_service_url}/users/${user_id}").SaveAs("user"),
                httpsteps.GET("${notification_service_url}/urgent/${user_id}").SaveAs("urgent_notifications"),
            ))).
        
        // Step 2: Send initial response
        Step("initial_response", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
            user, _ := ctx.GetMap("user")
            notifications, _ := ctx.GetMap("urgent_notifications")
            
            initial := map[string]interface{}{
                "user":                 user,
                "urgent_notifications": notifications,
                "loading_additional":   true,
            }
            
            ctx.Set("initial_data", initial)
            return nil
        })).
        
        // Step 3: Load additional data asynchronously
        Step("additional_data", flow.NewStepWrapper(
            base.NewParallelStep("additional",
                httpsteps.GET("${content_service_url}/feed/${user_id}").SaveAs("feed"),
                httpsteps.GET("${analytics_service_url}/recommendations/${user_id}").SaveAs("recommendations"),
                httpsteps.GET("${weather_service_url}/current").SaveAs("weather"),
            ))).
        
        // Step 4: Combine all data
        Transform("combine_data", func(ctx interfaces.ExecutionContext) error {
            initial, _ := ctx.GetMap("initial_data")
            feed, _ := ctx.GetMap("feed")
            recommendations, _ := ctx.GetMap("recommendations")
            weather, _ := ctx.GetMap("weather")
            
            complete := initial.(map[string]interface{})
            complete["feed"] = feed
            complete["recommendations"] = recommendations
            complete["weather"] = weather
            complete["loading_additional"] = false
            
            ctx.Set("complete_dashboard", complete)
            return nil
        })
    
    _, err := progressiveFlow.Execute(ctx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    dashboardData, _ := ctx.Get("complete_dashboard")
    c.JSON(http.StatusOK, dashboardData)
}
```

### 3. Error Handling with Fallbacks

```go
func (h *ContentHandler) GetResilientContent(c *gin.Context) {
    contentID := c.Param("contentId")
    
    ctx := flow.NewContextWithConfig(h.config)
    ctx.Set("content_id", contentID)
    
    resilientFlow := flow.NewFlow("resilient_content").
        // Step 1: Try primary content service
        Step("primary_content", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
            step := httpsteps.GET("${primary_content_url}/content/${content_id}").
                WithTimeout(5 * time.Second).
                SaveAs("primary_content")
            
            err := step.Run(ctx)
            if err != nil {
                ctx.Logger().Warn("Primary content service failed", zap.Error(err))
                // Don't return error - continue to fallback
            }
            return nil
        })).
        
        // Step 2: If primary failed, try fallback
        Choice("content_source").
            When(func(ctx interfaces.ExecutionContext) bool {
                return !ctx.Has("primary_content")
            }).
                Step("fallback_content", flow.NewStepWrapper(
                    httpsteps.GET("${fallback_content_url}/content/${content_id}").
                        WithTimeout(3 * time.Second).
                        SaveAs("fallback_content"))).
                Step("mark_fallback", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
                    ctx.Set("used_fallback", true)
                    return nil
                })).
            Otherwise().
                Step("mark_primary", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
                    ctx.Set("used_fallback", false)
                    return nil
                })).
        EndChoice().
        
        // Step 3: Format response
        Transform("format_response", func(ctx interfaces.ExecutionContext) error {
            var content interface{}
            usedFallback := false
            
            if ctx.Has("primary_content") {
                content, _ = ctx.Get("primary_content")
            } else if ctx.Has("fallback_content") {
                content, _ = ctx.Get("fallback_content")
                usedFallback = true
            } else {
                // Last resort - return cached or default content
                content = map[string]interface{}{
                    "id":      contentID,
                    "title":   "Content Unavailable",
                    "message": "Content is temporarily unavailable. Please try again later.",
                }
                usedFallback = true
            }
            
            response := map[string]interface{}{
                "success":       true,
                "data":          content,
                "used_fallback": usedFallback,
                "timestamp":     time.Now(),
            }
            
            ctx.Set("final_response", response)
            return nil
        })
    
    _, err := resilientFlow.Execute(ctx)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    response, _ := ctx.Get("final_response")
    c.JSON(http.StatusOK, response)
}
```

## Testing Your BFF

### 1. Unit Testing Flows

```go
// handlers/onboarding_test.go
package handlers

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/zap"

    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
    "github.com/venkatvghub/api-orchestration-framework/pkg/flow"
)

func TestOnboardingHandler_GetScreen(t *testing.T) {
    // Setup
    cfg := &config.FrameworkConfig{
        Timeouts: config.TimeoutConfig{
            FlowExecution: 10 * time.Second,
        },
    }
    logger := zap.NewNop()
    handler := NewOnboardingHandler(cfg, logger)
    
    // Create test context
    ctx := flow.NewContextWithConfig(cfg)
    ctx.Set("screen_id", "welcome")
    ctx.Set("user_id", "test-user-123")
    ctx.Set("version", "v1")
    
    // Mock screen data
    mockScreen := map[string]interface{}{
        "id":    "welcome",
        "title": "Welcome to Our App!",
        "type":  "welcome",
    }
    ctx.Set("screen_config", mockScreen)
    
    // Test mobile transformer
    transformer := handler.createMobileTransformer()
    err := transformer(ctx)
    
    // Assertions
    require.NoError(t, err)
    
    mobileScreen, exists := ctx.Get("mobile_screen")
    assert.True(t, exists)
    assert.NotNil(t, mobileScreen)
    
    screenMap := mobileScreen.(map[string]interface{})
    assert.Equal(t, "welcome", screenMap["id"])
    assert.Equal(t, "Welcome to Our App!", screenMap["title"])
}

func TestOnboardingService_GetScreenValidator(t *testing.T) {
    cfg := &config.FrameworkConfig{}
    logger := zap.NewNop()
    service := NewOnboardingService(cfg, logger)
    
    tests := []struct {
        name     string
        screenID string
        data     map[string]interface{}
        wantErr  bool
    }{
        {
            name:     "valid personal info",
            screenID: "personal_info",
            data: map[string]interface{}{
                "first_name": "John",
                "last_name":  "Doe",
                "email":      "john@example.com",
            },
            wantErr: false,
        },
        {
            name:     "invalid email",
            screenID: "personal_info",
            data: map[string]interface{}{
                "first_name": "John",
                "last_name":  "Doe",
                "email":      "invalid-email",
            },
            wantErr: true,
        },
        {
            name:     "missing required field",
            screenID: "personal_info",
            data: map[string]interface{}{
                "first_name": "John",
                // missing last_name and email
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := service.GetScreenValidator(tt.screenID)
            err := validator.Validate(tt.data)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Integration Testing

```go
// integration_test.go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/zap"

    "your-project/handlers"
    "your-project/models"
    "github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

func TestOnboardingIntegration(t *testing.T) {
    // Setup test server
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    cfg := &config.FrameworkConfig{}
    logger := zap.NewNop()
    
    // Setup routes
    api := router.Group("/api/v1")
    onboarding := api.Group("/onboarding")
    {
        handler := handlers.NewOnboardingHandler(cfg, logger)
        onboarding.GET("/screens/:screenId", handler.GetScreen)
        onboarding.POST("/screens/:screenId/submit", handler.SubmitScreen)
    }
    
    t.Run("get screen", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/onboarding/screens/welcome?user_id=test-user", nil)
        req.Header.Set("X-API-Version", "v1")
        req.Header.Set("X-Device-Type", "ios")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response models.APIResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)
        
        assert.True(t, response.Success)
        assert.NotNil(t, response.Data)
    })
    
    t.Run("submit screen", func(t *testing.T) {
        submission := models.ScreenSubmission{
            UserID: "test-user",
            Data: map[string]interface{}{
                "first_name": "John",
                "last_name":  "Doe",
                "email":      "john@example.com",
            },
        }
        
        body, _ := json.Marshal(submission)
        req := httptest.NewRequest("POST", "/api/v1/onboarding/screens/personal_info/submit", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-Device-Type", "ios")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response models.APIResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)
        
        assert.True(t, response.Success)
    })
}
```

### 3. Load Testing

```go
// load_test.go
package main

import (
    "fmt"
    "net/http"
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestLoadOnboardingEndpoint(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }
    
    const (
        numRequests = 1000
        concurrency = 50
    )
    
    // Start test server (you'd start your actual server here)
    serverURL := "http://localhost:8080"
    
    var wg sync.WaitGroup
    results := make(chan time.Duration, numRequests)
    errors := make(chan error, numRequests)
    
    // Create worker pool
    semaphore := make(chan struct{}, concurrency)
    
    start := time.Now()
    
    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            requestStart := time.Now()
            
            resp, err := http.Get(fmt.Sprintf("%s/api/v1/onboarding/screens/welcome?user_id=user-%d", serverURL, id))
            if err != nil {
                errors <- err
                return
            }
            defer resp.Body.Close()
            
            duration := time.Since(requestStart)
            results <- duration
            
            if resp.StatusCode != http.StatusOK {
                errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
            }
        }(i)
    }
    
    wg.Wait()
    close(results)
    close(errors)
    
    totalDuration := time.Since(start)
    
    // Collect results
    var durations []time.Duration
    for duration := range results {
        durations = append(durations, duration)
    }
    
    var errorCount int
    for range errors {
        errorCount++
    }
    
    // Calculate statistics
    if len(durations) > 0 {
        var total time.Duration
        for _, d := range durations {
            total += d
        }
        avgDuration := total / time.Duration(len(durations))
        
        fmt.Printf("Load Test Results:\n")
        fmt.Printf("Total Requests: %d\n", numRequests)
        fmt.Printf("Successful Requests: %d\n", len(durations))
        fmt.Printf("Failed Requests: %d\n", errorCount)
        fmt.Printf("Total Duration: %v\n", totalDuration)
        fmt.Printf("Average Response Time: %v\n", avgDuration)
        fmt.Printf("Requests per Second: %.2f\n", float64(numRequests)/totalDuration.Seconds())
        
        // Assertions
        assert.Less(t, errorCount, numRequests/10, "Error rate should be less than 10%")
        assert.Less(t, avgDuration, 100*time.Millisecond, "Average response time should be less than 100ms")
    }
}
```

## Deployment and Monitoring

### 1. Docker Configuration

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mobile-bff ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/mobile-bff .
COPY --from=builder /app/config ./config

EXPOSE 8080
CMD ["./mobile-bff"]
```

### 2. Docker Compose for Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  mobile-bff:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - USER_SERVICE_URL=http://user-service:8081
      - CONTENT_SERVICE_URL=http://content-service:8082
      - LOG_LEVEL=info
    depends_on:
      - user-service
      - content-service
      - redis
    networks:
      - bff-network

  user-service:
    build: ./cmd/user-service
    ports:
      - "8081:8081"
    networks:
      - bff-network

  content-service:
    build: ./cmd/content-service
    ports:
      - "8082:8082"
    networks:
      - bff-network

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    networks:
      - bff-network

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - bff-network

networks:
  bff-network:
    driver: bridge
```

### 3. Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mobile-bff
  labels:
    app: mobile-bff
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mobile-bff
  template:
    metadata:
      labels:
        app: mobile-bff
    spec:
      containers:
      - name: mobile-bff
        image: your-registry/mobile-bff:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: USER_SERVICE_URL
          value: "http://user-service:8081"
        - name: CONTENT_SERVICE_URL
          value: "http://content-service:8082"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: mobile-bff-service
spec:
  selector:
    app: mobile-bff
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 4. Monitoring Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'mobile-bff'
    static_configs:
      - targets: ['mobile-bff:8080']
    metrics_path: /metrics
    scrape_interval: 5s

  - job_name: 'user-service'
    static_configs:
      - targets: ['user-service:8081']

  - job_name: 'content-service'
    static_configs:
      - targets: ['content-service:8082']
```

### 5. Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Mobile BFF Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(bff_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(bff_request_duration_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(bff_request_duration_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(bff_errors_total[5m])",
            "legendFormat": "{{endpoint}}"
          }
        ]
      }
    ]
  }
}
```

## Common Pitfalls and Solutions

### 1. Over-fetching Data

**Problem**: Fetching too much data from backend services.

**Solution**: Use field selection and mobile transformers.

```go
// Bad: Fetching everything
httpsteps.GET("${user_service_url}/users/${user_id}").SaveAs("user")

// Good: Fetch only what you need
httpsteps.GET("${user_service_url}/users/${user_id}?fields=id,name,email,avatar").SaveAs("user")

// Even better: Use mobile transformer
Transform("mobile_optimize", func(ctx interfaces.ExecutionContext) error {
    transformer := transformers.NewMobileTransformer([]string{"id", "name", "email", "avatar"})
    userData, _ := ctx.GetMap("user")
    optimized, err := transformer.Transform(userData)
    if err != nil {
        return err
    }
    ctx.Set("mobile_user", optimized)
    return nil
})
```

### 2. Not Handling Failures Gracefully

**Problem**: One service failure breaks the entire response.

**Solution**: Use optional steps and fallbacks.

```go
// Bad: All steps are required
Parallel("data").
    Step("critical", criticalStep).
    Step("optional", optionalStep). // If this fails, everything fails
EndParallel()

// Good: Handle optional failures
Parallel("data").
    Step("critical", criticalStep).
    Step("optional", createOptionalStep(optionalStep, defaultValue)).
EndParallel()

func createOptionalStep(step interfaces.Step, defaultValue interface{}) interfaces.Step {
    return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
        if err := step.Run(ctx); err != nil {
            ctx.Logger().Warn("Optional step failed", zap.Error(err))
            ctx.Set("optional_data", defaultValue)
        }
        return nil // Don't fail the flow
    })
}
```

### 3. Poor Caching Strategy

**Problem**: Caching everything or nothing.

**Solution**: Cache strategically based on data characteristics.

```go
// Cache user profile (changes infrequently)
Step("cache_profile", flow.NewStepWrapper(
    core.NewCacheSetStep("profile_cache", "profile_${user_id}", "user_profile", 30*time.Minute)))

// Don't cache real-time data like notifications
Step("fetch_notifications", flow.NewStepWrapper(
    httpsteps.GET("${notification_service_url}/notifications/${user_id}").SaveAs("notifications")))

// Cache expensive computations
Step("cache_recommendations", flow.NewStepWrapper(
    core.NewCacheSetStep("rec_cache", "recommendations_${user_id}", "recommendations", 1*time.Hour)))
```

### 4. Blocking Operations

**Problem**: Sequential operations that could be parallel.

**Solution**: Use parallel steps for independent operations.

```go
// Bad: Sequential operations
Step("user", getUserStep).
Step("notifications", getNotificationsStep).
Step("activity", getActivityStep)

// Good: Parallel operations
Parallel("user_data").
    Step("user", getUserStep).
    Step("notifications", getNotificationsStep).
    Step("activity", getActivityStep).
EndParallel()
```

### 5. Inadequate Error Handling

**Problem**: Generic error messages that don't help debugging.

**Solution**: Structured error handling with context.

```go
func (h *Handler) handleError(c *gin.Context, err error, operation string) {
    h.logger.Error("Operation failed",
        zap.String("operation", operation),
        zap.String("user_id", c.Query("user_id")),
        zap.String("endpoint", c.Request.URL.Path),
        zap.Error(err))

    // Return structured error
    c.JSON(http.StatusInternalServerError, models.APIResponse{
        Success: false,
        Error: &models.APIError{
            Code:    fmt.Sprintf("%s_ERROR", strings.ToUpper(operation)),
            Message: fmt.Sprintf("Failed to %s", operation),
        },
        Timestamp: time.Now(),
    })
}
```

## Best Practices Summary

### 1. Design Principles
- **Single Responsibility**: Each BFF endpoint should serve a specific frontend need
- **Mobile-First**: Optimize for mobile constraints (bandwidth, battery, processing)
- **Fail Fast**: Validate inputs early and provide clear error messages
- **Graceful Degradation**: Handle service failures without breaking the user experience

### 2. Performance Optimization
- **Parallel Execution**: Use parallel steps for independent operations
- **Caching**: Cache frequently accessed, slowly changing data
- **Field Selection**: Only fetch and return data that the frontend needs
- **Compression**: Use response compression for large payloads

### 3. Monitoring and Observability
- **Structured Logging**: Use structured logs with correlation IDs
- **Metrics**: Track request rates, response times, and error rates
- **Health Checks**: Implement comprehensive health checks
- **Distributed Tracing**: Use tracing for complex flows

### 4. Security
- **Authentication**: Validate tokens on every request
- **Authorization**: Check permissions for data access
- **Input Validation**: Validate all inputs thoroughly
- **Rate Limiting**: Implement rate limiting to prevent abuse

### 5. Testing
- **Unit Tests**: Test individual components and transformers
- **Integration Tests**: Test complete flows end-to-end
- **Load Tests**: Verify performance under load
- **Contract Tests**: Ensure compatibility with backend services

This comprehensive guide provides everything a junior developer needs to build robust, scalable BFF layers using the API Orchestration Framework. Start with the basic patterns and gradually incorporate advanced features as your application grows. 