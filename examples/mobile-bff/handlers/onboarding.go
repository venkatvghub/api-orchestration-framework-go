package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/config"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/models"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/services"
	frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
	httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
	"github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
)

// OnboardingHandler handles onboarding API requests with mock API integration
type OnboardingHandler struct {
	config           *frameworkConfig.FrameworkConfig
	logger           *zap.Logger
	cacheService     *services.CacheService
	analyticsService *services.AnalyticsService
	mockAPIBaseURL   string
}

// NewOnboardingHandler creates a new onboarding handler
func NewOnboardingHandler(cfg *frameworkConfig.FrameworkConfig, logger *zap.Logger) *OnboardingHandler {
	mockAPIConfig := config.GetMockAPIConfig()

	return &OnboardingHandler{
		config:           cfg,
		logger:           logger,
		cacheService:     services.NewCacheService(cfg, logger),
		analyticsService: services.NewAnalyticsService(cfg, logger),
		mockAPIBaseURL:   mockAPIConfig.BaseURL,
	}
}

// GetScreen handles GET /api/v1/onboarding/screens/:screenId
func (h *OnboardingHandler) GetScreen(c *gin.Context) {
	start := time.Now()
	screenID := c.Param("screenId")
	userID := c.Query("user_id")
	deviceType := c.GetHeader("X-Device-Type")

	h.logger.Info("Getting onboarding screen",
		zap.String("screen_id", screenID),
		zap.String("user_id", userID),
		zap.String("device_type", deviceType))

	// Get flow context from middleware
	ctx, exists := flow.GetFlowContext(c)
	if !exists {
		h.logger.Error("Flow context not found in middleware")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "CONTEXT_ERROR", Message: "Flow context not available"},
			Timestamp: time.Now(),
		})
		return
	}

	// Create screen retrieval flow with direct API fetch
	screenFlow := flow.NewFlow("get_onboarding_screen").
		WithDescription("Retrieve onboarding screen directly from API").
		WithTimeout(60 * time.Second)

	screenFlow.
		// Direct API fetch
		Step("fetch_screen", h.createMockAPIScreenStep()).
		Step("extract_screen_data", h.createExtractScreenDataStep()).

		// Transform for mobile
		Transform("mobile_transform", h.createMobileTransformer()).

		// Track analytics
		Step("track_view", h.createAnalyticsStep("screen_view"))

	// Execute flow
	_, err := screenFlow.Execute(ctx)

	// Record metrics
	duration := time.Since(start)
	success := err == nil

	tags := map[string]string{
		"screen_id":   screenID,
		"device_type": deviceType,
		"success":     strconv.FormatBool(success),
		"source":      "flow_based",
	}

	metrics.RecordDuration("onboarding_screen_request", duration, tags)
	metrics.IncrementCounter("onboarding_screen_requests_total", tags)

	if err != nil {
		h.logger.Error("Failed to get onboarding screen", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "SCREEN_FETCH_ERROR", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	screenData, _ := ctx.Get("mobile_screen")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    screenData,
		Metadata: &models.Metadata{
			Version:     "2.0.0",
			ProcessTime: duration.String(),
			Source:      "flow_based",
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
		OSVersion:  c.GetHeader("X-OS-Version"),
		UserAgent:  c.GetHeader("User-Agent"),
	}

	h.logger.Info("Submitting onboarding screen",
		zap.String("screen_id", screenID),
		zap.String("user_id", submission.UserID))

	// Get flow context from middleware
	ctx, exists := flow.GetFlowContext(c)
	if !exists {
		h.logger.Error("Flow context not found in middleware")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "CONTEXT_ERROR", Message: "Flow context not available"},
			Timestamp: time.Now(),
		})
		return
	}

	// Set submission data in context
	ctx.Set("submission", submission)

	// Explicitly set variables for HTTP step substitution
	ctx.Set("user_id", submission.UserID)
	ctx.Set("screen_id", screenID)

	// Create submission flow with sequential processing
	submissionFlow := flow.NewFlow("submit_onboarding_screen").
		WithDescription("Process onboarding screen submission with sequential operations").
		WithTimeout(120 * time.Second)

	submissionFlow.
		// Step 1: Validate submission
		Step("validate", h.createValidationStep(screenID)).

		// Step 2: Submit to mock API
		Step("submit_to_api", h.createMockAPISubmissionStep()).

		// Post-processing steps
		Step("update_progress", h.createMockAPIUpdateProgressStep()).
		Step("track_submission", h.createAnalyticsStep("screen_submit")).
		Step("determine_next", h.createNextScreenStep()).

		// Format response
		Transform("format_response", h.createSubmissionResponseTransformer())

	// Execute flow
	_, err := submissionFlow.Execute(ctx)

	// Record metrics
	duration := time.Since(start)
	success := err == nil

	tags := map[string]string{
		"screen_id": screenID,
		"user_id":   submission.UserID,
		"success":   strconv.FormatBool(success),
		"source":    "flow_based",
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

	responseData, _ := ctx.Get("submission_response")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    responseData,
		Metadata: &models.Metadata{
			Version:     "2.0.0",
			ProcessTime: duration.String(),
			Source:      "flow_based",
		},
		Timestamp: time.Now(),
	})
}

