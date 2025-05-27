package registry

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// MockStep is a simple mock implementation of interfaces.Step for testing.
type MockStep struct {
	NameValue        string
	DescriptionValue string
	RunError         error
}

func (ms *MockStep) Run(ctx interfaces.ExecutionContext) error {
	return ms.RunError
}

func (ms *MockStep) Name() string {
	return ms.NameValue
}

func (ms *MockStep) Description() string {
	return ms.DescriptionValue
}

// Helper to create a mock StepFactory
func mockStepFactory(name, desc string, errOnCreate bool) StepFactory {
	return func(config map[string]interface{}) (interfaces.Step, error) {
		if errOnCreate {
			return nil, errors.New("factory error")
		}
		return &MockStep{NameValue: name, DescriptionValue: desc}, nil
	}
}

func TestNewStepRegistry(t *testing.T) {
	r := NewStepRegistry()
	assert.NotNil(t, r)
	assert.NotNil(t, r.steps)
	assert.Empty(t, r.steps)
}

func TestStepRegistry_Register(t *testing.T) {
	r := NewStepRegistry()
	mockFactory := mockStepFactory("test-step", "A test step", false)

	t.Run("successful registration", func(t *testing.T) {
		info := &StepInfo{Name: "test-step", Factory: mockFactory, Description: "desc", Category: "cat", Version: "v1"}
		err := r.Register(info)
		assert.NoError(t, err)
		assert.Contains(t, r.steps, "test-step")
		assert.Equal(t, info, r.steps["test-step"])
		r.Unregister("test-step") // Clean up
	})

	t.Run("empty name", func(t *testing.T) {
		info := &StepInfo{Name: "", Factory: mockFactory}
		err := r.Register(info)
		assert.Error(t, err)
		assert.EqualError(t, err, "step name cannot be empty")
	})

	t.Run("nil factory", func(t *testing.T) {
		info := &StepInfo{Name: "test-step-nil-factory", Factory: nil}
		err := r.Register(info)
		assert.Error(t, err)
		assert.EqualError(t, err, "step factory cannot be nil")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		info1 := &StepInfo{Name: "dup-step", Factory: mockFactory}
		r.Register(info1)
		info2 := &StepInfo{Name: "dup-step", Factory: mockFactory}
		err := r.Register(info2)
		assert.Error(t, err)
		assert.EqualError(t, err, "step 'dup-step' is already registered")
		r.Unregister("dup-step") // Clean up
	})
}

func TestStepRegistry_Unregister(t *testing.T) {
	r := NewStepRegistry()
	mockFactory := mockStepFactory("test-unregister", "Unregister test", false)
	info := &StepInfo{Name: "test-unregister", Factory: mockFactory}
	r.Register(info)

	t.Run("successful unregistration", func(t *testing.T) {
		err := r.Unregister("test-unregister")
		assert.NoError(t, err)
		assert.NotContains(t, r.steps, "test-unregister")
	})

	t.Run("unregister non-existent step", func(t *testing.T) {
		err := r.Unregister("non-existent-step")
		assert.Error(t, err)
		assert.EqualError(t, err, "step 'non-existent-step' is not registered")
	})
}

func TestStepRegistry_Create(t *testing.T) {
	r := NewStepRegistry()
	mockFactory := mockStepFactory("created-step", "Created for test", false)
	info := &StepInfo{Name: "creatable-step", Factory: mockFactory}
	r.Register(info)

	factoryWithError := mockStepFactory("error-step", "Errors on create", true)
	infoError := &StepInfo{Name: "error-step", Factory: factoryWithError}
	r.Register(infoError)

	t.Run("successful creation", func(t *testing.T) {
		step, err := r.Create("creatable-step", nil)
		assert.NoError(t, err)
		assert.NotNil(t, step)
		assert.Equal(t, "created-step", step.Name())
	})

	t.Run("create non-existent step", func(t *testing.T) {
		step, err := r.Create("non-existent-create", nil)
		assert.Error(t, err)
		assert.Nil(t, step)
		assert.EqualError(t, err, "step 'non-existent-create' is not registered")
	})

	t.Run("factory returns error", func(t *testing.T) {
		step, err := r.Create("error-step", nil)
		assert.Error(t, err)
		assert.Nil(t, step)
		assert.EqualError(t, err, "factory error")
	})
}

