package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/config"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/di"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/handlers"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/middleware"
	frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
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
func startServer(lc fx.Lifecycle, server *http.Server, router *gin.Engine, cfg *frameworkConfig.FrameworkConfig, logger *zap.Logger) {
	// Setup router with middleware and routes
	setupRouter(router, cfg, logger)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting Mobile Onboarding BFF v2 server",
					zap.String("addr", server.Addr),
					zap.String("version", "2.0.0"),
					zap.String("mock_api_integration", "enabled"))

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

func setupRouter(router *gin.Engine, cfg *frameworkConfig.FrameworkConfig, logger *zap.Logger) {
	// Get mock API configuration
	mockAPIConfig := config.GetMockAPIConfig()

	// Add framework middleware for proper context integration
	router.Use(flow.RequestIDMiddleware())
	router.Use(flow.TimingMiddleware())
	// router.Use(flow.TimeoutMiddleware(cfg.Timeouts.HTTPRequest)) // Temporarily disabled
	// router.Use(flow.CancellationMiddleware()) // Causing context cancellation
	router.Use(flow.ErrorHandlingMiddleware(logger))

	// Add application-specific middleware (includes flow context creation)
	router.Use(middleware.OnboardingFlowContextMiddleware(cfg, logger))
	router.Use(middleware.OnboardingValidationMiddleware())
	router.Use(middleware.OnboardingMetricsMiddleware())

	// Add standard middleware
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(cfg.Security.RateLimit))
	router.Use(middleware.DeviceInfoMiddleware())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "2.0.0",
			"service":   "mobile-onboarding-bff-v2",
			"mock_api":  "integrated",
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

				// Screen endpoints
				onboarding.GET("/screens/:screenId", handler.GetScreen)
				onboarding.POST("/screens/:screenId/submit", handler.SubmitScreen)

				// Flow management endpoints
				onboarding.GET("/flow/:userId", handler.GetOnboardingFlow)
				onboarding.POST("/flow/:userId/complete", handler.CompleteOnboarding)

				// Progress tracking
				onboarding.GET("/progress/:userId", handler.GetProgress)

				// Analytics endpoints
				onboarding.GET("/analytics/:userId", handler.GetAnalytics)
				onboarding.POST("/analytics/events", handler.TrackEvent)
			}
		}
	}

	// Demo endpoints for testing
	demo := router.Group("/demo")
	{
		demo.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Mobile Onboarding BFF v2 Demo",
				"endpoints": []string{
					"GET /api/v1/onboarding/screens/{screenId}?user_id={userId}",
					"POST /api/v1/onboarding/screens/{screenId}/submit",
					"GET /api/v1/onboarding/flow/{userId}",
					"POST /api/v1/onboarding/flow/{userId}/complete",
					"GET /api/v1/onboarding/progress/{userId}",
					"GET /api/v1/onboarding/analytics/{userId}",
					"POST /api/v1/onboarding/analytics/events",
				},
				"screens":      []string{"welcome", "personal_info", "preferences", "verification", "completion"},
				"mock_api_url": mockAPIConfig.BaseURL,
			})
		})

		demo.GET("/test/:screenId", func(c *gin.Context) {
			screenId := c.Param("screenId")
			c.JSON(http.StatusOK, gin.H{
				"test_curl": gin.H{
					"get_screen":    "curl -X GET \"http://localhost:8080/api/v1/onboarding/screens/" + screenId + "?user_id=demo123\" -H \"X-Device-Type: ios\"",
					"submit_screen": "curl -X POST \"http://localhost:8080/api/v1/onboarding/screens/" + screenId + "/submit\" -H \"Content-Type: application/json\" -H \"X-Device-Type: ios\" -d '{\"user_id\":\"demo123\",\"data\":{}}'",
				},
			})
		})

		demo.GET("/debug/context", func(c *gin.Context) {
			// Test if flow context is available
			flowCtx, exists := flow.GetFlowContext(c)
			if !exists {
				c.JSON(http.StatusOK, gin.H{
					"flow_context_exists": false,
					"gin_keys":            c.Keys,
				})
				return
			}

			userID, _ := flowCtx.GetString("user_id")
			screenID, _ := flowCtx.GetString("screen_id")
			deviceType, _ := flowCtx.GetString("device_type")

			c.JSON(http.StatusOK, gin.H{
				"flow_context_exists": true,
				"user_id":             userID,
				"screen_id":           screenID,
				"device_type":         deviceType,
				"context_keys":        flowCtx.Keys(),
				"gin_keys":            c.Keys,
			})
		})
	}
}
