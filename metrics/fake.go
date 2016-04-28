package metrics

type metric struct {
	Name string
	Tags map[string]string
	Rate float64
}

type intMetric struct {
	metric
	Value int64
}

type floatMetric struct {
	metric
	Value float64
}

type fakeMetricsReporter struct {
	LastCountMetric              *intMetric
	LastTimeInMillisecondsMetric *floatMetric
}

func newFakeMetricsReporter() *fakeMetricsReporter {
	return &fakeMetricsReporter{}
}

func (r *fakeMetricsReporter) Count(name string, value int64, tags map[string]string, rate float64) error {
	r.LastCountMetric = &intMetric{
		metric{name, tags, rate},
		value,
	}
	return nil
}

func (r *fakeMetricsReporter) Gauge(name string, value float64, tags map[string]string, rate float64) error {
	return nil
}

func (r *fakeMetricsReporter) Histogram(name string, value float64, tags map[string]string, rate float64) error {
	return nil
}

func (r *fakeMetricsReporter) Set(name string, value string, tags map[string]string, rate float64) error {
	return nil
}

func (r *fakeMetricsReporter) TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error {
	r.LastTimeInMillisecondsMetric = &floatMetric{
		metric{name, tags, rate},
		value,
	}
	return nil
}

func (r *fakeMetricsReporter) Close() error {
	return nil
}
