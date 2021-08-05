package metrics

import (
	"fmt"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
)

type DataDogMetricsReporter struct {
	client *statsd.Client
}

func NewDataDogMetricsReporter(addr string) (*DataDogMetricsReporter, error) {
	c, err := statsd.New(addr)
	if err != nil {
		return nil, fmt.Errorf("Could not create statsd client: %v", err)
	}
	return &DataDogMetricsReporter{c}, nil
}

func (c *DataDogMetricsReporter) Count(name string, value int64, tags map[string]string, rate float64) error {
	return c.client.Count(name, value, convertTags(tags), rate)
}

func (c *DataDogMetricsReporter) Gauge(name string, value float64, tags map[string]string, rate float64) error {
	return c.client.Gauge(name, value, convertTags(tags), rate)
}

func (c *DataDogMetricsReporter) Histogram(name string, value float64, tags map[string]string, rate float64) error {
	return c.client.Histogram(name, value, convertTags(tags), rate)
}

func (c *DataDogMetricsReporter) Distribution(name string, value float64, tags map[string]string, rate float64) error {
	return c.client.Distribution(name, value, convertTags(tags), rate)
}

func (c *DataDogMetricsReporter) Set(name string, value string, tags map[string]string, rate float64) error {
	return c.client.Set(name, value, convertTags(tags), rate)
}

func (c *DataDogMetricsReporter) TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error {
	return c.client.TimeInMilliseconds(name, value, convertTags(tags), rate)
}

func (c *DataDogMetricsReporter) Close() error {
	return c.client.Close()
}

// converts from {"TagName":"TagValue"} to ["tagname:tagvalue"]
func convertTags(tags map[string]string) []string {
	result := make([]string, 0, len(tags))
	for k, v := range tags {
		k := strings.ToLower(strings.TrimSpace(k))
		v := strings.ToLower(strings.TrimSpace(v))
		result = append(result, fmt.Sprintf("%s:%s", k, v))
	}
	return result
}
