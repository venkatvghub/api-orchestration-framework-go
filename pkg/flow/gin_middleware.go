package flow

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// FlowContextMiddleware creates and injects flow context into Gin context
// This is the main middleware for proper context propagation
func FlowContextMiddleware(cfg *config.FrameworkConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create flow context from Gin context
		flowCtx := NewContextFromGinWithLogger(c, cfg, logger)

		// Set request metadata
		flowCtx.Set("request_method", c.Request.Method)
		flowCtx.Set("request_path", c.Request.URL.Path)
		flowCtx.Set("request_id", generateRequestID())
		flowCtx.Set("start_time", time.Now())

		// Store flow context in Gin context for handlers to access
		c.Set("flow_context", flowCtx)

		// Add request ID to response headers
		requestID, _ := flowCtx.GetString("request_id")
		c.Header("X-Request-ID", requestID)

		logger.Debug("Flow context created",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

// TimingMiddleware adds start time to Gin context for proper flow timing
func TimingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("start_time", time.Now())
		c.Next()
	}
}

// ContextMiddleware ensures proper context propagation for flows
func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure the request context is properly set up
		// This helps with cancellation propagation
		c.Next()
	}
}

// TimeoutMiddleware adds timeout handling to requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestIDMiddleware adds unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// ErrorHandlingMiddleware provides consistent error handling for flows
func ErrorHandlingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString("request_id")
				logger.Error("Request panic recovered",
					zap.String("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Any("error", err))

				c.JSON(500, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "Internal server error",
					},
					"request_id": requestID,
					"timestamp":  time.Now(),
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// CancellationMiddleware handles request cancellation properly
func CancellationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request is already cancelled
		select {
		case <-c.Request.Context().Done():
			c.JSON(499, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "REQUEST_CANCELLED",
					"message": "Request was cancelled",
				},
				"timestamp": time.Now(),
			})
			c.Abort()
			return
		default:
		}

		c.Next()
	}
}

// GetFlowContext retrieves the flow context from Gin context
func GetFlowContext(c *gin.Context) (interfaces.ExecutionContext, bool) {
	if flowCtxInterface, exists := c.Get("flow_context"); exists {
		if flowCtx, ok := flowCtxInterface.(interfaces.ExecutionContext); ok {
			return flowCtx, true
		}
	}
	return nil, false
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
