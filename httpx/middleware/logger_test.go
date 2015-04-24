package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

func TestLogger(t *testing.T) {
	b := new(bytes.Buffer)

	h := LogTo(httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return nil
	}), stdLogger(b))

	ctx := context.Background()
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	if err := h.ServeHTTPContext(ctx, resp, req); err != nil {
		t.Fatal(err)
	}

	if got, want := b.String(), "request_id= request at=request method=GET path=\"/\"\n"; got != want {
		t.Fatalf("%s; want %s", got, want)
	}
}
