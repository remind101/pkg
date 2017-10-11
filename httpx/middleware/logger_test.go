// Thanks negroni!
package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/logger"
	"golang.org/x/net/context"
)

func TestLogger(t *testing.T) {
	b := new(bytes.Buffer)

	h := LogTo(httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(201)
		return nil
	}), stdLogger(logger.ALL, b))

	ctx := context.Background()
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	if err := h.ServeHTTPContext(ctx, resp, req); err != nil {
		t.Fatal(err)
	}

	got := b.String()

	// Missing duration to avoid timing false positives
	want := "request_id= request method=GET path=/ status=201"
	if strings.Contains(got, want) != true {
		t.Fatalf("%s; want %s", got, want)
	}
}

// set to warn, check it logs nothing
func TestLoggerUnderLevel(t *testing.T) {
	b := new(bytes.Buffer)

	h := LogTo(httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(201)
		return nil
	}), stdLogger(logger.WARN, b))

	ctx := context.Background()
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	if err := h.ServeHTTPContext(ctx, resp, req); err != nil {
		t.Fatal(err)
	}

	got := b.String()

	want := ""
	if strings.Contains(got, want) != true {
		t.Fatalf("%s; want %s", got, want)
	}
}
