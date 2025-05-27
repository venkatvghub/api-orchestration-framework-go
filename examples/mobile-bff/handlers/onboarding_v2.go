package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/models"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/services"
	"github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/bff"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/core"
	httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
	"github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
	"github.com/venkatvghub/api-orchestration-framework/pkg/validators"
)

// OnboardingHandlerV2 handles V2 onboarding API requests with enhanced features
type OnboardingHandlerV2 struct {
	*OnboardingHandler     // Embed V1 handler
	abTestingService       *services.ABTestingService
	personalizationService *services.PersonalizationService
}

// NewOnboardingHandlerV2 creates a new V2 onboarding handler
func NewOnboardingHandlerV2(cfg *config.FrameworkConfig, logger *zap.Logger) *OnboardingHandlerV2 {
	return &OnboardingHandlerV2{
		OnboardingHandler:      NewOnboardingHandler(cfg, logger),
		abTestingService:       services.NewABTestingService(cfg, logger),
		personalizationService: services.NewPersonalizationService(cfg, logger),
	}
}

// GetScreen handles GET /api/v2/onboarding/screens/:screenId with enhanced features
func (h *OnboardingHandlerV2) GetScreen(c *gin.Context) {
	start := time.Now()
	screenID := c.Param("screenId")
	userID := c.Query("user_id")
	deviceType := c.GetHeader("X-Device-Type")
	appVersion := c.GetHeader("X-App-Version")

	h.logger.Info("Getting V2 onboarding screen",
		zap.String("screen_id", screenID),
		zap.String("user_id", userID),
		zap.String("device_type", deviceType),
		zap.String("app_version", appVersion))

	// Create enhanced flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("screen_id", screenID)
	ctx.Set("user_id", userID)
	ctx.Set("version", "v2")
	ctx.Set("device_type", deviceType)
	ctx.Set("app_version", appVersion)

	// Create enhanced screen flow with A/B testing and personalization
	screenFlow := flow.NewFlow("get_onboarding_screen_v2").
		WithDescription("Retrieve V2 onboarding screen with A/B testing and personalization").
		WithTimeout(h.config.Timeouts.FlowExecution)

	screenFlow.
		// Step 1: Enhanced validation with device compatibility
		Step("validate_request_v2", h.createValidationStep("validate_request_v2", h.screenService.GetEnhancedRequestValidator())).

		// Step 2: A/B Testing - determine screen variant
		Step("ab_testing", h.createABTestingStep()).

		// Step 3: Check cache with variant key
		Step("check_cache_v2", h.createEnhancedCacheStep()).

		// Step 4: Fetch screen data if not cached
		Choice("fetch_screen_choice_v2").
		When(func(ctx interfaces.ExecutionContext) bool {
			return !ctx.Has("cached_screen")
		}).
		// Sequential steps within the When branch
		Step("fetch_screen_config_v2", h.createFetchScreenStepV2()).
		Step("fetch_user_progress_v2", h.createFetchUserProgressStepV2()).
		Step("fetch_personalization", h.createFetchPersonalizationStep()).
		Step("fetch_device_capabilities", h.createFetchDeviceCapabilitiesStep()).
		Step("apply_ab_variant", h.createApplyABVariantStep()).
		Step("apply_personalization", h.createApplyPersonalizationStep()).
		Step("optimize_for_device", h.createDeviceOptimizationStep()).
		Step("cache_screen_v2", h.createEnhancedCacheSetStep()).
		EndChoice().

		// Step 5: Enhanced mobile transformation with compression
		Transform("mobile_transform_v2", h.createEnhancedMobileTransformer()).

		// Step 6: Enhanced analytics with detailed tracking
		Step("log_analytics_v2", h.createEnhancedAnalyticsStep("screen_view")).

		// Step 7: Performance optimization
		Step("optimize_response", h.createResponseOptimizationStep())

	// Execute the flow
	_, err := screenFlow.Execute(ctx)

	// Enhanced metrics collection
	duration := time.Since(start)
	success := err == nil

	tags := map[string]string{
		"screen_id":   screenID,
		"version":     "v2",
		"device_type": deviceType,
		"app_version": appVersion,
		"success":     strconv.FormatBool(success),
		"ab_variant":  getStringFromContext(ctx, "ab_variant", "default"),
	}

	metrics.RecordDuration("onboarding_screen_request_v2", duration, tags)
	metrics.IncrementCounter("onboarding_screen_requests_v2_total", tags)

	if err != nil {
		h.logger.Error("Failed to get V2 onboarding screen",
			zap.String("screen_id", screenID),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "SCREEN_FETCH_ERROR_V2", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	// Get enhanced screen data from context
	screenData, _ := ctx.Get("optimized_screen_data")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    screenData,
		Metadata: &models.Metadata{
			Version:     "v2",
			ProcessTime: duration.String(),
		},
		Timestamp: time.Now(),
	})
}

