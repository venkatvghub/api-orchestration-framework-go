package registry

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// StepFactory is a function that creates a new step instance
type StepFactory func(config map[string]interface{}) (interfaces.Step, error)

// StepInfo holds metadata about a registered step
type StepInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Version     string                 `json:"version"`
	ConfigSpec  map[string]interface{} `json:"config_spec"`
	Factory     StepFactory            `json:"-"`
}

// StepRegistry manages step registration and discovery
type StepRegistry struct {
	mu    sync.RWMutex
	steps map[string]*StepInfo
}

// NewStepRegistry creates a new step registry
func NewStepRegistry() *StepRegistry {
	return &StepRegistry{
		steps: make(map[string]*StepInfo),
	}
}

// Register registers a new step type
func (r *StepRegistry) Register(info *StepInfo) error {
	if info.Name == "" {
		return fmt.Errorf("step name cannot be empty")
	}
	if info.Factory == nil {
		return fmt.Errorf("step factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.steps[info.Name]; exists {
		return fmt.Errorf("step '%s' is already registered", info.Name)
	}

	r.steps[info.Name] = info
	return nil
}

// Unregister removes a step type from the registry
func (r *StepRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.steps[name]; !exists {
		return fmt.Errorf("step '%s' is not registered", name)
	}

	delete(r.steps, name)
	return nil
}

// Create creates a new step instance by name
func (r *StepRegistry) Create(name string, config map[string]interface{}) (interfaces.Step, error) {
	r.mu.RLock()
	info, exists := r.steps[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("step '%s' is not registered", name)
	}

	return info.Factory(config)
}

// GetInfo returns information about a registered step
func (r *StepRegistry) GetInfo(name string) (*StepInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.steps[name]
	if !exists {
		return nil, fmt.Errorf("step '%s' is not registered", name)
	}

	// Return a copy to prevent modification
	return &StepInfo{
		Name:        info.Name,
		Description: info.Description,
		Category:    info.Category,
		Version:     info.Version,
		ConfigSpec:  info.ConfigSpec,
	}, nil
}

// List returns all registered step names
func (r *StepRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.steps))
	for name := range r.steps {
		names = append(names, name)
	}
	return names
}

// ListByCategory returns step names filtered by category
func (r *StepRegistry) ListByCategory(category string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, info := range r.steps {
		if info.Category == category {
			names = append(names, name)
		}
	}
	return names
}

// GetAllInfo returns information about all registered steps
func (r *StepRegistry) GetAllInfo() map[string]*StepInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*StepInfo)
	for name, info := range r.steps {
		result[name] = &StepInfo{
			Name:        info.Name,
			Description: info.Description,
			Category:    info.Category,
			Version:     info.Version,
			ConfigSpec:  info.ConfigSpec,
		}
	}
	return result
}

// Exists checks if a step is registered
func (r *StepRegistry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.steps[name]
	return exists
}

// Global registry instance
var globalRegistry = NewStepRegistry()

// Global registry functions
func Register(info *StepInfo) error {
	return globalRegistry.Register(info)
}

func Create(name string, config map[string]interface{}) (interfaces.Step, error) {
	return globalRegistry.Create(name, config)
}

func GetInfo(name string) (*StepInfo, error) {
	return globalRegistry.GetInfo(name)
}

func List() []string {
	return globalRegistry.List()
}

func ListByCategory(category string) []string {
	return globalRegistry.ListByCategory(category)
}

func GetAllInfo() map[string]*StepInfo {
	return globalRegistry.GetAllInfo()
}

func Exists(name string) bool {
	return globalRegistry.Exists(name)
}

// Helper function to register a step with reflection-based config spec
func RegisterWithReflection(name, description, category, version string, factory StepFactory, configType interface{}) error {
	configSpec := generateConfigSpec(configType)

	info := &StepInfo{
		Name:        name,
		Description: description,
		Category:    category,
		Version:     version,
		ConfigSpec:  configSpec,
		Factory:     factory,
	}

	return Register(info)
}

// generateConfigSpec uses reflection to generate a config specification
func generateConfigSpec(configType interface{}) map[string]interface{} {
	if configType == nil {
		return nil
	}

	t := reflect.TypeOf(configType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	spec := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldSpec := map[string]interface{}{
			"type":     field.Type.String(),
			"required": true,
		}

		// Check for json tags
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			fieldSpec["json_name"] = jsonTag
		}

		// Check for validation tags
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			fieldSpec["validation"] = validateTag
		}

		// Check for description tags
		if descTag := field.Tag.Get("description"); descTag != "" {
			fieldSpec["description"] = descTag
		}

		// Check for default tags
		if defaultTag := field.Tag.Get("default"); defaultTag != "" {
			fieldSpec["default"] = defaultTag
			fieldSpec["required"] = false
		}

		spec[field.Name] = fieldSpec
	}

	return spec
}

// Step categories
const (
	CategoryCore       = "core"
	CategoryHTTP       = "http"
	CategoryBFF        = "bff"
	CategoryAuth       = "auth"
	CategoryCache      = "cache"
	CategoryTransform  = "transform"
	CategoryValidation = "validation"
	CategoryUtility    = "utility"
)
