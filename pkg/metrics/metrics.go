package metrics

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	// Counters
	IncrementCounter(name string, tags map[string]string)
	IncrementCounterBy(name string, value int64, tags map[string]string)

	// Gauges
	SetGauge(name string, value float64, tags map[string]string)
	IncrementGauge(name string, tags map[string]string)
	DecrementGauge(name string, tags map[string]string)

	// Histograms
	RecordDuration(name string, duration time.Duration, tags map[string]string)
	RecordValue(name string, value float64, tags map[string]string)

	// Flow metrics
	RecordFlowExecution(flowName string, duration time.Duration, success bool)
	RecordStepExecution(stepName string, duration time.Duration, success bool)
	RecordHTTPRequest(method, url string, statusCode int, duration time.Duration)
	RecordCacheOperation(operation string, hit bool, duration time.Duration)
	RecordTransformation(transformer string, duration time.Duration, success bool)
	RecordValidation(validator string, duration time.Duration, success bool)
}

// PrometheusMetrics provides a Prometheus-based metrics implementation
type PrometheusMetrics struct {
	namespace  string
	subsystem  string
	mu         sync.RWMutex
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics(namespace, subsystem string) *PrometheusMetrics {
	return &PrometheusMetrics{
		namespace:  namespace,
		subsystem:  subsystem,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

func (p *PrometheusMetrics) getOrCreateCounter(name string, labelNames []string) *prometheus.CounterVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	if counter, exists := p.counters[name]; exists {
		return counter
	}

	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      fmt.Sprintf("Counter metric for %s", name),
	}, labelNames)

	p.counters[name] = counter
	return counter
}

func (p *PrometheusMetrics) getOrCreateGauge(name string, labelNames []string) *prometheus.GaugeVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	if gauge, exists := p.gauges[name]; exists {
		return gauge
	}

	gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      fmt.Sprintf("Gauge metric for %s", name),
	}, labelNames)

	p.gauges[name] = gauge
	return gauge
}

func (p *PrometheusMetrics) getOrCreateHistogram(name string, labelNames []string) *prometheus.HistogramVec {
	p.mu.Lock()
	defer p.mu.Unlock()

	if histogram, exists := p.histograms[name]; exists {
		return histogram
	}

	histogram := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      name,
		Help:      fmt.Sprintf("Histogram metric for %s", name),
		Buckets:   prometheus.DefBuckets,
	}, labelNames)

	p.histograms[name] = histogram
	return histogram
}

func (p *PrometheusMetrics) extractLabels(tags map[string]string) ([]string, []string) {
	if len(tags) == 0 {
		return []string{}, []string{}
	}

	labelNames := make([]string, 0, len(tags))
	labelValues := make([]string, 0, len(tags))

	for key, value := range tags {
		labelNames = append(labelNames, key)
		labelValues = append(labelValues, value)
	}

	return labelNames, labelValues
}

func (p *PrometheusMetrics) IncrementCounter(name string, tags map[string]string) {
	p.IncrementCounterBy(name, 1, tags)
}

func (p *PrometheusMetrics) IncrementCounterBy(name string, value int64, tags map[string]string) {
	labelNames, labelValues := p.extractLabels(tags)
	counter := p.getOrCreateCounter(name, labelNames)
	counter.WithLabelValues(labelValues...).Add(float64(value))
}

func (p *PrometheusMetrics) SetGauge(name string, value float64, tags map[string]string) {
	labelNames, labelValues := p.extractLabels(tags)
	gauge := p.getOrCreateGauge(name, labelNames)
	gauge.WithLabelValues(labelValues...).Set(value)
}

func (p *PrometheusMetrics) IncrementGauge(name string, tags map[string]string) {
	labelNames, labelValues := p.extractLabels(tags)
	gauge := p.getOrCreateGauge(name, labelNames)
	gauge.WithLabelValues(labelValues...).Inc()
}

func (p *PrometheusMetrics) DecrementGauge(name string, tags map[string]string) {
	labelNames, labelValues := p.extractLabels(tags)
	gauge := p.getOrCreateGauge(name, labelNames)
	gauge.WithLabelValues(labelValues...).Dec()
}

func (p *PrometheusMetrics) RecordDuration(name string, duration time.Duration, tags map[string]string) {
	p.RecordValue(name, duration.Seconds(), tags)
}

func (p *PrometheusMetrics) RecordValue(name string, value float64, tags map[string]string) {
	labelNames, labelValues := p.extractLabels(tags)
	histogram := p.getOrCreateHistogram(name, labelNames)
	histogram.WithLabelValues(labelValues...).Observe(value)
}

