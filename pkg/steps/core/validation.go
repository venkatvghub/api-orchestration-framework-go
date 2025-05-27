package core

import (
	"fmt"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/validators"
	"go.uber.org/zap"
)

// ValidationStep validates data using the validator system
type ValidationStep struct {
	*base.BaseStep
	validator     validators.Validator
	dataField     string
	continueOnErr bool
	storeResult   bool
	resultField   string
}

// NewValidationStep creates a new validation step
func NewValidationStep(name string, validator validators.Validator) *ValidationStep {
	return &ValidationStep{
		BaseStep:      base.NewBaseStep(name, "Data validation"),
		validator:     validator,
		dataField:     "", // Empty means validate entire context
		continueOnErr: false,
		storeResult:   true,
		resultField:   "validation_result",
	}
}

// WithDataField sets the specific field to validate
func (vs *ValidationStep) WithDataField(field string) *ValidationStep {
	vs.dataField = field
	return vs
}

// WithContinueOnError controls whether to continue execution on validation failure
func (vs *ValidationStep) WithContinueOnError(continueOnError bool) *ValidationStep {
	vs.continueOnErr = continueOnError
	return vs
}

// WithResultStorage controls whether to store validation results
func (vs *ValidationStep) WithResultStorage(store bool, field string) *ValidationStep {
	vs.storeResult = store
	if field != "" {
		vs.resultField = field
	}
	return vs
}

func (vs *ValidationStep) Run(ctx *flow.Context) error {
	// Get data to validate
	dataToValidate, err := vs.getDataToValidate(ctx)
	if err != nil {
		ctx.Logger().Error("Failed to get data for validation",
			zap.String("step", vs.Name()),
			zap.String("validator", vs.validator.Name()),
			zap.Error(err))
		return err
	}

	// Perform validation
	validationErr := vs.validator.Validate(dataToValidate)

	// Store validation result if requested
	if vs.storeResult {
		result := map[string]interface{}{
			"valid":     validationErr == nil,
			"validator": vs.validator.Name(),
		}

		if validationErr != nil {
			result["error"] = validationErr.Error()
			result["details"] = vs.extractValidationDetails(validationErr)
		}

		ctx.Set(vs.resultField, result)
	}

	// Log validation result
	if validationErr != nil {
		ctx.Logger().Warn("Validation failed",
			zap.String("step", vs.Name()),
			zap.String("validator", vs.validator.Name()),
			zap.String("data_field", vs.dataField),
			zap.Error(validationErr))

		if !vs.continueOnErr {
			return fmt.Errorf("validation failed: %w", validationErr)
		}
	} else {
		ctx.Logger().Info("Validation passed",
			zap.String("step", vs.Name()),
			zap.String("validator", vs.validator.Name()),
			zap.String("data_field", vs.dataField))
	}

	return nil
}

func (vs *ValidationStep) getDataToValidate(ctx *flow.Context) (map[string]interface{}, error) {
	if vs.dataField == "" {
		// Validate entire context
		result := make(map[string]interface{})
		for _, key := range ctx.Keys() {
			if value, ok := ctx.Get(key); ok {
				result[key] = value
			}
		}
		return result, nil
	}

	// Validate specific field
	value, exists := ctx.Get(vs.dataField)
	if !exists {
		return nil, fmt.Errorf("data field '%s' not found in context", vs.dataField)
	}

	// If the value is already a map, use it directly
	if mapValue, ok := value.(map[string]interface{}); ok {
		return mapValue, nil
	}

	// Otherwise, wrap it in a map
	return map[string]interface{}{
		vs.dataField: value,
	}, nil
}

func (vs *ValidationStep) extractValidationDetails(err error) interface{} {
	// Handle different types of validation errors
	switch e := err.(type) {
	case *validators.ValidationError:
		return map[string]interface{}{
			"field":     e.Field,
			"value":     e.Value,
			"message":   e.Message,
			"validator": e.Validator,
		}
	case *validators.MultiValidationError:
		var details []interface{}
		for _, subErr := range e.Errors {
			details = append(details, vs.extractValidationDetails(subErr))
		}
		return details
	default:
		return err.Error()
	}
}

// ValidationChainStep runs multiple validators in sequence
type ValidationChainStep struct {
	*base.BaseStep
	validators    []validators.Validator
	dataField     string
	stopOnFirst   bool
	continueOnErr bool
	storeResults  bool
	resultField   string
}

// NewValidationChainStep creates a new validation chain step
func NewValidationChainStep(name string, validators ...validators.Validator) *ValidationChainStep {
	return &ValidationChainStep{
		BaseStep:      base.NewBaseStep(name, "Validation chain"),
		validators:    validators,
		dataField:     "",
		stopOnFirst:   true,
		continueOnErr: false,
		storeResults:  true,
		resultField:   "validation_results",
	}
}

// WithDataField sets the specific field to validate
func (vcs *ValidationChainStep) WithDataField(field string) *ValidationChainStep {
	vcs.dataField = field
	return vcs
}

// WithStopOnFirst controls whether to stop on first validation failure
func (vcs *ValidationChainStep) WithStopOnFirst(stop bool) *ValidationChainStep {
	vcs.stopOnFirst = stop
	return vcs
}

// WithContinueOnError controls whether to continue execution on validation failure
func (vcs *ValidationChainStep) WithContinueOnError(continueOnError bool) *ValidationChainStep {
	vcs.continueOnErr = continueOnError
	return vcs
}

