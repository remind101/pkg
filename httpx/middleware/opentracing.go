package middleware

import (
	"fmt"
	"net/http"

	dd_opentracing "github.com/DataDog/dd-trace-go/opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

type OpentracingTracer struct {
	handler httpx.Handler
	router  *httpx.Router
}

func OpentracingTracing(h httpx.Handler, router *httpx.Router) *OpentracingTracer {
	return &OpentracingTracer{h, router}
}

func (h *OpentracingTracer) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	path := templatePath(h.router, r)
	route := fmt.Sprintf("%s %s", r.Method, path)

	var span opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		fmt.Printf("Missing or invalid wireContext: %s\nerr: %s\n", wireContext, err)
		span = opentracing.StartSpan("server.request")
	} else {
		span = opentracing.StartSpan("server.request", ext.RPCServerOption(wireContext))
	}
	span.SetTag(dd_opentracing.ResourceName, route)
	span.SetTag(dd_opentracing.SpanType, "web")

	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return h.handler.ServeHTTPContext(ctx, w, r)
}
