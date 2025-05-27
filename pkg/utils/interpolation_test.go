package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// mockContext implements interfaces.ExecutionContext for testing
type mockContext map[string]interface{}

func (m mockContext) Get(key string) (interface{}, bool) {
	v, ok := m[key]
	return v, ok
}

func (m mockContext) Set(key string, value interface{}) {
	m[key] = value
}

func (m mockContext) Delete(key string) {
	delete(m, key)
}

func (m mockContext) Has(key string) bool {
	_, exists := m[key]
	return exists
}

func (m mockContext) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m mockContext) GetString(key string) (string, error) {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}
	return "", nil
}

func (m mockContext) GetInt(key string) (int, error) {
	if val, ok := m[key]; ok {
		if i, ok := val.(int); ok {
			return i, nil
		}
	}
	return 0, nil
}

func (m mockContext) GetBool(key string) (bool, error) {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b, nil
		}
	}
	return false, nil
}

func (m mockContext) GetMap(key string) (map[string]interface{}, error) {
	if val, ok := m[key]; ok {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal, nil
		}
	}
	return nil, nil
}

func (m mockContext) Context() context.Context {
	return context.Background()
}

func (m mockContext) Clone() interfaces.ExecutionContext {
	newMock := make(mockContext)
	for k, v := range m {
		newMock[k] = v
	}
	return newMock
}

func (m mockContext) FlowName() string {
	return "test-flow"
}

func (m mockContext) ExecutionID() string {
	return "test-exec-id"
}

func (m mockContext) StartTime() time.Time {
	return time.Now()
}

func (m mockContext) Duration() time.Duration {
	return time.Second
}

func (m mockContext) Logger() *zap.Logger {
	return zap.NewNop()
}

