package context

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
)

// Copy copies common httpx values injected into a request context to another
// context.
//
// This is useful when performing work outside of the request lifecycle that was
// a result of a request. For instance, the request id and tracing spans are useful
// but the deadline of the request context does not apply.
func Copy(ctx context.Context) context.Context {
	return CopyToContext(context.Background(), ctx)
}

// CopyToContext copies common httpx values injected into a request context to a
// target context. See #Copy for an explanation of why this might be useful.
func CopyToContext(target, source context.Context) context.Context {
	// Copy logger
	if l, ok := logger.FromContext(source); ok {
		target = logger.WithLogger(target, l)
	}

	// Copy reporter
	if r, ok := reporter.FromContext(source); ok {
		target = reporter.WithReporter(target, r)
	}

	// Copy request id
	target = httpx.WithRequestID(target, httpx.RequestID(source))

	// Copy trace span
	if span := opentracing.SpanFromContext(source); span != nil {
		target = opentracing.ContextWithSpan(target, span)
	}

	return target
}
