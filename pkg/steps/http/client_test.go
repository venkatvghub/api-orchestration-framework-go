package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig()

	// Test base settings
	assert.Equal(t, 30*time.Second, config.BaseTimeout)
	assert.Equal(t, 100, config.MaxIdleConns)
	assert.Equal(t, 10, config.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, config.IdleConnTimeout)
	assert.False(t, config.DisableKeepAlives)
	assert.False(t, config.DisableCompression)

	// Test retry settings
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialRetryDelay)
	assert.Equal(t, 10*time.Second, config.MaxRetryDelay)
	assert.Equal(t, 2.0, config.RetryMultiplier)
	assert.True(t, config.RetryJitter)
	assert.Equal(t, []int{429, 502, 503, 504}, config.RetryableStatusCodes)

	// Test circuit breaker settings
	assert.Equal(t, uint(5), config.FailureThreshold)
	assert.Equal(t, uint(3), config.SuccessThreshold)
	assert.Equal(t, 5*time.Second, config.CircuitBreakerDelay)
	assert.Equal(t, uint(3), config.HalfOpenMaxCalls)

	// Test timeout settings
	assert.Equal(t, 15*time.Second, config.RequestTimeout)
	assert.Equal(t, 5*time.Second, config.ConnectionTimeout)
	assert.Equal(t, 30*time.Second, config.KeepAliveTimeout)

	// Test fallback settings
	assert.True(t, config.EnableFallback)
	assert.Equal(t, 503, config.FallbackStatusCode)
	assert.NotNil(t, config.FallbackBody)
	assert.NotNil(t, config.FallbackHeaders)
	assert.Equal(t, "application/json", config.FallbackHeaders["Content-Type"])

	// Test rate limiting
	assert.False(t, config.EnableRateLimit)
	assert.Equal(t, 100, config.RequestsPerSecond)
	assert.Equal(t, 10, config.BurstSize)

	// Test metrics
	assert.True(t, config.EnableMetrics)
	assert.Equal(t, "http_client", config.MetricsPrefix)
}

func TestNewResilientHTTPClient(t *testing.T) {
	config := DefaultClientConfig()
	client := NewResilientHTTPClient(config)

	assert.NotNil(t, client)
	assert.NotNil(t, client.baseClient)
	assert.NotNil(t, client.retryPolicy)
	assert.NotNil(t, client.circuitBreaker)
	assert.NotNil(t, client.timeoutPolicy)
	assert.NotNil(t, client.fallbackPolicy)
	assert.Equal(t, config, client.config)
	assert.Equal(t, config.BaseTimeout, client.baseClient.Timeout)
}

func TestNewResilientHTTPClient_NilConfig(t *testing.T) {
	client := NewResilientHTTPClient(nil)

	assert.NotNil(t, client)
	assert.NotNil(t, client.config)
	// Should use default config when nil is passed
	assert.Equal(t, 30*time.Second, client.config.BaseTimeout)
}

