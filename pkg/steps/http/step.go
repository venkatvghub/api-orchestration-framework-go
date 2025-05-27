package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/metrics"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
	"github.com/venkatvghub/api-orchestration-framework/pkg/validators"
)

// HTTPStep represents an enhanced HTTP request step with comprehensive features
type HTTPStep struct {
	*base.BaseStep
	method          string
	url             string
	headers         map[string]string
	body            interface{}
	bodyType        string // json, form, raw
	saveAs          string
	client          HTTPClient
	clientConfig    *ClientConfig
	transformer     transformers.Transformer
	validator       validators.Validator
	responseTimeout time.Duration
	userAgent       string
	basicAuth       *BasicAuth
	bearerToken     string
	queryParams     map[string]string
	cookies         []*http.Cookie
	expectedStatus  []int
}

// BasicAuth holds basic authentication credentials
type BasicAuth struct {
	Username string
	Password string
}

// NewHTTPStep creates a new enhanced HTTP step
func NewHTTPStep(method, url string) *HTTPStep {
	stepName := fmt.Sprintf("%s_%s", strings.ToLower(method), utils.SanitizeURL(url))

	return &HTTPStep{
		BaseStep:        base.NewBaseStep(stepName, fmt.Sprintf("%s request to %s", method, url)),
		method:          strings.ToUpper(method),
		url:             url,
		headers:         make(map[string]string),
		queryParams:     make(map[string]string),
		client:          DefaultClient,
		clientConfig:    DefaultClientConfig(),
		bodyType:        "json",
		responseTimeout: 30 * time.Second,
		expectedStatus:  []int{200, 201, 202, 204},
		userAgent:       "API-Orchestration-Framework/1.0",
	}
}

// HTTP method constructors
func GET(url string) *HTTPStep {
	return NewHTTPStep("GET", url)
}

func POST(url string) *HTTPStep {
	return NewHTTPStep("POST", url)
}

func PUT(url string) *HTTPStep {
	return NewHTTPStep("PUT", url)
}

func DELETE(url string) *HTTPStep {
	return NewHTTPStep("DELETE", url)
}

func PATCH(url string) *HTTPStep {
	return NewHTTPStep("PATCH", url)
}

func HEAD(url string) *HTTPStep {
	return NewHTTPStep("HEAD", url)
}

func OPTIONS(url string) *HTTPStep {
	return NewHTTPStep("OPTIONS", url)
}

// Configuration methods with fluent API

// WithHeader adds a single header
func (h *HTTPStep) WithHeader(key, value string) *HTTPStep {
	h.headers[key] = value
	return h
}

// WithHeaders adds multiple headers
func (h *HTTPStep) WithHeaders(headers map[string]string) *HTTPStep {
	for k, v := range headers {
		h.headers[k] = v
	}
	return h
}

// WithJSONBody sets JSON body content
func (h *HTTPStep) WithJSONBody(body interface{}) *HTTPStep {
	h.body = body
	h.bodyType = "json"
	h.headers["Content-Type"] = "application/json"
	return h
}

// WithFormBody sets form-encoded body content
func (h *HTTPStep) WithFormBody(data map[string]string) *HTTPStep {
	h.body = data
	h.bodyType = "form"
	h.headers["Content-Type"] = "application/x-www-form-urlencoded"
	return h
}

// WithRawBody sets raw body content
func (h *HTTPStep) WithRawBody(body []byte, contentType string) *HTTPStep {
	h.body = body
	h.bodyType = "raw"
	if contentType != "" {
		h.headers["Content-Type"] = contentType
	}
	return h
}

// WithQueryParam adds a query parameter
func (h *HTTPStep) WithQueryParam(key, value string) *HTTPStep {
	h.queryParams[key] = value
	return h
}

// WithQueryParams adds multiple query parameters
func (h *HTTPStep) WithQueryParams(params map[string]string) *HTTPStep {
	for k, v := range params {
		h.queryParams[k] = v
	}
	return h
}

// WithBasicAuth sets basic authentication
func (h *HTTPStep) WithBasicAuth(username, password string) *HTTPStep {
	h.basicAuth = &BasicAuth{
		Username: username,
		Password: password,
	}
	return h
}

