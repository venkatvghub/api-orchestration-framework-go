package main

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
)

// Simple models
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Simple cache
var cache = make(map[string]interface{})

func main() {
	// Setup
	logger, _ := zap.NewDevelopment()
	cfg := config.ConfigFromEnv()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add flow middleware
	router.Use(func(c *gin.Context) {
		flowCtx := flow.NewContextFromGinWithLogger(c, cfg, logger)
		c.Set("flow_context", flowCtx)
		c.Next()
	})

	// Routes
	router.GET("/health", healthCheck)
	router.GET("/users/:id", getUserWithFlow)
	router.GET("/dashboard/:userId", getDashboard)

	// Start server
	logger.Info("Starting simple BFF server on :8080")
	router.Run(":8080")
}

func healthCheck(c *gin.Context) {
	c.JSON(200, APIResponse{
		Success: true,
		Data:    map[string]string{"status": "healthy"},
	})
}

// Example 1: Simple user fetch with caching
func getUserWithFlow(c *gin.Context) {
	userID := c.Param("id")

	// Get flow context
	flowCtx, _ := flow.GetFlowContext(c)
	flowCtx.Set("user_id", userID)

	// Create flow
	userFlow := flow.NewFlow("get_user").
		WithDescription("Get user with caching").
		WithTimeout(30 * time.Second)

	userFlow.
		// Step 1: Check cache
		Step("check_cache", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
			userID, _ := ctx.GetString("user_id")
			if cachedUser, exists := cache["user_"+userID]; exists {
				ctx.Set("cached_user", cachedUser)
			}
			return nil
		})).

		// Step 2: Conditional fetch
		Choice("cache_status").
		When(func(ctx interfaces.ExecutionContext) bool {
			return !ctx.Has("cached_user")
		}).
		// Fetch from mock API
		Step("fetch_user", httpsteps.GET("https://jsonplaceholder.typicode.com/users/${user_id}").
			SaveAs("api_user")).
		Step("cache_user", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
			userID, _ := ctx.GetString("user_id")
			apiUser, _ := ctx.Get("api_user")
			cache["user_"+userID] = apiUser
			ctx.Set("user_data", apiUser)
			return nil
		})).
		Otherwise().
		Step("use_cached", flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
			cached, _ := ctx.Get("cached_user")
			ctx.Set("user_data", cached)
			return nil
		})).
		EndChoice().

		// Step 3: Transform for mobile
		Transform("mobile_transform", func(ctx interfaces.ExecutionContext) error {
			userData, _ := ctx.Get("user_data")

			// Parse the user data
			var user map[string]interface{}
			if userBytes, ok := userData.([]byte); ok {
				json.Unmarshal(userBytes, &user)
			} else {
				user = userData.(map[string]interface{})
			}

			// Create mobile-optimized response
			mobileUser := map[string]interface{}{
				"id":   user["id"],
				"name": user["name"],
			}

			ctx.Set("mobile_user", mobileUser)
			return nil
		})

	// Execute flow
	_, err := userFlow.Execute(flowCtx)
	if err != nil {
		c.JSON(500, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	mobileUser, _ := flowCtx.Get("mobile_user")
	c.JSON(200, APIResponse{
		Success: true,
		Data:    mobileUser,
	})
}

// Example 2: Dashboard with parallel data fetching
func getDashboard(c *gin.Context) {
	userID := c.Param("userId")

	// Get flow context
	flowCtx, _ := flow.GetFlowContext(c)
	flowCtx.Set("user_id", userID)

	// Create dashboard flow
	dashboardFlow := flow.NewFlow("dashboard").
		WithDescription("Get dashboard with parallel data fetching").
		WithTimeout(30 * time.Second)

	dashboardFlow.
		// Fetch multiple things in parallel
		Parallel("fetch_dashboard_data").
		Step("user", httpsteps.GET("https://jsonplaceholder.typicode.com/users/${user_id}").
			SaveAs("user_data")).
		Step("posts", httpsteps.GET("https://jsonplaceholder.typicode.com/users/${user_id}/posts").
			SaveAs("posts_data")).
		Step("todos", httpsteps.GET("https://jsonplaceholder.typicode.com/users/${user_id}/todos").
			SaveAs("todos_data")).
		EndParallel().

		// Combine and transform data
		Transform("combine_data", func(ctx interfaces.ExecutionContext) error {
			// Get all the data
			userData, _ := ctx.Get("user_data")
			postsData, _ := ctx.Get("posts_data")
			todosData, _ := ctx.Get("todos_data")

			// Parse user data
			var user map[string]interface{}
			if userBytes, ok := userData.([]byte); ok {
				json.Unmarshal(userBytes, &user)
			}

			// Parse posts data
			var posts []interface{}
			if postsBytes, ok := postsData.([]byte); ok {
				json.Unmarshal(postsBytes, &posts)
			}

			// Parse todos data
			var todos []interface{}
			if todosBytes, ok := todosData.([]byte); ok {
				json.Unmarshal(todosBytes, &todos)
			}

			// Create dashboard
			dashboard := map[string]interface{}{
				"user": map[string]interface{}{
					"id":   user["id"],
					"name": user["name"],
				},
				"stats": map[string]interface{}{
					"total_posts": len(posts),
					"total_todos": len(todos),
				},
				"recent_posts": getFirst3(posts),
			}

			ctx.Set("dashboard", dashboard)
			return nil
		})

	// Execute flow
	_, err := dashboardFlow.Execute(flowCtx)
	if err != nil {
		c.JSON(500, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	dashboard, _ := flowCtx.Get("dashboard")
	c.JSON(200, APIResponse{
		Success: true,
		Data:    dashboard,
	})
}

// Helper function
func getFirst3(items []interface{}) []interface{} {
	if len(items) <= 3 {
		return items
	}
	return items[:3]
}
