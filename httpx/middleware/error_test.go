package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remind101/pkg/httpx"
	"context"
)

func TestError(t *testing.T) {
	h := &Error{
		handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return errors.New("boom")
		}),
	}

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/path", nil)
	resp := httptest.NewRecorder()

	err := h.ServeHTTPContext(ctx, resp, req)

	if err != nil {
		t.Fatal("Expected no error to be returned because it was handled")
	}

	if got, want := resp.Body.String(), "boom\n"; got != want {
		t.Fatalf("Body => %s; want %s", got, want)
	}

	if got, want := resp.Code, 500; got != want {
		t.Fatalf("Status => %v; want %v", got, want)
	}
}

type tmpError string

func (te tmpError) Error() string {
	return string(te)
}

func (te tmpError) Temporary() bool {
	return true
}

func TestTemporaryError(t *testing.T) {
	h := &Error{
		handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return tmpError("Service unavailable")
		}),
	}

	req, _ := http.NewRequest("GET", "/path", nil)
	resp := httptest.NewRecorder()
	err := h.ServeHTTPContext(context.Background(), resp, req)

	if err != nil {
		t.Fatal("Expected no error to be returned because it was handled")
	}

	if got, want := resp.Body.String(), "Service unavailable\n"; got != want {
		t.Fatalf("Body => %s; want %s", got, want)
	}

	if got, want := resp.Code, 503; got != want {
		t.Fatalf("Status => %v; want %v", got, want)
	}
}

func TestErrorWithHandler(t *testing.T) {
	var called bool

	h := &Error{
		ErrorHandler: func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
			called = true
		},
		handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return errors.New("boom")
		}),
	}

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/path", nil)
	resp := httptest.NewRecorder()

	h.ServeHTTPContext(ctx, resp, req)

	if !called {
		t.Fatal("Expected the error handler to be called")
	}
}