// GetOnboardingFlow handles GET /api/v1/onboarding/flow/:userId
func (h *OnboardingHandler) GetOnboardingFlow(c *gin.Context) {
	userID := c.Param("userId")

	h.logger.Info("Getting onboarding flow", zap.String("user_id", userID))

	// Get flow context from middleware
	ctx, exists := flow.GetFlowContext(c)
	if !exists {
		h.logger.Error("Flow context not found in middleware")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "CONTEXT_ERROR", Message: "Flow context not available"},
			Timestamp: time.Now(),
		})
		return
	}

	// Create flow retrieval with sequential data fetching
	flowFlow := flow.NewFlow("get_onboarding_flow").
		WithDescription("Get user onboarding flow with sequential data fetching").
		WithTimeout(60 * time.Second)

	flowFlow.
		Step("fetch_flow", h.createMockAPIFlowStep()).
		Step("fetch_progress", h.createMockAPIProgressStep()).
		Transform("format_flow", h.createFlowResponseTransformer())

	// Execute flow
	_, err := flowFlow.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to get onboarding flow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "FLOW_FETCH_ERROR", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	flowData, _ := ctx.Get("flow_response")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      flowData,
		Timestamp: time.Now(),
	})
}

// GetProgress handles GET /api/v1/onboarding/progress/:userId
func (h *OnboardingHandler) GetProgress(c *gin.Context) {
	userID := c.Param("userId")

	h.logger.Info("Getting user progress", zap.String("user_id", userID))

	// Get flow context from middleware
	ctx, exists := flow.GetFlowContext(c)
	if !exists {
		h.logger.Error("Flow context not found in middleware")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "CONTEXT_ERROR", Message: "Flow context not available"},
			Timestamp: time.Now(),
		})
		return
	}

	// Create progress retrieval flow
	progressFlow := flow.NewFlow("get_user_progress").
		WithTimeout(60*time.Second).
		Step("fetch_progress", h.createMockAPIProgressStep()).
		Transform("format_progress", h.createProgressResponseTransformer())

	// Execute flow
	_, err := progressFlow.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to get user progress", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "PROGRESS_FETCH_ERROR", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	progressData, _ := ctx.Get("progress_response")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      progressData,
		Timestamp: time.Now(),
	})
}

