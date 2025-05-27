package core

import (
	"fmt"
	"strings"

	"github.com/venkatvghub/api-orchestration-framework/pkg/flow"
	"github.com/venkatvghub/api-orchestration-framework/pkg/steps/base"
	"github.com/venkatvghub/api-orchestration-framework/pkg/utils"
	"go.uber.org/zap"
)

// LogStep provides enhanced logging functionality
type LogStep struct {
	*base.BaseStep
	level       string
	message     string
	fields      map[string]string
	sanitize    bool
	includeCtx  bool
	contextKeys []string
}

// NewLogStep creates a new log step
func NewLogStep(name, level, message string) *LogStep {
	return &LogStep{
		BaseStep:   base.NewBaseStep(name, "Logging"),
		level:      level,
		message:    message,
		fields:     make(map[string]string),
		sanitize:   true,
		includeCtx: false,
	}
}

// WithFields adds custom fields to log
func (ls *LogStep) WithFields(fields map[string]string) *LogStep {
	for k, v := range fields {
		ls.fields[k] = v
	}
	return ls
}

// WithField adds a single field to log
func (ls *LogStep) WithField(key, value string) *LogStep {
	ls.fields[key] = value
	return ls
}

// WithSanitization controls whether to sanitize sensitive information
func (ls *LogStep) WithSanitization(sanitize bool) *LogStep {
	ls.sanitize = sanitize
	return ls
}

// WithContext includes context data in the log
func (ls *LogStep) WithContext(include bool, keys ...string) *LogStep {
	ls.includeCtx = include
	ls.contextKeys = keys
	return ls
}

func (ls *LogStep) Run(ctx *flow.Context) error {
	// Interpolate message
	logMessage, err := utils.InterpolateString(ls.message, ctx)
	if err != nil {
		logMessage = ls.message // Fallback to original message
	}

	// Sanitize message if enabled
	if ls.sanitize {
		logMessage = utils.SanitizeLogMessage(logMessage)
	}

	// Prepare log fields
	logFields := ls.prepareLogFields(ctx)

	// Log based on level
	switch strings.ToLower(ls.level) {
	case "debug":
		ctx.Logger().Debug(logMessage, logFields...)
	case "info":
		ctx.Logger().Info(logMessage, logFields...)
	case "warn", "warning":
		ctx.Logger().Warn(logMessage, logFields...)
	case "error":
		ctx.Logger().Error(logMessage, logFields...)
	case "fatal":
		ctx.Logger().Fatal(logMessage, logFields...)
	case "panic":
		ctx.Logger().Panic(logMessage, logFields...)
	default:
		ctx.Logger().Info(logMessage, logFields...)
	}

	return nil
}

func (ls *LogStep) prepareLogFields(ctx *flow.Context) []zap.Field {
	var fields []zap.Field

	// Add step name
	fields = append(fields, zap.String("step", ls.Name()))

	// Add custom fields
	for key, valueTemplate := range ls.fields {
		value, err := utils.InterpolateString(valueTemplate, ctx)
		if err != nil {
			value = valueTemplate // Fallback to template
		}

		if ls.sanitize {
			value = utils.SanitizeLogMessage(value)
		}

		fields = append(fields, zap.String(key, value))
	}

	// Add context data if requested
	if ls.includeCtx {
		contextData := ls.getContextData(ctx)
		if len(contextData) > 0 {
			fields = append(fields, zap.Any("context", contextData))
		}
	}

	return fields
}

func (ls *LogStep) getContextData(ctx *flow.Context) map[string]interface{} {
	result := make(map[string]interface{})

	if len(ls.contextKeys) > 0 {
		// Include only specified keys
		for _, key := range ls.contextKeys {
			if value, ok := ctx.Get(key); ok {
				if ls.sanitize && ls.isSensitiveKey(key) {
					result[key] = "***"
				} else {
					result[key] = value
				}
			}
		}
	} else {
		// Include all context keys
		for _, key := range ctx.Keys() {
			if value, ok := ctx.Get(key); ok {
				if ls.sanitize && ls.isSensitiveKey(key) {
					result[key] = "***"
				} else {
					result[key] = value
				}
			}
		}
	}

	return result
}

func (ls *LogStep) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "token", "secret", "key", "auth",
		"authorization", "cookie", "session", "credential",
	}

	keyLower := strings.ToLower(key)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(keyLower, sensitive) {
			return true
		}
	}

	return false
}