func TestStepRegistry_GetInfo(t *testing.T) {
	r := NewStepRegistry()
	mockFactory := mockStepFactory("info-step", "Info test", false)
	originalInfo := &StepInfo{
		Name:        "info-step",
		Factory:     mockFactory,
		Description: "Original Description",
		Category:    "TestCategory",
		Version:     "1.0.0",
		ConfigSpec:  map[string]interface{}{"key": "value"},
	}
	r.Register(originalInfo)

	t.Run("successful get info", func(t *testing.T) {
		retrievedInfo, err := r.GetInfo("info-step")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedInfo)
		assert.Equal(t, originalInfo.Name, retrievedInfo.Name)
		assert.Equal(t, originalInfo.Description, retrievedInfo.Description)
		assert.Equal(t, originalInfo.Category, retrievedInfo.Category)
		assert.Equal(t, originalInfo.Version, retrievedInfo.Version)
		assert.Equal(t, originalInfo.ConfigSpec, retrievedInfo.ConfigSpec)
		assert.Nil(t, retrievedInfo.Factory) // Factory should not be exposed

		// Test that it's a copy
		retrievedInfo.Description = "Modified Description"
		assert.NotEqual(t, retrievedInfo.Description, originalInfo.Description)
	})

	t.Run("get info for non-existent step", func(t *testing.T) {
		info, err := r.GetInfo("non-existent-info")
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.EqualError(t, err, "step 'non-existent-info' is not registered")
	})
}

func TestStepRegistry_List(t *testing.T) {
	r := NewStepRegistry()
	r.Register(&StepInfo{Name: "step1", Factory: mockStepFactory("s1", "d1", false)})
	r.Register(&StepInfo{Name: "step2", Factory: mockStepFactory("s2", "d2", false)})

	names := r.List()
	assert.ElementsMatch(t, []string{"step1", "step2"}, names)

	rEmpty := NewStepRegistry()
	assert.Empty(t, rEmpty.List())
}

func TestStepRegistry_ListByCategory(t *testing.T) {
	r := NewStepRegistry()
	r.Register(&StepInfo{Name: "stepA1", Category: "A", Factory: mockStepFactory("sA1", "", false)})
	r.Register(&StepInfo{Name: "stepB1", Category: "B", Factory: mockStepFactory("sB1", "", false)})
	r.Register(&StepInfo{Name: "stepA2", Category: "A", Factory: mockStepFactory("sA2", "", false)})

	categoryA := r.ListByCategory("A")
	assert.ElementsMatch(t, []string{"stepA1", "stepA2"}, categoryA)

	categoryB := r.ListByCategory("B")
	assert.ElementsMatch(t, []string{"stepB1"}, categoryB)

	categoryC := r.ListByCategory("C")
	assert.Empty(t, categoryC)
}

func TestStepRegistry_GetAllInfo(t *testing.T) {
	r := NewStepRegistry()
	info1 := &StepInfo{Name: "all1", Description: "d1", Category: "c1", Version: "v1", ConfigSpec: map[string]interface{}{"k1": "v1"}, Factory: mockStepFactory("all1", "", false)}
	info2 := &StepInfo{Name: "all2", Description: "d2", Category: "c2", Version: "v2", ConfigSpec: map[string]interface{}{"k2": "v2"}, Factory: mockStepFactory("all2", "", false)}
	r.Register(info1)
	r.Register(info2)

	allInfos := r.GetAllInfo()
	assert.Len(t, allInfos, 2)
	assert.Contains(t, allInfos, "all1")
	assert.Contains(t, allInfos, "all2")

	retrievedInfo1 := allInfos["all1"]
	assert.Equal(t, info1.Name, retrievedInfo1.Name)
	assert.Equal(t, info1.Description, retrievedInfo1.Description)
	assert.Nil(t, retrievedInfo1.Factory)

	// Test it's a copy
	retrievedInfo1.Description = "modified"
	assert.NotEqual(t, retrievedInfo1.Description, info1.Description)

	rEmpty := NewStepRegistry()
	assert.Empty(t, rEmpty.GetAllInfo())
}