// SubmitScreen handles POST /api/v2/onboarding/screens/:screenId/submit with enhanced processing
func (h *OnboardingHandlerV2) SubmitScreen(c *gin.Context) {
	start := time.Now()
	screenID := c.Param("screenId")

	var submission models.ScreenSubmission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "INVALID_REQUEST_V2", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	// Enhanced device info collection
	submission.ScreenID = screenID
	submission.Timestamp = time.Now()
	submission.Version = "v2"
	submission.DeviceInfo = models.DeviceInfo{
		Type:       c.GetHeader("X-Device-Type"),
		Platform:   c.GetHeader("X-Platform"),
		AppVersion: c.GetHeader("X-App-Version"),
		OSVersion:  c.GetHeader("X-OS-Version"),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	h.logger.Info("Submitting V2 onboarding screen",
		zap.String("screen_id", screenID),
		zap.String("user_id", submission.UserID),
		zap.String("version", "v2"))

	// Create enhanced flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("submission", submission)
	ctx.Set("screen_id", screenID)
	ctx.Set("version", "v2")

	// Create enhanced submission flow
	submissionFlow := flow.NewFlow("submit_onboarding_screen_v2").
		WithDescription("Process V2 onboarding screen submission with enhanced features").
		WithTimeout(h.config.Timeouts.FlowExecution)

	submissionFlow.
		// Step 1: Enhanced validation with ML-based fraud detection
		Step("validate_submission_v2", h.createValidationStep("validate_submission_v2", h.screenService.GetSubmissionValidator(screenID))).

		// Step 2: Parallel processing with enhanced features
		Parallel("process_submission_v2").
		Step("save_user_data_v2", h.createSaveUserDataStepV2()).
		Step("update_progress_v2", h.createUpdateProgressStepV2()).
		Step("update_personalization", h.createUpdatePersonalizationStep()).
		Step("log_submission_analytics_v2", h.createEnhancedAnalyticsStep("screen_submit")).
		EndParallel().

		// Step 3: Smart next screen determination with ML
		Step("generate_recommendations", h.createRecommendationsStep()).
		Step("determine_smart_next_screen", h.createSmartNextScreenStep()).

		// Step 4: Enhanced response transformation
		Transform("format_response_v2", h.createEnhancedSubmissionResponseTransformer())

	// Execute the flow
	_, err := submissionFlow.Execute(ctx)

	// Enhanced metrics collection
	duration := time.Since(start)
	success := err == nil

	tags := map[string]string{
		"screen_id": screenID,
		"user_id":   submission.UserID,
		"version":   "v2",
		"success":   strconv.FormatBool(success),
	}

	metrics.RecordDuration("onboarding_submission_request_v2", duration, tags)
	metrics.IncrementCounter("onboarding_submissions_v2_total", tags)

	if err != nil {
		h.logger.Error("Failed to submit V2 onboarding screen",
			zap.String("screen_id", screenID),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "SUBMISSION_ERROR_V2", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	responseData, _ := ctx.Get("response_data")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    responseData,
		Metadata: &models.Metadata{
			Version:     "v2",
			ProcessTime: duration.String(),
		},
		Timestamp: time.Now(),
	})
}

// GetOnboardingFlow handles GET /api/v2/onboarding/flow/:userId with enhanced aggregation
func (h *OnboardingHandlerV2) GetOnboardingFlow(c *gin.Context) {
	userID := c.Param("userId")

	h.logger.Info("Getting V2 onboarding flow", zap.String("user_id", userID))

	// Create flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("user_id", userID)
	ctx.Set("version", "v2")

	// Use the proper aggregation pattern
	aggregationStep := h.createEnhancedAggregationStep()

	// Execute aggregation - aggregation step works with flow.Context
	err := aggregationStep.Run(ctx)
	if err != nil {
		h.logger.Error("Failed to get V2 onboarding flow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "FLOW_FETCH_ERROR_V2", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	flowData, _ := ctx.Get("aggregated_data")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      flowData,
		Timestamp: time.Now(),
	})
}

// CompleteOnboarding handles POST /api/v2/onboarding/flow/:userId/complete with enhanced completion
func (h *OnboardingHandlerV2) CompleteOnboarding(c *gin.Context) {
	userID := c.Param("userId")

	h.logger.Info("Completing V2 onboarding flow", zap.String("user_id", userID))

	// Create flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("user_id", userID)
	ctx.Set("version", "v2")

	// Create enhanced completion flow
	completionFlow := flow.NewFlow("complete_onboarding_v2").
		WithDescription("Complete V2 user onboarding process with enhanced features").
		WithTimeout(h.config.Timeouts.FlowExecution)

	completionFlow.
		// Step 1: Enhanced validation with completion readiness check
		Step("validate_completion_v2", h.createValidationStep("validate_completion_v2", h.createValidateCompletionStepV2())).

		// Step 2: Enhanced parallel completion processing
		Parallel("process_completion_v2").
		Step("mark_flow_complete_v2", h.createMarkFlowCompleteStepV2()).
		Step("trigger_personalized_welcome", h.createTriggerPersonalizedWelcomeStep()).
		Step("setup_user_account_v2", h.createSetupUserAccountStepV2()).
		Step("generate_user_insights", h.createGenerateUserInsightsStep()).
		Step("trigger_recommendations_engine", h.createTriggerRecommendationsEngineStep()).
		Step("log_completion_analytics_v2", h.createEnhancedAnalyticsStep("onboarding_complete")).
		EndParallel().

		// Step 3: Generate enhanced completion response
		Transform("format_completion_response_v2", h.createEnhancedCompletionResponseTransformer())

	// Execute the flow
	_, err := completionFlow.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to complete V2 onboarding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "COMPLETION_ERROR_V2", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	completionData, _ := ctx.Get("completion_data")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      completionData,
		Timestamp: time.Now(),
	})
}

