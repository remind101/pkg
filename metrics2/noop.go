package metrics2

type NoopMetricsReporter struct{}

func (c *NoopMetricsReporter) Count(name string, value int64, tags map[string]string, rate float64) error {
	return nil
}

func (c *NoopMetricsReporter) Gauge(name string, value float64, tags map[string]string, rate float64) error {
	return nil
}

func (c *NoopMetricsReporter) Histogram(name string, value float64, tags map[string]string, rate float64) error {
	return nil
}

func (c *NoopMetricsReporter) Set(name string, value string, tags map[string]string, rate float64) error {
	return nil
}

func (c *NoopMetricsReporter) TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error {
	return nil
}

func (c *NoopMetricsReporter) Close() error {
	return nil
}