func TestResilientHTTPClient_Do(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	config := &ClientConfig{
		BaseTimeout:      10 * time.Second,
		MaxRetries:       1,
		RequestTimeout:   5 * time.Second,
		FailureThreshold: 3,
		EnableFallback:   false,
	}
	client := NewResilientHTTPClient(config)

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResilientHTTPClient_DoWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	config := &ClientConfig{
		BaseTimeout:      10 * time.Second,
		MaxRetries:       1,
		RequestTimeout:   5 * time.Second,
		FailureThreshold: 3,
		EnableFallback:   false,
	}
	client := NewResilientHTTPClient(config)

	ctx := context.Background()
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithContext(ctx, req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResilientHTTPClient_GetMetrics(t *testing.T) {
	config := DefaultClientConfig()
	client := NewResilientHTTPClient(config)

	metrics := client.GetMetrics()

	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "circuit_breaker")
	assert.Contains(t, metrics, "config")

	circuitBreakerMetrics, ok := metrics["circuit_breaker"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, circuitBreakerMetrics, "state")

	configMetrics, ok := metrics["config"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 3, configMetrics["max_retries"])
	assert.Equal(t, "15s", configMetrics["request_timeout"])
	assert.Equal(t, uint(5), configMetrics["failure_threshold"])
}

func TestResilientHTTPClient_IsCircuitBreakerOpen(t *testing.T) {
	config := DefaultClientConfig()
	client := NewResilientHTTPClient(config)

	// Initially circuit breaker should be closed
	assert.False(t, client.IsCircuitBreakerOpen())
}

func TestResilientHTTPClient_ResetCircuitBreaker(t *testing.T) {
	config := DefaultClientConfig()
	client := NewResilientHTTPClient(config)

	// Should not panic when resetting circuit breaker
	assert.NotPanics(t, func() {
		client.ResetCircuitBreaker()
	})
}

func TestClientConfig_CustomConfiguration(t *testing.T) {
	config := &ClientConfig{
		BaseTimeout:         60 * time.Second,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     120 * time.Second,
		DisableKeepAlives:   true,
		DisableCompression:  true,

		MaxRetries:           5,
		InitialRetryDelay:    2 * time.Second,
		MaxRetryDelay:        20 * time.Second,
		RetryMultiplier:      3.0,
		RetryJitter:          false,
		RetryableStatusCodes: []int{500, 502, 503},

		FailureThreshold:    10,
		SuccessThreshold:    5,
		CircuitBreakerDelay: 10 * time.Second,
		HalfOpenMaxCalls:    5,

		RequestTimeout:    30 * time.Second,
		ConnectionTimeout: 10 * time.Second,
		KeepAliveTimeout:  60 * time.Second,

		EnableFallback:     false,
		FallbackStatusCode: 500,
		FallbackBody:       []byte("custom fallback"),
		FallbackHeaders:    map[string]string{"Custom": "Header"},

		EnableRateLimit:   true,
		RequestsPerSecond: 50,
		BurstSize:         5,

		EnableMetrics: false,
		MetricsPrefix: "custom_prefix",
	}

	client := NewResilientHTTPClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, config, client.config)
	assert.Equal(t, config.BaseTimeout, client.baseClient.Timeout)

	// Verify custom configuration is applied
	assert.Equal(t, 60*time.Second, client.config.BaseTimeout)
	assert.Equal(t, 5, client.config.MaxRetries)
	assert.Equal(t, uint(10), client.config.FailureThreshold)
	assert.False(t, client.config.EnableFallback)
	assert.True(t, client.config.EnableRateLimit)
	assert.False(t, client.config.EnableMetrics)
}

func TestClientConfig_ZeroValues(t *testing.T) {
	config := &ClientConfig{}
	client := NewResilientHTTPClient(config)

	assert.NotNil(t, client)
	assert.NotNil(t, client.baseClient)

	// Should handle zero values gracefully
	assert.Equal(t, config, client.config)
}

func TestResilientHTTPClient_NetworkError(t *testing.T) {
	config := &ClientConfig{
		BaseTimeout:      1 * time.Second,
		MaxRetries:       2,
		RequestTimeout:   500 * time.Millisecond,
		FailureThreshold: 3,
		EnableFallback:   false,
	}
	client := NewResilientHTTPClient(config)

	req, err := http.NewRequest("GET", "http://nonexistent.example.com", nil)
	require.NoError(t, err)

	_, err = client.Do(req)
	assert.Error(t, err)
}

func TestResilientHTTPClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &ClientConfig{
		BaseTimeout:      10 * time.Second,
		MaxRetries:       0,
		RequestTimeout:   5 * time.Second,
		FailureThreshold: 3,
		EnableFallback:   false,
	}
	client := NewResilientHTTPClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	_, err = client.DoWithContext(ctx, req)
	assert.Error(t, err)
}

func TestResilientHTTPClient_HTTPSRequest(t *testing.T) {
	// Create HTTPS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("secure response"))
	}))
	defer server.Close()

	config := &ClientConfig{
		BaseTimeout:      10 * time.Second,
		MaxRetries:       1,
		RequestTimeout:   5 * time.Second,
		FailureThreshold: 3,
		EnableFallback:   false,
	}
	client := NewResilientHTTPClient(config)

	// Use the test server's client to handle TLS
	client.baseClient = server.Client()

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
