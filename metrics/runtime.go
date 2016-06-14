package metrics

import (
	"runtime"
	"time"
)

// sets how often the metrics Runtime will sample and
// emit stats.
var RuntimeMetricsSamplingInterval time.Duration = 30 * time.Second

// Runtime enters into a loop, sampling and outputing the runtime stats periodically.
// Usage:
//   func main() {
//     ...
//     go metrics.Runtime()
//     ...
//   }

func Runtime() {
	sampleEvery(RuntimeMetricsSamplingInterval)
}

func sampleEvery(t time.Duration) {
	c := time.Tick(t)
	for _ = range c {
		ReportRuntimeMetrics()
	}
}

// Runtime enters into a loop, sampling and outputing the runtime stats periodically.
func ReportRuntimeMetrics() {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)
	numGoroutines := runtime.NumGoroutine()

	r := map[string]float64{
		"goroutine":                     float64(numGoroutines),
		"runtime.MemStats.Alloc":        float64(memstats.Alloc),
		"runtime.MemStats.Frees":        float64(memstats.Frees),
		"runtime.MemStats.HeapAlloc":    float64(memstats.HeapAlloc),
		"runtime.MemStats.HeapIdle":     float64(memstats.HeapIdle),
		"runtime.MemStats.HeapObjects":  float64(memstats.HeapObjects),
		"runtime.MemStats.HeapReleased": float64(memstats.HeapReleased),
		"runtime.MemStats.HeapSys":      float64(memstats.HeapSys),
		"runtime.MemStats.LastGC":       float64(memstats.LastGC),
		"runtime.MemStats.Lookups":      float64(memstats.Lookups),
		"runtime.MemStats.Mallocs":      float64(memstats.Mallocs),
		"runtime.MemStats.MCacheInuse":  float64(memstats.MCacheInuse),
		"runtime.MemStats.MCacheSys":    float64(memstats.MCacheSys),
		"runtime.MemStats.MSpanInuse":   float64(memstats.MSpanInuse),
		"runtime.MemStats.MSpanSys":     float64(memstats.MSpanSys),
		"runtime.MemStats.NextGC":       float64(memstats.NextGC),
		"runtime.MemStats.NumGC":        float64(memstats.NumGC),
		"runtime.MemStats.StackInuse":   float64(memstats.StackInuse),
	}

	for name, value := range r {
		Gauge(name, value, nil, 1.0)
	}
}
