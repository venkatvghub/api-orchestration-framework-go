package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusConfig holds configuration for Prometheus metrics
type PrometheusConfig struct {
	Namespace string
	Subsystem string
	Registry  *prometheus.Registry
}

// NewPrometheusMetricsWithConfig creates a new Prometheus metrics collector with custom configuration
func NewPrometheusMetricsWithConfig(config *PrometheusConfig) *PrometheusMetrics {
	if config == nil {
		config = &PrometheusConfig{
			Namespace: "api_orchestration",
			Subsystem: "framework",
		}
	}

	metrics := &PrometheusMetrics{
		namespace:  config.Namespace,
		subsystem:  config.Subsystem,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}

	// Register Go runtime metrics
	if config.Registry != nil {
		config.Registry.MustRegister(collectors.NewGoCollector())
		config.Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	} else {
		prometheus.MustRegister(collectors.NewGoCollector())
		prometheus.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	return metrics
}

// MetricsHandler returns an HTTP handler for Prometheus metrics
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// MetricsHandlerWithRegistry returns an HTTP handler for Prometheus metrics with a custom registry
func MetricsHandlerWithRegistry(registry *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// SetupDefaultPrometheusMetrics sets up the global metrics with default Prometheus configuration
func SetupDefaultPrometheusMetrics() {
	config := &PrometheusConfig{
		Namespace: "api_orchestration",
		Subsystem: "framework",
	}
	SetGlobalMetrics(NewPrometheusMetricsWithConfig(config))
}

// SetupPrometheusMetricsWithConfig sets up the global metrics with custom Prometheus configuration
func SetupPrometheusMetricsWithConfig(config *PrometheusConfig) {
	SetGlobalMetrics(NewPrometheusMetricsWithConfig(config))
}

// Common metric names as constants for consistency
const (
	// Flow metrics
	FlowExecutionsTotal   = "flow_executions_total"
	FlowExecutionDuration = "flow_execution_duration_seconds"
	StepExecutionsTotal   = "step_executions_total"
	StepExecutionDuration = "step_execution_duration_seconds"

	// HTTP metrics
	HTTPRequestsTotal   = "http_requests_total"
	HTTPRequestDuration = "http_request_duration_seconds"

	// Cache metrics
	CacheOperationsTotal   = "cache_operations_total"
	CacheOperationDuration = "cache_operation_duration_seconds"

	// Transformation metrics
	TransformationsTotal   = "transformations_total"
	TransformationDuration = "transformation_duration_seconds"

	// Validation metrics
	ValidationsTotal   = "validations_total"
	ValidationDuration = "validation_duration_seconds"
)

// Common label names
const (
	LabelFlowName    = "flow_name"
	LabelStepName    = "step_name"
	LabelSuccess     = "success"
	LabelMethod      = "method"
	LabelURL         = "url"
	LabelStatusCode  = "status_code"
	LabelOperation   = "operation"
	LabelHit         = "hit"
	LabelTransformer = "transformer"
	LabelValidator   = "validator"
)
