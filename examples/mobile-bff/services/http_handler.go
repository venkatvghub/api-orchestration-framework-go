package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/config"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/models"
	frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

// MockAPIService handles interactions with mock APIs (WireMock)
type MockAPIService struct {
	config     *frameworkConfig.FrameworkConfig
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
}

// NewMockAPIService creates a new mock API service
func NewMockAPIService(cfg *frameworkConfig.FrameworkConfig, logger *zap.Logger) *MockAPIService {
	mockAPIConfig := config.GetMockAPIConfig()

	return &MockAPIService{
		config:  cfg,
		logger:  logger,
		baseURL: mockAPIConfig.BaseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Increased timeout to prevent cancellation
		},
	}
}

// GetScreenWithContext fetches screen configuration from mock API with context
func (s *MockAPIService) GetScreenWithContext(ctx context.Context, screenID, userID, deviceType string) (*models.OnboardingScreen, error) {
	url := fmt.Sprintf("%s/api/screens/%s", s.baseURL, screenID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("user_id", userID)
	q.Add("device_type", deviceType)
	req.URL.RawQuery = q.Encode()

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Version", "2.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var mockAPIResp models.MockAPIResponse
	if err := json.Unmarshal(body, &mockAPIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert mock API response to OnboardingScreen
	screen := &models.OnboardingScreen{
		ID:          screenID,
		Title:       mockAPIResp.Data["title"].(string),
		Description: mockAPIResp.Data["description"].(string),
		Type:        models.ScreenType(mockAPIResp.Data["type"].(string)),
		Fields:      []models.ScreenField{},
		Actions:     []models.ScreenAction{},
		Metadata: map[string]interface{}{
			"source":    "mock_api",
			"timestamp": time.Now(),
		},
	}

	return screen, nil
}

// GetScreen fetches screen configuration from mock API
func (s *MockAPIService) GetScreen(screenID, userID, deviceType string) (*models.OnboardingScreen, error) {
	return s.GetScreenWithContext(context.Background(), screenID, userID, deviceType)
}

// SubmitScreenWithContext submits screen data to mock API with context
func (s *MockAPIService) SubmitScreenWithContext(ctx context.Context, submission *models.ScreenSubmission) (*models.SubmissionResponse, error) {
	url := fmt.Sprintf("%s/api/screens/%s/submit", s.baseURL, submission.ScreenID)

	reqData := models.MockAPIRequest{
		UserID:    submission.UserID,
		ScreenID:  submission.ScreenID,
		Data:      submission.Data,
		Timestamp: submission.Timestamp,
		DeviceInfo: map[string]interface{}{
			"type":        submission.DeviceInfo.Type,
			"platform":    submission.DeviceInfo.Platform,
			"app_version": submission.DeviceInfo.AppVersion,
			"os_version":  submission.DeviceInfo.OSVersion,
		},
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Version", "2.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var mockAPIResp models.MockAPIResponse
	if err := json.Unmarshal(body, &mockAPIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to SubmissionResponse
	response := &models.SubmissionResponse{
		Success:    mockAPIResp.Success,
		NextScreen: mockAPIResp.Data["next_screen"].(string),
		Message:    mockAPIResp.Data["message"].(string),
		Metadata: map[string]interface{}{
			"source":    "mock_api",
			"timestamp": time.Now(),
		},
	}

	return response, nil
}

// SubmitScreen submits screen data to mock API
func (s *MockAPIService) SubmitScreen(submission *models.ScreenSubmission) (*models.SubmissionResponse, error) {
	return s.SubmitScreenWithContext(context.Background(), submission)
}

// GetUserProgressWithContext fetches user progress from mock API with context
func (s *MockAPIService) GetUserProgressWithContext(ctx context.Context, userID string) (*models.OnboardingProgress, error) {
	url := fmt.Sprintf("%s/api/users/%s/progress", s.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Version", "2.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var mockAPIResp models.MockAPIResponse
	if err := json.Unmarshal(body, &mockAPIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert to OnboardingProgress
	progress := &models.OnboardingProgress{
		UserID:          userID,
		CurrentScreen:   mockAPIResp.Data["current_screen"].(string),
		CompletedSteps:  int(mockAPIResp.Data["completed_steps"].(float64)),
		TotalSteps:      int(mockAPIResp.Data["total_steps"].(float64)),
		PercentComplete: mockAPIResp.Data["percent_complete"].(float64),
		LastUpdated:     time.Now(),
	}

	return progress, nil
}

// GetUserProgress fetches user progress from mock API
func (s *MockAPIService) GetUserProgress(userID string) (*models.OnboardingProgress, error) {
	return s.GetUserProgressWithContext(context.Background(), userID)
}

// UpdateUserProgressWithContext updates user progress in mock API with context
func (s *MockAPIService) UpdateUserProgressWithContext(ctx context.Context, userID, screenID string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/users/%s/progress", s.baseURL, userID)

	reqData := map[string]interface{}{
		"user_id":   userID,
		"screen_id": screenID,
		"data":      data,
		"timestamp": time.Now(),
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Version", "2.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateUserProgress updates user progress in mock API
func (s *MockAPIService) UpdateUserProgress(userID, screenID string, data map[string]interface{}) error {
	return s.UpdateUserProgressWithContext(context.Background(), userID, screenID, data)
}
