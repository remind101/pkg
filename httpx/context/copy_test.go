package context_test

import (
	"context"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/remind101/pkg/httpx"
	httpxcontext "github.com/remind101/pkg/httpx/context"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
)

func TestCopy(t *testing.T) {
	l := logger.DefaultLogger
	r := reporter.NewLogReporter()

	ctx := context.Background()
	ctx = logger.WithLogger(ctx, l)
	ctx = reporter.WithReporter(ctx, r)
	ctx = httpx.WithRequestID(ctx, "abc")
	_, ctx = opentracing.StartSpanFromContext(ctx, "test.span")
	ctx, cancel := context.WithCancel(ctx)

	cc := httpxcontext.Copy(ctx)
	cancel()

	if cc.Err() != nil {
		t.Fatal(cc.Err())
	}

	if got, want := httpx.RequestID(cc), "abc"; got != want {
		t.Errorf("got %v; expected %v", got, want)
	}

	if _, ok := logger.FromContext(cc); !ok {
		t.Error("expected logger in context")
	}

	if _, ok := reporter.FromContext(cc); !ok {
		t.Error("expected reporter in context")
	}

	if span := opentracing.SpanFromContext(cc); span == nil {
		t.Error("expected span in context")
	}
}
