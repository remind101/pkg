package metricshttpx

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/metrics"
)

// ResponseTimeReporter reports timing metrics using metrics package
//
// Usage:
//   r := httpx.NewRouter()
//   ...
//   r.HandleFunc("/", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//   	w.WriteHeader(http.StatusCreated)
//   	return nil
//   }).Methods("GET")
//   s := NewResponseTimeReporter(r, r)
//
func NewResponseTimeReporter(handler httpx.Handler, router *httpx.Router) *responseTimeReporter {
	if router == nil {
		panic("NewResponseTimeReporter: router is requred")
	}
	return &responseTimeReporter{handler, router}
}

type responseTimeReporter struct {
	handler httpx.Handler
	router  *httpx.Router
}

func (h *responseTimeReporter) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	t := metrics.ResponseTime()
	defer t.Done()

	rw := middleware.NewResponseWriter(w) // exposes status code
	err := h.handler.ServeHTTPContext(ctx, rw, r)

	route := fmt.Sprintf("%s %s", r.Method, templatePath(h.router, r))
	status := strconv.Itoa(rw.Status())
	t.SetTags(map[string]string{
		"route":  route,
		"status": status,
	})

	return err
}

func templatePath(router *httpx.Router, r *http.Request) string {
	route, _, _ := router.Handler(r)
	if route == nil {
		return "unknown"
	}
	return route.GetPathTemplate()
}