// CompleteOnboarding handles POST /api/v1/onboarding/flow/:userId/complete
func (h *OnboardingHandler) CompleteOnboarding(c *gin.Context) {
	userID := c.Param("userId")

	var completionData map[string]interface{}
	if err := c.ShouldBindJSON(&completionData); err != nil {
		completionData = make(map[string]interface{})
	}

	h.logger.Info("Completing onboarding flow", zap.String("user_id", userID))

	// Get flow context from middleware
	ctx, exists := flow.GetFlowContext(c)
	if !exists {
		h.logger.Error("Flow context not found in middleware")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "CONTEXT_ERROR", Message: "Flow context not available"},
			Timestamp: time.Now(),
		})
		return
	}

	// Set completion data in context
	ctx.Set("completion_data", completionData)

	// Create completion flow with sequential steps
	completionFlow := flow.NewFlow("complete_onboarding").
		WithTimeout(120*time.Second).
		Step("complete_flow", h.createMockAPICompletionStep()).
		Step("track_completion", h.createAnalyticsStep("onboarding_complete")).
		Step("cleanup_cache", h.createCacheCleanupStep()).
		Transform("format_completion", h.createCompletionResponseTransformer())

	// Execute flow
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

	responseData, _ := ctx.Get("completion_response")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      responseData,
		Timestamp: time.Now(),
	})
}

// GetAnalytics handles GET /api/v1/onboarding/analytics/:userId
func (h *OnboardingHandler) GetAnalytics(c *gin.Context) {
	userID := c.Param("userId")
	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "7d"
	}

	h.logger.Info("Getting analytics", zap.String("user_id", userID), zap.String("time_range", timeRange))

	// Get flow context from middleware
	ctx, exists := flow.GetFlowContext(c)
	if !exists {
		h.logger.Error("Flow context not found in middleware")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "CONTEXT_ERROR", Message: "Flow context not available"},
			Timestamp: time.Now(),
		})
		return
	}

	// Set time range in context
	ctx.Set("time_range", timeRange)

	// Create analytics flow
	analyticsFlow := flow.NewFlow("get_analytics").
		WithTimeout(60*time.Second).
		Step("fetch_analytics", h.createMockAPIAnalyticsStep()).
		Transform("format_analytics", h.createAnalyticsResponseTransformer())

	// Execute flow
	_, err := analyticsFlow.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to get analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "ANALYTICS_FETCH_ERROR", Message: err.Error()},
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

// TrackEvent handles POST /api/v1/onboarding/analytics/events
func (h *OnboardingHandler) TrackEvent(c *gin.Context) {
	var event models.AnalyticsEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "INVALID_EVENT", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	event.Timestamp = time.Now()

	h.logger.Info("Tracking analytics event",
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID))

	// Get flow context from middleware
	ctx, exists := flow.GetFlowContext(c)
	if !exists {
		h.logger.Error("Flow context not found in middleware")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "CONTEXT_ERROR", Message: "Flow context not available"},
			Timestamp: time.Now(),
		})
		return
	}

	// Set event in context
	ctx.Set("event", event)

	// Create event tracking flow
	eventFlow := flow.NewFlow("track_event").
		WithTimeout(60*time.Second).
		Step("send_event", h.createMockAPIEventStep()).
		Transform("format_event_response", h.createEventResponseTransformer())

	// Execute flow
	_, err := eventFlow.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to track event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     &models.APIError{Code: "EVENT_TRACK_ERROR", Message: err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	responseData, _ := ctx.Get("event_response")

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      responseData,
		Timestamp: time.Now(),
	})
}

// Helper methods for creating flow steps

func (h *OnboardingHandler) createMockAPIScreenStep() *httpsteps.HTTPStep {
	return httpsteps.GET(h.mockAPIBaseURL+"/api/screens/${screen_id}").
		WithQueryParam("user_id", "${user_id}").
		WithQueryParam("device_type", "${device_type}").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("mock_api_screen")
}

func (h *OnboardingHandler) createMockAPIProgressStep() *httpsteps.HTTPStep {
	return httpsteps.GET(h.mockAPIBaseURL+"/api/users/${user_id}/progress").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("user_progress")
}

func (h *OnboardingHandler) createCacheStoreStep(screenID, deviceType string) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		screenData, _ := ctx.Get("screen_data")
		if screenData != nil {
			cacheKey := h.cacheService.GetScreenCacheKey(screenID, deviceType)
			h.cacheService.Set(cacheKey, screenData)
		}
		return nil
	})
}

