package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remind101/pkg/httpx"
)

func TestBasicAuth(t *testing.T) {
	ctx := context.Background()
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/city/123.123.123.123", nil)
	req.SetBasicAuth("user", "pass")

	var h httpx.Handler

	const test_str = "sfaftfofsfhfi"

	h = httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte(test_str))
		return nil
	})

	h = BasicAuth(h, "user", "pass", "realm")
	if err := h.ServeHTTPContext(ctx, resp, req); err != nil {
		t.Fatal(err)
	}

	if got, want := resp.Body.String(), test_str; got != want {
		t.Fatalf("Body => %s; want %s", got, want)
	}
}

func TestBasicAuthUnauthorized(t *testing.T) {
	ctx := context.Background()
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/city/123.123.123.123", nil)

	var h httpx.Handler

	const test_str = "sfaftfofsfhfi"

	h = httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte(test_str))
		return nil
	})

	h = BasicAuth(h, "user", "pass", "realm")
	if err := h.ServeHTTPContext(ctx, resp, req); err != nil {
		t.Fatal(err)
	}

	if resp.Code != 401 {
		t.Fatalf("Expected code 401")
	}
}
