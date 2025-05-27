package utils

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// mockExecutionContext implements interfaces.ExecutionContext for testing
type mockExecutionContext map[string]interface{}

func (m mockExecutionContext) Get(key string) (interface{}, bool) {
	v, ok := m[key]
	return v, ok
}

func (m mockExecutionContext) Set(key string, value interface{}) {
	m[key] = value
}

func (m mockExecutionContext) Delete(key string) {
	delete(m, key)
}

func (m mockExecutionContext) Has(key string) bool {
	_, exists := m[key]
	return exists
}

func (m mockExecutionContext) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m mockExecutionContext) GetString(key string) (string, error) {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}
	return "", nil
}

func (m mockExecutionContext) GetInt(key string) (int, error) {
	if val, ok := m[key]; ok {
		if i, ok := val.(int); ok {
			return i, nil
		}
	}
	return 0, nil
}

func (m mockExecutionContext) GetBool(key string) (bool, error) {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b, nil
		}
	}
	return false, nil
}

func (m mockExecutionContext) GetMap(key string) (map[string]interface{}, error) {
	if val, ok := m[key]; ok {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal, nil
		}
	}
	return nil, nil
}

func (m mockExecutionContext) GetLogger() interface{} {
	return zap.NewNop()
}

func (m mockExecutionContext) GetMetrics() interface{} {
	return nil
}

func (m mockExecutionContext) Context() context.Context {
	return context.Background()
}

func (m mockExecutionContext) WithTimeout(timeout time.Duration) (interfaces.ExecutionContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = ctx
	return m, cancel
}

func (m mockExecutionContext) WithCancel() (interfaces.ExecutionContext, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	return m, cancel
}

func (m mockExecutionContext) WithValue(key, value interface{}) interfaces.ExecutionContext {
	newMock := make(mockExecutionContext)
	for k, v := range m {
		newMock[k] = v
	}
	if strKey, ok := key.(string); ok {
		newMock[strKey] = value
	}
	return newMock
}

func (m mockExecutionContext) IsTimedOut() bool {
	return false
}

func (m mockExecutionContext) IsCancelled() bool {
	return false
}

func (m mockExecutionContext) GetStartTime() time.Time {
	return time.Now()
}

func (m mockExecutionContext) GetElapsedTime() time.Duration {
	return time.Second
}

func (m mockExecutionContext) Clone() interfaces.ExecutionContext {
	newMock := make(mockExecutionContext)
	for k, v := range m {
		newMock[k] = v
	}
	return newMock
}

func (m mockExecutionContext) FlowName() string {
	return "test-flow"
}

func (m mockExecutionContext) ExecutionID() string {
	return "test-exec-id"
}

func (m mockExecutionContext) StartTime() time.Time {
	return time.Now()
}

func (m mockExecutionContext) Duration() time.Duration {
	return time.Second
}

func (m mockExecutionContext) Logger() *zap.Logger {
	return zap.NewNop()
}

func TestGetNestedValue(t *testing.T) {
	ctx := flow.NewContext()
	ctx.Set("foo", map[string]interface{}{
		"bar": map[string]interface{}{
			"baz": 42,
		},
	})
	ctx.Set("single", 99)
	ctx.Set("null", nil)

	t.Run("Nested value exists", func(t *testing.T) {
		result := GetNestedValue("foo.bar.baz", ctx)
		assert.Equal(t, "42", result)
	})

	t.Run("Single value exists", func(t *testing.T) {
		result := GetNestedValue("single", ctx)
		assert.Equal(t, "99", result)
	})

	t.Run("Nested value missing", func(t *testing.T) {
		result := GetNestedValue("foo.bar.missing", ctx)
		assert.Equal(t, "${foo.bar.missing}", result)
	})

	t.Run("Root value missing", func(t *testing.T) {
		result := GetNestedValue("missing", ctx)
		assert.Equal(t, "${missing}", result)
	})

	t.Run("Null value", func(t *testing.T) {
		result := GetNestedValue("null", ctx)
		assert.Equal(t, "${null}", result)
	})

	t.Run("Path through non-map", func(t *testing.T) {
		ctx.Set("notmap", "string")
		result := GetNestedValue("notmap.field", ctx)
		assert.Equal(t, "${notmap.field}", result)
	})
}

