package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/failsafe-go/failsafe-go/fallback"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/failsafe-go/failsafe-go/timeout"
)

// HTTPClient interface for dependency injection and testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error)
}

// ResilientHTTPClient provides HTTP client with comprehensive resilience patterns
type ResilientHTTPClient struct {
	baseClient     *http.Client
	retryPolicy    retrypolicy.RetryPolicy[*http.Response]
	circuitBreaker circuitbreaker.CircuitBreaker[*http.Response]
	timeoutPolicy  timeout.Timeout[*http.Response]
	fallbackPolicy fallback.Fallback[*http.Response]
	config         *ClientConfig
}

// ClientConfig holds comprehensive HTTP client configuration
type ClientConfig struct {
	// Base HTTP client settings
	BaseTimeout         time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	DisableKeepAlives   bool
	DisableCompression  bool

	// Retry policy settings
	MaxRetries           int
	InitialRetryDelay    time.Duration
	MaxRetryDelay        time.Duration
	RetryMultiplier      float64
	RetryJitter          bool
	RetryableStatusCodes []int

	// Circuit breaker settings
	FailureThreshold    uint
	SuccessThreshold    uint
	CircuitBreakerDelay time.Duration
	HalfOpenMaxCalls    uint

	// Timeout settings
	RequestTimeout    time.Duration
	ConnectionTimeout time.Duration
	KeepAliveTimeout  time.Duration

	// Fallback settings
	EnableFallback     bool
	FallbackStatusCode int
	FallbackBody       []byte
	FallbackHeaders    map[string]string

	// Rate limiting
	EnableRateLimit   bool
	RequestsPerSecond int
	BurstSize         int

	// Metrics and monitoring
	EnableMetrics bool
	MetricsPrefix string
}

// DefaultClientConfig returns production-ready default configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		// Base settings
		BaseTimeout:         30 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  false,

		// Retry settings
		MaxRetries:           3,
		InitialRetryDelay:    1 * time.Second,
		MaxRetryDelay:        10 * time.Second,
		RetryMultiplier:      2.0,
		RetryJitter:          true,
		RetryableStatusCodes: []int{429, 502, 503, 504},

		// Circuit breaker settings
		FailureThreshold:    5,
		SuccessThreshold:    3,
		CircuitBreakerDelay: 5 * time.Second,
		HalfOpenMaxCalls:    3,

		// Timeout settings
		RequestTimeout:    15 * time.Second,
		ConnectionTimeout: 5 * time.Second,
		KeepAliveTimeout:  30 * time.Second,

		// Fallback settings
		EnableFallback:     true,
		FallbackStatusCode: 503,
		FallbackBody:       []byte(`{"error":"service_unavailable","message":"Service temporarily unavailable"}`),
		FallbackHeaders:    map[string]string{"Content-Type": "application/json"},

		// Rate limiting
		EnableRateLimit:   false,
		RequestsPerSecond: 100,
		BurstSize:         10,

		// Metrics
		EnableMetrics: true,
		MetricsPrefix: "http_client",
	}
}

// NewResilientHTTPClient creates a new HTTP client with comprehensive resilience
func NewResilientHTTPClient(config *ClientConfig) *ResilientHTTPClient {
	if config == nil {
		config = DefaultClientConfig()
	}

	// Create base HTTP client with optimized transport
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
		DisableCompression:  config.DisableCompression,
	}

	// Create resilience policies
	retryPolicy := createAdvancedRetryPolicy(config)
	circuitBreaker := createAdvancedCircuitBreaker(config)
	timeoutPolicy := createAdvancedTimeoutPolicy(config)
	fallbackPolicy := createAdvancedFallbackPolicy(config)

	// Create failsafe RoundTripper with policy composition
	// Only include non-nil policies
	var policies []failsafe.Policy[*http.Response]
	if fallbackPolicy != nil {
		policies = append(policies, fallbackPolicy)
	}
	policies = append(policies, retryPolicy, circuitBreaker, timeoutPolicy)

	roundTripper := failsafehttp.NewRoundTripper(transport, policies...)

	resilientClient := &ResilientHTTPClient{
		baseClient: &http.Client{
			Transport: roundTripper,
			Timeout:   config.BaseTimeout,
		},
		retryPolicy:    retryPolicy,
		circuitBreaker: circuitBreaker,
		timeoutPolicy:  timeoutPolicy,
		fallbackPolicy: fallbackPolicy,
		config:         config,
	}

	return resilientClient
}

