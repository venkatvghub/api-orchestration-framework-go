package interfaces

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// ExecutionContext defines the interface for step execution context
// This breaks the circular dependency between steps and flow packages
type ExecutionContext interface {
	// Core context operations
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Has(key string) bool
	Delete(key string)
	Keys() []string

	// Typed getters
	GetString(key string) (string, error)
	GetInt(key string) (int, error)
	GetBool(key string) (bool, error)
	GetMap(key string) (map[string]interface{}, error)

	// Context operations
	Context() context.Context
	Clone() ExecutionContext

	// Metadata
	FlowName() string
	ExecutionID() string
	StartTime() time.Time
	Duration() time.Duration

	// Observability
	Logger() *zap.Logger
}

// Step represents a single operation in a flow
type Step interface {
	Run(ctx ExecutionContext) error
	Name() string
	Description() string
}
