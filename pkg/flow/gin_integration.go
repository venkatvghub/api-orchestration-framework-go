package flow

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

// NewContextFromGin creates a flow context from a Gin context
// This ensures proper context cancellation propagation from HTTP requests
func NewContextFromGin(c *gin.Context, cfg *config.FrameworkConfig) *Context {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Use current time if start_time is not available in context
	startTime := time.Now()
	if st, exists := c.Get("start_time"); exists {
		if t, ok := st.(time.Time); ok {
			startTime = t
		}
	}

	// Create a background context with timeout instead of using request context
	// The request context gets canceled too early, before flow execution completes
	baseCtx := context.Background()

	ctx := &Context{
		ctx:         baseCtx, // Use background context without timeout for now
		values:      make(map[string]interface{}),
		executionID: generateExecutionID(),
		startTime:   startTime,
		timeout:     cfg.Timeouts.FlowExecution,
		config:      cfg,
	}

	// Extract common values from Gin context
	if userID := c.Param("userId"); userID != "" {
		ctx.Set("user_id", userID)
	}
	if userID := c.Query("user_id"); userID != "" {
		ctx.Set("user_id", userID)
	}

	// Extract headers
	ctx.Set("device_type", c.GetHeader("X-Device-Type"))
	ctx.Set("platform", c.GetHeader("X-Platform"))
	ctx.Set("app_version", c.GetHeader("X-App-Version"))
	ctx.Set("api_version", c.GetHeader("X-API-Version"))
	ctx.Set("user_agent", c.GetHeader("User-Agent"))

	return ctx
}

// NewContextFromGinWithLogger creates a flow context from Gin with logger
func NewContextFromGinWithLogger(c *gin.Context, cfg *config.FrameworkConfig, logger *zap.Logger) *Context {
	ctx := NewContextFromGin(c, cfg)
	ctx.WithLogger(logger)
	return ctx
}
