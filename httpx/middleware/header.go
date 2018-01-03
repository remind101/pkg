package middleware

import (
	"net/http"

	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

type Header struct {
	// handler is the wrapped httpx.Handler.
	handler   httpx.Handler
	key       string
	extractor func(*http.Request) string
}

func ExtractHeader(h httpx.Handler, header string) *Header {
	return &Header{
		key:       header,
		handler:   h,
		extractor: HeaderExtractor([]string{header}),
	}
}

func (h *Header) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	e := h.extractor

	value := e(r)

	ctx = httpx.WithHeader(ctx, h.key, value)
	return h.handler.ServeHTTPContext(ctx, w, r)
}

// HeaderExtractor returns a function that can extract a value from a list
// of headers.
func HeaderExtractor(headers []string) func(*http.Request) string {
	return func(r *http.Request) string {
		for _, h := range headers {
			v := r.Header.Get(h)
			if v != "" {
				return v
			}
		}

		return ""
	}
}
