package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/handlers"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/middleware"
	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

func main() {
	// Load configuration
	cfg := config.ConfigFromEnv()

	// Setup logging
	logger, err := setupLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}
	defer logger.Sync()

	// Setup metrics
	setupMetrics()

	// Create Gin router
	router := setupRouter(cfg, logger)

	// Setup server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", getPort()),
		Handler:      router,
		ReadTimeout:  cfg.Timeouts.HTTPRequest,
		WriteTimeout: cfg.Timeouts.HTTPRequest,
		IdleTimeout:  cfg.HTTP.IdleConnTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting Mobile BFF server",
			zap.String("addr", server.Addr),
			zap.String("version", "1.0.0"))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func setupLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Format == "json" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	// Set log level
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return logger.WithOptions(zap.IncreaseLevel(level.Level())), nil
}

func setupMetrics() {
	// Setup Prometheus metrics
	metricsCollector := metrics.NewPrometheusMetrics("mobile_bff", "onboarding")
	metrics.SetGlobalMetrics(metricsCollector)
}

func setupRouter(cfg *config.FrameworkConfig, logger *zap.Logger) *gin.Engine {
	// Set Gin mode based on log level
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(cfg.Security.RateLimit))
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// API routes
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			onboarding := v1.Group("/onboarding")
			{
				handler := handlers.NewOnboardingHandler(cfg, logger)

				// Screen endpoints with versioning
				onboarding.GET("/screens/:screenId", handler.GetScreen)
				onboarding.POST("/screens/:screenId/submit", handler.SubmitScreen)
				onboarding.GET("/flow/:userId", handler.GetOnboardingFlow)
				onboarding.POST("/flow/:userId/complete", handler.CompleteOnboarding)
			}
		}

		v2 := api.Group("/v2")
		{
			onboarding := v2.Group("/onboarding")
			{
				handler := handlers.NewOnboardingHandlerV2(cfg, logger)

				// Enhanced endpoints for v2
				onboarding.GET("/screens/:screenId", handler.GetScreen)
				onboarding.POST("/screens/:screenId/submit", handler.SubmitScreen)
				onboarding.GET("/flow/:userId", handler.GetOnboardingFlow)
				onboarding.POST("/flow/:userId/complete", handler.CompleteOnboarding)
				onboarding.GET("/analytics/:userId", handler.GetAnalytics)
			}
		}
	}

	return router
}

func getPort() int {
	if port := os.Getenv("PORT"); port != "" {
		return parseInt(port, 8080)
	}
	return 8080
}

func parseInt(s string, defaultValue int) int {
	if i, err := fmt.Sscanf(s, "%d", &defaultValue); err != nil || i != 1 {
		return defaultValue
	}
	return defaultValue
}