// Do executes an HTTP request with resilience policies
func (c *ResilientHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.baseClient.Do(req)
}

// DoWithContext executes an HTTP request with context and resilience policies
func (c *ResilientHTTPClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return c.baseClient.Do(req)
}

// GetMetrics returns client metrics
func (c *ResilientHTTPClient) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Circuit breaker metrics
	if c.circuitBreaker != nil {
		metrics["circuit_breaker"] = map[string]interface{}{
			"state": c.circuitBreaker.State().String(),
		}
	}

	// Basic metrics
	metrics["config"] = map[string]interface{}{
		"max_retries":       c.config.MaxRetries,
		"request_timeout":   c.config.RequestTimeout.String(),
		"failure_threshold": c.config.FailureThreshold,
	}

	return metrics
}

// IsCircuitBreakerOpen returns true if circuit breaker is open
func (c *ResilientHTTPClient) IsCircuitBreakerOpen() bool {
	if c.circuitBreaker == nil {
		return false
	}
	return c.circuitBreaker.State() == circuitbreaker.OpenState
}

// ResetCircuitBreaker manually resets the circuit breaker
func (c *ResilientHTTPClient) ResetCircuitBreaker() {
	if c.circuitBreaker != nil {
		c.circuitBreaker.Close()
	}
}

// createAdvancedRetryPolicy creates a sophisticated retry policy
func createAdvancedRetryPolicy(config *ClientConfig) retrypolicy.RetryPolicy[*http.Response] {
	builder := retrypolicy.Builder[*http.Response]().
		HandleIf(func(response *http.Response, err error) bool {
			if err != nil {
				return true
			}
			// Retry on specific status codes
			for _, code := range config.RetryableStatusCodes {
				if response.StatusCode == code {
					return true
				}
			}
			return false
		}).
		WithBackoff(config.InitialRetryDelay, config.MaxRetryDelay).
		WithMaxRetries(config.MaxRetries)

	if config.RetryJitter {
		builder = builder.OnRetryScheduled(func(e failsafe.ExecutionScheduledEvent[*http.Response]) {
			// Add jitter logic here if needed
		})
	}

	return builder.Build()
}

// createAdvancedCircuitBreaker creates a sophisticated circuit breaker
func createAdvancedCircuitBreaker(config *ClientConfig) circuitbreaker.CircuitBreaker[*http.Response] {
	return circuitbreaker.Builder[*http.Response]().
		HandleIf(func(response *http.Response, err error) bool {
			if err != nil {
				return true
			}
			return response.StatusCode >= 500
		}).
		WithFailureThreshold(config.FailureThreshold).
		WithSuccessThreshold(config.SuccessThreshold).
		WithDelay(config.CircuitBreakerDelay).
		Build()
}

// createAdvancedTimeoutPolicy creates a comprehensive timeout policy
func createAdvancedTimeoutPolicy(config *ClientConfig) timeout.Timeout[*http.Response] {
	return timeout.Builder[*http.Response](config.RequestTimeout).Build()
}

// createAdvancedFallbackPolicy creates a sophisticated fallback policy
func createAdvancedFallbackPolicy(config *ClientConfig) fallback.Fallback[*http.Response] {
	if !config.EnableFallback {
		return nil
	}

	// Create fallback response
	fallbackResp := &http.Response{
		StatusCode: config.FallbackStatusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(string(config.FallbackBody))),
	}

	// Set fallback headers
	for key, value := range config.FallbackHeaders {
		fallbackResp.Header.Set(key, value)
	}

	return fallback.BuilderWithResult(fallbackResp).
		HandleIf(func(response *http.Response, err error) bool {
			if err != nil {
				return true
			}
			return response.StatusCode >= 500
		}).
		Build()
}

// Global default client instance
var DefaultClient = NewResilientHTTPClient(nil)
