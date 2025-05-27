package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

// OnboardingFlowContextMiddleware creates flow context for onboarding endpoints
func OnboardingFlowContextMiddleware(cfg *frameworkConfig.FrameworkConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create flow context from Gin context
		flowCtx := flow.NewContextFromGinWithLogger(c, cfg, logger)

		// Extract and set onboarding-specific parameters
		if userID := c.Param("userId"); userID != "" {
			flowCtx.Set("user_id", userID)
		}
		if userID := c.Query("user_id"); userID != "" {
			flowCtx.Set("user_id", userID)
		}
		if screenID := c.Param("screenId"); screenID != "" {
			flowCtx.Set("screen_id", screenID)
		}

		// Set device information
		flowCtx.Set("device_type", c.GetHeader("X-Device-Type"))
		flowCtx.Set("app_version", c.GetHeader("X-App-Version"))
		flowCtx.Set("platform", c.GetHeader("X-Platform"))
		flowCtx.Set("os_version", c.GetHeader("X-OS-Version"))

		// Store flow context in Gin context for handlers to access
		c.Set("flow_context", flowCtx)

		// Get user_id for logging (might be empty)
		userID, _ := flowCtx.GetString("user_id")

		logger.Debug("Onboarding flow context created",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("user_id", userID),
			zap.String("device_type", c.GetHeader("X-Device-Type")))

		c.Next()
	}
}

// OnboardingValidationMiddleware validates onboarding-specific parameters
func OnboardingValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate user_id for onboarding endpoints
		if c.FullPath() != "" {
			switch c.FullPath() {
			case "/api/v1/onboarding/screens/:screenId":
				if c.Request.Method == "GET" {
					userID := c.Query("user_id")
					if userID == "" {
						c.JSON(400, gin.H{
							"success": false,
							"error": gin.H{
								"code":    "MISSING_USER_ID",
								"message": "user_id parameter is required",
							},
							"timestamp": time.Now(),
						})
						c.Abort()
						return
					}
				}
			case "/api/v1/onboarding/flow/:userId",
				"/api/v1/onboarding/progress/:userId",
				"/api/v1/onboarding/analytics/:userId":
				userID := c.Param("userId")
				if userID == "" {
					c.JSON(400, gin.H{
						"success": false,
						"error": gin.H{
							"code":    "MISSING_USER_ID",
							"message": "userId parameter is required",
						},
						"timestamp": time.Now(),
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// OnboardingMetricsMiddleware records onboarding-specific metrics
func OnboardingMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Record onboarding-specific metrics
		if c.FullPath() != "" {
			switch c.FullPath() {
			case "/api/v1/onboarding/screens/:screenId":
				screenID := c.Param("screenId")
				if screenID != "" {
					tags := map[string]string{
						"screen_id":   screenID,
						"device_type": c.GetHeader("X-Device-Type"),
						"method":      c.Request.Method,
						"status":      strconv.Itoa(statusCode),
					}
					metrics.IncrementCounter("onboarding_screen_requests_total", tags)
					metrics.RecordDuration("onboarding_screen_request_duration", duration, tags)
				}
			case "/api/v1/onboarding/screens/:screenId/submit":
				screenID := c.Param("screenId")
				if screenID != "" {
					// Use same 4 labels as handler to avoid cardinality mismatch
					tags := map[string]string{
						"screen_id": screenID,
						"user_id":   "middleware", // placeholder since we don't have user_id in middleware
						"success":   strconv.FormatBool(statusCode < 400),
						"source":    "middleware",
					}
					metrics.IncrementCounter("onboarding_submissions_total", tags)
					metrics.RecordDuration("onboarding_submission_duration", duration, tags)
				}
			}
		}
	}
}

// GetOnboardingFlowContext retrieves the flow context from Gin context
func GetOnboardingFlowContext(c *gin.Context) (interfaces.ExecutionContext, bool) {
	return flow.GetFlowContext(c)
}
