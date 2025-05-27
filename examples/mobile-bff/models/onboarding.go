package models

import (
	"time"
)

// OnboardingScreen represents a single onboarding screen configuration
type OnboardingScreen struct {
	ID         string                 `json:"id"`
	Version    string                 `json:"version"`
	Title      string                 `json:"title"`
	Subtitle   string                 `json:"subtitle,omitempty"`
	Type       ScreenType             `json:"type"`
	Fields     []ScreenField          `json:"fields"`
	Actions    []ScreenAction         `json:"actions"`
	Validation ValidationRules        `json:"validation"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	NextScreen string                 `json:"next_screen,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// ScreenType defines the type of onboarding screen
type ScreenType string

const (
	ScreenTypeWelcome      ScreenType = "welcome"
	ScreenTypePersonalInfo ScreenType = "personal_info"
	ScreenTypePreferences  ScreenType = "preferences"
	ScreenTypePermissions  ScreenType = "permissions"
	ScreenTypeVerification ScreenType = "verification"
	ScreenTypeCompletion   ScreenType = "completion"
)

// ScreenField represents a form field in the screen
type ScreenField struct {
	ID           string                 `json:"id"`
	Type         FieldType              `json:"type"`
	Label        string                 `json:"label"`
	Placeholder  string                 `json:"placeholder,omitempty"`
	Required     bool                   `json:"required"`
	Options      []FieldOption          `json:"options,omitempty"`
	Validation   FieldValidation        `json:"validation,omitempty"`
	DefaultValue interface{}            `json:"default_value,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// FieldType defines the type of form field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeEmail    FieldType = "email"
	FieldTypePassword FieldType = "password"
	FieldTypePhone    FieldType = "phone"
	FieldTypeSelect   FieldType = "select"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeRadio    FieldType = "radio"
	FieldTypeDate     FieldType = "date"
	FieldTypeNumber   FieldType = "number"
	FieldTypeFile     FieldType = "file"
)

// FieldOption represents an option for select/radio fields
type FieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FieldValidation defines validation rules for a field
type FieldValidation struct {
	MinLength int      `json:"min_length,omitempty"`
	MaxLength int      `json:"max_length,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Custom    string   `json:"custom,omitempty"`
}

// ScreenAction represents an action button on the screen
type ScreenAction struct {
	ID       string     `json:"id"`
	Type     ActionType `json:"type"`
	Label    string     `json:"label"`
	Style    string     `json:"style,omitempty"`
	Endpoint string     `json:"endpoint,omitempty"`
	Method   string     `json:"method,omitempty"`
}

// ActionType defines the type of screen action
type ActionType string

const (
	ActionTypeNext     ActionType = "next"
	ActionTypeSkip     ActionType = "skip"
	ActionTypeSubmit   ActionType = "submit"
	ActionTypeBack     ActionType = "back"
	ActionTypeExternal ActionType = "external"
)

// ValidationRules defines validation rules for the entire screen
type ValidationRules struct {
	RequiredFields []string               `json:"required_fields,omitempty"`
	CustomRules    map[string]interface{} `json:"custom_rules,omitempty"`
}

// OnboardingFlow represents the complete onboarding flow for a user
type OnboardingFlow struct {
	UserID           string                 `json:"user_id"`
	FlowID           string                 `json:"flow_id"`
	Version          string                 `json:"version"`
	CurrentScreen    string                 `json:"current_screen"`
	CompletedScreens []string               `json:"completed_screens"`
	UserData         map[string]interface{} `json:"user_data"`
	Progress         float64                `json:"progress"`
	Status           FlowStatus             `json:"status"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	LastActivity     time.Time              `json:"last_activity"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// FlowStatus defines the status of an onboarding flow
type FlowStatus string

const (
	FlowStatusNotStarted FlowStatus = "not_started"
	FlowStatusInProgress FlowStatus = "in_progress"
	FlowStatusCompleted  FlowStatus = "completed"
	FlowStatusAbandoned  FlowStatus = "abandoned"
)

// ScreenSubmission represents data submitted for a screen
type ScreenSubmission struct {
	ScreenID   string                 `json:"screen_id"`
	UserID     string                 `json:"user_id"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	Version    string                 `json:"version"`
	DeviceInfo DeviceInfo             `json:"device_info"`
}

// DeviceInfo contains information about the user's device
type DeviceInfo struct {
	Type       string `json:"type"`
	Platform   string `json:"platform"`
	AppVersion string `json:"app_version"`
	OSVersion  string `json:"os_version,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
}

// OnboardingAnalytics represents analytics data for onboarding
type OnboardingAnalytics struct {
	UserID           string                  `json:"user_id"`
	FlowID           string                  `json:"flow_id"`
	TotalScreens     int                     `json:"total_screens"`
	CompletedScreens int                     `json:"completed_screens"`
	CompletionRate   float64                 `json:"completion_rate"`
	TimeSpent        time.Duration           `json:"time_spent"`
	DropOffPoints    []string                `json:"drop_off_points"`
	ScreenMetrics    map[string]ScreenMetric `json:"screen_metrics"`
	DeviceInfo       DeviceInfo              `json:"device_info"`
	CreatedAt        time.Time               `json:"created_at"`
}

// ScreenMetric contains metrics for a specific screen
type ScreenMetric struct {
	ScreenID    string        `json:"screen_id"`
	ViewCount   int           `json:"view_count"`
	SubmitCount int           `json:"submit_count"`
	AverageTime time.Duration `json:"average_time"`
	DropOffRate float64       `json:"drop_off_rate"`
	ErrorCount  int           `json:"error_count"`
	LastViewed  time.Time     `json:"last_viewed"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Metadata  *Metadata   `json:"metadata,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Metadata contains additional response metadata
type Metadata struct {
	Version     string `json:"version"`
	RequestID   string `json:"request_id,omitempty"`
	ProcessTime string `json:"process_time,omitempty"`
}

// ScreenVersionConfig represents version-specific configuration for screens
type ScreenVersionConfig struct {
	Version    string                 `json:"version"`
	Enabled    bool                   `json:"enabled"`
	Percentage float64                `json:"percentage"` // For A/B testing
	Features   []string               `json:"features"`
	Overrides  map[string]interface{} `json:"overrides,omitempty"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
}

// FlowConfiguration represents the overall flow configuration
type FlowConfiguration struct {
	ID        string                         `json:"id"`
	Name      string                         `json:"name"`
	Version   string                         `json:"version"`
	Screens   []string                       `json:"screens"`
	Versions  map[string]ScreenVersionConfig `json:"versions"`
	Rules     FlowRules                      `json:"rules"`
	Analytics AnalyticsConfig                `json:"analytics"`
	CreatedAt time.Time                      `json:"created_at"`
	UpdatedAt time.Time                      `json:"updated_at"`
}

// FlowRules defines rules for flow progression
type FlowRules struct {
	AllowSkip       bool     `json:"allow_skip"`
	RequiredScreens []string `json:"required_screens"`
	ConditionalFlow bool     `json:"conditional_flow"`
	TimeoutMinutes  int      `json:"timeout_minutes"`
}

// AnalyticsConfig defines analytics collection settings
type AnalyticsConfig struct {
	Enabled          bool     `json:"enabled"`
	TrackViews       bool     `json:"track_views"`
	TrackSubmissions bool     `json:"track_submissions"`
	TrackTiming      bool     `json:"track_timing"`
	CustomEvents     []string `json:"custom_events,omitempty"`
}
