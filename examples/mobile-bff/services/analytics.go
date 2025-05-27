package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/config"
	"github.com/venkatvghub/api-orchestration-framework/examples/mobile-bff/models"
	frameworkConfig "github.com/venkatvghub/api-orchestration-framework/pkg/config"
)

// AnalyticsService handles analytics tracking and reporting
type AnalyticsService struct {
	config     *frameworkConfig.FrameworkConfig
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
	events     []models.AnalyticsEvent
	mutex      sync.RWMutex
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(cfg *frameworkConfig.FrameworkConfig, logger *zap.Logger) *AnalyticsService {
	mockAPIConfig := config.GetMockAPIConfig()

	service := &AnalyticsService{
		config:  cfg,
		logger:  logger,
		baseURL: mockAPIConfig.BaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		events: make([]models.AnalyticsEvent, 0),
	}

	// Start batch processing goroutine
	go service.processBatch()

	return service
}

// TrackEvent tracks an analytics event
func (s *AnalyticsService) TrackEvent(event models.AnalyticsEvent) error {
	// Add timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add to local buffer
	s.mutex.Lock()
	s.events = append(s.events, event)
	s.mutex.Unlock()

	s.logger.Debug("Analytics event tracked",
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID),
		zap.String("screen_id", event.ScreenID))

	// Send immediately for critical events
	if s.isCriticalEvent(event.EventType) {
		return s.sendEventToMockAPI(event)
	}

	return nil
}

// TrackScreenView tracks a screen view event
func (s *AnalyticsService) TrackScreenView(userID, screenID, deviceType string, metadata map[string]interface{}) error {
	event := models.AnalyticsEvent{
		EventType: "screen_view",
		UserID:    userID,
		ScreenID:  screenID,
		Properties: map[string]interface{}{
			"device_type": deviceType,
			"timestamp":   time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Merge additional metadata
	for k, v := range metadata {
		event.Properties[k] = v
	}

	return s.TrackEvent(event)
}

// TrackScreenSubmit tracks a screen submission event
func (s *AnalyticsService) TrackScreenSubmit(userID, screenID string, success bool, metadata map[string]interface{}) error {
	event := models.AnalyticsEvent{
		EventType: "screen_submit",
		UserID:    userID,
		ScreenID:  screenID,
		Properties: map[string]interface{}{
			"success":   success,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Merge additional metadata
	for k, v := range metadata {
		event.Properties[k] = v
	}

	return s.TrackEvent(event)
}

// TrackOnboardingComplete tracks onboarding completion
func (s *AnalyticsService) TrackOnboardingComplete(userID string, duration time.Duration, metadata map[string]interface{}) error {
	event := models.AnalyticsEvent{
		EventType: "onboarding_complete",
		UserID:    userID,
		Properties: map[string]interface{}{
			"duration_seconds": duration.Seconds(),
			"timestamp":        time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Merge additional metadata
	for k, v := range metadata {
		event.Properties[k] = v
	}

	return s.TrackEvent(event)
}

// GetUserAnalytics retrieves analytics data for a user from mock API
func (s *AnalyticsService) GetUserAnalytics(userID string) (*models.UserAnalytics, error) {
	url := fmt.Sprintf("%s/api/analytics/users/%s", s.baseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
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

	// Convert to UserAnalytics
	analytics := &models.UserAnalytics{
		UserID:       userID,
		TotalEvents:  int(mockAPIResp.Data["total_events"].(float64)),
		ScreenViews:  int(mockAPIResp.Data["screen_views"].(float64)),
		Submissions:  int(mockAPIResp.Data["submissions"].(float64)),
		TimeSpent:    time.Duration(mockAPIResp.Data["time_spent_seconds"].(float64)) * time.Second,
		LastActivity: time.Now(), // Would be parsed from response in real implementation
		DeviceInfo:   map[string]interface{}{},
		CustomEvents: []models.AnalyticsEvent{},
	}

	return analytics, nil
}

// GetEventStats returns statistics about tracked events
func (s *AnalyticsService) GetEventStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	eventTypes := make(map[string]int)
	totalEvents := len(s.events)

	for _, event := range s.events {
		eventTypes[event.EventType]++
	}

	return map[string]interface{}{
		"total_events": totalEvents,
		"event_types":  eventTypes,
		"buffer_size":  totalEvents,
		"last_updated": time.Now(),
	}
}

// FlushEvents sends all buffered events to mock API
func (s *AnalyticsService) FlushEvents() error {
	s.mutex.Lock()
	events := make([]models.AnalyticsEvent, len(s.events))
	copy(events, s.events)
	s.events = s.events[:0] // Clear the buffer
	s.mutex.Unlock()

	if len(events) == 0 {
		return nil
	}

	return s.sendEventsBatch(events)
}

// sendEventToMockAPI sends a single event to mock API
func (s *AnalyticsService) sendEventToMockAPI(event models.AnalyticsEvent) error {
	url := fmt.Sprintf("%s/api/analytics/events", s.baseURL)

	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Debug("Analytics event sent to mock API",
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID))

	return nil
}

// sendEventsBatch sends multiple events to mock API in a batch
func (s *AnalyticsService) sendEventsBatch(events []models.AnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/api/analytics/events/batch", s.baseURL)

	batchData := map[string]interface{}{
		"events":    events,
		"timestamp": time.Now(),
		"count":     len(events),
	}

	jsonData, err := json.Marshal(batchData)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Info("Analytics batch sent to mock API",
		zap.Int("event_count", len(events)))

	return nil
}

// processBatch processes events in batches periodically
func (s *AnalyticsService) processBatch() {
	ticker := time.NewTicker(30 * time.Second) // Batch every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		if err := s.FlushEvents(); err != nil {
			s.logger.Error("Failed to flush analytics events", zap.Error(err))
		}
	}
}

// isCriticalEvent determines if an event should be sent immediately
func (s *AnalyticsService) isCriticalEvent(eventType string) bool {
	criticalEvents := map[string]bool{
		"onboarding_complete": true,
		"error":               true,
		"crash":               true,
	}

	return criticalEvents[eventType]
}
