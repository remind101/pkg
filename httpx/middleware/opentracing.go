package middleware

import (
	"context"
	"fmt"
	"net/http"

	dd_opentracing "github.com/DataDog/dd-trace-go/opentracing"
	dd_ext "github.com/DataDog/dd-trace-go/tracer/ext"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/remind101/pkg/httpx"
)

type OpentracingTracer struct {
	handler httpx.Handler
	router  *httpx.Router
}

func OpentracingTracing(h httpx.Handler, router *httpx.Router) *OpentracingTracer {
	return &OpentracingTracer{h, router}
}

func (h *OpentracingTracer) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	path := otTemplatePath(h.router, r)
	route := fmt.Sprintf("%s %s", r.Method, path)

	var span opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		span = opentracing.StartSpan("server.request")
	} else {
		span = opentracing.StartSpan("server.request", ext.RPCServerOption(wireContext))
	}
	span.SetTag(dd_opentracing.ResourceName, route)
	span.SetTag(dd_opentracing.SpanType, "web")
	span.SetTag(dd_ext.HTTPMethod, r.Method)
	span.SetTag(dd_ext.HTTPURL, r.RequestURI)

	if rid := httpx.RequestID(ctx); rid != "" {
		span.SetTag("request_id", rid)
	}

	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	r = r.WithContext(ctx)

	rw := NewResponseWriter(w)
	reqErr := h.handler.ServeHTTPContext(ctx, rw, r)
	if reqErr != nil {
		span.SetTag(dd_opentracing.Error, reqErr)
	}
	span.SetTag(dd_ext.HTTPCode, rw.Status())

	return reqErr
}

func otTemplatePath(router *httpx.Router, r *http.Request) string {
	var tpl string

	route, _, _ := router.Handler(r)
	if route != nil {
		tpl = route.GetPathTemplate()
	}

	if tpl == "" {
		tpl = "unknown"
	}

	return tpl
}
