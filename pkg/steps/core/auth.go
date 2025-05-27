package core

import (
	"fmt"
	"strings"
	"sync"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
	"go.uber.org/zap"
)

// TokenValidationStep validates authentication tokens
type TokenValidationStep struct {
	*base.BaseStep
	headerName    string
	tokenPrefix   string
	validTokens   *sync.Map // Thread-safe token storage
	validateFunc  func(string) bool
	extractClaims bool
}

// NewTokenValidationStep creates a new token validation step
func NewTokenValidationStep(name, headerName string) *TokenValidationStep {
	return &TokenValidationStep{
		BaseStep:      base.NewBaseStep(name, "Token validation"),
		headerName:    headerName,
		tokenPrefix:   "Bearer ",
		validTokens:   &sync.Map{},
		extractClaims: false,
	}
}

// WithTokenPrefix sets the token prefix (default: "Bearer ")
func (tvs *TokenValidationStep) WithTokenPrefix(prefix string) *TokenValidationStep {
	tvs.tokenPrefix = prefix
	return tvs
}

// WithValidationFunc sets a custom token validation function
func (tvs *TokenValidationStep) WithValidationFunc(fn func(string) bool) *TokenValidationStep {
	tvs.validateFunc = fn
	return tvs
}

// WithClaimsExtraction enables JWT claims extraction
func (tvs *TokenValidationStep) WithClaimsExtraction(extract bool) *TokenValidationStep {
	tvs.extractClaims = extract
	return tvs
}

// AddValidToken adds a valid token to the whitelist
func (tvs *TokenValidationStep) AddValidToken(token string) {
	tvs.validTokens.Store(token, true)
}

// RemoveValidToken removes a token from the whitelist
func (tvs *TokenValidationStep) RemoveValidToken(token string) {
	tvs.validTokens.Delete(token)
}

func (tvs *TokenValidationStep) Run(ctx *flow.Context) error {
	// Extract token from headers
	headers, ok := ctx.Get("headers")
	if !ok {
		ctx.Logger().Warn("No headers found in context",
			zap.String("step", tvs.Name()))
		return fmt.Errorf("no headers found for token validation")
	}

	headerMap, ok := headers.(map[string]interface{})
	if !ok {
		return fmt.Errorf("headers is not a map")
	}

	authHeader, ok := headerMap[tvs.headerName]
	if !ok {
		ctx.Logger().Warn("Authorization header not found",
			zap.String("step", tvs.Name()),
			zap.String("header_name", tvs.headerName))
		return fmt.Errorf("authorization header '%s' not found", tvs.headerName)
	}

	authValue, ok := authHeader.(string)
	if !ok {
		return fmt.Errorf("authorization header is not a string")
	}

	// Extract token
	token := strings.TrimSpace(authValue)
	if tvs.tokenPrefix != "" && strings.HasPrefix(token, tvs.tokenPrefix) {
		token = strings.TrimSpace(token[len(tvs.tokenPrefix):])
	}

	if token == "" {
		ctx.Logger().Warn("Empty token found",
			zap.String("step", tvs.Name()))
		return fmt.Errorf("empty token")
	}

	// Validate token
	if !tvs.isValidToken(token) {
		ctx.Logger().Warn("Invalid token",
			zap.String("step", tvs.Name()),
			zap.String("token_prefix", token[:min(len(token), 10)]+"..."))
		return fmt.Errorf("invalid token")
	}

	// Store validated token in context
	ctx.Set("auth_token", token)
	ctx.Set("authenticated", true)

	// Extract claims if enabled
	if tvs.extractClaims {
		if claims := tvs.extractTokenClaims(token); claims != nil {
			ctx.Set("auth_claims", claims)
		}
	}

	ctx.Logger().Info("Token validation successful",
		zap.String("step", tvs.Name()),
		zap.Bool("claims_extracted", tvs.extractClaims))

	return nil
}

func (tvs *TokenValidationStep) isValidToken(token string) bool {
	// Check whitelist first
	if _, exists := tvs.validTokens.Load(token); exists {
		return true
	}

	// Use custom validation function if provided
	if tvs.validateFunc != nil {
		return tvs.validateFunc(token)
	}

	// Default: non-empty token is valid
	return token != ""
}

func (tvs *TokenValidationStep) extractTokenClaims(_ string) map[string]interface{} {
	// Placeholder for JWT claims extraction
	// In a real implementation, you would decode the JWT token here
	return map[string]interface{}{
		"token_type": "bearer",
		"validated":  true,
	}
}

// HeaderExtractionStep extracts specific headers from the request
type HeaderExtractionStep struct {
	*base.BaseStep
	headerNames []string
	required    map[string]bool
	sanitize    bool
}

// NewHeaderExtractionStep creates a new header extraction step
func NewHeaderExtractionStep(name string, headerNames ...string) *HeaderExtractionStep {
	return &HeaderExtractionStep{
		BaseStep:    base.NewBaseStep(name, "Header extraction"),
		headerNames: headerNames,
		required:    make(map[string]bool),
		sanitize:    true,
	}
}

// WithRequired marks specific headers as required
func (hes *HeaderExtractionStep) WithRequired(headerNames ...string) *HeaderExtractionStep {
	for _, name := range headerNames {
		hes.required[name] = true
	}
	return hes
}

// WithSanitization controls whether to sanitize sensitive headers
func (hes *HeaderExtractionStep) WithSanitization(sanitize bool) *HeaderExtractionStep {
	hes.sanitize = sanitize
	return hes
}

func (hes *HeaderExtractionStep) Run(ctx *flow.Context) error {
	headers, ok := ctx.Get("headers")
	if !ok {
		return fmt.Errorf("no headers found in context")
	}

	headerMap, ok := headers.(map[string]interface{})
	if !ok {
		return fmt.Errorf("headers is not a map")
	}

	extractedHeaders := make(map[string]interface{})
	missingRequired := make([]string, 0)

	for _, headerName := range hes.headerNames {
		if value, exists := headerMap[headerName]; exists {
			extractedHeaders[headerName] = value
		} else if hes.required[headerName] {
			missingRequired = append(missingRequired, headerName)
		}
	}

	if len(missingRequired) > 0 {
		ctx.Logger().Warn("Required headers missing",
			zap.String("step", hes.Name()),
			zap.Strings("missing_headers", missingRequired))
		return fmt.Errorf("required headers missing: %v", missingRequired)
	}

	// Sanitize if enabled
	if hes.sanitize {
		extractedHeaders = utils.SanitizeHeaders(extractedHeaders)
	}

	// Store extracted headers
	ctx.Set("extracted_headers", extractedHeaders)

	ctx.Logger().Info("Headers extracted successfully",
		zap.String("step", hes.Name()),
		zap.Int("header_count", len(extractedHeaders)))

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
