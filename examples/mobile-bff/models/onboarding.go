package models

import (
	"time"
)

// OnboardingScreen represents a single onboarding screen configuration
type OnboardingScreen struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Subtitle    string                 `json:"subtitle,omitempty"`
	Type        ScreenType             `json:"type"`
	Fields      []ScreenField          `json:"fields"`
	Actions     []ScreenAction         `json:"actions"`
	Validation  ValidationRules        `json:"validation"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	NextScreen  string                 `json:"next_screen,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ScreenType defines the type of onboarding screen
type ScreenType string

const (
	ScreenTypeWelcome      ScreenType = "welcome"
	ScreenTypePersonalInfo ScreenType = "personal_info"
	ScreenTypePreferences  ScreenType = "preferences"
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
	ActionTypeAction   ActionType = "action"
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
	CurrentScreen    string                 `json:"current_screen"`
	CompletedScreens []string               `json:"completed_screens"`
	UserData         map[string]interface{} `json:"user_data"`
	Progress         float64                `json:"progress"`
	Status           FlowStatus             `json:"status"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	LastActivity     time.Time              `json:"last_activity"`
	TotalScreens     int                    `json:"total_screens"`
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
	Source      string `json:"source,omitempty"`
}

// MockAPIRequest represents a request to mock API
type MockAPIRequest struct {
	UserID     string                 `json:"user_id"`
	ScreenID   string                 `json:"screen_id,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	DeviceInfo map[string]interface{} `json:"device_info,omitempty"`
}

// MockAPIResponse represents a response from mock API
type MockAPIResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// SubmissionResponse represents the response after submitting a screen
type SubmissionResponse struct {
	Success    bool                   `json:"success"`
	NextScreen string                 `json:"next_screen,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// OnboardingProgress represents user progress through onboarding
type OnboardingProgress struct {
	UserID          string    `json:"user_id"`
	CurrentScreen   string    `json:"current_screen"`
	CompletedSteps  int       `json:"completed_steps"`
	TotalSteps      int       `json:"total_steps"`
	PercentComplete float64   `json:"percent_complete"`
	LastUpdated     time.Time `json:"last_updated"`
}

// UserAnalytics represents analytics data for a user
type UserAnalytics struct {
	UserID       string                 `json:"user_id"`
	TotalEvents  int                    `json:"total_events"`
	ScreenViews  int                    `json:"screen_views"`
	Submissions  int                    `json:"submissions"`
	TimeSpent    time.Duration          `json:"time_spent"`
	LastActivity time.Time              `json:"last_activity"`
	DeviceInfo   map[string]interface{} `json:"device_info"`
	CustomEvents []AnalyticsEvent       `json:"custom_events"`
}

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	EventType  string                 `json:"event_type"`
	UserID     string                 `json:"user_id"`
	ScreenID   string                 `json:"screen_id,omitempty"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
}

// VerificationRequest represents a verification request
type VerificationRequest struct {
	UserID           string `json:"user_id"`
	Type             string `json:"type"`   // email, phone
	Target           string `json:"target"` // email address or phone number
	Code             string `json:"code,omitempty"`
	VerificationCode string `json:"verification_code,omitempty"`
}

// VerificationResponse represents a verification response
type VerificationResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	Verified     bool   `json:"verified"`
	CodeSent     bool   `json:"code_sent,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	AttemptsLeft int    `json:"attempts_left,omitempty"`
}
