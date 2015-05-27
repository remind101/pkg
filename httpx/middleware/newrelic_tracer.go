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
	txName := fmt.Sprintf("%s %s", r.Method, path)

	tx := h.createTx(txName, r.URL.String(), h.tracer)
	ctx = nra.WithTx(ctx, tx)

	tx.Start()
	defer tx.End()

	return h.handler.ServeHTTPContext(ctx, w, r)
}

func templatePath(router *httpx.Router, r *http.Request) string {
	var tpl string

	route, _, _ := router.Handler(r)
	if route != nil {
		tpl = route.GetPathTemplate()
	}

	if tpl == "" {
		tpl = r.URL.Path
	}

	return tpl
}

func createTx(name, url string, tracer nra.TxTracer) nra.Tx {
	return nra.NewRequestTx(name, url, tracer)
}