func TestInterpolateString(t *testing.T) {
	t.Run("Basic interpolation", func(t *testing.T) {
		ctx := mockContext{"foo": "bar", "num": 42}
		result, err := InterpolateString("hello ${foo} ${num}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "hello bar 42", result)
	})

	t.Run("No variables", func(t *testing.T) {
		ctx := mockContext{}
		result, err := InterpolateString("hello world", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("Missing variable", func(t *testing.T) {
		ctx := mockContext{"foo": "bar"}
		result, err := InterpolateString("${foo} ${missing}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "bar ${missing}", result)
	})

	t.Run("Nested variable", func(t *testing.T) {
		ctx := mockContext{
			"user": map[string]interface{}{
				"name": "John",
				"age":  30,
			},
		}
		result, err := InterpolateString("Hello ${user.name}, age ${user.age}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Hello John, age 30", result)
	})

	t.Run("Unclosed variable", func(t *testing.T) {
		ctx := mockContext{"foo": "bar"}
		_, err := InterpolateString("hello ${foo", ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unclosed variable reference")
	})

	t.Run("Empty variable name", func(t *testing.T) {
		ctx := mockContext{}
		result, err := InterpolateString("hello ${}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "hello ${}", result)
	})

	t.Run("Multiple same variables", func(t *testing.T) {
		ctx := mockContext{"name": "Alice"}
		result, err := InterpolateString("${name} says hello to ${name}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Alice says hello to Alice", result)
	})

	t.Run("Variable at start and end", func(t *testing.T) {
		ctx := mockContext{"start": "Begin", "end": "Finish"}
		result, err := InterpolateString("${start} middle ${end}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Begin middle Finish", result)
	})

	t.Run("Nested braces", func(t *testing.T) {
		ctx := mockContext{"foo": "bar"}
		result, err := InterpolateString("${foo} and {not a var}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "bar and {not a var}", result)
	})

	t.Run("Boolean and nil values", func(t *testing.T) {
		ctx := mockContext{"bool": true, "null": nil}
		result, err := InterpolateString("bool: ${bool}, null: ${null}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "bool: true, null: <nil>", result)
	})
}

func TestInterpolateMap(t *testing.T) {
	t.Run("Basic map interpolation", func(t *testing.T) {
		ctx := mockContext{"foo": "bar", "num": 42}
		input := map[string]string{
			"greeting": "hello ${foo}",
			"count":    "number: ${num}",
			"static":   "no variables here",
		}
		result, err := InterpolateMap(input, ctx)
		assert.NoError(t, err)
		expected := map[string]string{
			"greeting": "hello bar",
			"count":    "number: 42",
			"static":   "no variables here",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Empty map", func(t *testing.T) {
		ctx := mockContext{}
		input := map[string]string{}
		result, err := InterpolateMap(input, ctx)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Map with error", func(t *testing.T) {
		ctx := mockContext{"foo": "bar"}
		input := map[string]string{
			"valid":   "${foo}",
			"invalid": "${foo",
		}
		_, err := InterpolateMap(input, ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to interpolate key invalid")
	})

	t.Run("Map with missing variables", func(t *testing.T) {
		ctx := mockContext{"exists": "value"}
		input := map[string]string{
			"found":    "${exists}",
			"notfound": "${missing}",
		}
		result, err := InterpolateMap(input, ctx)
		assert.NoError(t, err)
		expected := map[string]string{
			"found":    "value",
			"notfound": "${missing}",
		}
		assert.Equal(t, expected, result)
	})
}

func TestHasVariables(t *testing.T) {
	t.Run("Has variables", func(t *testing.T) {
		testCases := []string{
			"${foo}",
			"hello ${world}",
			"${start} middle ${end}",
			"text ${var} more text",
		}
		for _, tc := range testCases {
			assert.True(t, HasVariables(tc), "Should detect variables in: %s", tc)
		}
	})

	t.Run("No variables", func(t *testing.T) {
		testCases := []string{
			"no variables here",
			"just text",
			"",
			"{ not a variable }",
			"$notavar",
			"${incomplete",
			"incomplete}",
		}
		for _, tc := range testCases {
			assert.False(t, HasVariables(tc), "Should not detect variables in: %s", tc)
		}
	})
}

func TestExtractVariables(t *testing.T) {
	t.Run("Extract single variable", func(t *testing.T) {
		vars := ExtractVariables("hello ${world}")
		expected := []string{"world"}
		assert.Equal(t, expected, vars)
	})

	t.Run("Extract multiple variables", func(t *testing.T) {
		vars := ExtractVariables("${foo} and ${bar} and ${baz}")
		expected := []string{"foo", "bar", "baz"}
		assert.Equal(t, expected, vars)
	})

	t.Run("Extract nested variables", func(t *testing.T) {
		vars := ExtractVariables("${user.name} is ${user.age} years old")
		expected := []string{"user.name", "user.age"}
		assert.Equal(t, expected, vars)
	})

	t.Run("No variables", func(t *testing.T) {
		vars := ExtractVariables("no variables here")
		assert.Empty(t, vars)
	})

	t.Run("Empty string", func(t *testing.T) {
		vars := ExtractVariables("")
		assert.Empty(t, vars)
	})

	t.Run("Malformed variables", func(t *testing.T) {
		vars := ExtractVariables("${incomplete and ${}")
		expected := []string{""}
		assert.Equal(t, expected, vars)
	})

	t.Run("Duplicate variables", func(t *testing.T) {
		vars := ExtractVariables("${foo} and ${foo} again")
		expected := []string{"foo", "foo"}
		assert.Equal(t, expected, vars)
	})

	t.Run("Variables with special characters", func(t *testing.T) {
		vars := ExtractVariables("${var_name} and ${var-name} and ${var.name}")
		expected := []string{"var_name", "var-name", "var.name"}
		assert.Equal(t, expected, vars)
	})
}

func TestInterpolateStringEdgeCases(t *testing.T) {
	t.Run("Consecutive variables", func(t *testing.T) {
		ctx := mockContext{"a": "A", "b": "B"}
		result, err := InterpolateString("${a}${b}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "AB", result)
	})

	t.Run("Variable with spaces in name", func(t *testing.T) {
		ctx := mockContext{"var with spaces": "value"}
		result, err := InterpolateString("${var with spaces}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "value", result)
	})

	t.Run("Complex nested structure", func(t *testing.T) {
		ctx := mockContext{
			"config": map[string]interface{}{
				"database": map[string]interface{}{
					"host": "localhost",
					"port": 5432,
				},
			},
		}
		result, err := InterpolateString("Connect to ${config.database.host}:${config.database.port}", ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Connect to localhost:5432", result)
	})
}

func TestMockContextMethods(t *testing.T) {
	ctx := mockContext{"test": "value", "num": 42}

	t.Run("Set and Get", func(t *testing.T) {
		ctx.Set("new", "newvalue")
		val, ok := ctx.Get("new")
		assert.True(t, ok)
		assert.Equal(t, "newvalue", val)
	})

	t.Run("Has", func(t *testing.T) {
		assert.True(t, ctx.Has("test"))
		assert.False(t, ctx.Has("missing"))
	})

	t.Run("Delete", func(t *testing.T) {
		ctx.Set("todelete", "value")
		assert.True(t, ctx.Has("todelete"))
		ctx.Delete("todelete")
		assert.False(t, ctx.Has("todelete"))
	})

	t.Run("Keys", func(t *testing.T) {
		keys := ctx.Keys()
		assert.Contains(t, keys, "test")
		assert.Contains(t, keys, "num")
	})
}
