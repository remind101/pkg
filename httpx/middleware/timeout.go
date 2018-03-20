package middleware

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/remind101/pkg/httpx"
)

// TimeoutHandler returns a Handler that runs h with the given time limit.
//
// The new Handler calls h.ServeHTTPContext to handle each request, but if a
// call runs for longer than its time limit, the handler will return an
// error that satisfies the timeoutError interface, to integrate with the
// error middleware.
//
// After such a timeout, writes by h to its ResponseWriter will return
// ErrHandlerTimeout.
//
// TimeoutHandler buffers all Handler writes to memory and does not
// support the Hijacker or Flusher interfaces.
//
// NOTE This is a modified version of https://godoc.org/net/http#TimeoutHandler
func TimeoutHandler(h httpx.Handler, dt time.Duration) httpx.Handler {
	return &timeoutHandler{
		handler: h,
		dt:      dt,
	}
}

type handlerTimeout string

func (e handlerTimeout) Timeout() bool {
	return true
}
func (e handlerTimeout) Error() string {
	return string(e)
}

// ErrHandlerTimeout is returned on ResponseWriter Write calls
// in handlers which have timed out.
var ErrHandlerTimeout = handlerTimeout("http: handler timeout")

type timeoutHandler struct {
	handler httpx.Handler
	dt      time.Duration
}

func (h *timeoutHandler) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	ctx, cancelCtx := context.WithTimeout(ctx, h.dt)
	defer cancelCtx()

	r = r.WithContext(ctx)
	done := make(chan struct{})
	tw := &timeoutWriter{
		w: w,
		h: make(http.Header),
	}
	panicChan := make(chan interface{}, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()
		err = h.handler.ServeHTTPContext(ctx, tw, r)
		close(done)
	}()
	select {
	case p := <-panicChan:
		panic(p)
	case <-done:
		tw.mu.Lock()
		defer tw.mu.Unlock()
		dst := w.Header()
		for k, vv := range tw.h {
			dst[k] = vv
		}
		if !tw.wroteHeader {
			tw.code = http.StatusOK
		}
		w.WriteHeader(tw.code)
		w.Write(tw.wbuf.Bytes())
	case <-ctx.Done():
		tw.mu.Lock()
		defer tw.mu.Unlock()
		tw.timedOut = true
		err = ErrHandlerTimeout
	}
	return err
}

type timeoutWriter struct {
	w    http.ResponseWriter
	h    http.Header
	wbuf bytes.Buffer

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	code        int
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.writeHeader(http.StatusOK)
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.writeHeader(code)
}

func (tw *timeoutWriter) writeHeader(code int) {
	tw.wroteHeader = true
	tw.code = code
}