func TestSetNestedValue(t *testing.T) {
	t.Run("Set simple value", func(t *testing.T) {
		data := make(map[string]interface{})
		err := SetNestedValue(data, "key", "value")
		assert.NoError(t, err)
		assert.Equal(t, "value", data["key"])
	})

	t.Run("Set nested value", func(t *testing.T) {
		data := make(map[string]interface{})
		err := SetNestedValue(data, "a.b.c", 123)
		assert.NoError(t, err)

		nested := data["a"].(map[string]interface{})["b"].(map[string]interface{})
		assert.Equal(t, 123, nested["c"])
	})

	t.Run("Overwrite existing nested value", func(t *testing.T) {
		data := map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"c": "old",
				},
			},
		}
		err := SetNestedValue(data, "a.b.c", "new")
		assert.NoError(t, err)

		nested := data["a"].(map[string]interface{})["b"].(map[string]interface{})
		assert.Equal(t, "new", nested["c"])
	})

	t.Run("Error when intermediate is not map", func(t *testing.T) {
		data := map[string]interface{}{
			"a": "not a map",
		}
		err := SetNestedValue(data, "a.b.c", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a map")
	})

	t.Run("Create intermediate maps", func(t *testing.T) {
		data := make(map[string]interface{})
		err := SetNestedValue(data, "deep.nested.path.value", "test")
		assert.NoError(t, err)

		result := data["deep"].(map[string]interface{})["nested"].(map[string]interface{})["path"].(map[string]interface{})["value"]
		assert.Equal(t, "test", result)
	})
}

func TestHasNestedValue(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "value",
			},
		},
		"simple": "value",
		"null":   nil,
	}

	t.Run("Has nested value", func(t *testing.T) {
		assert.True(t, HasNestedValue(data, "a.b.c"))
	})

	t.Run("Has simple value", func(t *testing.T) {
		assert.True(t, HasNestedValue(data, "simple"))
	})

	t.Run("Has null value", func(t *testing.T) {
		assert.True(t, HasNestedValue(data, "null"))
	})

	t.Run("Missing nested value", func(t *testing.T) {
		assert.False(t, HasNestedValue(data, "a.b.missing"))
	})

	t.Run("Missing root value", func(t *testing.T) {
		assert.False(t, HasNestedValue(data, "missing"))
	})

	t.Run("Path through non-map", func(t *testing.T) {
		assert.False(t, HasNestedValue(data, "simple.field"))
	})
}

func TestFlattenMap(t *testing.T) {
	t.Run("Simple map", func(t *testing.T) {
		data := map[string]interface{}{
			"a": 1,
			"b": 2,
		}
		result := FlattenMap(data, "")
		expected := map[string]interface{}{
			"a": 1,
			"b": 2,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Nested map", func(t *testing.T) {
		data := map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"c": 1,
				},
				"d": 2,
			},
			"e": 3,
		}
		result := FlattenMap(data, "")
		expected := map[string]interface{}{
			"a.b.c": 1,
			"a.d":   2,
			"e":     3,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("With prefix", func(t *testing.T) {
		data := map[string]interface{}{
			"a": map[string]interface{}{
				"b": 1,
			},
		}
		result := FlattenMap(data, "prefix")
		expected := map[string]interface{}{
			"prefix.a.b": 1,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Interface map conversion", func(t *testing.T) {
		data := map[string]interface{}{
			"a": map[interface{}]interface{}{
				"b": 1,
				"c": 2,
			},
		}
		result := FlattenMap(data, "")
		expected := map[string]interface{}{
			"a.b": 1,
			"a.c": 2,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Empty map", func(t *testing.T) {
		data := map[string]interface{}{}
		result := FlattenMap(data, "")
		assert.Empty(t, result)
	})
}

func TestUnflattenMap(t *testing.T) {
	t.Run("Simple unflatten", func(t *testing.T) {
		data := map[string]interface{}{
			"a.b.c": 1,
			"a.d":   2,
			"e":     3,
		}
		result := UnflattenMap(data)
		expected := map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"c": 1,
				},
				"d": 2,
			},
			"e": 3,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Single level", func(t *testing.T) {
		data := map[string]interface{}{
			"a": 1,
			"b": 2,
		}
		result := UnflattenMap(data)
		assert.Equal(t, data, result)
	})

	t.Run("Error handling", func(t *testing.T) {
		data := map[string]interface{}{
			"a":   1,
			"a.b": 2, // This should cause an error and fall back to key as-is
		}
		result := UnflattenMap(data)
		// Should contain both keys as-is due to conflict
		assert.Contains(t, result, "a")
		assert.Contains(t, result, "a.b")
	})
}

func TestGetNestedValueFromContext(t *testing.T) {
	ctx := mockExecutionContext{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "John",
				"age":  30,
			},
		},
		"simple": "value",
		"null":   nil,
	}

	t.Run("Get nested value", func(t *testing.T) {
		result := GetNestedValueFromContext("user.profile.name", ctx)
		assert.Equal(t, "John", result)
	})

	t.Run("Get simple value", func(t *testing.T) {
		result := GetNestedValueFromContext("simple", ctx)
		assert.Equal(t, "value", result)
	})

	t.Run("Missing root", func(t *testing.T) {
		result := GetNestedValueFromContext("missing", ctx)
		assert.Equal(t, "", result)
	})

	t.Run("Missing nested", func(t *testing.T) {
		result := GetNestedValueFromContext("user.profile.missing", ctx)
		assert.Equal(t, "", result)
	})

	t.Run("Empty path", func(t *testing.T) {
		result := GetNestedValueFromContext("", ctx)
		assert.Equal(t, "", result)
	})

	t.Run("Null value", func(t *testing.T) {
		result := GetNestedValueFromContext("null", ctx)
		assert.Equal(t, "<nil>", result)
	})
}

