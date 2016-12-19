## Description

Reports custom metrics to DataDog.

## Packages

* [metricshttpx](./metricshttpx) - implements [github.com/remind101/pkg/httpx](https://github.com/remind101/pkg/tree/master/httpx)
  middleware that instruments `response.time` metric.
* [metricsmartini](./metricsmartini) - implements [github.com/go-martini/martini](https://github.com/go-martini/martini)
  middleware that instruments `response.time` metric.

## Usage

    metrics.SetEmpireDefaultTags() // sets empire.app.name, empire.app.process, empire.app.release and
                                   // container_id tags on every metric.

    metrics.Reporter, _ = NewDataDogMetricsReporter("statsd:2026")
    defer metrics.Close()
    ...
    metrics.Count("mycount", 1, map[string]string{"feature_version":"v1"}, 1.0)

See [metrics.go](https://github.com/remind101/pkg/blob/master/metrics/metrics.go) for more examples.
