package middleware

import (
	"fmt"
	"net/http"

	"github.com/remind101/nra"
	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

type NewRelicTracer struct {
	handler  httpx.Handler
	tracer   nra.TxTracer
	router   *httpx.Router
	createTx func(string, string, nra.TxTracer) nra.Tx
}

func NewRelicTracing(h httpx.Handler, router *httpx.Router, tracer nra.TxTracer) *NewRelicTracer {
	return &NewRelicTracer{h, tracer, router, createTx}
}

func (h *NewRelicTracer) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	path := templatePath(h.router, r)
	txName := fmt.Sprintf("%s (%s)", path, r.Method)

	tx := h.createTx(txName, r.URL.String(), h.tracer)
	ctx = nra.WithTx(ctx, tx)

	tx.Start()
	defer tx.End()

	return h.handler.ServeHTTPContext(ctx, w, r)
}

func templatePath(router *httpx.Router, r *http.Request) string {
	route, _, vars := router.Handler(r)
	var templatePath string

	if route != nil {
		var pairs []string
		for k, _ := range vars {
			pairs = append(pairs, k, fmt.Sprintf(":%s", k))
		}
		url, err := route.URLPath(pairs...)
		if err == nil {
			templatePath = url.String()
		}
	}

	if templatePath == "" {
		templatePath = r.URL.Path
	}

	return templatePath
}

func createTx(name, url string, tracer nra.TxTracer) nra.Tx {
	return nra.NewRequestTx(name, url, tracer)
}
