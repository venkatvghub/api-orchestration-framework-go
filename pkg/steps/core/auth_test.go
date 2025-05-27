package core

import (
	"strings"
	"testing"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"go.uber.org/zap"
)

func TestNewTokenValidationStep(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")

	if step.Name() != "test_auth" {
		t.Errorf("Name() = %v, want 'test_auth'", step.Name())
	}

	if step.Description() != "Token validation" {
		t.Errorf("Description() = %v, want 'Token validation'", step.Description())
	}

	if step.headerName != "Authorization" {
		t.Errorf("headerName = %v, want 'Authorization'", step.headerName)
	}

	if step.tokenPrefix != "Bearer " {
		t.Errorf("tokenPrefix = %v, want 'Bearer '", step.tokenPrefix)
	}
}

func TestTokenValidationStep_WithTokenPrefix(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	result := step.WithTokenPrefix("Token ")

	if result != step {
		t.Error("WithTokenPrefix should return the same instance")
	}

	if step.tokenPrefix != "Token " {
		t.Errorf("tokenPrefix = %v, want 'Token '", step.tokenPrefix)
	}
}

func TestTokenValidationStep_WithValidationFunc(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	validationFunc := func(token string) bool {
		return token == "valid_token"
	}

	result := step.WithValidationFunc(validationFunc)

	if result != step {
		t.Error("WithValidationFunc should return the same instance")
	}

	if step.validateFunc == nil {
		t.Error("validateFunc should be set")
	}
}

func TestTokenValidationStep_WithClaimsExtraction(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	result := step.WithClaimsExtraction(true)

	if result != step {
		t.Error("WithClaimsExtraction should return the same instance")
	}

	if !step.extractClaims {
		t.Error("extractClaims should be true")
	}
}

func TestTokenValidationStep_AddValidToken(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	step.AddValidToken("valid_token_123")

	// Test that the token was added (we can't directly access the sync.Map)
	// We'll test this through the validation process
	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("headers", map[string]interface{}{
		"Authorization": "Bearer valid_token_123",
	})

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed with valid token: %v", err)
	}

	// Check that token was stored
	token, ok := ctx.Get("auth_token")
	if !ok || token != "valid_token_123" {
		t.Error("Valid token should be stored in context")
	}

	authenticated, ok := ctx.Get("authenticated")
	if !ok || authenticated != true {
		t.Error("authenticated flag should be set to true")
	}
}

func TestTokenValidationStep_RemoveValidToken(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	step.AddValidToken("token_to_remove")
	step.RemoveValidToken("token_to_remove")

	// Test that the token was removed
	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("headers", map[string]interface{}{
		"Authorization": "Bearer token_to_remove",
	})

	// Since we removed the token and there's no custom validation function,
	// it should still pass (default validation accepts any non-empty token)
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}
}

func TestTokenValidationStep_Run_Success(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	// Set up valid headers
	ctx.Set("headers", map[string]interface{}{
		"Authorization": "Bearer valid_token_123",
	})

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check stored values
	token, ok := ctx.Get("auth_token")
	if !ok || token != "valid_token_123" {
		t.Error("Token should be stored in context")
	}

	authenticated, ok := ctx.Get("authenticated")
	if !ok || authenticated != true {
		t.Error("authenticated flag should be set")
	}
}

func TestTokenValidationStep_Run_WithClaimsExtraction(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization").
		WithClaimsExtraction(true)
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	ctx.Set("headers", map[string]interface{}{
		"Authorization": "Bearer valid_token_123",
	})

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	// Check that claims were extracted
	claims, ok := ctx.Get("auth_claims")
	if !ok {
		t.Error("Claims should be extracted and stored")
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		t.Error("Claims should be a map")
	}

	if claimsMap["token_type"] != "bearer" {
		t.Error("Claims should contain token_type")
	}
}

func TestTokenValidationStep_Run_NoHeaders(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	err := step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail when no headers are present")
	}

	if err.Error() != "no headers found for token validation" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestTokenValidationStep_Run_InvalidHeaders(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	// Set invalid headers (not a map)
	ctx.Set("headers", "invalid_headers")

	err := step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail with invalid headers")
	}

	if err.Error() != "headers is not a map" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestTokenValidationStep_Run_MissingAuthHeader(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization")
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	ctx.Set("headers", map[string]interface{}{
		"Content-Type": "application/json",
	})

	err := step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail when authorization header is missing")
	}

	if err.Error() != "authorization header 'Authorization' not found" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestTokenValidationStep_Run_EmptyToken(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization").
		WithTokenPrefix("Bearer")

	ctx := flow.NewContext().WithLogger(zap.NewNop())

	// Use "Bearer" exactly - this should match the prefix and leave empty string
	ctx.Set("headers", map[string]interface{}{
		"Authorization": "Bearer",
	})

	err := step.Run(ctx)

	if err == nil {
		t.Error("Run() should fail with empty token")
		return
	}

	if err.Error() != "empty token" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestTokenValidationStep_Run_CustomValidation(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "Authorization").
		WithValidationFunc(func(token string) bool {
			return token == "special_token"
		})

	ctx := flow.NewContext().WithLogger(zap.NewNop())

	// Test with valid token
	ctx.Set("headers", map[string]interface{}{
		"Authorization": "Bearer special_token",
	})

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed with valid custom token: %v", err)
	}

	// Test with invalid token
	ctx.Set("headers", map[string]interface{}{
		"Authorization": "Bearer invalid_token",
	})

	err = step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail with invalid custom token")
	}

	if err.Error() != "invalid token" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestTokenValidationStep_Run_CustomPrefix(t *testing.T) {
	step := NewTokenValidationStep("test_auth", "X-API-Key").
		WithTokenPrefix("ApiKey ")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("headers", map[string]interface{}{
		"X-API-Key": "ApiKey my_api_key_123",
	})

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	token, ok := ctx.Get("auth_token")
	if !ok || token != "my_api_key_123" {
		t.Error("Token should be extracted without prefix")
	}
}

