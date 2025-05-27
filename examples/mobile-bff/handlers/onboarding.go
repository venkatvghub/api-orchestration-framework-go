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

// OnboardingHandler handles onboarding API requests
type OnboardingHandler struct {
	config           *config.FrameworkConfig
	logger           *zap.Logger
	screenService    *services.ScreenService
	analyticsService *services.AnalyticsService
	cacheService     *services.CacheService
}

// NewOnboardingHandler creates a new onboarding handler
func NewOnboardingHandler(cfg *config.FrameworkConfig, logger *zap.Logger) *OnboardingHandler {
	return &OnboardingHandler{
		config:           cfg,
		logger:           logger,
		screenService:    services.NewScreenService(cfg, logger),
		analyticsService: services.NewAnalyticsService(cfg, logger),
		cacheService:     services.NewCacheService(cfg, logger),
	}
}

// GetScreen handles GET /api/v1/onboarding/screens/:screenId
func (h *OnboardingHandler) GetScreen(c *gin.Context) {
	start := time.Now()
	screenID := c.Param("screenId")
	userID := c.Query("user_id")
	version := c.GetHeader("X-API-Version")
	deviceType := c.GetHeader("X-Device-Type")

	if version == "" {
		version = "v1"
	}

	h.logger.Info("Getting onboarding screen",
		zap.String("screen_id", screenID),
		zap.String("user_id", userID),
		zap.String("version", version),
		zap.String("device_type", deviceType))

	// Create flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("screen_id", screenID)
	ctx.Set("user_id", userID)
	ctx.Set("version", version)
	ctx.Set("device_type", deviceType)

	// Create flow for screen retrieval
	screenFlow := flow.NewFlow("get_onboarding_screen").
		WithDescription("Retrieve onboarding screen with version support").
		WithTimeout(h.config.Timeouts.FlowExecution)

	// Add steps to the flow
	screenFlow.
		// Step 1: Validate request parameters
		Step("validate_request", h.createValidationStep("validate_request", h.screenService.GetRequestValidator())).

		// Step 2: Check cache for screen data
		Step("check_cache", h.createCacheGetStep("check_cache", fmt.Sprintf("screen:%s:%s", screenID, version), "cached_screen")).

		// Step 3: Fetch screen configuration if not cached
		Choice("fetch_screen_choice").
		When(func(ctx interfaces.ExecutionContext) bool {
			return !ctx.Has("cached_screen")
		}).
		Step("fetch_screen_config", h.createFetchScreenStep()).
		Step("fetch_user_progress", h.createFetchUserProgressStep()).
		Step("apply_version_logic", h.createVersionLogicStep()).
		Step("cache_screen", h.createCacheSetStep("cache_screen", fmt.Sprintf("screen:%s:%s", screenID, version), "screen_data", 5*time.Minute)).
		EndChoice().

		// Step 4: Transform for mobile response
		Transform("mobile_transform", h.createMobileTransformer()).

		// Step 5: Log analytics
		Step("log_analytics", h.createAnalyticsStep("screen_view"))

	// Execute the flow
	_, err := screenFlow.Execute(ctx)

	// Record metrics
	duration := time.Since(start)
	success := err == nil

	tags := map[string]string{
		"screen_id":   screenID,
		"version":     version,
		"device_type": deviceType,
		"success":     strconv.FormatBool(success),
	}

	metrics.RecordDuration("onboarding_screen_request", duration, tags)
	metrics.IncrementCounter("onboarding_screen_requests_total", tags)

	if err != nil {
		h.logger.Error("Failed to get onboarding screen",
			zap.String("screen_id", screenID),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "SCREEN_FETCH_ERROR", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	// Get screen data from context
	screenData, _ := ctx.Get("screen_data")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    screenData,
		Metadata: &models.Metadata{
			Version:     version,
			ProcessTime: duration.String(),
		},
		Timestamp: time.Now(),
	})
}

