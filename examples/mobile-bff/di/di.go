package di

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
	"github.com/venkatvghub/api-orchestration-framework/pkg/registry"
	httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
	"github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
	"github.com/venkatvghub/api-orchestration-framework/pkg/validators"
)

// Module returns an fx.Option that provides all framework dependencies for mobile onboarding
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			// Core dependencies
			provideConfig,
			provideLogger,
			provideMetrics,

			// Framework components
			provideStepRegistry,
			provideFlowContext,
			provideHTTPClient,
			provideFlowFactory,

			// Transformers - mobile-optimized for onboarding
			provideMobileTransformer,

			// Validators - named providers to avoid conflicts
			fx.Annotated{
				Name:   "required_fields",
				Target: provideRequiredFieldsValidator,
			},
			fx.Annotated{
				Name:   "email",
				Target: provideEmailValidator,
			},

			// Infrastructure
			provideRouter,
			provideServer,
		),
	)
}

// Core Dependencies

// provideConfig loads the framework config
func provideConfig() *config.FrameworkConfig {
	return config.ConfigFromEnv()
}

// provideLogger creates a zap.Logger based on config
func provideLogger(cfg *config.FrameworkConfig) (*zap.Logger, error) {
	if cfg.Logging.Format == "json" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// provideMetrics creates and sets up the metrics collector for mobile onboarding
func provideMetrics() metrics.MetricsCollector {
	metricsCollector := metrics.NewPrometheusMetrics("mobile_onboarding", "v2")
	metrics.SetGlobalMetrics(metricsCollector)
	return metricsCollector
}

// Framework Components

// provideStepRegistry creates a step registry with onboarding-specific steps
func provideStepRegistry(logger *zap.Logger) *registry.StepRegistry {
	stepRegistry := registry.NewStepRegistry()

	// Register onboarding-specific steps
	logger.Info("Registering mobile onboarding steps in step registry")

	return stepRegistry
}

// provideFlowContext creates a flow execution context
func provideFlowContext(cfg *config.FrameworkConfig, logger *zap.Logger) interfaces.ExecutionContext {
	ctx := flow.NewContextWithConfig(cfg)
	ctx.WithLogger(logger)
	return ctx
}

// provideHTTPClient creates a resilient HTTP client for mock API calls
func provideHTTPClient(config *config.FrameworkConfig) httpsteps.HTTPClient {
	// Create custom HTTP client config with increased timeouts
	clientConfig := &httpsteps.ClientConfig{
		// Base settings - increased timeouts
		BaseTimeout:         120 * time.Second, // Increased from 30s to 120s
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

		// Timeout settings - increased timeouts
		RequestTimeout:    60 * time.Second, // Increased from 15s to 60s
		ConnectionTimeout: 10 * time.Second, // Increased from 5s to 10s
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

	return httpsteps.NewResilientHTTPClient(clientConfig)
}

// Transformers

// provideMobileTransformer creates a mobile-optimized transformer for onboarding screens
func provideMobileTransformer() transformers.Transformer {
	return transformers.NewMobileTransformer([]string{"id", "title", "description", "type", "fields", "actions", "next_screen"})
}

// Validators

// provideRequiredFieldsValidator creates a required fields validator for onboarding data
func provideRequiredFieldsValidator() validators.Validator {
	return validators.NewRequiredFieldsValidator("user_id", "screen_id")
}

// provideEmailValidator creates an email validator for user registration
func provideEmailValidator() validators.Validator {
	return validators.EmailRequiredValidator("email")
}

// Infrastructure

// provideRouter creates a Gin router for the mobile onboarding BFF
func provideRouter(cfg *config.FrameworkConfig) *gin.Engine {
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	return gin.New()
}

// provideServer creates the HTTP server for mobile onboarding
func provideServer(cfg *config.FrameworkConfig, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", getPort()),
		Handler:      router,
		ReadTimeout:  cfg.Timeouts.HTTPRequest,
		WriteTimeout: cfg.Timeouts.HTTPRequest,
		IdleTimeout:  cfg.HTTP.IdleConnTimeout,
	}
}

// Utility functions

// getPort returns the port from environment or default
func getPort() int {
	if port := os.Getenv("PORT"); port != "" {
		return parseInt(port, 8080)
	}
	return 8080
}

// parseInt parses string to int with fallback
func parseInt(s string, defaultValue int) int {
	if i, err := fmt.Sscanf(s, "%d", &defaultValue); err != nil || i != 1 {
		return defaultValue
	}
	return defaultValue
}

// Factory functions for creating instances with DI

// FlowFactory creates flows with injected dependencies for mobile onboarding
type FlowFactory struct {
	Config   *config.FrameworkConfig
	Logger   *zap.Logger
	Registry *registry.StepRegistry
}

// NewFlowFactory creates a flow factory with DI
func NewFlowFactory(cfg *config.FrameworkConfig, logger *zap.Logger, registry *registry.StepRegistry) *FlowFactory {
	return &FlowFactory{
		Config:   cfg,
		Logger:   logger,
		Registry: registry,
	}
}

// CreateFlow creates a new flow with injected dependencies
func (ff *FlowFactory) CreateFlow(name string) *flow.Flow {
	return flow.NewFlow(name)
}

// CreateContext creates a new execution context with injected dependencies
func (ff *FlowFactory) CreateContext() interfaces.ExecutionContext {
	ctx := flow.NewContextWithConfig(ff.Config)
	ctx.WithLogger(ff.Logger)
	return ctx
}

// Add FlowFactory to DI
func provideFlowFactory(cfg *config.FrameworkConfig, logger *zap.Logger, registry *registry.StepRegistry) *FlowFactory {
	return NewFlowFactory(cfg, logger, registry)
}
