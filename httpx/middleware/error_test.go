package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/remind101/pkg/httpx"
)

func TestErrorWithHandler(t *testing.T) {
	var called bool
	boomErr := errors.New("boom")

	h := &Error{
		ErrorHandler: func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
			called = true
		},
		handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return boomErr
		}),
	}

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/path", nil)
	resp := httptest.NewRecorder()

	err := h.ServeHTTPContext(ctx, resp, req)
	if err != boomErr {
		t.Fatal("Expected error to be returned")
	}

	if !called {
		t.Fatal("Expected the error handler to be called")
	}
}