func TestStepRegistry_Exists(t *testing.T) {
	r := NewStepRegistry()
	r.Register(&StepInfo{Name: "exists-step", Factory: mockStepFactory("exists", "", false)})

	assert.True(t, r.Exists("exists-step"))
	assert.False(t, r.Exists("not-exists-step"))
}

// Test global registry functions (they wrap the globalRegistry instance)
// We need to be careful as globalRegistry is shared state. Reset it for tests.
func resetGlobalRegistry() {
	globalRegistry.mu.Lock()
	globalRegistry.steps = make(map[string]*StepInfo)
	globalRegistry.mu.Unlock()
}

func TestGlobalRegistryFunctions(t *testing.T) {
	// Ensure a clean global registry for these tests
	originalGlobalSteps := globalRegistry.steps
	globalRegistry.steps = make(map[string]*StepInfo)             // Temporary clean slate
	defer func() { globalRegistry.steps = originalGlobalSteps }() // Restore

	mockFactory := mockStepFactory("global-step", "Global test", false)
	info := &StepInfo{Name: "global-step", Factory: mockFactory, Category: "globalCat"}

	t.Run("Global Register and Exists", func(t *testing.T) {
		resetGlobalRegistry()
		err := Register(info)
		assert.NoError(t, err)
		assert.True(t, Exists("global-step"))
	})

	t.Run("Global Create", func(t *testing.T) {
		resetGlobalRegistry()
		Register(info)
		step, err := Create("global-step", nil)
		assert.NoError(t, err)
		assert.NotNil(t, step)
		assert.Equal(t, "global-step", step.Name())
	})

	t.Run("Global GetInfo", func(t *testing.T) {
		resetGlobalRegistry()
		Register(info)
		retrieved, err := GetInfo("global-step")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "global-step", retrieved.Name)
	})

	t.Run("Global List", func(t *testing.T) {
		resetGlobalRegistry()
		Register(info)
		Register(&StepInfo{Name: "global-step2", Factory: mockFactory})
		list := List()
		assert.ElementsMatch(t, []string{"global-step", "global-step2"}, list)
	})

	t.Run("Global ListByCategory", func(t *testing.T) {
		resetGlobalRegistry()
		Register(info) // Category: "globalCat"
		Register(&StepInfo{Name: "global-step-other", Factory: mockFactory, Category: "otherCat"})
		list := ListByCategory("globalCat")
		assert.ElementsMatch(t, []string{"global-step"}, list)
	})

	t.Run("Global GetAllInfo", func(t *testing.T) {
		resetGlobalRegistry()
		Register(info)
		all := GetAllInfo()
		assert.Len(t, all, 1)
		assert.Contains(t, all, "global-step")
	})
}

type SampleConfig struct {
	FieldString  string `json:"field_string" validate:"required" description:"A string field"`
	FieldInt     int    `json:"field_int,omitempty" default:"10"`
	FieldBool    bool
	privateField string
}

func TestGenerateConfigSpec(t *testing.T) {
	t.Run("nil config type", func(t *testing.T) {
		spec := generateConfigSpec(nil)
		assert.Nil(t, spec)
	})

	t.Run("non-struct type", func(t *testing.T) {
		var i int
		spec := generateConfigSpec(i)
		assert.Nil(t, spec)
		spec = generateConfigSpec(&i) // Pointer to non-struct
		assert.Nil(t, spec)
	})

	t.Run("struct type", func(t *testing.T) {
		config := SampleConfig{}
		spec := generateConfigSpec(config)

		assert.NotNil(t, spec)
		assert.Len(t, spec, 3) // FieldString, FieldInt, FieldBool (privateField is ignored)

		// FieldString
		fieldStringSpec, ok := spec["FieldString"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "string", fieldStringSpec["type"])
		assert.Equal(t, true, fieldStringSpec["required"])
		assert.Equal(t, "field_string", fieldStringSpec["json_name"])
		assert.Equal(t, "required", fieldStringSpec["validation"])
		assert.Equal(t, "A string field", fieldStringSpec["description"])

		// FieldInt
		fieldIntSpec, ok := spec["FieldInt"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "int", fieldIntSpec["type"])
		assert.Equal(t, false, fieldIntSpec["required"]) // Has default
		assert.Equal(t, "field_int,omitempty", fieldIntSpec["json_name"])
		assert.Equal(t, "10", fieldIntSpec["default"])

		// FieldBool
		fieldBoolSpec, ok := spec["FieldBool"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "bool", fieldBoolSpec["type"])
		assert.Equal(t, true, fieldBoolSpec["required"]) // No default
	})

	t.Run("pointer to struct type", func(t *testing.T) {
		config := &SampleConfig{}
		spec := generateConfigSpec(config)
		assert.NotNil(t, spec)
		assert.Len(t, spec, 3)
	})
}