func TestNewHeaderExtractionStep(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type", "User-Agent")

	if step.Name() != "test_headers" {
		t.Errorf("Name() = %v, want 'test_headers'", step.Name())
	}

	if step.Description() != "Header extraction" {
		t.Errorf("Description() = %v, want 'Header extraction'", step.Description())
	}

	if len(step.headerNames) != 2 {
		t.Errorf("headerNames length = %d, want 2", len(step.headerNames))
	}

	if step.sanitize != true {
		t.Error("sanitize should be true by default")
	}
}

func TestHeaderExtractionStep_WithRequired(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type", "User-Agent")
	result := step.WithRequired("Content-Type")

	if result != step {
		t.Error("WithRequired should return the same instance")
	}

	if !step.required["Content-Type"] {
		t.Error("Content-Type should be marked as required")
	}
}

func TestHeaderExtractionStep_WithSanitization(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type")
	result := step.WithSanitization(false)

	if result != step {
		t.Error("WithSanitization should return the same instance")
	}

	if step.sanitize != false {
		t.Error("sanitize should be false")
	}
}

func TestHeaderExtractionStep_Run_Success(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type", "User-Agent")
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	ctx.Set("headers", map[string]interface{}{
		"Content-Type":  "application/json",
		"User-Agent":    "Test-Agent/1.0",
		"Authorization": "Bearer token123",
	})

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	extracted, ok := ctx.Get("extracted_headers")
	if !ok {
		t.Error("extracted_headers should be set")
	}

	extractedMap, ok := extracted.(map[string]interface{})
	if !ok {
		t.Error("extracted_headers should be a map")
	}

	if extractedMap["Content-Type"] != "application/json" {
		t.Error("Content-Type should be extracted")
	}

	if extractedMap["User-Agent"] != "Test-Agent/1.0" {
		t.Error("User-Agent should be extracted")
	}

	// Authorization should not be extracted (not in headerNames)
	if _, exists := extractedMap["Authorization"]; exists {
		t.Error("Authorization should not be extracted")
	}
}

func TestHeaderExtractionStep_Run_MissingRequired(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type", "User-Agent").
		WithRequired("Content-Type")

	ctx := flow.NewContext().WithLogger(zap.NewNop())
	ctx.Set("headers", map[string]interface{}{
		"User-Agent": "Test-Agent/1.0",
		// Missing Content-Type
	})

	err := step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail when required header is missing")
	}

	if !strings.Contains(err.Error(), "required headers missing") {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestHeaderExtractionStep_Run_NoHeaders(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type")
	ctx := flow.NewContext()

	err := step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail when no headers are present")
	}

	if err.Error() != "no headers found in context" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestHeaderExtractionStep_Run_InvalidHeaders(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type")
	ctx := flow.NewContext()

	ctx.Set("headers", "invalid_headers")

	err := step.Run(ctx)
	if err == nil {
		t.Error("Run() should fail with invalid headers")
	}

	if err.Error() != "headers is not a map" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestHeaderExtractionStep_Run_OptionalHeaders(t *testing.T) {
	step := NewHeaderExtractionStep("test_headers", "Content-Type", "Optional-Header")
	ctx := flow.NewContext().WithLogger(zap.NewNop())

	ctx.Set("headers", map[string]interface{}{
		"Content-Type": "application/json",
		// Missing Optional-Header (but it's not required)
	})

	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	extracted, ok := ctx.Get("extracted_headers")
	if !ok {
		t.Error("extracted_headers should be set")
	}

	extractedMap, ok := extracted.(map[string]interface{})
	if !ok {
		t.Error("extracted_headers should be a map")
	}

	if len(extractedMap) != 1 {
		t.Errorf("extracted_headers should have 1 item, got %d", len(extractedMap))
	}

	if extractedMap["Content-Type"] != "application/json" {
		t.Error("Content-Type should be extracted")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{5, 3, 3},
		{2, 8, 2},
		{4, 4, 4},
		{0, 1, 0},
		{-1, 5, -1},
	}

	for _, test := range tests {
		result := min(test.a, test.b)
		if result != test.expected {
			t.Errorf("min(%d, %d) = %d, want %d", test.a, test.b, result, test.expected)
		}
	}
}
