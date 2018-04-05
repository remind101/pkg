package context

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
)

// See CopyToContext
func Copy(ctx context.Context) context.Context {
	return CopyToContext(context.Background(), ctx)
}

// CopyToContext provides a hook to copy an event context before publishing.
// This is important if the given event context is actually a request context
// whose lifecycle likely ends before subscribers have a chance to process the
// event.
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
