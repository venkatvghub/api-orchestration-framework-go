package di

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/interfaces"
	"github.com/venkatvghub/api-orchestration-framework/pkg/registry"
	httpsteps "github.com/venkatvghub/api-orchestration-framework/pkg/steps/http"
	"github.com/venkatvghub/api-orchestration-framework/pkg/transformers"
	"github.com/venkatvghub/api-orchestration-framework/pkg/validators"
)

// ExampleFlowService demonstrates how to use DI for flow orchestration
type ExampleFlowService struct {
	flowFactory *FlowFactory
	logger      *zap.Logger
	registry    *registry.StepRegistry
}

// NewExampleFlowService creates a service with injected dependencies
func NewExampleFlowService(
	flowFactory *FlowFactory,
	logger *zap.Logger,
	registry *registry.StepRegistry,
) *ExampleFlowService {
	return &ExampleFlowService{
		flowFactory: flowFactory,
		logger:      logger,
		registry:    registry,
	}
}

// CreateUserOnboardingFlow demonstrates creating a flow with DI
func (s *ExampleFlowService) CreateUserOnboardingFlow() *flow.Flow {
	return s.flowFactory.CreateFlow("user_onboarding").
		WithDescription("User onboarding flow with DI").
		WithTimeout(30*time.Second).
		StepFunc("start", func(ctx interfaces.ExecutionContext) error {
			s.logger.Info("Starting user onboarding", zap.String("flow", ctx.FlowName()))
			return nil
		}).
		StepFunc("validate_input", func(ctx interfaces.ExecutionContext) error {
			s.logger.Info("Validating user input", zap.String("flow", ctx.FlowName()))
			// Add validation logic here
			return nil
		}).
		StepFunc("create_user", func(ctx interfaces.ExecutionContext) error {
			s.logger.Info("Creating user", zap.String("execution_id", ctx.ExecutionID()))
			ctx.Set("user_id", "12345")
			return nil
		}).
		Step("welcome_api", httpsteps.POST("https://api.example.com/users/{{user_id}}/welcome").
			WithJSONBody(map[string]interface{}{
				"user_id": "{{user_id}}",
				"message": "Welcome to our platform!",
			}).
			SaveAs("welcome_response")).
		StepFunc("complete", func(ctx interfaces.ExecutionContext) error {
			s.logger.Info("User onboarding completed", zap.String("execution_id", ctx.ExecutionID()))
			return nil
		})
}

// ExecuteFlow demonstrates executing a flow with DI context
func (s *ExampleFlowService) ExecuteFlow(flowName string) error {
	ctx := s.flowFactory.CreateContext()

	// Create and execute flow
	userFlow := s.CreateUserOnboardingFlow()
	result, err := userFlow.Execute(ctx)
	if err != nil {
		s.logger.Error("Flow execution failed",
			zap.String("flow", flowName),
			zap.Error(err))
		return err
	}

	s.logger.Info("Flow execution completed",
		zap.String("flow", result.FlowName),
		zap.Duration("duration", result.Duration),
		zap.Bool("success", result.Success))

	return nil
}

// ExampleTransformerService demonstrates using transformers with DI
type ExampleTransformerService struct {
	fieldTransformer   transformers.Transformer
	mobileTransformer  transformers.Transformer
	flattenTransformer transformers.Transformer
	logger             *zap.Logger
}

// NewExampleTransformerService creates a transformer service with DI
func NewExampleTransformerService(
	fieldTransformer transformers.Transformer,
	mobileTransformer transformers.Transformer,
	flattenTransformer transformers.Transformer,
	logger *zap.Logger,
) *ExampleTransformerService {
	return &ExampleTransformerService{
		fieldTransformer:   fieldTransformer,
		mobileTransformer:  mobileTransformer,
		flattenTransformer: flattenTransformer,
		logger:             logger,
	}
}

