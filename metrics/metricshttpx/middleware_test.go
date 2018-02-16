package metricshttpx_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/metrics"
	"github.com/remind101/pkg/metrics/metricshttpx"
	"context"
)

func TestMiddlewareReportsResponseTimeMetrics(t *testing.T) {
	fakeReporter := &fakeMetricsReporter{}
	cleanup := withReporter(fakeReporter)
	defer cleanup()

	r := httpx.NewRouter()
	r.HandleFunc("/", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusCreated)
		return nil
	}).Methods("GET")

	s := metricshttpx.NewResponseTimeReporter(r, r)
	b := middleware.BackgroundContext(s)

	ts := httptest.NewServer(b)
	defer ts.Close()

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	gotName, wantName := fakeReporter.LatestTimingMetricName, "response.time"
	if gotName != wantName {
		t.Errorf("expected response time metric name to be %s, got %s", wantName, gotName)
	}

	gotTime := fakeReporter.LatestTimingMetricValue
	if gotTime == 0 {
		t.Errorf("expected response time to be reported")
	}

	gotTags := fakeReporter.LatestTimingMetricTags
	wantTags := map[string]string{
		"route":  "GET /",
		"status": "201",
	}

	if !reflect.DeepEqual(gotTags, wantTags) {
		t.Errorf("expected tags:\n\t%v\\ngot tags:%v\n\t", wantTags, gotTags)
	}
}

func withReporter(reporter metrics.MetricsReporter) func() {
	oldReporter := metrics.Reporter
	metrics.Reporter = reporter
	return func() {
		metrics.Reporter = oldReporter
	}
}

type fakeMetricsReporter struct {
	metrics.NoopMetricsReporter
	LatestTimingMetricName  string
	LatestTimingMetricValue float64
	LatestTimingMetricTags  map[string]string
}

func (r *fakeMetricsReporter) TimeInMilliseconds(name string, value float64, tags map[string]string, rate float64) error {
	r.LatestTimingMetricName = name
	r.LatestTimingMetricValue = value
	r.LatestTimingMetricTags = tags
	return nil
}