func TestGetNestedValueFromInterface(t *testing.T) {
	t.Run("Map string interface", func(t *testing.T) {
		data := map[string]interface{}{"key": "value"}
		result := getNestedValueFromInterface(data, "key")
		assert.Equal(t, "value", result)
	})

	t.Run("Map interface interface", func(t *testing.T) {
		data := map[interface{}]interface{}{"key": "value"}
		result := getNestedValueFromInterface(data, "key")
		assert.Equal(t, "value", result)
	})

	t.Run("Map string string", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		result := getNestedValueFromInterface(data, "key")
		assert.Equal(t, "value", result)
	})

	t.Run("Nil data", func(t *testing.T) {
		result := getNestedValueFromInterface(nil, "key")
		assert.Nil(t, result)
	})

	t.Run("Struct with reflection", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		data := TestStruct{Name: "John", Age: 30}
		result := getNestedValueFromInterface(data, "Name")
		assert.Equal(t, "John", result)
	})

	t.Run("Unsupported type", func(t *testing.T) {
		result := getNestedValueFromInterface("string", "key")
		assert.Nil(t, result)
	})
}

func TestGetValueFromStruct(t *testing.T) {
	type TestStruct struct {
		Name    string
		Age     int
		private string
	}

	t.Run("Get public field", func(t *testing.T) {
		data := TestStruct{Name: "John", Age: 30}
		result := getValueFromStruct(data, "Name")
		assert.Equal(t, "John", result)
	})

	t.Run("Get field case insensitive", func(t *testing.T) {
		data := TestStruct{Name: "John", Age: 30}
		result := getValueFromStruct(data, "name")
		assert.Equal(t, "John", result)
	})

	t.Run("Missing field", func(t *testing.T) {
		data := TestStruct{Name: "John", Age: 30}
		result := getValueFromStruct(data, "Missing")
		assert.Nil(t, result)
	})

	t.Run("Nil data", func(t *testing.T) {
		result := getValueFromStruct(nil, "Name")
		assert.Nil(t, result)
	})

	t.Run("Pointer to struct", func(t *testing.T) {
		data := &TestStruct{Name: "John", Age: 30}
		result := getValueFromStruct(data, "Name")
		assert.Equal(t, "John", result)
	})

	t.Run("Non-struct type", func(t *testing.T) {
		result := getValueFromStruct("not a struct", "field")
		assert.Nil(t, result)
	})
}

func TestFlattenUnflattenRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "value",
				"other":  42,
			},
			"simple": "test",
		},
		"root": "rootvalue",
	}

	flattened := FlattenMap(original, "")
	unflattened := UnflattenMap(flattened)

	assert.True(t, reflect.DeepEqual(original, unflattened), "Round trip should preserve structure")
}
