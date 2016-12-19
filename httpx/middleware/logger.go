package middleware

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/logger"
)

// StdoutLogger is a logger.Logger generator that generates a logger that writes
// to stdout.
var StdoutLogger = stdLogger(os.Stdout)

// LogTo is an http middleware that wraps the handler to insert a logger and
// log the request to it.
func LogTo(h http.Handler, f func(context.Context, *http.Request) logger.Logger) http.Handler {
	return InsertLogger(Log(h), f)
}

// InsertLogger returns an http.Handler middleware that will call f to generate
// a logger, then insert it into the context.
func InsertLogger(h http.Handler, f func(context.Context, *http.Request) logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := f(ctx, r)
		r = r.WithContext(logger.WithLogger(ctx, l))
		h.ServeHTTP(w, r)
	})
}

func stdLogger(out io.Writer) func(context.Context, *http.Request) logger.Logger {
	return func(ctx context.Context, r *http.Request) logger.Logger {
		return logger.New(log.New(out, fmt.Sprintf("request_id=%s ", httpx.RequestID(ctx)), 0))
	}
}

// Logger is middleware that logs the request details to the logger.Logger
// embedded within the context.
type Logger struct {
	// handler is the wrapped httpx.Handler
	handler http.Handler
}

func Log(h http.Handler) *Logger {
	return &Logger{
		handler: h,
	}
}

func (h *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := NewResponseWriter(w)

	logger.Info(ctx, "request.start",
		"method", r.Method,
		"path", r.URL.Path,
	)

	h.handler.ServeHTTP(rw, r)

	logger.Info(ctx, "request.complete",
		"status", rw.Status(),
	)
}
