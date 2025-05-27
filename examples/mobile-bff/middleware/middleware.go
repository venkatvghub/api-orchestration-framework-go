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
			zap.String("device_type", param.Request.Header.Get("X-Device-Type")),
			zap.String("app_version", param.Request.Header.Get("X-App-Version")),
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
			"method":      c.Request.Method,
			"endpoint":    c.FullPath(),
			"status":      strconv.Itoa(statusCode),
			"device_type": c.GetHeader("X-Device-Type"),
			"app_version": c.GetHeader("X-App-Version"),
			"service":     "mobile-bff",
		}

		metrics.IncrementCounter("bff_requests_total", tags)
		metrics.RecordDuration("bff_request_duration", duration, tags)

		if statusCode >= 400 {
			metrics.IncrementCounter("bff_errors_total", tags)
		}

		// Record onboarding-specific metrics
		if c.FullPath() != "" && len(c.FullPath()) > 0 {
			if c.FullPath() == "/api/v1/onboarding/screens/:screenId" {
				screenId := c.Param("screenId")
				if screenId != "" {
					screenTags := map[string]string{
						"screen_id":   screenId,
						"device_type": c.GetHeader("X-Device-Type"),
						"method":      c.Request.Method,
						"status":      strconv.Itoa(statusCode),
					}
					metrics.IncrementCounter("onboarding_screen_requests_total", screenTags)
				}
			}
		}
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Device-Type, X-App-Version, X-Platform, X-OS-Version")
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
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Rate limit exceeded. Please try again later.",
				},
				"retry_after": cfg.WindowSize.Seconds(),
				"timestamp":   time.Now(),
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
		osVersion := c.GetHeader("X-OS-Version")

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
		c.Set("os_version", osVersion)

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
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

// ValidationMiddleware validates common request parameters
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate user_id parameter for onboarding endpoints
		if c.FullPath() != "" && len(c.FullPath()) > 0 {
			if c.FullPath() == "/api/v1/onboarding/screens/:screenId" && c.Request.Method == "GET" {
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
				c.Set("user_id", userID)
			}
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