// WithBearerToken sets bearer token authentication
func (h *HTTPStep) WithBearerToken(token string) *HTTPStep {
	h.bearerToken = token
	return h
}

// WithCookie adds a cookie
func (h *HTTPStep) WithCookie(cookie *http.Cookie) *HTTPStep {
	h.cookies = append(h.cookies, cookie)
	return h
}

// WithUserAgent sets the user agent
func (h *HTTPStep) WithUserAgent(userAgent string) *HTTPStep {
	h.userAgent = userAgent
	return h
}

// WithTimeout sets request timeout
func (h *HTTPStep) WithTimeout(timeout time.Duration) *HTTPStep {
	h.responseTimeout = timeout
	return h
}

// WithExpectedStatus sets expected HTTP status codes
func (h *HTTPStep) WithExpectedStatus(codes ...int) *HTTPStep {
	h.expectedStatus = codes
	return h
}

// SaveAs sets the context key to save response data
func (h *HTTPStep) SaveAs(key string) *HTTPStep {
	h.saveAs = key
	return h
}

// WithClient sets a custom HTTP client
func (h *HTTPStep) WithClient(client HTTPClient) *HTTPStep {
	h.client = client
	return h
}

// WithClientConfig sets HTTP client configuration
func (h *HTTPStep) WithClientConfig(config *ClientConfig) *HTTPStep {
	h.clientConfig = config
	h.client = NewResilientHTTPClient(config)
	return h
}

// WithTransformer sets response transformer
func (h *HTTPStep) WithTransformer(transformer transformers.Transformer) *HTTPStep {
	h.transformer = transformer
	return h
}

// WithValidator sets response validator
func (h *HTTPStep) WithValidator(validator validators.Validator) *HTTPStep {
	h.validator = validator
	return h
}

// Run executes the HTTP step
func (h *HTTPStep) Run(ctx interfaces.ExecutionContext) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		metrics.RecordStepExecution(h.Name(), duration, true)
	}()

	// Get buffer from pool for request body
	bodyBuffer := utils.GetBuffer()
	defer utils.PutBuffer(bodyBuffer)

	// Prepare request
	req, err := h.prepareRequest(ctx, bodyBuffer)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %w", err)
	}

	// Execute request with resilient client
	resp, err := h.client.Do(req)
	if err != nil {
		metrics.RecordHTTPRequest(h.method, h.url, 0, time.Since(start))
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Record HTTP metrics
	metrics.RecordHTTPRequest(h.method, h.url, resp.StatusCode, time.Since(start))

	// Process response
	return h.processResponse(ctx, resp)
}

// Helper methods

// interpolateURL interpolates variables in the URL
func (h *HTTPStep) interpolateURL(ctx interfaces.ExecutionContext) (string, error) {
	return utils.InterpolateString(h.url, ctx)
}

// interpolateHeaders interpolates variables in headers
func (h *HTTPStep) interpolateHeaders(ctx interfaces.ExecutionContext) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range h.headers {
		interpolatedValue, err := utils.InterpolateString(value, ctx)
		if err != nil {
			return nil, fmt.Errorf("header interpolation failed for %s: %w", key, err)
		}
		result[key] = interpolatedValue
	}
	return result, nil
}