// WithResultStorage controls whether to store validation results
func (vcs *ValidationChainStep) WithResultStorage(store bool, field string) *ValidationChainStep {
	vcs.storeResults = store
	if field != "" {
		vcs.resultField = field
	}
	return vcs
}

func (vcs *ValidationChainStep) Run(ctx *flow.Context) error {
	// Get data to validate
	dataToValidate, err := vcs.getDataToValidate(ctx)
	if err != nil {
		ctx.Logger().Error("Failed to get data for validation chain",
			zap.String("step", vcs.Name()),
			zap.Error(err))
		return err
	}

	var results []map[string]interface{}
	var firstError error
	validCount := 0

	// Run each validator
	for i, validator := range vcs.validators {
		validationErr := validator.Validate(dataToValidate)

		result := map[string]interface{}{
			"validator": validator.Name(),
			"valid":     validationErr == nil,
			"index":     i,
		}

		if validationErr != nil {
			result["error"] = validationErr.Error()
			result["details"] = vcs.extractValidationDetails(validationErr)

			if firstError == nil {
				firstError = validationErr
			}

			ctx.Logger().Warn("Validator failed in chain",
				zap.String("step", vcs.Name()),
				zap.String("validator", validator.Name()),
				zap.Int("index", i),
				zap.Error(validationErr))

			if vcs.stopOnFirst {
				results = append(results, result)
				break
			}
		} else {
			validCount++
			ctx.Logger().Debug("Validator passed in chain",
				zap.String("step", vcs.Name()),
				zap.String("validator", validator.Name()),
				zap.Int("index", i))
		}

		results = append(results, result)
	}

	// Store results if requested
	if vcs.storeResults {
		chainResult := map[string]interface{}{
			"valid":            firstError == nil,
			"total_validators": len(vcs.validators),
			"valid_count":      validCount,
			"failed_count":     len(vcs.validators) - validCount,
			"results":          results,
		}

		if firstError != nil {
			chainResult["first_error"] = firstError.Error()
		}

		ctx.Set(vcs.resultField, chainResult)
	}

	// Log overall result
	if firstError != nil {
		ctx.Logger().Warn("Validation chain failed",
			zap.String("step", vcs.Name()),
			zap.Int("total_validators", len(vcs.validators)),
			zap.Int("valid_count", validCount),
			zap.Error(firstError))

		if !vcs.continueOnErr {
			return fmt.Errorf("validation chain failed: %w", firstError)
		}
	} else {
		ctx.Logger().Info("Validation chain passed",
			zap.String("step", vcs.Name()),
			zap.Int("total_validators", len(vcs.validators)))
	}

	return nil
}

func (vcs *ValidationChainStep) getDataToValidate(ctx *flow.Context) (map[string]interface{}, error) {
	if vcs.dataField == "" {
		// Validate entire context
		result := make(map[string]interface{})
		for _, key := range ctx.Keys() {
			if value, ok := ctx.Get(key); ok {
				result[key] = value
			}
		}
		return result, nil
	}

	// Validate specific field
	value, exists := ctx.Get(vcs.dataField)
	if !exists {
		return nil, fmt.Errorf("data field '%s' not found in context", vcs.dataField)
	}

	// If the value is already a map, use it directly
	if mapValue, ok := value.(map[string]interface{}); ok {
		return mapValue, nil
	}

	// Otherwise, wrap it in a map
	return map[string]interface{}{
		vcs.dataField: value,
	}, nil
}

func (vcs *ValidationChainStep) extractValidationDetails(err error) interface{} {
	// Handle different types of validation errors
	switch e := err.(type) {
	case *validators.ValidationError:
		return map[string]interface{}{
			"field":     e.Field,
			"value":     e.Value,
			"message":   e.Message,
			"validator": e.Validator,
		}
	case *validators.MultiValidationError:
		var details []interface{}
		for _, subErr := range e.Errors {
			details = append(details, vcs.extractValidationDetails(subErr))
		}
		return details
	default:
		return err.Error()
	}
}

// Helper functions for creating common validation steps

// NewRequiredFieldsValidationStep creates a step that validates required fields
func NewRequiredFieldsValidationStep(name string, fields ...string) *ValidationStep {
	validator := validators.NewRequiredFieldsValidator(fields...)
	return NewValidationStep(name, validator)
}

// NewRequiredStringFieldsValidationStep creates a step that validates required string fields
func NewRequiredStringFieldsValidationStep(name string, fields ...string) *ValidationStep {
	validator := validators.NewRequiredStringFieldsValidator(fields...)
	return NewValidationStep(name, validator)
}

// NewEmailValidationStep creates a step that validates email fields
func NewEmailValidationStep(name string, fields ...string) *ValidationStep {
	validator := validators.EmailRequiredValidator(fields...)
	return NewValidationStep(name, validator)
}

// NewIDValidationStep creates a step that validates ID fields
func NewIDValidationStep(name string, fields ...string) *ValidationStep {
	validator := validators.IDRequiredValidator(fields...)
	return NewValidationStep(name, validator)
}

// NewCustomValidationStep creates a step with a custom validation function
func NewCustomValidationStep(name string, validationFunc func(map[string]interface{}) error) *ValidationStep {
	validator := validators.NewFuncValidator(name+"_validator", validationFunc)
	return NewValidationStep(name, validator)
}
