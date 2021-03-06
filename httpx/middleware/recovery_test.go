package middleware

import (
	gerrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/errors"
	"github.com/remind101/pkg/reporter"
)

func TestRecovery(t *testing.T) {
	var (
		errBoom = gerrors.New("boom")
	)

	h := &Recovery{
		Reporter: reporter.ReporterFunc(func(ctx context.Context, level string, err error) error {
			e := err.(*errors.Error)

			if e.Err != errBoom {
				t.Fatalf("err => %v; want %v", err, errBoom)
			}

			if got, want := e.ContextData()["request_id"], "1234"; got != want {
				t.Fatalf("RequestID => %s; want %s", got, want)
			}

			return nil
		}),
		handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			panic(errBoom)
		}),
	}

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	ctx = httpx.WithRequestID(ctx, "1234")

	defer func() {
		if err := recover(); err != nil {
			t.Fatal("Expected the panic to be handled.")
		}
	}()

	err := h.ServeHTTPContext(ctx, resp, req)

	if err.Error() != errBoom.Error() {
		t.Fatalf("err => %v; want %v", err.Error(), errBoom.Error())
	}
}

func TestRecoveryPanicString(t *testing.T) {
	h := &Recovery{
		Reporter: reporter.ReporterFunc(func(ctx context.Context, level string, err error) error {
			return nil
		}),
		handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			panic("boom")
		}),
	}

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()

	defer func() {
		if err := recover(); err != nil {
			t.Fatal("Expected the panic to be handled.")
		}
	}()

	err := h.ServeHTTPContext(ctx, resp, req)

	if got, want := err.Error(), "boom"; got != want {
		t.Fatalf("err => %v; want %v", got, want)
	}
}
