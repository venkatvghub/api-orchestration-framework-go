package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockTransformer for testing
type MockTransformer struct {
	mock.Mock
}

func (m *MockTransformer) Name() string {
	return "mock_transformer"
}

func (m *MockTransformer) Transform(data map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(data)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockValidator for testing
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Name() string {
	return "mock_validator"
}

func (m *MockValidator) Validate(data map[string]interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

// MockExecutionContext for testing
type MockExecutionContext struct {
	data map[string]interface{}
}

func NewMockExecutionContext() *MockExecutionContext {
	return &MockExecutionContext{
		data: make(map[string]interface{}),
	}
}

func (m *MockExecutionContext) Get(key string) (interface{}, bool) {
	val, exists := m.data[key]
	return val, exists
}

func (m *MockExecutionContext) Set(key string, value interface{}) {
	m.data[key] = value
}

func (m *MockExecutionContext) Delete(key string) {
	delete(m.data, key)
}

func (m *MockExecutionContext) Has(key string) bool {
	_, exists := m.data[key]
	return exists
}

func (m *MockExecutionContext) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *MockExecutionContext) GetString(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("value is not a string")
	}
	return "", fmt.Errorf("key not found")
}

func (m *MockExecutionContext) GetInt(key string) (int, error) {
	if val, ok := m.data[key]; ok {
		if i, ok := val.(int); ok {
			return i, nil
		}
		return 0, fmt.Errorf("value is not an int")
	}
	return 0, fmt.Errorf("key not found")
}

func (m *MockExecutionContext) GetBool(key string) (bool, error) {
	if val, ok := m.data[key]; ok {
		if b, ok := val.(bool); ok {
			return b, nil
		}
		return false, fmt.Errorf("value is not a bool")
	}
	return false, fmt.Errorf("key not found")
}

func (m *MockExecutionContext) GetMap(key string) (map[string]interface{}, error) {
	if val, ok := m.data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m, nil
		}
		return nil, fmt.Errorf("value is not a map")
	}
	return nil, fmt.Errorf("key not found")
}

func (m *MockExecutionContext) GetLogger() interface{} {
	return nil
}

func (m *MockExecutionContext) GetMetrics() interface{} {
	return nil
}

func (m *MockExecutionContext) Context() context.Context {
	return context.Background()
}

func (m *MockExecutionContext) WithTimeout(timeout time.Duration) (interfaces.ExecutionContext, context.CancelFunc) {
	_, cancel := context.WithTimeout(context.Background(), timeout)
	newMock := &MockExecutionContext{data: make(map[string]interface{})}
	for k, v := range m.data {
		newMock.data[k] = v
	}
	return newMock, cancel
}

func (m *MockExecutionContext) WithCancel() (interfaces.ExecutionContext, context.CancelFunc) {
	_, cancel := context.WithCancel(context.Background())
	newMock := &MockExecutionContext{data: make(map[string]interface{})}
	for k, v := range m.data {
		newMock.data[k] = v
	}
	return newMock, cancel
}

func (m *MockExecutionContext) WithValue(key, value interface{}) interfaces.ExecutionContext {
	newMock := &MockExecutionContext{data: make(map[string]interface{})}
	for k, v := range m.data {
		newMock.data[k] = v
	}
	if strKey, ok := key.(string); ok {
		newMock.data[strKey] = value
	}
	return newMock
}

func (m *MockExecutionContext) IsTimedOut() bool {
	return false
}

func (m *MockExecutionContext) IsCancelled() bool {
	return false
}

func (m *MockExecutionContext) GetStartTime() time.Time {
	return time.Now()
}

func (m *MockExecutionContext) GetElapsedTime() time.Duration {
	return time.Second
}

func (m *MockExecutionContext) Clone() interfaces.ExecutionContext {
	newMock := &MockExecutionContext{data: make(map[string]interface{})}
	for k, v := range m.data {
		newMock.data[k] = v
	}
	return newMock
}

func (m *MockExecutionContext) FlowName() string {
	return "test-flow"
}

func (m *MockExecutionContext) ExecutionID() string {
	return "test-exec-id"
}

func (m *MockExecutionContext) StartTime() time.Time {
	return time.Now()
}

func (m *MockExecutionContext) Duration() time.Duration {
	return time.Second
}

func (m *MockExecutionContext) Logger() *zap.Logger {
	return zap.NewNop()
}