// SubmitScreen handles POST /api/v1/onboarding/screens/:screenId/submit
func (h *OnboardingHandler) SubmitScreen(c *gin.Context) {
	start := time.Now()
	screenID := c.Param("screenId")

	var submission models.ScreenSubmission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "INVALID_REQUEST", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	submission.ScreenID = screenID
	submission.Timestamp = time.Now()
	submission.DeviceInfo = models.DeviceInfo{
		Type:       c.GetHeader("X-Device-Type"),
		Platform:   c.GetHeader("X-Platform"),
		AppVersion: c.GetHeader("X-App-Version"),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	h.logger.Info("Submitting onboarding screen",
		zap.String("screen_id", screenID),
		zap.String("user_id", submission.UserID))

	// Create flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("submission", submission)
	ctx.Set("screen_id", screenID)

	// Create submission flow
	submissionFlow := flow.NewFlow("submit_onboarding_screen").
		WithDescription("Process onboarding screen submission").
		WithTimeout(h.config.Timeouts.FlowExecution)

	submissionFlow.
		// Step 1: Validate submission data
		Step("validate_submission", h.createValidationStep("validate_submission", h.screenService.GetSubmissionValidator(screenID))).

		// Step 2: Process submission in parallel
		Parallel("process_submission").
		Step("save_user_data", h.createSaveUserDataStep()).
		Step("update_progress", h.createUpdateProgressStep()).
		Step("log_submission_analytics", h.createAnalyticsStep("screen_submit")).
		EndParallel().

		// Step 3: Determine next screen
		Step("determine_next_screen", h.createNextScreenStep()).

		// Step 4: Transform response
		Transform("format_response", h.createSubmissionResponseTransformer())

	// Execute the flow
	_, err := submissionFlow.Execute(ctx)

	// Record metrics
	duration := time.Since(start)
	success := err == nil

	tags := map[string]string{
		"screen_id": screenID,
		"user_id":   submission.UserID,
		"success":   strconv.FormatBool(success),
	}

	metrics.RecordDuration("onboarding_submission_request", duration, tags)
	metrics.IncrementCounter("onboarding_submissions_total", tags)

	if err != nil {
		h.logger.Error("Failed to submit onboarding screen",
			zap.String("screen_id", screenID),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "SUBMISSION_ERROR", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	responseData, _ := ctx.Get("response_data")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    responseData,
		Metadata: &models.Metadata{
			Version:     "v1",
			ProcessTime: duration.String(),
		},
		Timestamp: time.Now(),
	})
}

// GetOnboardingFlow handles GET /api/v1/onboarding/flow/:userId
func (h *OnboardingHandler) GetOnboardingFlow(c *gin.Context) {
	userID := c.Param("userId")

	h.logger.Info("Getting onboarding flow", zap.String("user_id", userID))

	// Create aggregation step for flow data
	aggregationStep := bff.NewAggregationStep("get_onboarding_flow").
		WithParallel(true).
		WithTimeout(h.config.Timeouts.FlowExecution)

	// Add required steps
	aggregationStep.
		AddRequiredStep(flow.NewStepWrapper(h.createGetUserFlowStep())).
		AddRequiredStep(flow.NewStepWrapper(h.createGetFlowConfigStep())).
		AddOptionalStep(flow.NewStepWrapper(h.createGetUserPreferencesStep()), map[string]interface{}{
			"preferences": map[string]interface{}{},
		})

	// Add transformer for mobile optimization
	aggregationStep.WithTransformer(transformers.NewMobileTransformer([]string{
		"flow_id", "current_screen", "progress", "completed_screens",
	}))

	// Create flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("user_id", userID)

	// Execute aggregation
	err := aggregationStep.Run(ctx)
	if err != nil {
		h.logger.Error("Failed to get onboarding flow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "FLOW_FETCH_ERROR", Message: err.Error()},
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

// CompleteOnboarding handles POST /api/v1/onboarding/flow/:userId/complete
func (h *OnboardingHandler) CompleteOnboarding(c *gin.Context) {
	userID := c.Param("userId")

	h.logger.Info("Completing onboarding flow", zap.String("user_id", userID))

	// Create flow context
	ctx := flow.NewContextWithConfig(h.config)
	ctx.WithLogger(h.logger)
	ctx.Set("user_id", userID)

	// Create completion flow
	completionFlow := flow.NewFlow("complete_onboarding").
		WithDescription("Complete user onboarding process").
		WithTimeout(h.config.Timeouts.FlowExecution)

	completionFlow.
		// Step 1: Validate completion eligibility
		Step("validate_completion", h.createValidateCompletionStep()).

		// Step 2: Process completion in parallel
		Parallel("process_completion").
		Step("mark_flow_complete", h.createMarkFlowCompleteStep()).
		Step("trigger_welcome_email", h.createTriggerWelcomeEmailStep()).
		Step("setup_user_account", h.createSetupUserAccountStep()).
		Step("log_completion_analytics", h.createAnalyticsStep("onboarding_complete")).
		EndParallel().

		// Step 3: Generate completion response
		Transform("format_completion_response", h.createCompletionResponseTransformer())

	// Execute the flow
	_, err := completionFlow.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to complete onboarding", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "COMPLETION_ERROR", Message: err.Error()},
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

// Helper methods to create flow steps

func (h *OnboardingHandler) createFetchScreenStep() interfaces.Step {
	return httpsteps.GET("${config.backend_url}/screens/${screen_id}").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("screen_config")
}

func (h *OnboardingHandler) createFetchUserProgressStep() interfaces.Step {
	return httpsteps.GET("${config.backend_url}/users/${user_id}/progress").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		WithTimeout(h.config.Timeouts.HTTPRequest).
		SaveAs("user_progress")
}

func (h *OnboardingHandler) createVersionLogicStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewTransformValueStep("apply_version_logic", "screen_config",
			func(data interface{}) interface{} {
				// Apply version-specific logic here
				return data
			})
		err := step.Run(flowCtx)
		if err == nil {
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandler) createMobileTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		transformer := transformers.NewMobileTransformer([]string{
			"id", "title", "subtitle", "type", "fields", "actions",
		})

		screenData, _ := ctx.GetMap("screen_config")
		transformed, err := transformer.Transform(screenData)
		if err != nil {
			return err
		}

		ctx.Set("screen_data", transformed)
		return nil
	}
}

func (h *OnboardingHandler) createAnalyticsStep(eventType string) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewInfoLogStep("log_analytics", fmt.Sprintf("Analytics: %s", eventType)).
			WithContext(true, "user_id", "screen_id", "device_type")
		err := step.Run(flowCtx)
		if err == nil {
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandler) createSaveUserDataStep() interfaces.Step {
	return httpsteps.POST("${config.backend_url}/users/${submission.user_id}/data").
		WithJSONBody("${submission.data}").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("save_result")
}

func (h *OnboardingHandler) createUpdateProgressStep() interfaces.Step {
	return httpsteps.PUT("${config.backend_url}/users/${submission.user_id}/progress").
		WithJSONBody(map[string]interface{}{
			"completed_screen": "${submission.screen_id}",
			"timestamp":        time.Now(),
		}).
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("progress_result")
}

func (h *OnboardingHandler) createNextScreenStep() interfaces.Step {
	return httpsteps.GET("${config.backend_url}/flows/next-screen").
		WithQueryParam("user_id", "${submission.user_id}").
		WithQueryParam("current_screen", "${submission.screen_id}").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("next_screen")
}

func (h *OnboardingHandler) createSubmissionResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		nextScreen, _ := ctx.GetMap("next_screen")

		response := map[string]interface{}{
			"success":     true,
			"next_screen": nextScreen,
			"timestamp":   time.Now(),
		}

		ctx.Set("response_data", response)
		return nil
	}
}

func (h *OnboardingHandler) createGetUserFlowStep() interfaces.Step {
	return httpsteps.GET("${config.backend_url}/users/${user_id}/flow").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("user_flow")
}

func (h *OnboardingHandler) createGetFlowConfigStep() interfaces.Step {
	return httpsteps.GET("${config.backend_url}/flows/config").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("flow_config")
}

func (h *OnboardingHandler) createGetUserPreferencesStep() interfaces.Step {
	return httpsteps.GET("${config.backend_url}/users/${user_id}/preferences").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("user_preferences")
}

func (h *OnboardingHandler) createValidateCompletionStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewConditionStep("validate_completion", "user_flow.status", "equals", "ready_to_complete")
		err := step.Run(flowCtx)
		if err == nil {
			h.syncContexts(flowCtx, ctx)
		}
		return err
	})
}