func (p *PrometheusMetrics) RecordFlowExecution(flowName string, duration time.Duration, success bool) {
	tags := map[string]string{
		"flow_name": flowName,
		"success":   boolToString(success),
	}
	p.IncrementCounter("flow_executions_total", tags)
	p.RecordDuration("flow_execution_duration_seconds", duration, tags)
}

func (p *PrometheusMetrics) RecordStepExecution(stepName string, duration time.Duration, success bool) {
	tags := map[string]string{
		"step_name": stepName,
		"success":   boolToString(success),
	}
	p.IncrementCounter("step_executions_total", tags)
	p.RecordDuration("step_execution_duration_seconds", duration, tags)
}

func (p *PrometheusMetrics) RecordHTTPRequest(method, url string, statusCode int, duration time.Duration) {
	tags := map[string]string{
		"method":      method,
		"url":         url,
		"status_code": strconv.Itoa(statusCode),
	}
	p.IncrementCounter("http_requests_total", tags)
	p.RecordDuration("http_request_duration_seconds", duration, tags)
}

func (p *PrometheusMetrics) RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	tags := map[string]string{
		"operation": operation,
		"hit":       boolToString(hit),
	}
	p.IncrementCounter("cache_operations_total", tags)
	p.RecordDuration("cache_operation_duration_seconds", duration, tags)
}

func (p *PrometheusMetrics) RecordTransformation(transformer string, duration time.Duration, success bool) {
	tags := map[string]string{
		"transformer": transformer,
		"success":     boolToString(success),
	}
	p.IncrementCounter("transformations_total", tags)
	p.RecordDuration("transformation_duration_seconds", duration, tags)
}

func (p *PrometheusMetrics) RecordValidation(validator string, duration time.Duration, success bool) {
	tags := map[string]string{
		"validator": validator,
		"success":   boolToString(success),
	}
	p.IncrementCounter("validations_total", tags)
	p.RecordDuration("validation_duration_seconds", duration, tags)
}

// InMemoryMetrics provides an in-memory metrics implementation
type InMemoryMetrics struct {
	mu         sync.RWMutex
	counters   map[string]int64
	gauges     map[string]float64
	histograms map[string][]float64
}

// NewInMemoryMetrics creates a new in-memory metrics collector
func NewInMemoryMetrics() *InMemoryMetrics {
	return &InMemoryMetrics{
		counters:   make(map[string]int64),
		gauges:     make(map[string]float64),
		histograms: make(map[string][]float64),
	}
}

func (m *InMemoryMetrics) IncrementCounter(name string, tags map[string]string) {
	m.IncrementCounterBy(name, 1, tags)
}

func (m *InMemoryMetrics) IncrementCounterBy(name string, value int64, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.buildKey(name, tags)
	m.counters[key] += value
}

func (m *InMemoryMetrics) SetGauge(name string, value float64, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.buildKey(name, tags)
	m.gauges[key] = value
}

func (m *InMemoryMetrics) IncrementGauge(name string, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.buildKey(name, tags)
	m.gauges[key]++
}

func (m *InMemoryMetrics) DecrementGauge(name string, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.buildKey(name, tags)
	m.gauges[key]--
}

func (m *InMemoryMetrics) RecordDuration(name string, duration time.Duration, tags map[string]string) {
	m.RecordValue(name, float64(duration.Milliseconds()), tags)
}

func (m *InMemoryMetrics) RecordValue(name string, value float64, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.buildKey(name, tags)
	m.histograms[key] = append(m.histograms[key], value)
}

func (m *InMemoryMetrics) RecordFlowExecution(flowName string, duration time.Duration, success bool) {
	tags := map[string]string{
		"flow_name": flowName,
		"success":   boolToString(success),
	}
	m.IncrementCounter("flow_executions_total", tags)
	m.RecordDuration("flow_execution_duration_ms", duration, tags)
}

func (m *InMemoryMetrics) RecordStepExecution(stepName string, duration time.Duration, success bool) {
	tags := map[string]string{
		"step_name": stepName,
		"success":   boolToString(success),
	}
	m.IncrementCounter("step_executions_total", tags)
	m.RecordDuration("step_execution_duration_ms", duration, tags)
}

func (m *InMemoryMetrics) RecordHTTPRequest(method, url string, statusCode int, duration time.Duration) {
	tags := map[string]string{
		"method":      method,
		"url":         url,
		"status_code": string(rune(statusCode)),
	}
	m.IncrementCounter("http_requests_total", tags)
	m.RecordDuration("http_request_duration_ms", duration, tags)
}

