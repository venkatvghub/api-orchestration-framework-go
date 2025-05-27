package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

// LoggingMiddleware provides structured logging for all requests
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("body_size", param.BodySize),
		)
		return ""
	})
}

// MetricsMiddleware records HTTP metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Record HTTP request metrics
		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			statusCode,
			duration,
		)

		// Record custom BFF metrics
		tags := map[string]string{
			"method":   c.Request.Method,
			"endpoint": c.FullPath(),
			"status":   strconv.Itoa(statusCode),
			"version":  getAPIVersion(c),
		}

		metrics.IncrementCounter("bff_requests_total", tags)
		metrics.RecordDuration("bff_request_duration", duration, tags)

		if statusCode >= 400 {
			metrics.IncrementCounter("bff_errors_total", tags)
		}
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Version, X-Device-Type, X-App-Version")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting per IP
func RateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": cfg.WindowSize.Seconds(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// DeviceInfoMiddleware extracts device information from headers
func DeviceInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceType := c.GetHeader("X-Device-Type")
		appVersion := c.GetHeader("X-App-Version")
		platform := c.GetHeader("X-Platform")

		if deviceType == "" {
			deviceType = "unknown"
		}
		if appVersion == "" {
			appVersion = "unknown"
		}
		if platform == "" {
			platform = "unknown"
		}

		c.Set("device_type", deviceType)
		c.Set("app_version", appVersion)
		c.Set("platform", platform)

		c.Next()
	}
}

// VersionMiddleware extracts API version from headers or path
func VersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("X-API-Version")
		if version == "" {
			// Extract from path
			version = getAPIVersion(c)
		}

		c.Set("api_version", version)
		c.Next()
	}
}

// AuthMiddleware validates authentication tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{
				"error": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Simple token validation (in production, use proper JWT validation)
		if len(token) < 10 {
			c.JSON(401, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		c.Set("user_token", token)
		c.Next()
	}
}

// getAPIVersion extracts API version from the request path
func getAPIVersion(c *gin.Context) string {
	path := c.FullPath()
	if len(path) > 7 && path[:7] == "/api/v1" {
		return "v1"
	}
	if len(path) > 7 && path[:7] == "/api/v2" {
		return "v2"
	}
	return "unknown"
}
