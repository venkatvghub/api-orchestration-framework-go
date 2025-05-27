package config

import (
	"os"
	"strconv"
	"time"
)

// FrameworkConfig holds all framework configuration
type FrameworkConfig struct {
	HTTP     HTTPConfig     `json:"http"`
	Cache    CacheConfig    `json:"cache"`
	Logging  LoggingConfig  `json:"logging"`
	Security SecurityConfig `json:"security"`
	Timeouts TimeoutConfig  `json:"timeouts"`
	Mobile   MobileConfig   `json:"mobile"`
}

// HTTPConfig holds HTTP client configuration
type HTTPConfig struct {
	MaxIdleConns        int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `json:"idle_conn_timeout"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	MaxRetryDelay       time.Duration `json:"max_retry_delay"`
	FailureThreshold    uint          `json:"failure_threshold"`
	SuccessThreshold    uint          `json:"success_threshold"`
	CircuitBreakerDelay time.Duration `json:"circuit_breaker_delay"`
	RequestTimeout      time.Duration `json:"request_timeout"`
	EnableFallback      bool          `json:"enable_fallback"`
	UserAgent           string        `json:"user_agent"`
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	DefaultTTL    time.Duration `json:"default_ttl"`
	MaxSize       int           `json:"max_size"`
	CleanupPeriod time.Duration `json:"cleanup_period"`
	EnableMetrics bool          `json:"enable_metrics"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level            string   `json:"level"`
	Format           string   `json:"format"` // json, console
	EnableStackTrace bool     `json:"enable_stack_trace"`
	SanitizeFields   []string `json:"sanitize_fields"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	TokenValidation TokenValidationConfig `json:"token_validation"`
	RateLimit       RateLimitConfig       `json:"rate_limit"`
}

// TokenValidationConfig holds token validation settings
type TokenValidationConfig struct {
	Algorithm    string        `json:"algorithm"`
	SecretKey    string        `json:"secret_key"`
	Issuer       string        `json:"issuer"`
	ExpiryBuffer time.Duration `json:"expiry_buffer"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// TimeoutConfig holds various timeout settings
type TimeoutConfig struct {
	FlowExecution time.Duration `json:"flow_execution"`
	StepExecution time.Duration `json:"step_execution"`
	HTTPRequest   time.Duration `json:"http_request"`
	CacheOp       time.Duration `json:"cache_op"`
}

// MobileConfig holds mobile-specific optimizations
type MobileConfig struct {
	MaxPayloadSize    int      `json:"max_payload_size"`
	CompressionLevel  int      `json:"compression_level"`
	DefaultFields     []string `json:"default_fields"`
	EnableCompression bool     `json:"enable_compression"`
	OptimizeImages    bool     `json:"optimize_images"`
}

// DefaultConfig returns the default framework configuration
func DefaultConfig() *FrameworkConfig {
	return &FrameworkConfig{
		HTTP: HTTPConfig{
			MaxIdleConns:        getEnvInt("HTTP_MAX_IDLE_CONNS", 100),
			MaxIdleConnsPerHost: getEnvInt("HTTP_MAX_IDLE_CONNS_PER_HOST", 10),
			IdleConnTimeout:     getEnvDuration("HTTP_IDLE_CONN_TIMEOUT", 90*time.Second),
			MaxRetries:          getEnvInt("HTTP_MAX_RETRIES", 3),
			RetryDelay:          getEnvDuration("HTTP_RETRY_DELAY", 1*time.Second),
			MaxRetryDelay:       getEnvDuration("HTTP_MAX_RETRY_DELAY", 10*time.Second),
			FailureThreshold:    uint(getEnvInt("HTTP_FAILURE_THRESHOLD", 5)),
			SuccessThreshold:    uint(getEnvInt("HTTP_SUCCESS_THRESHOLD", 3)),
			CircuitBreakerDelay: getEnvDuration("HTTP_CIRCUIT_BREAKER_DELAY", 5*time.Second),
			RequestTimeout:      getEnvDuration("HTTP_REQUEST_TIMEOUT", 15*time.Second),
			EnableFallback:      getEnvBool("HTTP_ENABLE_FALLBACK", true),
			UserAgent:           getEnvString("HTTP_USER_AGENT", "API-Orchestration-Framework/2.0"),
		},
		Cache: CacheConfig{
			DefaultTTL:    getEnvDuration("CACHE_DEFAULT_TTL", 5*time.Minute),
			MaxSize:       getEnvInt("CACHE_MAX_SIZE", 1000),
			CleanupPeriod: getEnvDuration("CACHE_CLEANUP_PERIOD", 10*time.Minute),
			EnableMetrics: getEnvBool("CACHE_ENABLE_METRICS", true),
		},
		Logging: LoggingConfig{
			Level:            getEnvString("LOG_LEVEL", "info"),
			Format:           getEnvString("LOG_FORMAT", "json"),
			EnableStackTrace: getEnvBool("LOG_ENABLE_STACK_TRACE", false),
			SanitizeFields:   []string{"password", "token", "secret", "key", "authorization"},
		},
		Security: SecurityConfig{
			TokenValidation: TokenValidationConfig{
				Algorithm:    getEnvString("TOKEN_ALGORITHM", "HS256"),
				SecretKey:    getEnvString("TOKEN_SECRET_KEY", ""),
				Issuer:       getEnvString("TOKEN_ISSUER", "api-orchestration-framework"),
				ExpiryBuffer: getEnvDuration("TOKEN_EXPIRY_BUFFER", 5*time.Minute),
			},
			RateLimit: RateLimitConfig{
				RequestsPerSecond: getEnvInt("RATE_LIMIT_RPS", 100),
				BurstSize:         getEnvInt("RATE_LIMIT_BURST", 200),
				WindowSize:        getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
			},
		},
		Timeouts: TimeoutConfig{
			FlowExecution: getEnvDuration("TIMEOUT_FLOW_EXECUTION", 30*time.Second),
			StepExecution: getEnvDuration("TIMEOUT_STEP_EXECUTION", 10*time.Second),
			HTTPRequest:   getEnvDuration("TIMEOUT_HTTP_REQUEST", 15*time.Second),
			CacheOp:       getEnvDuration("TIMEOUT_CACHE_OP", 1*time.Second),
		},
		Mobile: MobileConfig{
			MaxPayloadSize:    getEnvInt("MOBILE_MAX_PAYLOAD_SIZE", 1024*1024), // 1MB
			CompressionLevel:  getEnvInt("MOBILE_COMPRESSION_LEVEL", 6),
			DefaultFields:     []string{"id", "name", "status", "timestamp"},
			EnableCompression: getEnvBool("MOBILE_ENABLE_COMPRESSION", true),
			OptimizeImages:    getEnvBool("MOBILE_OPTIMIZE_IMAGES", true),
		},
	}
}

// Environment variable helpers
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ConfigFromEnv creates configuration from environment variables
func ConfigFromEnv() *FrameworkConfig {
	return DefaultConfig()
}

// Validate validates the configuration
func (c *FrameworkConfig) Validate() error {
	// Add validation logic here
	return nil
}