// TransformData demonstrates using injected transformers
func (s *ExampleTransformerService) TransformData(data map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("Starting data transformation")

	// Apply field transformation
	result, err := s.fieldTransformer.Transform(data)
	if err != nil {
		s.logger.Error("Field transformation failed", zap.Error(err))
		return nil, err
	}

	// Apply mobile optimization
	result, err = s.mobileTransformer.Transform(result)
	if err != nil {
		s.logger.Error("Mobile transformation failed", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Data transformation completed")
	return result, nil
}

// ExampleValidatorService demonstrates using validators with DI
type ExampleValidatorService struct {
	requiredFieldsValidator validators.Validator
	emailValidator          validators.Validator
	logger                  *zap.Logger
}

// NewExampleValidatorService creates a validator service with DI
func NewExampleValidatorService(
	requiredFieldsValidator validators.Validator,
	emailValidator validators.Validator,
	logger *zap.Logger,
) *ExampleValidatorService {
	return &ExampleValidatorService{
		requiredFieldsValidator: requiredFieldsValidator,
		emailValidator:          emailValidator,
		logger:                  logger,
	}
}

// ValidateUserData demonstrates using injected validators
func (s *ExampleValidatorService) ValidateUserData(data map[string]interface{}) error {
	s.logger.Info("Starting data validation")

	// Validate required fields
	if err := s.requiredFieldsValidator.Validate(data); err != nil {
		s.logger.Error("Required fields validation failed", zap.Error(err))
		return err
	}

	// Validate email if present
	if _, hasEmail := data["email"]; hasEmail {
		if err := s.emailValidator.Validate(data); err != nil {
			s.logger.Error("Email validation failed", zap.Error(err))
			return err
		}
	}

	s.logger.Info("Data validation completed")
	return nil
}

// ExampleHTTPService demonstrates using HTTP client with DI
type ExampleHTTPService struct {
	httpClient httpsteps.HTTPClient
	logger     *zap.Logger
}

// NewExampleHTTPService creates an HTTP service with DI
func NewExampleHTTPService(
	httpClient httpsteps.HTTPClient,
	logger *zap.Logger,
) *ExampleHTTPService {
	return &ExampleHTTPService{
		httpClient: httpClient,
		logger:     logger,
	}
}

// MakeAPICall demonstrates using injected HTTP client
func (s *ExampleHTTPService) MakeAPICall(url string) error {
	s.logger.Info("Making API call", zap.String("url", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		return err
	}

	// Add headers
	req.Header.Set("User-Agent", "API-Orchestration-Framework/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("API call failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	s.logger.Info("API call completed",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode))

	return nil
}

// ExampleStepService demonstrates creating custom steps with DI
type ExampleStepService struct {
	registry *registry.StepRegistry
	logger   *zap.Logger
}

// NewExampleStepService creates a step service with DI
func NewExampleStepService(
	registry *registry.StepRegistry,
	logger *zap.Logger,
) *ExampleStepService {
	return &ExampleStepService{
		registry: registry,
		logger:   logger,
	}
}

// RegisterCustomSteps demonstrates registering custom steps
func (s *ExampleStepService) RegisterCustomSteps() {
	s.logger.Info("Registering custom steps")

	// Register a custom HTTP step
	httpStepInfo := &registry.StepInfo{
		Name:        "api_call",
		Description: "Custom API call step",
		Category:    "http",
		Version:     "1.0.0",
		ConfigSpec: map[string]interface{}{
			"url":    "string",
			"method": "string",
		},
		Factory: func(config map[string]interface{}) (interfaces.Step, error) {
			url, ok := config["url"].(string)
			if !ok {
				return nil, fmt.Errorf("url is required for api_call step")
			}
			method, ok := config["method"].(string)
			if !ok {
				method = "GET"
			}
			return httpsteps.NewHTTPStep(method, url), nil
		},
	}

	if err := s.registry.Register(httpStepInfo); err != nil {
		s.logger.Error("Failed to register api_call step", zap.Error(err))
	}

	s.logger.Info("Custom steps registered successfully")
}

// ExampleModule demonstrates how to create a module-specific DI configuration
func ExampleModule() fx.Option {
	return fx.Options(
		fx.Provide(
			NewExampleFlowService,
			NewExampleTransformerService,
			NewExampleValidatorService,
			NewExampleHTTPService,
			NewExampleStepService,
		),
		fx.Invoke(func(
			flowService *ExampleFlowService,
			transformerService *ExampleTransformerService,
			validatorService *ExampleValidatorService,
			httpService *ExampleHTTPService,
			stepService *ExampleStepService,
			logger *zap.Logger,
		) {
			logger.Info("Example services initialized with DI")

			// Register custom steps
			stepService.RegisterCustomSteps()
		}),
	)
}

// DemoApplication shows how to use all services together
func DemoApplication() fx.Option {
	return fx.Options(
		Module(),        // Core framework dependencies
		ExampleModule(), // Example services
		fx.Invoke(func(
			lc fx.Lifecycle,
			flowService *ExampleFlowService,
			transformerService *ExampleTransformerService,
			validatorService *ExampleValidatorService,
			httpService *ExampleHTTPService,
			logger *zap.Logger,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Starting demo application")

					// Execute example flow
					go func() {
						if err := flowService.ExecuteFlow("demo_flow"); err != nil {
							logger.Error("Demo flow failed", zap.Error(err))
						}
					}()

					// Test data transformation
					go func() {
						testData := map[string]interface{}{
							"id":    "123",
							"name":  "John Doe",
							"email": "john@example.com",
							"age":   30,
						}

						// Validate data
						if err := validatorService.ValidateUserData(testData); err != nil {
							logger.Error("Validation failed", zap.Error(err))
							return
						}

						// Transform data
						if _, err := transformerService.TransformData(testData); err != nil {
							logger.Error("Transformation failed", zap.Error(err))
							return
						}

						logger.Info("Demo data processing completed")
					}()

					// Test HTTP call
					go func() {
						if err := httpService.MakeAPICall("https://httpbin.org/get"); err != nil {
							logger.Error("HTTP call failed", zap.Error(err))
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Stopping demo application")
					return nil
				},
			})
		}),
	)
}