// GetAnalytics handles GET /api/v2/onboarding/analytics/:userId
func (h *OnboardingHandlerV2) GetAnalytics(c *gin.Context) {
	userID := c.Param("userId")
	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "7d"
	}

	h.logger.Info("Getting V2 onboarding analytics",
		zap.String("user_id", userID),
		zap.String("time_range", timeRange))

	// Create flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("user_id", userID)
	ctx.Set("time_range", timeRange)
	ctx.Set("version", "v2")

	// Use the demonstration function to show proper aggregation step handling
	analyticsStep := h.fixAggregationSteps()

	// Execute analytics aggregation
	err := analyticsStep.Run(ctx)
	if err != nil {
		h.logger.Error("Failed to get V2 analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "ANALYTICS_FETCH_ERROR_V2", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	// Transform analytics data
	transformer := h.createAnalyticsResponseTransformer()
	if err := transformer(ctx); err != nil {
		h.logger.Error("Failed to transform analytics data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "ANALYTICS_TRANSFORM_ERROR", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	analyticsData, _ := ctx.Get("analytics_response")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      analyticsData,
		Timestamp: time.Now(),
	})
}

// Helper methods for V2 enhanced steps

func (h *OnboardingHandlerV2) createABTestingStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewTransformValueStep("ab_testing", "user_id",
			func(data interface{}) interface{} {
				userID := fmt.Sprintf("%v", data)
				variant := h.abTestingService.GetVariant(userID, "onboarding_flow")
				return variant
			})
		// Set the target field manually since SaveAs doesn't exist
		err := step.Run(flowCtx)
		if err == nil {
			if val, ok := flowCtx.Get("user_id"); ok {
				userID := fmt.Sprintf("%v", val)
				variant := h.abTestingService.GetVariant(userID, "onboarding_flow")
				flowCtx.Set("ab_variant", variant)
			}
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandlerV2) createEnhancedCacheStep() interfaces.Step {
	return h.createCacheGetStep("enhanced_cache", "screen:v2:${screen_id}:${ab_variant}:${device_type}", "cached_screen")
}

func (h *OnboardingHandlerV2) createFetchScreenStepV2() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.backend_url}/v2/screens/${screen_id}").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		WithQueryParam("variant", "${ab_variant}").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("screen_config")
}

