package metrics

import (
	"context"
	"os"
	"time"
)

// Usage:
//
//	metrics.SetEmpireDefaultTags()
//	metrics.Reporter, _ = NewDataDogMetricsReporter("statsd:2026")
//	defer metrics.Close()
//	...
//	metrics.Count("mycount", 1, map[string]string{"feature_version":"v1"}, 1.0)
var Reporter MetricsReporter
var defaultTags map[string]string

func init() {
	resetReporter()
	resetDefaultTags()
}

// MetricsReporter defines the interface for metrics reporting
type MetricsReporter interface {
	Count(name string, value int64, tags map[string]string, rate float64) error
	Gauge(name string, value float64, tags map[string]string, rate float64) error
	Histogram(name string, value float64, tags map[string]string, rate float64) error
	Distribution(name string, value float64, tags map[string]string, rate float64) error
	Set(name string, value string, tags map[string]string, rate float64) error
	TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error
	Close() error
}

// SetEmpireDefaultTags sets default tags reflecting the empire environment to each metric
func SetEmpireDefaultTags() {
	defaultTags["empire.app.name"] = os.Getenv("EMPIRE_APPNAME")
	defaultTags["empire.app.process"] = os.Getenv("EMPIRE_PROCESS")
	defaultTags["empire.app.release"] = os.Getenv("EMPIRE_RELEASE")
}

func resetDefaultTags() {
	defaultTags = make(map[string]string, 1)
}

func resetReporter() {
	Reporter = &NoopMetricsReporter{}
}

// Count reports a count metric
func Count(name string, value int64, tags map[string]string, rate float64) error {
	return Reporter.Count(name, value, withDefaultTags(tags), rate)
}

// CountWithContext reports a count metric with context
func CountWithContext(ctx context.Context, name string, value int64, tags map[string]string, rate float64) error {
	return Reporter.Count(name, value, withContextTags(ctx, tags), rate)
}

// Gauge reports a gauge metric
func Gauge(name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Gauge(name, value, withDefaultTags(tags), rate)
}

// GaugeWithContext reports a gauge metric with context
func GaugeWithContext(ctx context.Context, name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Gauge(name, value, withContextTags(ctx, tags), rate)
}

// Histogram reports a histogram metric
func Histogram(name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Histogram(name, value, withDefaultTags(tags), rate)
}

// HistogramWithContext reports a histogram metric with context
func HistogramWithContext(ctx context.Context, name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Histogram(name, value, withContextTags(ctx, tags), rate)
}

// Distribution reports a distribution metric
func Distribution(name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Distribution(name, value, withDefaultTags(tags), rate)
}

// DistributionWithContext reports a distribution metric with context
func DistributionWithContext(ctx context.Context, name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Distribution(name, value, withContextTags(ctx, tags), rate)
}

// Set reports a set metric
func Set(name string, value string, tags map[string]string, rate float64) error {
	return Reporter.Set(name, value, withDefaultTags(tags), rate)
}

// SetWithContext reports a set metric with context
func SetWithContext(ctx context.Context, name string, value string, tags map[string]string, rate float64) error {
	return Reporter.Set(name, value, withContextTags(ctx, tags), rate)
}

// TimeInMilliseconds reports a timing metric
func TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.TimeInMilliseconds(name, value, withDefaultTags(tags), rate)
}

// TimeInMillisecondsWithContext reports a timing metric with context
func TimeInMillisecondsWithContext(ctx context.Context, name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.TimeInMilliseconds(name, value, withContextTags(ctx, tags), rate)
}

// Close closes the backend connection cleanly
func Close() error {
	return Reporter.Close()
}

// Time is a shorthand for TimeInMilliseconds for easy code block instrumentation
//
// Usage:
//
//	t := metrics.Time("foo.bar", map[string]string{"baz":"qux"}, 1.0)
//	defer t.Done()
//	...
//	t.SetTags(map[string]string{"foo":"bar"}) // totally optional
func Time(name string, tags map[string]string, rate float64) *timer {
	t := &timer{name: name, tags: tags, rate: rate}
	t.Start()
	return t
}

// TimeWithContext is a shorthand for TimeInMilliseconds with context
func TimeWithContext(ctx context.Context, name string, tags map[string]string, rate float64) *timerWithContext {
	t := &timerWithContext{
		ctx:   ctx,
		name:  name,
		tags:  tags,
		rate:  rate,
		start: time.Now(),
	}
	return t
}

// ResponseTime is a shorthand for reporting web response time.
//
// Usage:
//
//	t := metrics.ResponseTime()
//	defer t.Done()
//	...
//	t.SetTags(map[string]string{"route":"GET /foo/bar"})
func ResponseTime() *timer {
	t := &timer{name: "response.time", rate: 1.0}
	t.Start()
	return t
}

// ResponseTimeWithContext is a shorthand for reporting web response time with context
func ResponseTimeWithContext(ctx context.Context) *timerWithContext {
	return TimeWithContext(ctx, "response.time", nil, 1.0)
}

// timerWithContext is a timer with context support
type timerWithContext struct {
	ctx   context.Context
	name  string
	tags  map[string]string
	rate  float64
	start time.Time
}

// SetTags sets the tags for the timer
func (t *timerWithContext) SetTags(tags map[string]string) {
	t.tags = tags
}

// Done reports the elapsed time since the timer was created
func (t *timerWithContext) Done() {
	elapsed := time.Since(t.start)
	TimeInMillisecondsWithContext(t.ctx, t.name, float64(elapsed/time.Millisecond), t.tags, t.rate)
}

func withDefaultTags(tags map[string]string) map[string]string {
	if tags == nil && defaultTags == nil {
		return nil
	}
	result := make(map[string]string, len(tags)+len(defaultTags))
	for k, v := range defaultTags {
		result[k] = v
	}
	for k, v := range tags {
		result[k] = v
	}
	return result
}

// withContextTags adds context-specific tags to the provided tags
func withContextTags(ctx context.Context, tags map[string]string) map[string]string {
	// Start with default tags
	result := withDefaultTags(tags)
	if result == nil {
		result = make(map[string]string)
	}

	// Add context-specific tags if available
	if ctx != nil {
		if traceID := getTraceIDFromContext(ctx); traceID != "" {
			result["trace_id"] = traceID
		}
		if spanID := getSpanIDFromContext(ctx); spanID != "" {
			result["span_id"] = spanID
		}
	}

	return result
}

// getTraceIDFromContext extracts a trace ID from context if available
func getTraceIDFromContext(ctx context.Context) string {
	// This is a placeholder - implement based on your tracing system
	// For example, if using OpenTelemetry:
	// span := trace.SpanFromContext(ctx)
	// if span.SpanContext().IsValid() {
	//     return span.SpanContext().TraceID().String()
	// }
	return ""
}

// getSpanIDFromContext extracts a span ID from context if available
func getSpanIDFromContext(ctx context.Context) string {
	// This is a placeholder - implement based on your tracing system
	return ""
}
