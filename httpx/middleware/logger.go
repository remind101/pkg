package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/logger"
	"golang.org/x/net/context"
)

type loggerGenerator func(context.Context, *http.Request) logger.Logger

// 
func StdoutLoggerWithLevel(lvl string) loggerGenerator {
	l := logger.ParseLevel(lvl)
	return stdLogger(l, os.Stdout)
}

// StdoutLogger is a logger.Logger generator that generates a logger that writes
// to stdout with level debug
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

	logger.Info(ctx, "request",
		"method", r.Method,
		"path", r.URL.Path,
		"status", rw.Status(),
		"ms", ms,
	)

	return err
}