func TestNewHTTPStep(t *testing.T) {
	step := NewHTTPStep("GET", "https://api.example.com/users")

	assert.Equal(t, "GET", step.method)
	assert.Equal(t, "https://api.example.com/users", step.url)
	assert.NotNil(t, step.headers)
	assert.NotNil(t, step.queryParams)
	assert.Equal(t, "json", step.bodyType)
	assert.Equal(t, 30*time.Second, step.responseTimeout)
	assert.Equal(t, []int{200, 201, 202, 204}, step.expectedStatus)
	assert.Equal(t, "API-Orchestration-Framework/1.0", step.userAgent)
	assert.Contains(t, step.Name(), "get_")
	assert.Contains(t, step.Description(), "GET request to")
}

func TestHTTPMethodConstructors(t *testing.T) {
	url := "https://api.example.com/test"

	tests := []struct {
		name     string
		step     *HTTPStep
		expected string
	}{
		{"GET", GET(url), "GET"},
		{"POST", POST(url), "POST"},
		{"PUT", PUT(url), "PUT"},
		{"DELETE", DELETE(url), "DELETE"},
		{"PATCH", PATCH(url), "PATCH"},
		{"HEAD", HEAD(url), "HEAD"},
		{"OPTIONS", OPTIONS(url), "OPTIONS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.step.method)
			assert.Equal(t, url, tt.step.url)
		})
	}
}

func TestHTTPStepConfiguration(t *testing.T) {
	step := NewHTTPStep("POST", "https://api.example.com/users")

	// Test header configuration
	step.WithHeader("Authorization", "Bearer token123")
	step.WithHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	})

	assert.Equal(t, "Bearer token123", step.headers["Authorization"])
	assert.Equal(t, "application/json", step.headers["Content-Type"])
	assert.Equal(t, "application/json", step.headers["Accept"])

	// Test query parameters
	step.WithQueryParam("page", "1")
	step.WithQueryParams(map[string]string{
		"limit":  "10",
		"filter": "active",
	})

	assert.Equal(t, "1", step.queryParams["page"])
	assert.Equal(t, "10", step.queryParams["limit"])
	assert.Equal(t, "active", step.queryParams["filter"])

	// Test authentication
	step.WithBasicAuth("user", "pass")
	assert.NotNil(t, step.basicAuth)
	assert.Equal(t, "user", step.basicAuth.Username)
	assert.Equal(t, "pass", step.basicAuth.Password)

	step.WithBearerToken("token456")
	assert.Equal(t, "token456", step.bearerToken)

	// Test other configurations
	step.WithUserAgent("Custom-Agent/1.0")
	assert.Equal(t, "Custom-Agent/1.0", step.userAgent)

	step.WithTimeout(60 * time.Second)
	assert.Equal(t, 60*time.Second, step.responseTimeout)

	step.WithExpectedStatus(200, 201, 400)
	assert.Equal(t, []int{200, 201, 400}, step.expectedStatus)

	step.SaveAs("response_data")
	assert.Equal(t, "response_data", step.saveAs)

	// Test cookie
	cookie := &http.Cookie{Name: "session", Value: "abc123"}
	step.WithCookie(cookie)
	assert.Len(t, step.cookies, 1)
	assert.Equal(t, "session", step.cookies[0].Name)
	assert.Equal(t, "abc123", step.cookies[0].Value)
}

func TestHTTPStepBodyConfiguration(t *testing.T) {
	step := NewHTTPStep("POST", "https://api.example.com/users")

	// Test JSON body
	jsonData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	step.WithJSONBody(jsonData)
	assert.Equal(t, jsonData, step.body)
	assert.Equal(t, "json", step.bodyType)
	assert.Equal(t, "application/json", step.headers["Content-Type"])

	// Test form body
	formData := map[string]string{
		"username": "johndoe",
		"password": "secret",
	}
	step.WithFormBody(formData)
	assert.Equal(t, formData, step.body)
	assert.Equal(t, "form", step.bodyType)
	assert.Equal(t, "application/x-www-form-urlencoded", step.headers["Content-Type"])

	// Test raw body
	rawData := []byte("raw content")
	step.WithRawBody(rawData, "text/plain")
	assert.Equal(t, rawData, step.body)
	assert.Equal(t, "raw", step.bodyType)
	assert.Equal(t, "text/plain", step.headers["Content-Type"])
}

