package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/handlers"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/middleware"
	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/di"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
)

func main() {
	app := fx.New(
		di.Module(),
		fx.Invoke(startServer),
	)
	app.Run()
}

// startServer is invoked by Fx with DI
func startServer(lc fx.Lifecycle, server *http.Server, router *gin.Engine, cfg *config.FrameworkConfig, logger *zap.Logger) {
	// Setup router with middleware and routes
	setupRouter(router, cfg, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting Mobile BFF server",
					zap.String("addr", server.Addr),
					zap.String("version", "1.0.0"))

				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("Failed to start server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down server...")
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		},
	})
}

func setupRouter(router *gin.Engine, cfg *config.FrameworkConfig, logger *zap.Logger) {
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
				onboarding.GET("/screens/:screenId", handler.GetScreen)
				onboarding.POST("/screens/:screenId/submit", handler.SubmitScreen)
				onboarding.GET("/flow/:userId", handler.GetOnboardingFlow)
				onboarding.POST("/flow/:userId/complete", handler.CompleteOnboarding)
				onboarding.GET("/analytics/:userId", handler.GetAnalytics)
			}
		}
	}
}
