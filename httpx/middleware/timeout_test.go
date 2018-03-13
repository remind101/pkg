package middleware

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/remind101/pkg/httpx"
)

func TestTimeoutFailure(t *testing.T) {
	h := httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
		time.Sleep(100 * time.Millisecond)
		rw.WriteHeader(http.StatusOK)
		return nil
	})
	th := TimeoutHandler(h, 50*time.Millisecond)

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	err := th.ServeHTTPContext(ctx, resp, req)

	if got, want := err.Error(), "http: handler timeout"; got != want {
		t.Fatalf("err => %v; want %v", got, want)
	}
}

func TestTimeoutSuccess(t *testing.T) {
	h := httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintln(rw, "Hello")
		return nil
	})
	th := TimeoutHandler(h, 50*time.Millisecond)

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	err := th.ServeHTTPContext(ctx, resp, req)

	if err != nil {
		t.Fatalf("expected no error, got %#v", err)
	}

	if got, want := resp.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("err => %v; want %v", got, want)
	}

	b, _ := ioutil.ReadAll(resp.Result().Body)
	if got, want := string(b), "Hello\n"; got != want {
		t.Fatalf("err => %v; want %v", got, want)
	}
}