func TestHTTPStepClientConfiguration(t *testing.T) {
	step := NewHTTPStep("GET", "https://api.example.com/test")

	// Test custom client
	mockClient := &MockHTTPClient{}
	step.WithClient(mockClient)
	assert.Equal(t, mockClient, step.client)

	// Test client config
	config := &ClientConfig{
		BaseTimeout: 45 * time.Second,
		MaxRetries:  5,
	}
	step.WithClientConfig(config)
	assert.Equal(t, config, step.clientConfig)

	// Test transformer
	mockTransformer := &MockTransformer{}
	step.WithTransformer(mockTransformer)
	assert.Equal(t, mockTransformer, step.transformer)

	// Test validator
	mockValidator := &MockValidator{}
	step.WithValidator(mockValidator)
	assert.Equal(t, mockValidator, step.validator)
}

func TestHTTPStepRun_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "Custom-Agent/1.0", r.Header.Get("User-Agent"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   1,
			"name": "Test User",
		})
	}))
	defer server.Close()

	step := GET(server.URL+"/test").
		WithHeader("Authorization", "Bearer token123").
		WithUserAgent("Custom-Agent/1.0").
		SaveAs("user_data")

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)

	// Check saved response
	userData, exists := ctx.Get("user_data")
	assert.True(t, exists)
	assert.NotNil(t, userData)

	// Check response data structure
	responseMap, ok := userData.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, responseMap, "body")
	assert.Contains(t, responseMap, "headers")
	assert.Contains(t, responseMap, "status_code")
}

func TestHTTPStepRun_WithJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read and verify body
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var requestData map[string]interface{}
		err = json.Unmarshal(body, &requestData)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", requestData["name"])
		assert.Equal(t, "john@example.com", requestData["email"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      123,
			"message": "User created",
		})
	}))
	defer server.Close()

	step := POST(server.URL + "/users").
		WithJSONBody(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}).
		SaveAs("create_response")

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)

	response, exists := ctx.Get("create_response")
	assert.True(t, exists)
	assert.NotNil(t, response)
}

func TestHTTPStepRun_WithFormBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		err := r.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, "johndoe", r.FormValue("username"))
		assert.Equal(t, "secret", r.FormValue("password"))

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Login successful")
	}))
	defer server.Close()

	step := POST(server.URL + "/login").
		WithFormBody(map[string]string{
			"username": "johndoe",
			"password": "secret",
		})

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestHTTPStepRun_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Equal(t, "active", r.URL.Query().Get("status"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "name": "User 1"},
			{"id": 2, "name": "User 2"},
		})
	}))
	defer server.Close()

	step := GET(server.URL+"/users").
		WithQueryParam("page", "1").
		WithQueryParams(map[string]string{
			"limit":  "10",
			"status": "active",
		})

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestHTTPStepRun_WithBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "testuser", username)
		assert.Equal(t, "testpass", password)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Authenticated")
	}))
	defer server.Close()

	step := GET(server.URL+"/protected").
		WithBasicAuth("testuser", "testpass")

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestHTTPStepRun_WithCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		assert.NoError(t, err)
		assert.Equal(t, "abc123", cookie.Value)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Cookie received")
	}))
	defer server.Close()

	step := GET(server.URL + "/session").
		WithCookie(&http.Cookie{
			Name:  "session",
			Value: "abc123",
		})

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestHTTPStepRun_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	// Use a simple client config without fallback to get the actual status code
	config := &ClientConfig{
		BaseTimeout:         10 * time.Second,
		EnableFallback:      false,
		MaxRetries:          1,
		FailureThreshold:    3,
		SuccessThreshold:    2,
		CircuitBreakerDelay: 1 * time.Second,
		RequestTimeout:      5 * time.Second,
	}
	step := GET(server.URL+"/error").
		WithExpectedStatus(200, 201).
		WithClientConfig(config)

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestHTTPStepRun_WithTransformer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":   123,
			"user_name": "John Doe",
		})
	}))
	defer server.Close()

	mockTransformer := &MockTransformer{}
	mockTransformer.On("Transform", mock.Anything).Return(map[string]interface{}{
		"body": map[string]interface{}{
			"id":   123,
			"name": "John Doe",
		},
		"status_code":  200,
		"headers":      make(http.Header),
		"content_type": "application/json",
	}, nil)

	step := GET(server.URL + "/user").
		WithTransformer(mockTransformer).
		SaveAs("transformed_user")

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)
	mockTransformer.AssertExpectations(t)

	transformedData, exists := ctx.Get("transformed_user")
	assert.True(t, exists)

	responseMap, ok := transformedData.(map[string]interface{})
	assert.True(t, ok)
	bodyData, ok := responseMap["body"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 123, bodyData["id"])
	assert.Equal(t, "John Doe", bodyData["name"])
}

