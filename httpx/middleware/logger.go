package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"context"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/logger"
)

type loggerGenerator func(context.Context, *http.Request) logger.Logger

func LoggerWithRequestID(ctx context.Context, r *http.Request) logger.Logger {
	return logger.DefaultLogger.With("request_id", httpx.RequestID(ctx))
}

// returns a loggerGenerator that generates a loggers that write to STDOUT
// with the level parsed from the string (eg "info")
// If the string isnt parsable, it defaults to "debug"
func StdoutLoggerWithLevel(lvl string) loggerGenerator {
	l := logger.ParseLevel(lvl)
	return stdLogger(l, os.Stdout)
}

// Legacy interface, returns a loggerGenerator that logs at DEBUG to stdout.
// Use StdoutLoggerWithLevel() instead.
var StdoutLogger = stdLogger(logger.DEBUG, os.Stdout)

// LogTo is an httpx middleware that wraps the handler to insert a logger and
// log the request to it.
func LogTo(h httpx.Handler, g loggerGenerator) httpx.Handler {
	return InsertLogger(Log(h), g)
}

// InsertLogger returns an httpx.Handler middleware that will call f to generate
// a logger, then insert it into the context.
func InsertLogger(h httpx.Handler, g loggerGenerator) httpx.Handler {
	return httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		l := g(ctx, r)

		ctx = logger.WithLogger(ctx, l)
		r = r.WithContext(ctx)

		return h.ServeHTTPContext(ctx, w, r)
	})
}

func stdLogger(level logger.Level, out io.Writer) loggerGenerator {
	return func(ctx context.Context, r *http.Request) logger.Logger {
		return logger.New(
			log.New(out, fmt.Sprintf("request_id=%s ", httpx.RequestID(ctx)), 0),
			level,
		)
	}
}

// Logger is middleware that logs the request details to the logger.Logger
// embedded within the context.
type Logger struct {
	// handler is the wrapped httpx.Handler
	handler httpx.Handler
}

func Log(h httpx.Handler) *Logger {
	return &Logger{
		handler: h,
	}
}

func (h *Logger) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	rw := NewResponseWriter(w)

	t := time.Now()

	err := h.handler.ServeHTTPContext(ctx, rw, r)

	ms := fmt.Sprintf("%d", (int(time.Now().Sub(t).Seconds() * 1000)))

	logger.Debug(ctx, "request",
		"method", r.Method,
		"path", r.URL.Path,
		"status", rw.Status(),
		"ms", ms,
	)

	return err
}