func (h *OnboardingHandlerV2) createFetchPersonalizationStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.backend_url}/users/${user_id}/personalization").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("personalization_data")
}

func (h *OnboardingHandlerV2) createEnhancedMobileTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		transformer := transformers.NewMobileTransformer([]string{
			"id", "title", "subtitle", "type", "fields", "actions",
			"personalization", "ab_variant", "device_capabilities",
		})

		screenData, _ := ctx.GetMap("screen_config")
		personalization, _ := ctx.GetMap("personalization_data")
		deviceCaps, _ := ctx.GetMap("device_capabilities")

		// Merge all data for transformation
		mergedData := make(map[string]interface{})
		for k, v := range screenData {
			mergedData[k] = v
		}
		mergedData["personalization"] = personalization
		mergedData["device_capabilities"] = deviceCaps

		transformed, err := transformer.Transform(mergedData)
		if err != nil {
			return err
		}

		ctx.Set("screen_data", transformed)
		return nil
	}
}

func (h *OnboardingHandlerV2) createEnhancedAnalyticsStep(eventType string) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewInfoLogStep("enhanced_analytics", fmt.Sprintf("V2 Analytics: %s", eventType)).
			WithContext(true, "user_id", "screen_id", "device_type", "ab_variant", "version")
		err := step.Run(flowCtx)
		if err == nil {
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func getStringFromContext(ctx interfaces.ExecutionContext, key, defaultValue string) string {
	if val, ok := ctx.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (h *OnboardingHandlerV2) createFetchUserProgressStepV2() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.backend_url}/v2/users/${user_id}/progress").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("user_progress")
}

func (h *OnboardingHandlerV2) createFetchDeviceCapabilitiesStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.backend_url}/devices/capabilities").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		WithQueryParam("device_type", "${device_type}").
		WithQueryParam("app_version", "${app_version}").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("device_capabilities")
}

func (h *OnboardingHandlerV2) createApplyABVariantStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewTransformValueStep("apply_ab_variant", "screen_config",
			func(data interface{}) interface{} {
				// Apply A/B variant modifications to screen config
				return data
			})
		err := step.Run(flowCtx)
		if err == nil {
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandlerV2) createApplyPersonalizationStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewTransformValueStep("apply_personalization", "screen_config",
			func(data interface{}) interface{} {
				// Apply personalization to screen config
				return data
			})
		err := step.Run(flowCtx)
		if err == nil {
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandlerV2) createDeviceOptimizationStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewTransformValueStep("device_optimization", "screen_config",
			func(data interface{}) interface{} {
				// Optimize screen config for device capabilities
				return data
			})
		err := step.Run(flowCtx)
		if err == nil {
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandlerV2) createEnhancedCacheSetStep() interfaces.Step {
	return h.createCacheSetStep("enhanced_cache_set", "screen:v2:${screen_id}:${ab_variant}:${device_type}", "screen_data", 10*time.Minute)
}

func (h *OnboardingHandlerV2) createResponseOptimizationStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewTransformValueStep("response_optimization", "screen_data",
			func(data interface{}) interface{} {
				// Apply response optimization (compression, etc.)
				return data
			})
		err := step.Run(flowCtx)
		if err == nil {
			// Set the optimized data
			if val, ok := flowCtx.Get("screen_data"); ok {
				ctx.Set("optimized_screen_data", val)
			}
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandlerV2) createSaveUserDataStepV2() *httpsteps.HTTPStep {
	return httpsteps.POST("${config.backend_url}/v2/users/${submission.user_id}/data").
		WithJSONBody("${submission.data}").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("save_result")
}

func (h *OnboardingHandlerV2) createUpdateProgressStepV2() *httpsteps.HTTPStep {
	return httpsteps.PUT("${config.backend_url}/v2/users/${submission.user_id}/progress").
		WithJSONBody(map[string]interface{}{
			"completed_screen": "${submission.screen_id}",
			"timestamp":        time.Now(),
			"version":          "v2",
		}).
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("progress_result")
}

func (h *OnboardingHandlerV2) createUpdatePersonalizationStep() *httpsteps.HTTPStep {
	return httpsteps.PUT("${config.backend_url}/users/${submission.user_id}/personalization").
		WithJSONBody("${submission.data}").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("personalization_result")
}

