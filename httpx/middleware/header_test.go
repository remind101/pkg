package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"context"
	"github.com/remind101/pkg/httpx"
)

func TestHeader(t *testing.T) {
	tests := []struct {
		header http.Header
		key    string
		val    string
	}{
		{http.Header{http.CanonicalHeaderKey("X-Some-Version"): []string{"versionversion"}}, "X-Some-Version", "versionversion"},
		{http.Header{http.CanonicalHeaderKey("X-Client-Geo"): []string{"clientgeo"}}, "X-Client-Geo", "clientgeo"},
		{http.Header{http.CanonicalHeaderKey("Foo"): []string{"1234"}}, "Bar", ""},
	}

	for _, tt := range tests {
		m := ExtractHeader(
			httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				data := httpx.Header(ctx, tt.key)

				if got, want := data, tt.val; got != want {
					t.Fatalf("%s => %s; want %s", tt.key, got, want)
				}

				return nil
			}),
			tt.key,
		)

		ctx := context.Background()
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header = tt.header

		if err := m.ServeHTTPContext(ctx, resp, req); err != nil {
			t.Fatal(err)
		}
	}
}