// MetricsLogStep logs metrics and performance data
type MetricsLogStep struct {
	*base.BaseStep
	metricName   string
	metricValue  string
	metricType   string // counter, gauge, histogram, timer
	tags         map[string]string
	includeStats bool
}

// NewMetricsLogStep creates a new metrics log step
func NewMetricsLogStep(name, metricName, metricType string) *MetricsLogStep {
	return &MetricsLogStep{
		BaseStep:     base.NewBaseStep(name, "Metrics logging"),
		metricName:   metricName,
		metricType:   metricType,
		tags:         make(map[string]string),
		includeStats: true,
	}
}

// WithValue sets the metric value template
func (mls *MetricsLogStep) WithValue(valueTemplate string) *MetricsLogStep {
	mls.metricValue = valueTemplate
	return mls
}

// WithTags adds tags to the metric
func (mls *MetricsLogStep) WithTags(tags map[string]string) *MetricsLogStep {
	for k, v := range tags {
		mls.tags[k] = v
	}
	return mls
}

// WithStats controls whether to include additional statistics
func (mls *MetricsLogStep) WithStats(include bool) *MetricsLogStep {
	mls.includeStats = include
	return mls
}

func (mls *MetricsLogStep) Run(ctx *flow.Context) error {
	// Interpolate metric value
	var metricValue interface{} = 1 // Default value for counters
	if mls.metricValue != "" {
		if interpolated, err := utils.InterpolateString(mls.metricValue, ctx); err == nil {
			metricValue = interpolated
		}
	}

	// Prepare tags
	tags := make(map[string]string)
	for key, valueTemplate := range mls.tags {
		if interpolated, err := utils.InterpolateString(valueTemplate, ctx); err == nil {
			tags[key] = interpolated
		} else {
			tags[key] = valueTemplate
		}
	}

	// Prepare log fields
	fields := []zap.Field{
		zap.String("step", mls.Name()),
		zap.String("metric_name", mls.metricName),
		zap.String("metric_type", mls.metricType),
		zap.Any("metric_value", metricValue),
		zap.Any("tags", tags),
	}

	// Add statistics if enabled
	if mls.includeStats {
		stats := mls.gatherStats(ctx)
		fields = append(fields, zap.Any("stats", stats))
	}

	ctx.Logger().Info("Metric logged", fields...)

	// Store metric in context for potential aggregation
	ctx.Set(fmt.Sprintf("metric_%s", mls.metricName), map[string]interface{}{
		"name":  mls.metricName,
		"type":  mls.metricType,
		"value": metricValue,
		"tags":  tags,
	})

	return nil
}

func (mls *MetricsLogStep) gatherStats(ctx *flow.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	// Add context size
	stats["context_keys"] = len(ctx.Keys())

	// Add execution time if available
	if startTime, ok := ctx.Get("start_time"); ok {
		stats["start_time"] = startTime
	}

	return stats
}

// Helper functions for creating common log steps

// NewInfoLogStep creates an info level log step
func NewInfoLogStep(name, message string) *LogStep {
	return NewLogStep(name, "info", message)
}

// NewErrorLogStep creates an error level log step
func NewErrorLogStep(name, message string) *LogStep {
	return NewLogStep(name, "error", message)
}

// NewDebugLogStep creates a debug level log step
func NewDebugLogStep(name, message string) *LogStep {
	return NewLogStep(name, "debug", message)
}

// NewWarnLogStep creates a warning level log step
func NewWarnLogStep(name, message string) *LogStep {
	return NewLogStep(name, "warn", message)
}

// NewContextLogStep creates a log step that includes context data
func NewContextLogStep(name, message string, contextKeys ...string) *LogStep {
	return NewLogStep(name, "info", message).
		WithContext(true, contextKeys...)
}

// NewCounterMetricStep creates a counter metric log step
func NewCounterMetricStep(name, metricName string) *MetricsLogStep {
	return NewMetricsLogStep(name, metricName, "counter")
}

// NewGaugeMetricStep creates a gauge metric log step
func NewGaugeMetricStep(name, metricName, valueTemplate string) *MetricsLogStep {
	return NewMetricsLogStep(name, metricName, "gauge").
		WithValue(valueTemplate)
}

// NewTimerMetricStep creates a timer metric log step
func NewTimerMetricStep(name, metricName, durationField string) *MetricsLogStep {
	return NewMetricsLogStep(name, metricName, "timer").
		WithValue(fmt.Sprintf("${%s}", durationField))
}