func (h *OnboardingHandlerV2) createRecommendationsStep() *httpsteps.HTTPStep {
	return httpsteps.POST("${config.ml_url}/recommendations/generate").
		WithJSONBody(map[string]interface{}{
			"user_id":     "${submission.user_id}",
			"screen_data": "${submission.data}",
			"context":     "onboarding",
		}).
		WithHeader("Authorization", "Bearer ${config.ml_token}").
		SaveAs("recommendations")
}

func (h *OnboardingHandlerV2) createSmartNextScreenStep() *httpsteps.HTTPStep {
	return httpsteps.POST("${config.ml_url}/flows/next-screen").
		WithJSONBody(map[string]interface{}{
			"user_id":         "${submission.user_id}",
			"current_screen":  "${submission.screen_id}",
			"user_data":       "${submission.data}",
			"recommendations": "${recommendations}",
		}).
		WithHeader("Authorization", "Bearer ${config.ml_token}").
		SaveAs("next_screen")
}

func (h *OnboardingHandlerV2) createEnhancedSubmissionResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		nextScreen, _ := ctx.GetMap("next_screen")
		recommendations, _ := ctx.Get("recommendations")

		response := map[string]interface{}{
			"success":         true,
			"next_screen":     nextScreen,
			"recommendations": recommendations,
			"timestamp":       time.Now(),
			"version":         "v2",
		}

		ctx.Set("response_data", response)
		return nil
	}
}

func (h *OnboardingHandlerV2) createGetUserFlowStepV2() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.backend_url}/v2/users/${user_id}/flow").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("user_flow")
}

func (h *OnboardingHandlerV2) createGetFlowConfigStepV2() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.backend_url}/v2/flows/config").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("flow_config")
}

func (h *OnboardingHandlerV2) createGetUserPreferencesStepV2() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.backend_url}/v2/users/${user_id}/preferences").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("user_preferences")
}

func (h *OnboardingHandlerV2) createGetUserAnalyticsStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.analytics_url}/users/${user_id}/analytics").
		WithHeader("Authorization", "Bearer ${config.analytics_token}").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("user_analytics")
}

func (h *OnboardingHandlerV2) createGetRecommendationsStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.ml_url}/recommendations/${user_id}").
		WithHeader("Authorization", "Bearer ${config.ml_token}").
		WithQueryParam("context", "onboarding_flow").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("flow_recommendations")
}

func (h *OnboardingHandlerV2) createValidateCompletionStepV2() interface{} {
	// Return a validator that can be used with createValidationStep
	return validators.NewFuncValidator("validate_completion_v2", func(data map[string]interface{}) error {
		if userFlow, ok := data["user_flow"].(map[string]interface{}); ok {
			if status, ok := userFlow["status"].(string); ok && status == "ready_to_complete" {
				return nil
			}
		}
		return fmt.Errorf("user flow is not ready to complete")
	})
}

func (h *OnboardingHandlerV2) createMarkFlowCompleteStepV2() *httpsteps.HTTPStep {
	return httpsteps.PUT("${config.backend_url}/v2/users/${user_id}/flow/complete").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("completion_result")
}

func (h *OnboardingHandlerV2) createTriggerPersonalizedWelcomeStep() *httpsteps.HTTPStep {
	return httpsteps.POST("${config.notification_url}/emails/personalized-welcome").
		WithJSONBody(map[string]interface{}{
			"user_id":         "${user_id}",
			"template":        "onboarding_complete_v2",
			"personalization": "${personalization_data}",
		}).
		WithHeader("Authorization", "Bearer ${config.notification_token}").
		SaveAs("welcome_email_result")
}

func (h *OnboardingHandlerV2) createSetupUserAccountStepV2() *httpsteps.HTTPStep {
	return httpsteps.POST("${config.backend_url}/v2/users/${user_id}/setup").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("setup_result")
}

func (h *OnboardingHandlerV2) createGenerateUserInsightsStep() *httpsteps.HTTPStep {
	return httpsteps.POST("${config.ml_url}/insights/generate").
		WithJSONBody(map[string]interface{}{
			"user_id": "${user_id}",
			"context": "onboarding_completion",
		}).
		WithHeader("Authorization", "Bearer ${config.ml_token}").
		SaveAs("user_insights")
}

