package metrics

// Usage:
//   metrics.SetAppName("myFancyApp")
//   metrics.Reporter, _ = NewDataDogMetricsReporter("statsd:2026")
//   defer metrics.Close()
//   ...
//   metrics.Count("mycount", 1, map[string]string{"feature_version":"v1"}, 1.0)
//
var Reporter MetricsReporter
var defaultTags map[string]string

func init() {
	resetReporter()
	resetDefaultTags()
}

type MetricsReporter interface {
	Count(name string, value int64, tags map[string]string, rate float64) error
	Gauge(name string, value float64, tags map[string]string, rate float64) error
	Histogram(name string, value float64, tags map[string]string, rate float64) error
	Set(name string, value string, tags map[string]string, rate float64) error
	TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error
	Close() error
}

// SetAppName adds a "app:<name>" tag to each metric
func SetAppName(appName string) {
	defaultTags["app"] = appName
}

// SetProcessName adds a "process:<name>" tag to each metric
func SetProcessName(processName string) {
	defaultTags["process"] = processName
}

func resetDefaultTags() {
	defaultTags = make(map[string]string, 1)
}

func resetReporter() {
	Reporter = &NoopMetricsReporter{}
}

func Count(name string, value int64, tags map[string]string, rate float64) error {
	return Reporter.Count(name, value, withDefaultTags(tags), rate)
}

func Gauge(name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Gauge(name, value, withDefaultTags(tags), rate)
}

func Histogram(name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.Histogram(name, value, withDefaultTags(tags), rate)
}

func Set(name string, value string, tags map[string]string, rate float64) error {
	return Reporter.Set(name, value, withDefaultTags(tags), rate)
}

func TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error {
	return Reporter.TimeInMilliseconds(name, value, withDefaultTags(tags), rate)
}

// Close closes the backend connection cleanly
func Close() error {
	return Reporter.Close()
}

// Time is a shorthand for TimeInMilliseconds for easy code block instrumentation
//
// Usage:
//   t := metrics.Time("foo.bar", map[string]string{"baz":"qux"}, 1.0)
//   defer t.Done()
//   ...
//   t.SetTags(map[string]string{"foo":"bar"}) // totally optional
func Time(name string, tags map[string]string, rate float64) *timer {
	t := &timer{name: name, tags: tags, rate: rate}
	t.Start()
	return t
}

// ResponseTime is a shorthand for reporting web response time.
//
// Usage:
//   t := metrics.ResponseTime()
//   defer t.Done()
//   ...
//   t.SetTags(map[string]string{"route":"GET /foo/bar"})
func ResponseTime() *timer {
	t := &timer{name: "response.time", rate: 1.0}
	t.Start()
	return t
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