// prepareRequest creates and configures the HTTP request
func (h *HTTPStep) prepareRequest(ctx interfaces.ExecutionContext, bodyBuffer *bytes.Buffer) (*http.Request, error) {
	// Interpolate URL with context variables
	finalURL, err := h.interpolateURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("URL interpolation failed: %w", err)
	}

	// Interpolate headers
	finalHeaders, err := h.interpolateHeaders(ctx)
	if err != nil {
		return nil, fmt.Errorf("header interpolation failed: %w", err)
	}

	// Prepare request body
	var bodyReader io.Reader
	if h.body != nil {
		if err := h.prepareBody(ctx, bodyBuffer); err != nil {
			return nil, fmt.Errorf("body preparation failed: %w", err)
		}
		bodyReader = bodyBuffer
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx.Context(), h.method, finalURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	// Set headers
	for key, value := range finalHeaders {
		req.Header.Set(key, value)
	}

	// Set query parameters
	if len(h.queryParams) > 0 {
		q := req.URL.Query()
		for key, value := range h.queryParams {
			interpolatedValue, err := utils.InterpolateString(value, ctx)
			if err != nil {
				return nil, fmt.Errorf("query param interpolation failed for %s: %w", key, err)
			}
			q.Set(key, interpolatedValue)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

// prepareBody prepares the request body using the buffer pool
func (h *HTTPStep) prepareBody(ctx interfaces.ExecutionContext, buffer *bytes.Buffer) error {
	switch body := h.body.(type) {
	case string:
		interpolated, err := utils.InterpolateString(body, ctx)
		if err != nil {
			return fmt.Errorf("body interpolation failed: %w", err)
		}
		buffer.WriteString(interpolated)
	case map[string]interface{}:
		// Interpolate map values
		interpolatedBody := make(map[string]interface{})
		for k, v := range body {
			if str, ok := v.(string); ok {
				interpolated, err := utils.InterpolateString(str, ctx)
				if err != nil {
					return fmt.Errorf("body field interpolation failed for %s: %w", k, err)
				}
				interpolatedBody[k] = interpolated
			} else {
				interpolatedBody[k] = v
			}
		}

		jsonData, err := json.Marshal(interpolatedBody)
		if err != nil {
			return fmt.Errorf("JSON marshaling failed: %w", err)
		}
		buffer.Write(jsonData)
	default:
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("JSON marshaling failed: %w", err)
		}
		buffer.Write(jsonData)
	}
	return nil
}

// processResponse processes the HTTP response
func (h *HTTPStep) processResponse(ctx interfaces.ExecutionContext, resp *http.Response) error {
	start := time.Now() // Define start time for this method

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response based on content type
	var responseData map[string]interface{}
	contentType := resp.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
			// If JSON parsing fails, store as string
			responseData = map[string]interface{}{
				"raw_response": string(bodyBytes),
			}
		}
	} else {
		// Store non-JSON responses as string
		responseData = map[string]interface{}{
			"raw_response": string(bodyBytes),
		}
	}

	// Apply transformer if configured
	if h.transformer != nil {
		transformedData, err := h.transformer.Transform(responseData)
		if err != nil {
			return fmt.Errorf("response transformation failed: %w", err)
		}
		responseData = transformedData
	}

	// Apply validator if configured
	if h.validator != nil {
		if err := h.validator.Validate(responseData); err != nil {
			return fmt.Errorf("response validation failed: %w", err)
		}
	}

	// Save response data to context
	if h.saveAs != "" {
		ctx.Set(h.saveAs, responseData)
	} else {
		// Default save location
		ctx.Set("http_response", responseData)
	}

	// Store HTTP metadata
	metadata := map[string]interface{}{
		"status_code":    resp.StatusCode,
		"headers":        resp.Header,
		"content_length": resp.ContentLength,
		"duration":       time.Since(start),
		"url":            h.url,
		"method":         h.method,
	}
	ctx.Set("http_metadata", metadata)

	ctx.Logger().Info("HTTP request completed successfully",
		zap.String("step", h.Name()),
		zap.String("method", h.method),
		zap.String("url", h.url),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(start)),
		zap.String("save_as", h.saveAs))

	return nil
}

// Helper functions for creating common HTTP steps with transformers and validators

// NewJSONAPIStep creates an HTTP step configured for JSON APIs
func NewJSONAPIStep(method, url string) *HTTPStep {
	return NewHTTPStep(method, url).
		WithHeader("Content-Type", "application/json").
		WithHeader("Accept", "application/json")
}

// NewMobileAPIStep creates an HTTP step optimized for mobile responses
func NewMobileAPIStep(method, url string, fields []string) *HTTPStep {
	return NewJSONAPIStep(method, url).
		WithTransformer(transformers.NewMobileTransformer(fields)).
		WithValidator(validators.NewRequiredFieldsValidator("status"))
}

// NewAuthenticatedStep creates an HTTP step with authentication
func NewAuthenticatedStep(method, url, tokenField string) *HTTPStep {
	return NewJSONAPIStep(method, url).
		WithBearerToken("${" + tokenField + "}")
}