func (h *OnboardingHandler) createCacheCleanupStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		// Clear the entire cache as a simple cleanup approach
		h.cacheService.Clear()
		h.logger.Info("Cache cleared during onboarding completion")
		return nil
	})
}

func (h *OnboardingHandler) createMobileTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		transformer := transformers.NewMobileTransformer([]string{
			"id", "title", "subtitle", "type", "fields", "actions", "next_screen",
			"verification_sent", "email", "completion_message", "recommendations",
		})

		screenData, _ := ctx.GetMap("screen_data")
		if screenData == nil {
			screenData, _ = ctx.GetMap("cached_screen")
		}

		transformed, err := transformer.Transform(screenData)
		if err != nil {
			return err
		}

		ctx.Set("mobile_screen", transformed)
		return nil
	}
}

func (h *OnboardingHandler) createAnalyticsStep(eventType string) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		userID, _ := ctx.GetString("user_id")
		screenID, _ := ctx.GetString("screen_id")
		deviceType, _ := ctx.GetString("device_type")

		event := models.AnalyticsEvent{
			EventType: eventType,
			UserID:    userID,
			ScreenID:  screenID,
			Properties: map[string]interface{}{
				"device_type": deviceType,
				"timestamp":   time.Now(),
			},
			Timestamp: time.Now(),
		}

		return h.analyticsService.TrackEvent(event)
	})
}

func (h *OnboardingHandler) createValidationStep(screenID string) interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		submissionData, exists := ctx.Get("submission")
		if !exists {
			return fmt.Errorf("submission data not found")
		}

		submission, ok := submissionData.(models.ScreenSubmission)
		if !ok {
			return fmt.Errorf("invalid submission data type")
		}

		if submission.UserID == "" {
			return fmt.Errorf("user_id is required")
		}

		if submission.Data == nil {
			return fmt.Errorf("submission data is required")
		}

		// Screen-specific validation
		switch screenID {
		case "personal_info":
			if _, ok := submission.Data["email"]; !ok {
				return fmt.Errorf("email is required for personal_info screen")
			}
		case "verification":
			if _, ok := submission.Data["verification_code"]; !ok {
				return fmt.Errorf("verification_code is required for verification screen")
			}
		}

		return nil
	})
}

func (h *OnboardingHandler) createMockAPISubmissionStep() *httpsteps.HTTPStep {
	return httpsteps.POST(h.mockAPIBaseURL+"/api/screens/${screen_id}/submit").
		WithJSONBody("${submission}").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("submission_result")
}

func (h *OnboardingHandler) createMockAPIUpdateProgressStep() *httpsteps.HTTPStep {
	return httpsteps.POST(h.mockAPIBaseURL+"/api/users/${user_id}/progress").
		WithJSONBody(map[string]interface{}{
			"user_id":   "${user_id}",
			"screen_id": "${screen_id}",
			"data": map[string]interface{}{
				"completed": true,
			},
			"timestamp": time.Now(),
		}).
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("progress_update")
}

func (h *OnboardingHandler) createNextScreenStep() *httpsteps.HTTPStep {
	return httpsteps.GET(h.mockAPIBaseURL+"/api/screens/${screen_id}/next").
		WithQueryParam("user_id", "${user_id}").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("next_screen")
}

func (h *OnboardingHandler) createSubmissionResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		submissionResult, _ := ctx.GetMap("submission_result")
		nextScreen, _ := ctx.GetMap("next_screen")

		response := map[string]interface{}{
			"success":   true,
			"timestamp": time.Now(),
		}

		if nextScreen != nil {
			response["next_screen"] = nextScreen
		}

		if submissionResult != nil {
			if progress, ok := submissionResult["progress"]; ok {
				response["progress"] = progress
			}
			if message, ok := submissionResult["message"]; ok {
				response["message"] = message
			}
		}

		ctx.Set("submission_response", response)
		return nil
	}
}

func (h *OnboardingHandler) createMockAPIFlowStep() *httpsteps.HTTPStep {
	return httpsteps.GET(h.mockAPIBaseURL+"/api/users/${user_id}/flow").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("flow_data")
}

