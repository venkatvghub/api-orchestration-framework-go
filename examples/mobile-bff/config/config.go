package config

import "os"

// MockAPIConfig holds mock API configuration
type MockAPIConfig struct {
	BaseURL string
}

// GetMockAPIConfig returns the mock API configuration
func GetMockAPIConfig() *MockAPIConfig {
	baseURL := os.Getenv("MOCK_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}

	return &MockAPIConfig{
		BaseURL: baseURL,
	}
}

// Default configuration constants
const (
	DefaultMockAPIBaseURL = "http://localhost:8082"
)
