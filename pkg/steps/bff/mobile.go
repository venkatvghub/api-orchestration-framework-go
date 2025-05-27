package bff

import (
	"fmt"
	"time"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
	"github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
	"github.com/venkatvghub/api-orchestration-framework/pkg/validators"
	"go.uber.org/zap"
)

// MobileAPIStep represents a mobile-optimized API step with BFF patterns
type MobileAPIStep struct {
	*base.BaseStep
	httpStep     *http.HTTPStep
	fields       []string
	cacheKey     string
	cacheTTL     time.Duration
	fallbackData map[string]interface{}
	metrics      *MobileMetrics
}

// MobileMetrics tracks mobile-specific metrics
type MobileMetrics struct {
	RequestCount   int64
	CacheHits      int64
	CacheMisses    int64
	FallbackUsed   int64
	AverageLatency time.Duration
	ErrorCount     int64
	TransformTime  time.Duration
	ValidationTime time.Duration
}

// NewMobileAPIStep creates a new mobile-optimized API step
func NewMobileAPIStep(name, method, url string, fields []string) *MobileAPIStep {
	httpStep := http.NewJSONAPIStep(method, url).
		WithTransformer(transformers.NewMobileTransformer(fields)).
		WithValidator(validators.NewRequiredFieldsValidator("status")).
		WithTimeout(10*time.Second). // Mobile-optimized timeout
		WithExpectedStatus(200, 201, 202, 204)

	return &MobileAPIStep{
		BaseStep: base.NewBaseStep(name, fmt.Sprintf("Mobile API: %s %s", method, url)),
		httpStep: httpStep,
		fields:   fields,
		cacheTTL: 5 * time.Minute, // Default mobile cache TTL
		metrics:  &MobileMetrics{},
	}
}

// Configuration methods with mobile-specific optimizations

// WithCaching enables response caching for mobile optimization
func (m *MobileAPIStep) WithCaching(key string, ttl time.Duration) *MobileAPIStep {
	m.cacheKey = key
	m.cacheTTL = ttl
	return m
}

// WithFallback sets fallback data for offline scenarios
func (m *MobileAPIStep) WithFallback(data map[string]interface{}) *MobileAPIStep {
	m.fallbackData = data
	return m
}

// WithMobileHeaders adds mobile-specific headers
func (m *MobileAPIStep) WithMobileHeaders(deviceType, appVersion, platform string) *MobileAPIStep {
	m.httpStep.
		WithHeader("X-Device-Type", deviceType).
		WithHeader("X-App-Version", appVersion).
		WithHeader("X-Platform", platform).
		WithHeader("X-Mobile-Optimized", "true")
	return m
}

// WithAuth adds authentication for mobile APIs
func (m *MobileAPIStep) WithAuth(tokenField string) *MobileAPIStep {
	m.httpStep.WithBearerToken("${" + tokenField + "}")
	return m
}

// WithRetry configures mobile-optimized retry policy
func (m *MobileAPIStep) WithRetry(maxRetries int, delay time.Duration) *MobileAPIStep {
	config := &http.ClientConfig{
		MaxRetries:           maxRetries,
		InitialRetryDelay:    delay,
		MaxRetryDelay:        delay * 4,
		RetryableStatusCodes: []int{429, 502, 503, 504, 408}, // Include timeout
		EnableFallback:       true,
		RequestTimeout:       10 * time.Second,
	}
	m.httpStep.WithClientConfig(config)
	return m
}

// SaveAs sets the context key to save response data
func (m *MobileAPIStep) SaveAs(key string) *MobileAPIStep {
	m.httpStep.SaveAs(key)
	return m
}

// Run executes the mobile API step with BFF optimizations
func (m *MobileAPIStep) Run(ctx *flow.Context) error {
	startTime := time.Now()
	m.metrics.RequestCount++

	// Check cache first if enabled
	if m.cacheKey != "" {
		if _, found := m.checkCache(ctx); found {
			m.metrics.CacheHits++
			ctx.Set(m.httpStep.Name()+"_cached", true)

			ctx.Logger().Info("Mobile API cache hit",
				zap.String("step", m.Name()),
				zap.String("cache_key", m.cacheKey),
				zap.Duration("duration", time.Since(startTime)))

			return nil
		}
		m.metrics.CacheMisses++
	}

	// Execute HTTP request
	err := m.httpStep.Run(ctx)
	if err != nil {
		m.metrics.ErrorCount++

		// Try fallback if available
		if m.fallbackData != nil {
			m.metrics.FallbackUsed++
			ctx.Set("mobile_response", m.fallbackData)
			ctx.Set("mobile_fallback_used", true)

			ctx.Logger().Warn("Mobile API fallback used",
				zap.String("step", m.Name()),
				zap.Error(err),
				zap.Duration("duration", time.Since(startTime)))

			return nil // Don't fail on fallback
		}

		ctx.Logger().Error("Mobile API request failed",
			zap.String("step", m.Name()),
			zap.Error(err),
			zap.Duration("duration", time.Since(startTime)))

		return err
	}

	// Cache response if caching is enabled
	if m.cacheKey != "" {
		m.cacheResponse(ctx)
	}

	// Update metrics
	duration := time.Since(startTime)
	m.updateMetrics(duration)

	ctx.Logger().Info("Mobile API request completed",
		zap.String("step", m.Name()),
		zap.Duration("duration", duration),
		zap.Int("field_count", len(m.fields)),
		zap.Bool("cached", m.cacheKey != ""))

	return nil
}