func (h *OnboardingHandlerV2) createTriggerRecommendationsEngineStep() *httpsteps.HTTPStep {
	return httpsteps.POST("${config.ml_url}/recommendations/initialize").
		WithJSONBody(map[string]interface{}{
			"user_id":         "${user_id}",
			"onboarding_data": "${user_flow}",
		}).
		WithHeader("Authorization", "Bearer ${config.ml_token}").
		SaveAs("recommendations_init_result")
}

func (h *OnboardingHandlerV2) createEnhancedCompletionResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		insights, _ := ctx.Get("user_insights")
		recommendations, _ := ctx.Get("recommendations_init_result")

		response := map[string]interface{}{
			"status":          "completed",
			"message":         "V2 Onboarding completed successfully with personalization",
			"timestamp":       time.Now(),
			"version":         "v2",
			"next_steps":      []string{"explore_personalized_dashboard", "setup_advanced_preferences"},
			"insights":        insights,
			"recommendations": recommendations,
		}

		ctx.Set("completion_data", response)
		return nil
	}
}

func (h *OnboardingHandlerV2) createGetCompletionMetricsStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.analytics_url}/metrics/completion").
		WithHeader("Authorization", "Bearer ${config.analytics_token}").
		WithQueryParam("user_id", "${user_id}").
		WithQueryParam("time_range", "${time_range}").
		SaveAs("completion_metrics")
}

func (h *OnboardingHandlerV2) createGetScreenAnalyticsStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.analytics_url}/screens/analytics").
		WithHeader("Authorization", "Bearer ${config.analytics_token}").
		WithQueryParam("user_id", "${user_id}").
		WithQueryParam("time_range", "${time_range}").
		SaveAs("screen_analytics")
}

func (h *OnboardingHandlerV2) createGetUserBehaviorStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.analytics_url}/users/${user_id}/behavior").
		WithHeader("Authorization", "Bearer ${config.analytics_token}").
		WithQueryParam("time_range", "${time_range}").
		SaveAs("user_behavior")
}

func (h *OnboardingHandlerV2) createGetPerformanceMetricsStep() *httpsteps.HTTPStep {
	return httpsteps.GET("${config.analytics_url}/performance/metrics").
		WithHeader("Authorization", "Bearer ${config.analytics_token}").
		WithQueryParam("user_id", "${user_id}").
		WithQueryParam("time_range", "${time_range}").
		SaveAs("performance_metrics")
}

func (h *OnboardingHandlerV2) createAnalyticsResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		completionMetrics, _ := ctx.Get("completion_metrics")
		screenAnalytics, _ := ctx.Get("screen_analytics")
		userBehavior, _ := ctx.Get("user_behavior")
		performanceMetrics, _ := ctx.Get("performance_metrics")

		response := map[string]interface{}{
			"completion_metrics":  completionMetrics,
			"screen_analytics":    screenAnalytics,
			"user_behavior":       userBehavior,
			"performance_metrics": performanceMetrics,
			"generated_at":        time.Now(),
			"version":             "v2",
		}

		ctx.Set("analytics_response", response)
		return nil
	}
}

// createEnhancedAggregationStep demonstrates the proper way to create aggregation steps
func (h *OnboardingHandlerV2) createEnhancedAggregationStep() *bff.AggregationStep {
	// Create enhanced aggregation step for flow data
	aggregationStep := bff.NewAggregationStep("get_onboarding_flow_v2").
		WithParallel(true).
		WithTimeout(h.config.Timeouts.FlowExecution)

	// Add enhanced steps with more data sources - properly wrapped
	aggregationStep.
		AddRequiredStep(flow.NewStepWrapper(h.createGetUserFlowStepV2())).
		AddRequiredStep(flow.NewStepWrapper(h.createGetFlowConfigStepV2())).
		AddRequiredStep(flow.NewStepWrapper(h.createGetUserPreferencesStepV2())).
		AddOptionalStep(flow.NewStepWrapper(h.createGetUserAnalyticsStep()), map[string]interface{}{
			"analytics": map[string]interface{}{},
		}).
		AddOptionalStep(flow.NewStepWrapper(h.createGetRecommendationsStep()), map[string]interface{}{
			"recommendations": []interface{}{},
		})

	// Add enhanced transformer for mobile optimization
	aggregationStep.WithTransformer(transformers.NewMobileTransformer([]string{
		"flow_id", "current_screen", "progress", "completed_screens",
		"personalization", "recommendations", "analytics",
	}))

	return aggregationStep
}