func (m *InMemoryMetrics) RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	tags := map[string]string{
		"operation": operation,
		"hit":       boolToString(hit),
	}
	m.IncrementCounter("cache_operations_total", tags)
	m.RecordDuration("cache_operation_duration_ms", duration, tags)
}

func (m *InMemoryMetrics) RecordTransformation(transformer string, duration time.Duration, success bool) {
	tags := map[string]string{
		"transformer": transformer,
		"success":     boolToString(success),
	}
	m.IncrementCounter("transformations_total", tags)
	m.RecordDuration("transformation_duration_ms", duration, tags)
}

func (m *InMemoryMetrics) RecordValidation(validator string, duration time.Duration, success bool) {
	tags := map[string]string{
		"validator": validator,
		"success":   boolToString(success),
	}
	m.IncrementCounter("validations_total", tags)
	m.RecordDuration("validation_duration_ms", duration, tags)
}

func (m *InMemoryMetrics) buildKey(name string, tags map[string]string) string {
	key := name
	for k, v := range tags {
		key += "," + k + "=" + v
	}
	return key
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// GetCounters returns a copy of all counters
func (m *InMemoryMetrics) GetCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]int64)
	for k, v := range m.counters {
		result[k] = v
	}
	return result
}

// GetGauges returns a copy of all gauges
func (m *InMemoryMetrics) GetGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]float64)
	for k, v := range m.gauges {
		result[k] = v
	}
	return result
}

// GetHistograms returns a copy of all histograms
func (m *InMemoryMetrics) GetHistograms() map[string][]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string][]float64)
	for k, v := range m.histograms {
		result[k] = make([]float64, len(v))
		copy(result[k], v)
	}
	return result
}

// Reset clears all metrics
func (m *InMemoryMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters = make(map[string]int64)
	m.gauges = make(map[string]float64)
	m.histograms = make(map[string][]float64)
}

// NoOpMetrics provides a no-op implementation for production when metrics are disabled
type NoOpMetrics struct{}

func NewNoOpMetrics() *NoOpMetrics {
	return &NoOpMetrics{}
}

func (n *NoOpMetrics) IncrementCounter(name string, tags map[string]string)                         {}
func (n *NoOpMetrics) IncrementCounterBy(name string, value int64, tags map[string]string)          {}
func (n *NoOpMetrics) SetGauge(name string, value float64, tags map[string]string)                  {}
func (n *NoOpMetrics) IncrementGauge(name string, tags map[string]string)                           {}
func (n *NoOpMetrics) DecrementGauge(name string, tags map[string]string)                           {}
func (n *NoOpMetrics) RecordDuration(name string, duration time.Duration, tags map[string]string)   {}
func (n *NoOpMetrics) RecordValue(name string, value float64, tags map[string]string)               {}
func (n *NoOpMetrics) RecordFlowExecution(flowName string, duration time.Duration, success bool)    {}
func (n *NoOpMetrics) RecordStepExecution(stepName string, duration time.Duration, success bool)    {}
func (n *NoOpMetrics) RecordHTTPRequest(method, url string, statusCode int, duration time.Duration) {}
func (n *NoOpMetrics) RecordCacheOperation(operation string, hit bool, duration time.Duration)      {}
func (n *NoOpMetrics) RecordTransformation(transformer string, duration time.Duration, success bool) {
}
func (n *NoOpMetrics) RecordValidation(validator string, duration time.Duration, success bool) {}

// Global metrics instance - now defaults to Prometheus
var globalMetrics MetricsCollector = NewPrometheusMetrics("api_orchestration", "framework")

// SetGlobalMetrics sets the global metrics collector
func SetGlobalMetrics(collector MetricsCollector) {
	globalMetrics = collector
}

// GetGlobalMetrics returns the global metrics collector
func GetGlobalMetrics() MetricsCollector {
	return globalMetrics
}

// Convenience functions for global metrics
func IncrementCounter(name string, tags map[string]string) {
	globalMetrics.IncrementCounter(name, tags)
}

func RecordDuration(name string, duration time.Duration, tags map[string]string) {
	globalMetrics.RecordDuration(name, duration, tags)
}

func RecordFlowExecution(flowName string, duration time.Duration, success bool) {
	globalMetrics.RecordFlowExecution(flowName, duration, success)
}

func RecordStepExecution(stepName string, duration time.Duration, success bool) {
	globalMetrics.RecordStepExecution(stepName, duration, success)
}

func RecordHTTPRequest(method, url string, statusCode int, duration time.Duration) {
	globalMetrics.RecordHTTPRequest(method, url, statusCode, duration)
}

// Global convenience function for cache operations
func RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	globalMetrics.RecordCacheOperation(operation, hit, duration)
}
