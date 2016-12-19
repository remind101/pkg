package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remind101/pkg/httpx"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		header http.Header
		id     string
	}{
		{http.Header{http.CanonicalHeaderKey("X-Request-ID"): []string{"1234"}}, "1234"},
		{http.Header{http.CanonicalHeaderKey("Request-ID"): []string{"1234"}}, "1234"},
		{http.Header{http.CanonicalHeaderKey("Foo"): []string{"1234"}}, ""},
	}

	for i, tt := range tests {
		t.Log(i)
		m := &RequestID{
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestID := httpx.RequestID(r.Context())

				if got, want := requestID, tt.id; got != want {
					t.Fatalf("RequestID => %s; want %s", got, want)
				}
			}),
		}

		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header = tt.header

		m.ServeHTTP(resp, req)
	}
}