func (h *OnboardingHandler) createMarkFlowCompleteStep() interfaces.Step {
	return httpsteps.PUT("${config.backend_url}/users/${user_id}/flow/complete").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("completion_result")
}

func (h *OnboardingHandler) createTriggerWelcomeEmailStep() interfaces.Step {
	return httpsteps.POST("${config.notification_url}/emails/welcome").
		WithJSONBody(map[string]interface{}{
			"user_id":  "${user_id}",
			"template": "onboarding_complete",
		}).
		WithHeader("Authorization", "Bearer ${config.notification_token}").
		SaveAs("email_result")
}

func (h *OnboardingHandler) createSetupUserAccountStep() interfaces.Step {
	return httpsteps.POST("${config.backend_url}/users/${user_id}/setup").
		WithHeader("Authorization", "Bearer ${config.api_token}").
		SaveAs("setup_result")
}

func (h *OnboardingHandler) createCompletionResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		response := map[string]interface{}{
			"status":     "completed",
			"message":    "Onboarding completed successfully",
			"timestamp":  time.Now(),
			"next_steps": []string{"explore_app", "setup_preferences"},
		}

		ctx.Set("completion_data", response)
		return nil
	}
}

// Helper wrapper functions for core steps

func (h *OnboardingHandler) createValidationStep(name string, validator interface{}) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		// Create a flow context wrapper
		flowCtx := h.createFlowContext(ctx)
		step := core.NewValidationStep(name, validator.(validators.Validator))
		return step.Run(flowCtx)
	})
}

func (h *OnboardingHandler) createCacheGetStep(name, keyTemplate, targetField string) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewCacheGetStep(name, keyTemplate, targetField)
		return step.Run(flowCtx)
	})
}

func (h *OnboardingHandler) createCacheSetStep(name, keyTemplate, valueField string, ttl time.Duration) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		flowCtx := h.createFlowContext(ctx)
		step := core.NewCacheSetStep(name, keyTemplate, valueField, ttl)
		return step.Run(flowCtx)
	})
}

func (h *OnboardingHandler) createFlowContext(ctx interfaces.ExecutionContext) *flow.Context {
	// Create a flow context and copy data from execution context
	flowCtx := flow.NewContextWithConfig(h.config)
	flowCtx.WithLogger(ctx.Logger())

	// Copy all data from execution context
	for _, key := range ctx.Keys() {
		if val, ok := ctx.Get(key); ok {
			flowCtx.Set(key, val)
		}
	}

	return flowCtx
}

func (h *OnboardingHandler) syncContexts(from *flow.Context, to interfaces.ExecutionContext) {
	// Copy data back from flow context to execution context
	for _, key := range from.Keys() {
		if val, ok := from.Get(key); ok {
			to.Set(key, val)
		}
	}
}