func TestRegisterWithReflection(t *testing.T) {
	resetGlobalRegistry()
	defer resetGlobalRegistry()

	mockFactory := mockStepFactory("reflected-step", "Reflected", false)

	t.Run("successful registration with reflection", func(t *testing.T) {
		err := RegisterWithReflection("reflected-step", "A step registered via reflection", "ReflectCategory", "v0.1", mockFactory, SampleConfig{})
		assert.NoError(t, err)

		info, err := GetInfo("reflected-step")
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "reflected-step", info.Name)
		assert.Equal(t, "A step registered via reflection", info.Description)
		assert.Equal(t, "ReflectCategory", info.Category)
		assert.Equal(t, "v0.1", info.Version)
		assert.NotNil(t, info.ConfigSpec)
		assert.Len(t, info.ConfigSpec, 3)
		fieldStringSpec, ok := info.ConfigSpec["FieldString"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "string", fieldStringSpec["type"])
	})

	t.Run("registration with nil config type", func(t *testing.T) {
		err := RegisterWithReflection("reflected-nil-config", "Nil config", "ReflectCategory", "v0.1", mockFactory, nil)
		assert.NoError(t, err)
		info, err := GetInfo("reflected-nil-config")
		assert.NoError(t, err)
		assert.Nil(t, info.ConfigSpec)
	})

	t.Run("registration with non-struct config type", func(t *testing.T) {
		var i int
		err := RegisterWithReflection("reflected-non-struct", "Non-struct config", "ReflectCategory", "v0.1", mockFactory, i)
		assert.NoError(t, err)
		info, err := GetInfo("reflected-non-struct")
		assert.NoError(t, err)
		assert.Nil(t, info.ConfigSpec)
	})

	// Test error propagation from Register (e.g., duplicate name)
	t.Run("duplicate registration with reflection", func(t *testing.T) {
		RegisterWithReflection("dup-reflect", "desc", "cat", "v1", mockFactory, nil)          // First registration
		err := RegisterWithReflection("dup-reflect", "desc2", "cat2", "v2", mockFactory, nil) // Attempt duplicate
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is already registered")
	})
}

// Test concurrency safety for registry operations
func TestStepRegistry_Concurrency(t *testing.T) {
	r := NewStepRegistry()
	numGoroutines := 100
	var wg sync.WaitGroup

	// Concurrent Registrations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent-step-%d", idx)
			info := &StepInfo{Name: name, Factory: mockStepFactory(name, "", false)}
			r.Register(info)
		}(i)
	}
	wg.Wait()
	assert.Len(t, r.steps, numGoroutines, "All registrations should succeed concurrently")

	// Concurrent Creates
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent-step-%d", idx)
			step, err := r.Create(name, nil)
			assert.NoError(t, err)
			assert.NotNil(t, step)
		}(i)
	}
	wg.Wait()

	// Concurrent GetInfo
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent-step-%d", idx)
			info, err := r.GetInfo(name)
			assert.NoError(t, err)
			assert.NotNil(t, info)
		}(i)
	}
	wg.Wait()

	// Concurrent List
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			list := r.List()
			assert.Len(t, list, numGoroutines)
		}()
	}
	wg.Wait()

	// Concurrent Unregistrations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent-step-%d", idx)
			r.Unregister(name)
		}(i)
	}
	wg.Wait()
	assert.Empty(t, r.steps, "All unregistrations should succeed concurrently")
}
