package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/remind101/pkg/metrics"
)

// ResponseTimeReporter reports timing metrics using metrics package
//
// Usage:
//   r := http.NewRouter()
//   ...
//   r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//   	w.WriteHeader(http.StatusCreated)
//   }).Methods("GET")
//   s := NewResponseTimeReporter(r, r)
//
func NewResponseTimeReporter(handler http.Handler) *responseTimeReporter {
	return &responseTimeReporter{handler}
}

type responseTimeReporter struct {
	handler http.Handler
}

func (h *responseTimeReporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := metrics.ResponseTime()
	defer t.Done()

	rw := NewResponseWriter(w) // exposes status code
	h.handler.ServeHTTP(rw, r)

	route := fmt.Sprintf("%s %s", r.Method, templatePath(r))
	status := strconv.Itoa(rw.Status())
	t.SetTags(map[string]string{
		"route":  route,
		"status": status,
	})
}

func templatePath(r *http.Request) string {
	path := "unknown"

	if route := mux.CurrentRoute(r); route != nil {
		path, _ = route.GetPathTemplate()
	}

	return path
}