func (h *OnboardingHandler) createFlowResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		flowData, _ := ctx.GetMap("flow_data")
		progressData, _ := ctx.GetMap("user_progress")

		response := make(map[string]interface{})
		for k, v := range flowData {
			response[k] = v
		}

		if progressData != nil {
			response["progress"] = progressData
		}

		ctx.Set("flow_response", response)
		return nil
	}
}

func (h *OnboardingHandler) createProgressResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		progressData, _ := ctx.GetMap("user_progress")
		ctx.Set("progress_response", progressData)
		return nil
	}
}

func (h *OnboardingHandler) createMockAPICompletionStep() *httpsteps.HTTPStep {
	return httpsteps.POST(h.mockAPIBaseURL+"/api/users/${user_id}/complete").
		WithJSONBody("${completion_data}").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("completion_result")
}

func (h *OnboardingHandler) createCompletionResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		completionResult, _ := ctx.GetMap("completion_result")

		response := map[string]interface{}{
			"status":     "completed",
			"message":    "Onboarding completed successfully!",
			"timestamp":  time.Now(),
			"next_steps": []string{"explore_dashboard", "setup_profile", "connect_friends"},
		}

		for k, v := range completionResult {
			response[k] = v
		}

		ctx.Set("completion_response", response)
		return nil
	}
}

func (h *OnboardingHandler) createMockAPIAnalyticsStep() *httpsteps.HTTPStep {
	return httpsteps.GET(h.mockAPIBaseURL+"/api/analytics/events").
		WithQueryParam("user_id", "${user_id}").
		WithQueryParam("time_range", "${time_range}").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("analytics_data")
}

func (h *OnboardingHandler) createAnalyticsResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		analyticsData, _ := ctx.GetMap("analytics_data")
		ctx.Set("analytics_response", analyticsData)
		return nil
	}
}

func (h *OnboardingHandler) createMockAPIEventStep() *httpsteps.HTTPStep {
	return httpsteps.POST(h.mockAPIBaseURL+"/api/analytics/events").
		WithJSONBody("${event}").
		WithHeader("Content-Type", "application/json").
		WithHeader("X-API-Version", "2.0").
		WithTimeout(60 * time.Second).
		SaveAs("event_result")
}

func (h *OnboardingHandler) createEventResponseTransformer() func(interfaces.ExecutionContext) error {
	return func(ctx interfaces.ExecutionContext) error {
		eventResult, _ := ctx.GetMap("event_result")

		response := map[string]interface{}{
			"tracked":   true,
			"timestamp": time.Now(),
		}

		for k, v := range eventResult {
			response[k] = v
		}

		ctx.Set("event_response", response)
		return nil
	}
}

func (h *OnboardingHandler) createExtractScreenDataStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		screenResponse, _ := ctx.GetMap("mock_api_screen")
		screenData := make(map[string]interface{})

		if screenResponse != nil {
			// HTTP response structure: {body: {data: {...}}, headers: {...}, status_code: 200}
			if body, ok := screenResponse["body"].(map[string]interface{}); ok {
				if data, ok := body["data"].(map[string]interface{}); ok {
					screenData = data
				} else if body["success"] == true {
					// If no data field, use the whole body
					screenData = body
				}
			} else {
				// Fallback to direct data extraction
				if data, ok := screenResponse["data"].(map[string]interface{}); ok {
					screenData = data
				} else {
					screenData = screenResponse
				}
			}
		}

		ctx.Set("screen_data", screenData)
		return nil
	})
}

func (h *OnboardingHandler) createDebugRawDataStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		screenResponse, _ := ctx.GetMap("mock_api_screen")
		ctx.Set("debug_raw_data", screenResponse)
		return nil
	})
}

func (h *OnboardingHandler) createDebugContextStep() interfaces.Step {
	return flow.StepFunc(func(ctx interfaces.ExecutionContext) error {
		// This step is intentionally empty as the debug context values are already captured in the flow
		return nil
	})
}