func TestHTTPStepRun_WithValidator(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    123,
			"email": "john@example.com",
		})
	}))
	defer server.Close()

	mockValidator := &MockValidator{}
	mockValidator.On("Validate", mock.Anything).Return(nil)

	step := GET(server.URL + "/user").
		WithValidator(mockValidator)

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)
	mockValidator.AssertExpectations(t)
}

func TestHTTPStepRun_ValidatorError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"invalid": "data",
		})
	}))
	defer server.Close()

	mockValidator := &MockValidator{}
	mockValidator.On("Validate", mock.Anything).Return(fmt.Errorf("validation failed"))

	step := GET(server.URL + "/user").
		WithValidator(mockValidator)

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	mockValidator.AssertExpectations(t)
}

func TestHTTPStepRun_ContextInterpolation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/123", r.URL.Path)
		assert.Equal(t, "Bearer token456", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   123,
			"name": "User 123",
		})
	}))
	defer server.Close()

	step := GET(server.URL+"/users/${user_id}").
		WithHeader("Authorization", "Bearer ${auth_token}")

	ctx := NewMockExecutionContext()
	ctx.Set("user_id", "123")
	ctx.Set("auth_token", "token456")

	err := step.Run(ctx)
	assert.NoError(t, err)
}

func TestHTTPStepRun_NetworkError(t *testing.T) {
	step := GET("http://nonexistent.example.com/test")

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.Error(t, err)
}

func TestHTTPStepRun_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &ClientConfig{
		BaseTimeout:         50 * time.Millisecond,
		RequestTimeout:      50 * time.Millisecond,
		EnableFallback:      false,
		MaxRetries:          1,
		FailureThreshold:    3,
		SuccessThreshold:    2,
		CircuitBreakerDelay: 1 * time.Second,
	}
	step := GET(server.URL + "/slow").
		WithClientConfig(config)

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.Error(t, err)
}

func TestNewJSONAPIStep(t *testing.T) {
	step := NewJSONAPIStep("POST", "https://api.example.com/data")

	assert.Equal(t, "POST", step.method)
	assert.Equal(t, "https://api.example.com/data", step.url)
	assert.Equal(t, "application/json", step.headers["Content-Type"])
	assert.Equal(t, "application/json", step.headers["Accept"])
}

func TestNewMobileAPIStep(t *testing.T) {
	fields := []string{"id", "name", "email"}
	step := NewMobileAPIStep("GET", "https://api.example.com/users", fields)

	assert.Equal(t, "GET", step.method)
	assert.Equal(t, "https://api.example.com/users", step.url)
	assert.Equal(t, "application/json", step.headers["Accept"])
	assert.Equal(t, "mobile-app/1.0", step.userAgent)
	assert.Equal(t, "id,name,email", step.queryParams["fields"])
}

func TestNewAuthenticatedStep(t *testing.T) {
	step := NewAuthenticatedStep("GET", "https://api.example.com/profile", "access_token")

	assert.Equal(t, "GET", step.method)
	assert.Equal(t, "https://api.example.com/profile", step.url)

	// The step should be configured to use the token from context
	ctx := NewMockExecutionContext()
	ctx.Set("access_token", "test_token_123")

	// We can't easily test the interpolation without running the step,
	// but we can verify the step was created correctly
	assert.NotNil(t, step)
	assert.Contains(t, step.Name(), "get_")
}

func TestHTTPStepRun_RawBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "raw text content", string(body))

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Raw body received")
	}))
	defer server.Close()

	step := POST(server.URL+"/raw").
		WithRawBody([]byte("raw text content"), "text/plain")

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)
}

func TestHTTPStepRun_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	step := DELETE(server.URL + "/resource").
		SaveAs("delete_response")

	ctx := NewMockExecutionContext()
	err := step.Run(ctx)

	assert.NoError(t, err)

	response, exists := ctx.Get("delete_response")
	assert.True(t, exists)

	responseMap, ok := response.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 204, responseMap["status_code"])
}
