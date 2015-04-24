package middleware

import (
	"net/http"

	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

// HeaderRequestID is the default name of the header to extract request ids
// from.
var HeaderRequestID = []string{"X-Request-Id", "Request-Id"}

// RequestID is middleware that extracts a request id from the headers and
// inserts it into the context.
type RequestID struct {
	// Extractor is a function that can extract a request id from an
	// http.Request. The zero value is a function that will pull a request
	// id from the `X-Request-ID` or `Request-ID` headers.
	Extractor func(*http.Request) string

	// handler is the wrapped httpx.Handler.
	handler httpx.Handler
}

func ExtractRequestID(h httpx.Handler) *RequestID {
	return &RequestID{
		handler: h,
	}
}

// ServeHTTPContext implements the httpx.Handler interface. It extracts a
// request id from the headers and inserts it into the context.
func (h *RequestID) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	e := h.Extractor
	if e == nil {
		e = extractRequestID
	}
	requestID := e(r)

	ctx = httpx.WithRequestID(ctx, requestID)
	return h.handler.ServeHTTPContext(ctx, w, r)
}

func extractRequestID(r *http.Request) string {
	for _, h := range HeaderRequestID {
		v := r.Header.Get(h)
		if v != "" {
			return v
		}
	}

	return ""
}
