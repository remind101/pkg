package middleware

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/errors"
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

type handlerTimeout struct {
	message string
}

func (e handlerTimeout) Timeout() bool {
	return true
}

func (e handlerTimeout) Error() string {
	return e.message
}

// ErrHandlerTimeout is returned on ResponseWriter Write calls
// in handlers which have timed out.
var ErrHandlerTimeout = &handlerTimeout{"http: handler timeout"}

type timeoutHandler struct {
	handler httpx.Handler
	dt      time.Duration
}

func (h *timeoutHandler) ServeHTTPContext(ctx context.Context, rw http.ResponseWriter, r *http.Request) (err error) {
	ctx, cancelCtx := context.WithTimeout(ctx, h.dt)
	defer cancelCtx()

	r = r.WithContext(ctx)
	done := make(chan struct{})
	tw := &timeoutWriter{
		h: make(http.Header),
	}
	panicChan := make(chan interface{}, 1)
	go func() {
		defer func() {
			if p := errors.Recover(ctx, recover()); p != nil {
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

		// If timeout writer was written to by request handler, we write the buffered
		// response to the response writer.
		//
		// It is possible that the handler merely returned an error in which case
		// a middleware may write a response, so we must not always write one here.
		if tw.isModified() {
			dst := rw.Header()
			for k, vv := range tw.h {
				dst[k] = vv
			}
			if !tw.wroteHeader {
				tw.code = http.StatusOK
			}
			rw.WriteHeader(tw.code)
			rw.Write(tw.wbuf.Bytes())
		}
	case <-ctx.Done():
		tw.mu.Lock()
		defer tw.mu.Unlock()
		tw.timedOut = true
		err = errors.New(ctx, ErrHandlerTimeout, 0)
	}
	return err
}

type timeoutWriter struct {
	h    http.Header
	wbuf bytes.Buffer

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	modified    bool
	code        int
}

func (tw *timeoutWriter) Header() http.Header {
	return tw.h
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.writeHeader(http.StatusOK)
	}
	tw.modified = true
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.modified = true
	tw.writeHeader(code)
}

func (tw *timeoutWriter) writeHeader(code int) {
	tw.wroteHeader = true
	tw.code = code
}

func (tw *timeoutWriter) isModified() bool {
	return tw.modified || len(tw.Header()) > 0
}