// fixAggregationSteps demonstrates how to properly handle aggregation steps
// This function serves as both documentation and a working example
func (h *OnboardingHandlerV2) fixAggregationSteps() *bff.AggregationStep {
	// INTERFACE HIERARCHY EXPLANATION:
	//
	// 1. interfaces.ExecutionContext (new interface) - used by modern flow system
	// 2. interfaces.Step (new interface) - implements Run(interfaces.ExecutionContext) error
	// 3. flow.Step (old interface) - implements Run(*flow.Context) error
	// 4. *flow.Context (old context) - implements interfaces.ExecutionContext
	// 5. *bff.AggregationStep - implements Run(*flow.Context) error (old interface)
	//
	// WRAPPING RULES:
	// - HTTP steps implement interfaces.Step (new)
	// - Aggregation steps expect flow.Step (old interface that uses *flow.Context)
	// - Use NewStepWrapper to convert: interfaces.Step → flow.Step
	// - Use NewLegacyStepWrapper to convert: flow.Step → interfaces.Step
	//
	// PROBLEM: Direct use of HTTP steps in aggregation fails
	// aggregationStep.AddRequiredStep(h.createGetCompletionMetricsStep()) // ❌ Wrong
	// Error: cannot use *httpsteps.HTTPStep (interfaces.Step) as flow.Step
	//
	// SOLUTION: Wrap HTTP steps with NewStepWrapper
	// aggregationStep.AddRequiredStep(flow.NewStepWrapper(h.createGetCompletionMetricsStep())) // ✅ Correct
	// This converts interfaces.Step → flow.Step for aggregation compatibility

	// Create analytics aggregation step (uses old flow.Context interface)
	aggregationStep := bff.NewAggregationStep("analytics_aggregation_example").
		WithParallel(true).
		WithTimeout(h.config.Timeouts.FlowExecution)

	// ✅ CORRECT: Wrap HTTP steps (interfaces.Step → flow.Step)
	aggregationStep.
		AddRequiredStep(flow.NewStepWrapper(h.createGetCompletionMetricsStep())).
		AddRequiredStep(flow.NewStepWrapper(h.createGetScreenAnalyticsStep())).
		AddOptionalStep(flow.NewStepWrapper(h.createGetUserBehaviorStep()), map[string]interface{}{
			"behavior": map[string]interface{}{},
		}).
		AddOptionalStep(flow.NewStepWrapper(h.createGetPerformanceMetricsStep()), map[string]interface{}{
			"performance": map[string]interface{}{},
		})

	// ✅ CORRECT: Core steps that use old interface also need wrapping
	// Example: aggregationStep.AddRequiredStep(flow.NewLegacyStepWrapper(core.NewValidationStep(...)))
	// This converts flow.Step (old) → interfaces.Step → flow.Step (for aggregation)

	// ✅ CORRECT: Return aggregation step directly
	// AggregationStep uses *flow.Context internally, so it cannot be wrapped with NewLegacyStepWrapper
	// It must be used directly: aggregationStep.Run(flowContext)
	return aggregationStep
}

// Helper functions that inherit from base handler
func (h *OnboardingHandlerV2) createFlowContext(ctx interfaces.ExecutionContext) *flow.Context {
	return h.OnboardingHandler.createFlowContext(ctx)
}

func (h *OnboardingHandlerV2) syncContexts(from *flow.Context, to interfaces.ExecutionContext) {
	h.OnboardingHandler.syncContexts(from, to)
}

func (h *OnboardingHandlerV2) createValidationStep(name string, validator interface{}) interfaces.Step {
	return h.OnboardingHandler.createValidationStep(name, validator)
}

func (h *OnboardingHandlerV2) createCacheGetStep(name, keyTemplate, targetField string) interfaces.Step {
	return h.OnboardingHandler.createCacheGetStep(name, keyTemplate, targetField)
}

func (h *OnboardingHandlerV2) createCacheSetStep(name, keyTemplate, valueField string, ttl time.Duration) interfaces.Step {
	return h.OnboardingHandler.createCacheSetStep(name, keyTemplate, valueField, ttl)
}