// Helper methods

func (m *MobileAPIStep) checkCache(ctx *flow.Context) (map[string]interface{}, bool) {
	cacheData, exists := ctx.Get("cache_" + m.cacheKey)
	if !exists {
		return nil, false
	}

	// Check TTL (simplified - in production use proper cache with TTL)
	if cacheInfo, ok := cacheData.(map[string]interface{}); ok {
		if data, hasData := cacheInfo["data"].(map[string]interface{}); hasData {
			ctx.Set("mobile_response", data)
			return data, true
		}
	}

	return nil, false
}

func (m *MobileAPIStep) cacheResponse(ctx *flow.Context) {
	if responseData, exists := ctx.Get("mobile_response"); exists {
		cacheData := map[string]interface{}{
			"data":      responseData,
			"timestamp": time.Now(),
			"ttl":       m.cacheTTL,
		}
		ctx.Set("cache_"+m.cacheKey, cacheData)
	}
}

func (m *MobileAPIStep) updateMetrics(duration time.Duration) {
	// Update average latency (simplified calculation)
	if m.metrics.AverageLatency == 0 {
		m.metrics.AverageLatency = duration
	} else {
		m.metrics.AverageLatency = (m.metrics.AverageLatency + duration) / 2
	}
}

// GetMetrics returns mobile-specific metrics
func (m *MobileAPIStep) GetMetrics() *MobileMetrics {
	return m.metrics
}

// Convenience constructors for common mobile API patterns

// NewMobileUserProfileStep creates a step for fetching user profile data
func NewMobileUserProfileStep(baseURL string) *MobileAPIStep {
	fields := []string{"id", "name", "email", "avatar", "preferences", "status"}
	return NewMobileAPIStep("user_profile", "GET", baseURL+"/api/v1/user/profile", fields).
		WithCaching("user_profile", 10*time.Minute).
		WithMobileHeaders("mobile", "1.0", "ios").
		WithAuth("access_token")
}

// NewMobileNotificationsStep creates a step for fetching notifications
func NewMobileNotificationsStep(baseURL string) *MobileAPIStep {
	fields := []string{"id", "title", "message", "type", "timestamp", "read", "priority"}
	return NewMobileAPIStep("notifications", "GET", baseURL+"/api/v1/notifications", fields).
		WithCaching("notifications", 2*time.Minute).
		WithMobileHeaders("mobile", "1.0", "ios").
		WithAuth("access_token").
		WithRetry(2, 1*time.Second)
}

// NewMobileContentStep creates a step for fetching content/feed data
func NewMobileContentStep(baseURL string) *MobileAPIStep {
	fields := []string{"id", "title", "summary", "image", "author", "timestamp", "category"}
	return NewMobileAPIStep("content", "GET", baseURL+"/api/v1/content/feed", fields).
		WithCaching("content_feed", 5*time.Minute).
		WithMobileHeaders("mobile", "1.0", "ios").
		WithAuth("access_token").
		WithFallback(map[string]interface{}{
			"items":  []interface{}{},
			"total":  0,
			"status": "offline",
		})
}

// NewMobileSearchStep creates a step for search functionality
func NewMobileSearchStep(baseURL string) *MobileAPIStep {
	fields := []string{"results", "total", "page", "suggestions", "filters"}
	return NewMobileAPIStep("search", "GET", baseURL+"/api/v1/search", fields).
		WithMobileHeaders("mobile", "1.0", "ios").
		WithAuth("access_token").
		WithRetry(1, 500*time.Millisecond) // Fast retry for search
}

// NewMobileAnalyticsStep creates a step for sending analytics data
func NewMobileAnalyticsStep(baseURL string) *MobileAPIStep {
	fields := []string{"status", "event_id", "timestamp"}
	return NewMobileAPIStep("analytics", "POST", baseURL+"/api/v1/analytics/events", fields).
		WithMobileHeaders("mobile", "1.0", "ios").
		WithAuth("access_token").
		WithRetry(3, 2*time.Second) // More retries for analytics
}
