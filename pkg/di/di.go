package di

import (
	"fmt"
	"net/http"
	"os"

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

// Module returns an fx.Option that provides all framework dependencies
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

			// Transformers
			provideFieldTransformer,
			provideMobileTransformer,
			provideFlattenTransformer,

			// Validators
			provideRequiredFieldsValidator,
			provideEmailValidator,

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

// provideMetrics creates and sets up the metrics collector
func provideMetrics() metrics.MetricsCollector {
	metricsCollector := metrics.NewPrometheusMetrics("mobile_bff", "onboarding")
	metrics.SetGlobalMetrics(metricsCollector)
	return metricsCollector
}

// Framework Components

// provideStepRegistry creates a step registry with common steps
func provideStepRegistry(logger *zap.Logger) *registry.StepRegistry {
	stepRegistry := registry.NewStepRegistry()

	// Register common steps
	logger.Info("Registering common steps in step registry")

	return stepRegistry
}

// provideFlowContext creates a flow execution context
func provideFlowContext(cfg *config.FrameworkConfig, logger *zap.Logger) interfaces.ExecutionContext {
	ctx := flow.NewContextWithConfig(cfg)
	ctx.WithLogger(logger)
	return ctx
}

// provideHTTPClient creates a resilient HTTP client
func provideHTTPClient(cfg *config.FrameworkConfig) httpsteps.HTTPClient {
	clientConfig := &httpsteps.ClientConfig{
		BaseTimeout:         cfg.Timeouts.HTTPRequest,
		MaxIdleConns:        cfg.HTTP.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.HTTP.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.HTTP.IdleConnTimeout,
		MaxRetries:          cfg.HTTP.MaxRetries,
		InitialRetryDelay:   cfg.HTTP.RetryDelay,
		MaxRetryDelay:       cfg.HTTP.MaxRetryDelay,
		FailureThreshold:    cfg.HTTP.FailureThreshold,
		SuccessThreshold:    cfg.HTTP.SuccessThreshold,
		CircuitBreakerDelay: cfg.HTTP.CircuitBreakerDelay,
		RequestTimeout:      cfg.Timeouts.HTTPRequest,
		EnableFallback:      cfg.HTTP.EnableFallback,
	}

	return httpsteps.NewResilientHTTPClient(clientConfig)
}

// Transformers

// provideFieldTransformer creates a field transformer
func provideFieldTransformer() transformers.Transformer {
	return transformers.NewFieldTransformer("default_field_transformer",
		[]string{"id", "name", "status", "timestamp"})
}

// provideMobileTransformer creates a mobile-optimized transformer
func provideMobileTransformer() transformers.Transformer {
	return transformers.NewMobileTransformer([]string{"id", "name", "avatar", "status"})
}

// provideFlattenTransformer creates a flatten transformer
func provideFlattenTransformer() transformers.Transformer {
	return transformers.NewFlattenTransformer("default_flatten", "mobile")
}

// Validators

// provideRequiredFieldsValidator creates a required fields validator
func provideRequiredFieldsValidator() validators.Validator {
	return validators.NewRequiredFieldsValidator("id", "name")
}

// provideEmailValidator creates an email validator
func provideEmailValidator() validators.Validator {
	return validators.EmailRequiredValidator("email")
}

// Infrastructure

// provideRouter creates a basic Gin router
func provideRouter(cfg *config.FrameworkConfig) *gin.Engine {
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	return gin.New()
}

// provideServer creates the HTTP server
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

// FlowFactory creates flows with injected dependencies
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
